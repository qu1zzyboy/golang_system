package wsReConn

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hhh500/quantGoInfra/define/defineEmoji"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/infra/ws/wsPingPong"
	"github.com/hhh500/quantGoInfra/infra/ws/wsRead"
	"github.com/hhh500/quantGoInfra/infra/ws/wsSub"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/jpillora/backoff"
)

type ReConnPrivate struct {
	resourceId   string                      // resource ID
	subParam     wsSub.SubParam              // 参数管理器,用于获取订阅参数
	sigChan      chan wsDefine.ReConnType    // 信号通道,用于接收重连信号
	read         wsDefine.ReadPrivateHandler // 读取消息处理函数,必选
	thisCancel   context.CancelFunc          // 本轮连接资源的释放函数
	thisConn     *wsDefine.SafeWrite         // websocket连接
	isConnOk     atomic.Bool                 // 连接状态,是否可用
	retryCount   atomic.Int32                // 重试次数
	exType       exchangeEnum.ExchangeType   // 交易所类型,方便设置pong
	resourceType resourceEnum.ResourceType   // 资源类型,方便监控
}

func NewReConnPrivate(
	exType exchangeEnum.ExchangeType,
	resourceType resourceEnum.ResourceType,
	subParam wsSub.SubParam,
	read wsDefine.ReadPrivateHandler, resourceId string) *ReConnPrivate {
	return &ReConnPrivate{
		sigChan:      make(chan wsDefine.ReConnType, 1),
		subParam:     subParam,
		read:         read,
		exType:       exType,
		resourceType: resourceType,
		resourceId:   resourceId,
	}
}

func (c *ReConnPrivate) IsConnOk() bool { return c.isConnOk.Load() }

func (c *ReConnPrivate) WriteAsync(data []byte) error {
	return c.thisConn.SafeWriteMsg(websocket.TextMessage, data)
}

func (c *ReConnPrivate) ReConnLoop(ctxStop context.Context) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "reConn"), func() { c.reConnLoop(ctxStop) })
}

func (c *ReConnPrivate) reConnLoop(ctxStop context.Context) {
	dynamicLog.Log.GetLog().Infof("进入[%s] private连接循环 %s", c.resourceId, defineEmoji.Rocket)
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s] private连接循环", c.resourceId)
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

func (c *ReConnPrivate) ReceiveSig(sig wsDefine.ReConnType) {
	select {
	case c.sigChan <- sig:
	default:
	}
}

func (c *ReConnPrivate) CloseSub(ctx context.Context) {
	c.thisCancel()
}

// 带指数退避的重连
func (c *ReConnPrivate) startReconnect(ctx context.Context, b *backoff.Backoff) {
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

func (c *ReConnPrivate) connect(ctx context.Context) error {
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
		case resourceEnum.ORDER_WRITE, resourceEnum.PAYLOAD_READ:
			safeRead := wsRead.NewPrivateRead(conn.GetConn(), c.resourceId)
			safeRead.ReadPrivateLoop(ctxStopChild, c.read, c.sigChan)
		default:
		}
	case exchangeEnum.BYBIT:
		// read = pingPong.WrapByBitPongHandler(c.read)
	default:
	}
	//开启ping-pong循环
	// pingPong.PingPongLoop(ctxStopChild, c.sigChan)
	return nil
}
