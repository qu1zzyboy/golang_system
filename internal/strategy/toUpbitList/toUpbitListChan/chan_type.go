package toUpbitListChan

import "upbitBnServer/internal/infra/systemx"

type TrigOrderInfo struct {
	ClientOrderId systemx.WsId16B
	T             int64
	P             uint64
}

type MonitorResp struct {
	P float64 // 探针返回价格上下界限
}
