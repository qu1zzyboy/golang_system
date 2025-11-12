package wsReConn

import (
	"context"
	"sync/atomic"
	"time"

	"upbitBnServer/internal/define/defineEmoji"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsPingPong"
	"upbitBnServer/internal/infra/ws/wsRead"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
)

type ReConnAuto struct {
	resourceId   string                    // resource ID
	subParam     wsDefine.SubDial          // 参数管理器,用于获取订阅参数
	sigChan      chan wsDefine.ReConnType  // 信号通道,用于接收重连信号
	read         wsDefine.ReadAutoHandler  // 读取消息处理函数,必选
	thisCancel   context.CancelFunc        // 本轮连接资源的释放函数
	thisConn     *wsDefine.SafeWrite       // websocket连接
	isConnOk     atomic.Bool               // 连接状态,是否可用
	retryCount   atomic.Int32              // 重试次数
	exType       exchangeEnum.ExchangeType // 交易所类型,方便设置pong
	resourceType resourceEnum.ResourceType // 资源类型,方便监控
}

func NewReConnAuto(
	exType exchangeEnum.ExchangeType,
	resourceType resourceEnum.ResourceType,
	subParam wsDefine.SubDial,
	read wsDefine.ReadAutoHandler, resourceId string) *ReConnAuto {
	return &ReConnAuto{
		sigChan:      make(chan wsDefine.ReConnType, 1),
		subParam:     subParam,
		read:         read,
		exType:       exType,
		resourceType: resourceType,
		resourceId:   resourceId,
	}
}

func (c *ReConnAuto) IsConnOk() bool { return c.isConnOk.Load() }

func (c *ReConnAuto) WriteAsync(data []byte) error {
	return c.thisConn.SafeWriteMsg(websocket.TextMessage, data)
}

func (c *ReConnAuto) ReConnLoop(ctxStop context.Context) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "reConn"), func() { c.reConnLoop(ctxStop) })
}

func (c *ReConnAuto) reConnLoop(ctxStop context.Context) {
	dynamicLog.Log.GetLog().Infof("进入[%s] auto_连接循环 %s", c.resourceId, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s] auto_连接循环", c.resourceId)
			if c.thisCancel != nil {
				c.thisCancel() // 释放上次连接的资源
			}
			return
		case sig := <-c.sigChan:
			if c.thisCancel != nil {
				c.thisCancel() // 释放上次连接的资源
			}
			if c.thisConn != nil {
				c.thisConn.SafeClose() // 关闭上次连接
			}
			c.isConnOk.Store(false)
			dynamicLog.Log.GetLog().Infof("[%s] 接收到连接信号[%s] %s", c.resourceId, sig, defineEmoji.YesBox)
			c.startReconnect(context.Background(), b)
			c.isConnOk.Store(true)
		}
	}
}

func (c *ReConnAuto) ReceiveSig(sig wsDefine.ReConnType) {
	select {
	case c.sigChan <- sig:
	default:
	}
}

func (c *ReConnAuto) CloseSub(ctx context.Context) {
	c.thisCancel()
}

// 带指数退避的重连
func (c *ReConnAuto) startReconnect(ctx context.Context, b *backoff.Backoff) {
	c.retryCount.Add(1)
	for {
		if err := c.connect(ctx); err != nil {
			delay := b.Duration()
			dynamicLog.Error.GetLog().Errorf("[%s]重连失败,错误: %v,重试次数: %d,等待时间: %s", c.resourceId, err, c.retryCount.Load(), delay)
			time.Sleep(delay)
			continue
		} else {
			b.Reset()
			break
		}
	}
}

func (c *ReConnAuto) connect(ctx context.Context) error {
	conn, err := c.subParam.DialTo(ctx)
	if err != nil {
		return err
	}
	/*******连接成功*******/
	c.thisConn = conn

	ctxStopChild, cancel := context.WithCancel(context.Background())
	c.thisCancel = cancel

	switch c.exType {
	case exchangeEnum.BINANCE:
		conn.GetConn().SetPingHandler(func(msg string) error {
			wsPingPong.PongBn(msg, conn)
			return nil
		})
		switch c.resourceType {
		case resourceEnum.WS_REQUEST_PRIVATE, resourceEnum.PAYLOAD_READ, resourceEnum.DELTA_DEPTH, resourceEnum.WS_REQUEST_MARKET:
			safeRead := wsRead.NewAutoRead(conn.GetConn(), c.resourceId)
			safeRead.ReadAutoLoop(ctxStopChild, c.read, c.sigChan)
		default:
		}
	case exchangeEnum.BYBIT:
		switch c.resourceType {
		case resourceEnum.MARK_PRICE, resourceEnum.AGG_TRADE, resourceEnum.SYMBOL_SUB_AUTO, resourceEnum.WS_REQUEST_PRIVATE:
			safeRead := wsRead.NewAutoRead(conn.GetConn(), c.resourceId)
			safeRead.ReadAutoLoop(ctxStopChild, c.read, c.sigChan)
		default:
		}
	case exchangeEnum.TREE_NEWS:
		safeRead := wsRead.NewAutoRead(conn.GetConn(), c.resourceId)
		safeRead.ReadAutoLoop(ctxStopChild, c.read, c.sigChan)
	default:
	}
	//开启ping-pong循环
	// pingPong.PingPongLoop(ctxStopChild, c.sigChan)
	return nil
}
