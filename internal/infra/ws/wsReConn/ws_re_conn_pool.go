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
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/idGen"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
)

const (
	reconnectMinInterval = 100 * time.Millisecond //最小重连间隔
	reconnectMaxInterval = 2 * time.Second        //最大重连间隔
)

var (
	b = &backoff.Backoff{
		Min:    reconnectMinInterval,
		Max:    reconnectMaxInterval,
		Factor: 1.8,
		Jitter: false,
	}
)

type ReConnPool struct {
	resourceId   string                    // resource ID
	subParam     wsDefine.SubDial          // 参数管理器,用于获取订阅参数
	sigChan      chan wsDefine.ReConnType  // 信号通道,用于接收重连信号
	read         wsDefine.ReadPoolHandler  // 读取消息处理函数,必选
	thisCancel   context.CancelFunc        // 本轮连接资源的释放函数
	thisConn     *wsDefine.SafeWrite       // websocket连接
	isConnOk     atomic.Bool               // 连接状态,是否可用
	retryCount   atomic.Int32              // 重试次数
	exType       exchangeEnum.ExchangeType // 交易所类型,方便设置pong
	resourceType resourceEnum.ResourceType // 资源类型,方便监控
}

func NewReConnPool(
	exType exchangeEnum.ExchangeType,
	resourceType resourceEnum.ResourceType,
	subParam wsDefine.SubDial,
	read wsDefine.ReadPoolHandler, resourceId string) *ReConnPool {
	return &ReConnPool{
		sigChan:      make(chan wsDefine.ReConnType, 1),
		subParam:     subParam,
		read:         read,
		exType:       exType,
		resourceType: resourceType,
		resourceId:   resourceId,
	}
}

func (c *ReConnPool) IsConnOk() bool { return c.isConnOk.Load() }

func (c *ReConnPool) WriteAsync(data []byte) error {
	return c.thisConn.SafeWriteMsg(websocket.TextMessage, data)
}

func (c *ReConnPool) ReConnLoop(ctxStop context.Context) {
	safex.SafeGo(idGen.BuildName2(c.resourceId, "reConn"), func() { c.reConnLoop(ctxStop) })
}

func (c *ReConnPool) reConnLoop(ctxStop context.Context) {
	if c.resourceType != resourceEnum.SYMBOL_SUB_256 {
		dynamicLog.Log.GetLog().Infof("进入[%s] market连接循环 %s", c.resourceId, defineEmoji.Rocket)
	}
	for {
		select {
		case <-ctxStop.Done():
			dynamicLog.Log.GetLog().Infof("主动退出[%s] market连接循环", c.resourceId)
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
			if c.resourceType != resourceEnum.SYMBOL_SUB_256 {
				dynamicLog.Log.GetLog().Infof("[%s] 接收到连接信号[%s] %s", c.resourceId, sig, defineEmoji.YesBox)
			}
			c.startReconnect(context.Background(), b)
			c.isConnOk.Store(true)
		}
	}
}

func (c *ReConnPool) ReceiveSig(sig wsDefine.ReConnType) {
	select {
	case c.sigChan <- sig:
	default:
	}
}

func (c *ReConnPool) CloseSub(ctx context.Context) {
	c.thisCancel()
}

// 带指数退避的重连
func (c *ReConnPool) startReconnect(ctx context.Context, b *backoff.Backoff) {
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

func (c *ReConnPool) connect(ctx context.Context) error {
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
		case resourceEnum.MARK_PRICE, resourceEnum.BOOK_TICK, resourceEnum.AGG_TRADE,
			resourceEnum.SYMBOL_SUB_256, resourceEnum.FORCE_ORDER:
			safeRead := wsRead.NewReadPool(conn.GetConn(), byteBufPool.SIZE_256, c.resourceId)
			safeRead.ReadPoolLoop(ctxStopChild, c.read, c.sigChan)
		default:
		}
	case exchangeEnum.BYBIT:
		//创建ping-pong对象
		pingPong := wsPingPong.NewPingPong(wsPingPong.PingByBit, conn, c.resourceId)
		pingPong.PingPongLoop(ctxStopChild, c.sigChan)
		switch c.resourceType {
		case resourceEnum.BOOK_TICK:
			safeRead := wsRead.NewReadPool(conn.GetConn(), byteBufPool.SIZE_256, c.resourceId)
			safeRead.ReadPoolLoop(ctxStopChild, c.read, c.sigChan)
		case resourceEnum.MARK_PRICE, resourceEnum.AGG_TRADE, resourceEnum.SYMBOL_SUB_256:
			safeRead := wsRead.NewReadPool(conn.GetConn(), byteBufPool.SIZE_1024, c.resourceId)
			safeRead.ReadPoolLoop(ctxStopChild, c.read, c.sigChan)
		default:
		}
	default:
	}
	return nil
}
