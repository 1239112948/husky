package main

import (
	"encoding/json"
	"github.com/guogeer/husky/cmd"
	"github.com/guogeer/husky/log"
	"net"
)

type Args struct {
	ServerName string
	ServerAddr string
	ServerData json.RawMessage
	ServerType string
	Weight     int
}

func init() {
	cmd.Bind(C2S_Register, (*Args)(nil))
	cmd.Bind(C2S_GetServerAddr, (*Args)(nil))
	cmd.Bind(C2S_Concurrent, (*Args)(nil))
	cmd.Bind(C2S_Route, (*cmd.ForwardArgs)(nil))

	cmd.Bind(C2S_Broadcast, (*cmd.Package)(nil))
	cmd.Bind(FUNC_Close, (*Args)(nil))
}

// ServerAddr == "" 无服务
func C2S_Register(ctx *cmd.Context, data interface{}) {
	args := data.(*Args)
	host, port, _ := net.SplitHostPort(args.ServerAddr)
	if host == "" {
		host, _, _ = net.SplitHostPort(ctx.Out.RemoteAddr())
	}

	addr := ""
	if port != "" {
		addr = host + ":" + port
	}
	log.Info("register", args.ServerName, addr)
	// TODO
	ctx.Out.WriteJSON("C2S_RegisterOk", struct{}{})

	newServer := &Server{
		out:  ctx.Out,
		name: args.ServerName,
		addr: addr,
		data: args.ServerData,
		typ:  args.ServerType,
	}
	gRouter.AddServer(newServer)
	// center server
	if newServer.typ == "center" {
		for _, server := range gRouter.servers {
			ctx.Out.WriteJSON("S2C_AddGame", map[string]interface{}{
				"Name": server.name,
				"Data": server.data,
			})
		}
	}
	for _, server := range gRouter.servers {
		if server.typ == "center" && server.name != newServer.name {
			server.out.WriteJSON("S2C_AddGame", map[string]interface{}{
				"Name": newServer.name,
				"Data": newServer.data,
			})
		}
	}

	// 向网关注册服务
	if newServer.typ == "gateway" {
		for _, server := range gRouter.servers {
			ctx.Out.WriteJSON("FUNC_RegisterServiceInGateway", map[string]interface{}{
				"Name": server.name,
			})
		}
	} else if newServer.addr != "" {
		for _, gw := range gRouter.gateways {
			gw.out.WriteJSON("FUNC_RegisterServiceInGateway", map[string]interface{}{
				"Name": newServer.name,
			})
		}
	}
}

func C2S_GetServerAddr(ctx *cmd.Context, data interface{}) {
	args := data.(*Args)
	name := args.ServerName
	addr := gRouter.GetServerAddr(name)
	log.Debug("get addr", name, addr)
	response := map[string]string{"ServerName": name, "ServerAddr": addr}
	ctx.Out.WriteJSON("S2C_GetServerAddr", response)
}

func C2S_Broadcast(ctx *cmd.Context, data interface{}) {
	pkg := data.(*cmd.Package)
	for _, gw := range gRouter.gateways {
		gw.out.WriteJSON("FUNC_Broadcast", pkg)
	}
}

// 更新网关负载
func C2S_Concurrent(ctx *cmd.Context, data interface{}) {
	args := data.(*Args)
	for _, gw := range gRouter.gateways {
		if gw.out == ctx.Out {
			// log.Debug("test ", gw.addr, gw.weight)
			gw.weight = args.Weight
		}
	}

	addr := gRouter.GetBestGateway()
	// log.Debug("concurrent", addr, args.Weight)
	if s := gRouter.GetServer("login"); s != nil {
		response := map[string]interface{}{"Address": addr}
		s.out.WriteJSON("S2C_GetBestGateway", response)
	}
}

func C2S_Route(ctx *cmd.Context, data interface{}) {
	args := data.(*cmd.ForwardArgs)
	servers := args.ServerList
	if len(servers) == 1 && servers[0] == "*" {
		prefixMap := make(map[string]bool)
		for _, server := range gRouter.servers {
			prefixMap[server.name] = true
		}
		servers = servers[:0]
		for s := range prefixMap {
			servers = append(servers, s)
		}
	}

	for _, name := range servers {
		if s := gRouter.GetServer(name); s != nil {
			s.out.WriteJSON(args.Name, args.Data)
		}
	}
}

func FUNC_Close(ctx *cmd.Context, data interface{}) {
	// args := data.(*Args)
	// gRouter.Remove(ctx.Out)
	// TODO
	// log.Info("server lose connection")
}
