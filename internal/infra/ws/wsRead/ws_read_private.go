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
