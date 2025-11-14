package wsRead

import (
	"context"

	"upbitBnServer/internal/define/defineEmoji"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/gorilla/websocket"
)

type ReadAuto struct {
	resourceId string          // resource ID
	conn       *websocket.Conn // websocket连接
}

func NewAutoRead(conn *websocket.Conn, resourceId string) *ReadAuto {
	return &ReadAuto{
		conn:       conn,
		resourceId: resourceId,
	}
}

func (c *ReadAuto) readAutoLoop(ctxStop context.Context, onMsg wsDefine.ReadAutoHandler, sigChan chan wsDefine.ReConnType) {
	dynamicLog.Log.GetLog().Debugf("进入[%s] read_auto循环 %s", c.resourceId, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s]ws read_auto循环", c.resourceId)
			return
		default:
			// do nothing here
		}
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			dynamicLog.Log.GetLog().Infof("读取出错退出[%s]ws read_auto循环,%v", c.resourceId, err)
			select {
			case sigChan <- wsDefine.READ_ERROR:
			default:
			}
			return
		}
		onMsg(msg) // 调用回调函数处理消息
	}
}

func (c *ReadAuto) ReadAutoLoop(ctxBreak context.Context, onMsg wsDefine.ReadAutoHandler, sigChan chan wsDefine.ReConnType) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "readAutoLoop"), func() { c.readAutoLoop(ctxBreak, onMsg, sigChan) })
}
