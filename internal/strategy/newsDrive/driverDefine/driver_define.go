package driverDefine

type StopType uint8

const (
	StopByTreeNews StopType = iota
	StopByMoveStopLoss
	StopByTakeProfit
	StopByGetCmcFailure
	StopByGetRemoteFailure
	StopByTimeLine
	StopByStopLoss
	TotalLen
)

var (
	StopReasonArr = [TotalLen]string{
		"未触发TreeNews",
		"%5移动止损触发",
		"BookTick止盈触发",
		"获取cmc_id失败",
		"获取远程参数失败",
		"触发平仓时间",
		"止损触发",
	}
)
