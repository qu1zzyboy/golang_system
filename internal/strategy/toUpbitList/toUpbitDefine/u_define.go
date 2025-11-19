package toUpbitDefine

type StopType uint8

const (
	StopByTreeNews StopType = iota
	StopByMoveStopLoss
	StopByBtTakeProfit
	StopByGetCmcFailure
	StopByGetRemoteFailure
	StopByMetaError

	TotalLen
)

var (
	StopReasonArr = [TotalLen]string{
		"未触发TreeNews",
		"移动止损触发",
		"BookTick止盈触发",
		"获取cmc_id失败",
		"获取远程参数失败",
		"交易元数据失败",
	}
)
