package wsPingPong

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hhh500/quantGoInfra/define/defineEmoji"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
)

type PingPong struct {
	thisPing   time.Time           // 这次ping消息的时间
	thisPong   time.Time           // 这次pong消息的时间
	resourceId string              // 资源标识
	traceId    string              // trace ID
	conn       *wsDefine.SafeWrite // websocket连接
	pingOut    wsDefine.PingFunc   // ping-pong函数
	pongRw     sync.RWMutex        // 读写锁
}

func NewPingPong(ping wsDefine.PingFunc, conn *wsDefine.SafeWrite, resourceId string) *PingPong {
	return &PingPong{
		thisPing:   time.Now(),
		thisPong:   time.Now(),
		resourceId: resourceId,
		traceId:    idGen.BuildName2(resourceId, "pingPongLoop"),
		conn:       conn,
		pingOut:    ping,
	}
}

func (c *PingPong) pingIn() error {
	c.thisPing = time.Now()
	return c.pingOut(c.conn)
}

func (c *PingPong) UpdatePong() {
	c.pongRw.Lock()
	defer c.pongRw.Unlock()
	c.thisPong = time.Now()
}

func (c *PingPong) isPongTimeout() (bool, float64) {
	c.pongRw.RLock()
	defer c.pongRw.RUnlock()
	if time.Since(c.thisPong) > wsDefine.KeepAliveInterval*2 {
		return true, time.Since(c.thisPong).Seconds()
	} else {
		return false, 0
	}
}

func (c *PingPong) pingPongLoop(ctxStop context.Context, sigChan chan string) {
	dynamicLog.Log.GetLog().Infof("进入[%s] ping-pong循环 %s", c.resourceId, defineEmoji.Rocket)
	ticker := time.NewTicker(wsDefine.KeepAliveInterval) //15s
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s]ws ping-pong循环", c.resourceId)
			return
		case <-ticker.C:
			if err := c.pingIn(); err != nil {
				dynamicLog.Error.GetLog().Errorf("ping出错退出[%s]ws ping-pong循环,%v", c.resourceId, err)

				select {
				case sigChan <- "PING_ERROR":
				default:
					// already signaled
				}

				return
			}
			isPongOut, diff := c.isPongTimeout()
			if isPongOut {
				pongOutMsg := fmt.Sprintf("PONG_TIMEOUT %.1f 秒, 链接不可用", diff)
				dynamicLog.Error.GetLog().Errorf("pong超时退出[%s]ws ping-pong循环,%s", c.resourceId, pongOutMsg)

				select {
				case sigChan <- pongOutMsg:
				default:
					// already signaled
				}

				return
			}
		}
	}
}

//
//func (c *PingPong) PingPongLoop(ctxBreak context.Context, sigChan chan string) {
//	observePanic.GoSafe(context.Background(), map[string]string{defineJson.ProtectId: c.traceId}, func() {
//		c.pingPongLoop(ctxBreak, sigChan)
//	})
//}
