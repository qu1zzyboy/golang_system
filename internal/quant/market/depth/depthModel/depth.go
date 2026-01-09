package depthModel

// DepthUpdate 订单簿深度更新（核心信息）
type DepthUpdate struct {
	Symbol            string     // 交易对
	EventTime         int64      // 事件时间 (E)
	TransactionTime   int64      // 交易时间 (T)
	FirstUpdateId     int64      // 首次更新ID (U)
	FinalUpdateId     int64      // 最终更新ID (u)
	PrevFinalUpdateId int64      // 上次流的最终更新ID (pu)
	Bids              [][]string // 买单深度 [[价格, 数量], ...]
	Asks              [][]string // 卖单深度 [[价格, 数量], ...]
}
