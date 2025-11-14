package wsPingPong

import (
	"context"
	"time"

	"upbitBnServer/internal/define/defineEmoji"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/utils/idGen"
)

type PingPong struct {
	resourceId string              // 资源标识
	conn       *wsDefine.SafeWrite // websocket连接
	ping       wsDefine.PingFunc   // ping-pong函数
}

func NewPingPong(ping wsDefine.PingFunc, conn *wsDefine.SafeWrite, resourceId string) *PingPong {
	return &PingPong{
		resourceId: resourceId,
		conn:       conn,
		ping:       ping,
	}
}

func (c *PingPong) pingPongLoop(ctxBreak context.Context, sigChan chan wsDefine.ReConnType) {
	dynamicLog.Log.GetLog().Infof("进入[%s] ping-pong循环 %s", c.resourceId, defineEmoji.Rocket)
	ticker := time.NewTicker(wsDefine.KeepAliveInterval) //15s
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ctxBreak.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s] ping-pong循环", c.resourceId)
			return
		case <-ticker.C:
			if err := c.ping(c.conn); err != nil {
				dynamicLog.Error.GetLog().Errorf("ping出错退出[%s]ws ping-pong循环,%v", c.resourceId, err)
				select {
				case sigChan <- wsDefine.PING_ERROR:
				default:
					// already signaled
				}
				return
			}
		}
	}
}

func (c *PingPong) PingPongLoop(ctxBreak context.Context, sigChan chan wsDefine.ReConnType) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "pingPongLoop"), func() { c.pingPongLoop(ctxBreak, sigChan) })
}
