package wsRead

import (
	"context"
	"io"

	"upbitBnServer/internal/define/defineEmoji"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/gorilla/websocket"
)

type ReadMarket struct {
	resourceId string          // resource ID
	conn       *websocket.Conn // websocket连接
	poolSize   int
}

func NewReadMarket(conn *websocket.Conn, poolSize int, resourceId string) *ReadMarket {
	return &ReadMarket{
		conn:       conn,
		poolSize:   poolSize,
		resourceId: resourceId,
	}
}

func (c *ReadMarket) readMarketLoop(ctxStop context.Context, onMsg wsDefine.ReadMarketHandler, sigChan chan wsDefine.ReConnType) {
	dynamicLog.Log.GetLog().Infof("进入[%s] ReadMarket循环 %s", c.resourceId, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s]ws ReadMarket循环", c.resourceId)
			return
		default:
			// do nothing here
		}
		// NextReader() → 定位“下一条消息的开始”
		_, r, err := c.conn.NextReader()
		if err != nil {
			dynamicLog.Log.GetLog().Infof("读取出错退出[%s]ws read循环,%v", c.resourceId, err)
			select {
			case sigChan <- wsDefine.READ_ERROR:
			default:
			}
			return
		}
		// 1️⃣ 从池中获取 buffer
		bufPtr := byteBufPool.AcquireBuffer(c.poolSize)
		b := (*bufPtr)[:0]

		total := 0
		for {
			// r.Read() → 读取“当前消息的内容”
			n, err := r.Read(b[total:cap(b)])
			total += n
			if err == io.EOF {
				break
			}
			if err != nil {
				byteBufPool.ReleaseBuffer(bufPtr)
				return
			}
		}
		// 所有权转交给 onMsg
		onMsg(total, bufPtr)
		// 3️⃣ 调用业务逻辑
		// onMsg(b[:total])
		// 4️⃣ 回收 buffer
		// byteBufPool.ReleaseBuffer(bufPtr)
	}
}

func (c *ReadMarket) ReadMarketLoop(ctxBreak context.Context, onMsg wsDefine.ReadMarketHandler, sigChan chan wsDefine.ReConnType) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "ReadMarketLoop"), func() { c.readMarketLoop(ctxBreak, onMsg, sigChan) })
}
