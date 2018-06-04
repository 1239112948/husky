package gateway

import (
	"github.com/husky/cmd"
	// "third/log"
	"github.com/husky/util"
	"time"
)

var (
	sessionLocation = map[string]string{}
)

func init() {
	util.NewPeriodTimer(func() {
		// log.Debug("tick")
		counter := cmd.GetSessionManage().Count()
		cmd.Route(cmd.ServerRouter, "C2S_Concurrent", map[string]interface{}{"Weight": counter})
	}, "2001-01-01", 10*time.Second)
}