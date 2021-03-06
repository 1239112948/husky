package cmd

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/guogeer/husky/log"
	"io"
	"net"
	"reflect"
	"regexp"
	"sync"
	"time"
)

var errInvalidMessageID = errors.New("invalid message ID")

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 32 << 10 // 32K
	sendQueueSize  = 16 << 10
)

const (
	RawMessage   = 0x01
	CloseMessage = 0xf0
	PingMessage  = 0xf1
	PongMessage  = 0xf2
	AuthMessage  = 0xf3
)

const (
	StateClosed = iota
	StateConecting
	StateConnected
	StateClosing
)

type TCPConn struct {
	rwc     net.Conn
	ssid    string
	send    chan []byte
	isClose bool
}

func (c *TCPConn) Close() {
	if c.isClose == true {
		return
	}
	c.isClose = true
	close(c.send)
}

func (c *TCPConn) RemoteAddr() string {
	return c.rwc.RemoteAddr().String()
}

func (c *TCPConn) ReadMessage() (mt uint8, buf []byte, err error) {
	var head [3]byte
	// read message
	if _, err = io.ReadFull(c.rwc, head[:3]); err != nil {
		return
	}

	// 0x01~0x0f 表示版本
	// 0xf0 写队列尾部标识
	// 0xf1 PING
	// 0xf2 PONG
	n := int(binary.BigEndian.Uint16(head[1:3]))

	// 消息
	mt = uint8(head[0])
	switch mt {
	case PingMessage, PongMessage, CloseMessage:
		return
	case AuthMessage, RawMessage:
		if n > 0 && n < maxMessageSize {
			buf = make([]byte, n)
			if _, err = io.ReadFull(c.rwc, buf); err == nil {
				return
			}
		}
	}
	err = errors.New("invalid data")
	return
}

func (c *TCPConn) NewMessageBytes(mt int, data []byte) ([]byte, error) {
	if len(data) > maxMessageSize {
		return nil, errTooLargeMessage
	}
	buf := make([]byte, len(data)+3)
	// 协议头
	copy(buf, []byte{byte(mt), 0x0, 0x0})
	binary.BigEndian.PutUint16(buf[1:3], uint16(len(data)))
	// 协议数据
	copy(buf[3:], data)
	return buf, nil
}

func (c *TCPConn) WriteJSON(name string, i interface{}) error {
	// 消息格式
	pkg := &Package{Id: name, Body: i}
	buf, err := defaultRawParser.Encode(pkg)
	if err != nil {
		return err
	}
	return c.Write(buf)
}

func (c *TCPConn) Write(data []byte) error {
	if c.isClose == true {
		return errors.New("connection is closed")
	}
	select {
	case c.send <- data:
	default:
		return errors.New("write too busy")
	}
	return nil
}

func (c *TCPConn) writeMsg(mt int, msg []byte) (int, error) {
	buf, err := c.NewMessageBytes(mt, msg)
	if err != nil {
		return 0, err
	}
	return c.rwc.Write(buf)
}

type Handler func(*Context, interface{})

type cmdEntry struct {
	h     Handler
	type_ reflect.Type
}

type CmdSet struct {
	services map[string]bool // 内部服务
	e        map[string]*cmdEntry
	mu       sync.RWMutex
}

var defaultCmdSet = &CmdSet{
	services: make(map[string]bool),
	e:        make(map[string]*cmdEntry),
}

func (s *CmdSet) RemoveService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[name] = false
}

func (s *CmdSet) RegisterService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[name] = true
}

// 恢复服务
func (s *CmdSet) RecoverService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[name]; ok {
		s.services[name] = true
	}
}

func (s *CmdSet) Bind(name string, h Handler, i interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.e[name]; ok {
		log.Warnf("%s exist", name)
	}
	type_ := reflect.TypeOf(i)
	s.e[name] = &cmdEntry{h: h, type_: type_}
}

func (s *CmdSet) Handle(ctx *Context, messageID string, data []byte) error {
	// 空数据使用默认JSON格式数据
	if data == nil || len(data) == 0 {
		data = []byte("{}")
	}

	serverName, name := routeMessage("", messageID)
	// 网关转发的消息ID仅允许包含字母、数字
	if ctx.isGateway == true {
		match, err := regexp.MatchString("^[A-Za-z0-9]+$", name)
		if err == nil && !match {
			return errors.New("invalid message id")
		}
	}

	s.mu.RLock()
	e := s.e[name]
	isService := s.services[serverName]
	s.mu.RUnlock()
	// router
	if len(serverName) > 0 {
		if ctx.isGateway == true {
			// 网关仅允许转发已注册的逻辑服务器
			if isService == false {
				return errors.New("gateway try to route invalid service")
			}
		}

		if ss := GetSession(ctx.Ssid); ss != nil {
			ss.Route(serverName, name, data)
		}
		return nil
	}

	if e == nil {
		return errInvalidMessageID
	}

	// unmarshal argument
	args := reflect.New(e.type_.Elem()).Interface()
	if err := json.Unmarshal(data, args); err != nil {
		return err
	}

	Enqueue(ctx, e.h, args)
	return nil
}

func funcClose(ctx *Context, i interface{}) {
	ctx.Out.Close()
}
