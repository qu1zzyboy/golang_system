package wsRead

import (
	"context"
	"io"

	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/gorilla/websocket"
)

type ReadPool struct {
	resourceId string          // resource ID
	conn       *websocket.Conn // websocket连接
	poolSize   int
}

func NewReadPool(conn *websocket.Conn, poolSize int, resourceId string) *ReadPool {
	return &ReadPool{
		conn:       conn,
		poolSize:   poolSize,
		resourceId: resourceId,
	}
}

func (c *ReadPool) readPoolLoop(ctxStop context.Context, onMsg wsDefine.ReadPoolHandler, sigChan chan wsDefine.ReConnType) {
	// dynamicLog.Log.GetLog().Infof("进入[%s] read_pool_%d 循环 %s", c.resourceId, c.poolSize, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			return
		default:
			// do nothing here
		}
		// NextReader() → 定位“下一条消息的开始”
		_, r, err := c.conn.NextReader()
		if err != nil {
			dynamicLog.Log.GetLog().Infof("读取出错退出[%s]ws read_pool循环,%v", c.resourceId, err)
			select {
			case sigChan <- wsDefine.READ_ERROR:
			default:
			}
			return
		}
		// 1 从池中获取 buffer
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
		onMsg(uint16(total), bufPtr)
		// 3 调用业务逻辑
		// onMsg(b[:total])
		// 4 回收 buffer
		// byteBufPool.ReleaseBuffer(bufPtr)
	}
}

func (c *ReadPool) ReadPoolLoop(ctxBreak context.Context, onMsg wsDefine.ReadPoolHandler, sigChan chan wsDefine.ReConnType) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "ReadPoolLoop"), func() { c.readPoolLoop(ctxBreak, onMsg, sigChan) })
}
