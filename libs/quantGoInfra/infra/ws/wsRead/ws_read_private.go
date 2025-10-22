package wsRead

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/hhh500/quantGoInfra/define/defineEmoji"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
)

type ReadPrivate struct {
	resourceId string          // resource ID
	conn       *websocket.Conn // websocket连接
}

func NewPrivateRead(conn *websocket.Conn, resourceId string) *ReadPrivate {
	return &ReadPrivate{
		conn:       conn,
		resourceId: resourceId,
	}
}

func (c *ReadPrivate) readPrivateLoop(ctxStop context.Context, onMsg wsDefine.ReadPrivateHandler, sigChan chan wsDefine.ReConnType) {
	dynamicLog.Log.GetLog().Infof("进入[%s] read_Private循环 %s", c.resourceId, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s]ws read_Private循环", c.resourceId)
			return
		default:
			// do nothing here
		}
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			dynamicLog.Log.GetLog().Infof("读取出错退出[%s]ws read_Private循环,%v", c.resourceId, err)
			select {
			case sigChan <- wsDefine.READ_ERROR:
			default:
			}
			return
		}
		onMsg(msg) // 调用回调函数处理消息
	}
}

func (c *ReadPrivate) ReadPrivateLoop(ctxBreak context.Context, onMsg wsDefine.ReadPrivateHandler, sigChan chan wsDefine.ReConnType) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "readPrivateLoop"), func() { c.readPrivateLoop(ctxBreak, onMsg, sigChan) })
}
