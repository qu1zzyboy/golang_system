package toUpbitListChan

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

type Sig uint8

const (
	ReceiveTreeNews Sig = iota
	ReceiveNoTreeNews
	CancelOrderReturn
	ReceiveTreeNewsAndOpen
	FailureOrder
	QUERY_ACCOUNT_RETURN
)

type Special struct {
	Amount       decimal.Decimal
	ErrCode      int64 //错误码
	SigType      Sig
	AccountKeyId uint8
}

type Job struct {
	Buf *[]byte
	Len int
}

var (
	chanTrigOrderArr []chan TrigOrderInfo
	chanMonitorArr   []chan MonitorResp
	chanSpecialArr   []chan Special
	chanSuOrderArr   []chan toUpBitListDataAfter.OnSuccessEvt
)

func InitChanArr(size int) {
	chanTrigOrderArr = make([]chan TrigOrderInfo, size)
	chanMonitorArr = make([]chan MonitorResp, size)
	chanSpecialArr = make([]chan Special, size)
	chanSuOrderArr = make([]chan toUpBitListDataAfter.OnSuccessEvt, size)
}

func RegisterSpecial(symbolIndex systemx.SymbolIndex16I, chanSpecial chan Special) {
	chanSpecialArr[symbolIndex] = chanSpecial
}

func RegisterBnPrivate(symbolIndex systemx.SymbolIndex16I, trigOrderChan chan TrigOrderInfo, monitorChan chan MonitorResp, chanSuOrder chan toUpBitListDataAfter.OnSuccessEvt) {
	chanTrigOrderArr[symbolIndex] = trigOrderChan
	chanMonitorArr[symbolIndex] = monitorChan
	chanSuOrderArr[symbolIndex] = chanSuOrder
}

func RegisterByBitPrivate(symbolIndex systemx.SymbolIndex16I, trigOrderChan chan TrigOrderInfo, chanSuOrder chan toUpBitListDataAfter.OnSuccessEvt) {
	chanTrigOrderArr[symbolIndex] = trigOrderChan
	chanSuOrderArr[symbolIndex] = chanSuOrder
}

func SendTradeLite(symbolIndex systemx.SymbolIndex16I, trig TrigOrderInfo) {
	chanTrigOrderArr[symbolIndex] <- trig
}

func SendMonitorData(symbolIndex systemx.SymbolIndex16I, data MonitorResp) {
	chanMonitorArr[symbolIndex] <- data
}

func SendSuOrder(symbolIndex systemx.SymbolIndex16I, evt toUpBitListDataAfter.OnSuccessEvt) {
	chanSuOrderArr[symbolIndex] <- evt
}

func SendSpecial(symbolIndex systemx.SymbolIndex16I, spec Special) {
	chanSpecialArr[symbolIndex] <- spec
}
