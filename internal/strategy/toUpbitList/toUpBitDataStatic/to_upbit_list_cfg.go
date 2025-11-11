package toUpBitDataStatic

type ConfigVir struct {
	MaxQty        float64  `json:"max_qty"`         // 最大开仓数量,USDT
	MaxCap        float64  `json:"max_cap"`         // 最大订阅市值,USDT
	MaxCap330     float64  `json:"max_cap_330"`     // 最大订阅市值,USDT
	MaxOi         float64  `json:"max_oi"`          // 最大持仓量,USDT
	PriceRiceTrig float64  `json:"price_rice_trig"` // 价格触发阈值,当价格变化超过该值时触发
	OrderRiceTrig float64  `json:"order_rice_trig"` // 下单触发阈值,当价格变化超过该值时下单
	MonthLen      int      `json:"month_len"`       // 持有时间,单位:月
	BlackList     []uint32 `json:"black_list"`      // 黑名单
	WhiteList     []uint32 `json:"white_list"`      // 白名单
}
