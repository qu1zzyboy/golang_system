package driverListChan

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute/order/bnOrderDedup"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/newsDrive/common/driverOrderDedup"

	"github.com/shopspring/decimal"
)

type Sig uint8

const (
	ReceiveTreeNews Sig = iota
	ReceiveNoTreeNews
	CANCEL_ORDER_RETURN
	ReceiveTreeNewsAndOpen
	FailureOrder
	QUERY_ACCOUNT_RETURN
)

type Special struct {
	Amount       decimal.Decimal
	ErrCode      int64
	SigType      Sig
	AccountKeyId uint8
}

var (
	bookTickChanArr    []chan systemx.Job
	aggTradeChanArr    []chan systemx.Job
	markPriceChanArr   []chan systemx.Job
	chanTradeLiteArr   []chan []byte
	chanOrderUpdateArr []chan []byte
	chanMonitorArr     []chan []byte
	chanSuOrderArr     []chan orderModel.OnSuccessEvt
	chanFaOrderArr     []chan orderModel.OnFailedEvt
	chanSpecialArr     []chan Special
)

func InitUpBit(size int) {
	bookTickChanArr = make([]chan systemx.Job, size)
	aggTradeChanArr = make([]chan systemx.Job, size)
	markPriceChanArr = make([]chan systemx.Job, size)
	chanTradeLiteArr = make([]chan []byte, size)
	chanOrderUpdateArr = make([]chan []byte, size)
	chanMonitorArr = make([]chan []byte, size)
	chanSuOrderArr = make([]chan orderModel.OnSuccessEvt, size)
	chanSpecialArr = make([]chan Special, size)
}

func RegisterMarket(symbolIndex int, bookTickChan chan Job, aggTradeChan chan Job, markPriceChan chan Job) {
	bookTickChanArr[symbolIndex] = bookTickChan
	aggTradeChanArr[symbolIndex] = aggTradeChan
	markPriceChanArr[symbolIndex] = markPriceChan
}

func RegisterPrivate(symbolIndex int,
	tradeLiteChan chan []byte,
	deltaOrderChan chan []byte,
	monitorChan chan []byte,
	chanSpecial chan Special,
	chanSuOrder chan orderModel.OnSuccessEvt) {
	chanTradeLiteArr[symbolIndex] = tradeLiteChan
	chanOrderUpdateArr[symbolIndex] = deltaOrderChan
	chanMonitorArr[symbolIndex] = monitorChan
	chanSpecialArr[symbolIndex] = chanSpecial
	chanSuOrderArr[symbolIndex] = chanSuOrder
}

func SendBookTick(symbolIndex int, buf *[]byte, len int) {
	bookTickChanArr[symbolIndex] <- Job{Buf: buf, Len: len}
}

func SendAggTrade(symbolIndex int, buf *[]byte, len int) {
	aggTradeChanArr[symbolIndex] <- Job{Buf: buf, Len: len}
}

func SendMarkPrice(symbolIndex int, buf *[]byte, len int) {
	markPriceChanArr[symbolIndex] <- Job{Buf: buf, Len: len}
}

func SendTradeLite(symbolIndex int, buf []byte) {
	chanTradeLiteArr[symbolIndex] <- buf
}

func SendDeltaOrder(symbolIndex int, buf []byte) {
	chanOrderUpdateArr[symbolIndex] <- buf
}

func SendMonitorData(symbolIndex int, data []byte) {
	chanMonitorArr[symbolIndex] <- data
}

func SendSuOrder(symbolIndex systemx.SymbolIndex16I, evt orderModel.OnSuccessEvt) {
	//唯一性去重校验
	if driverOrderDedup.Deduper.ExistsOrInsert(bnOrderDedup.OrderUnique{T: evt.T, ID: bnOrderDedup.GetId(evt.ClientOrderId, evt.OrderStatus)}) {
		return
	}
	chanSuOrderArr[symbolIndex] <- evt
}

func SendFaOrder(symbolIndex systemx.SymbolIndex16I, evt orderModel.OnFailedEvt) {
	chanFaOrderArr[symbolIndex] <- evt
}

func SendSpecial(symbolIndex systemx.SymbolIndex16I, spec Special) {
	chanSpecialArr[symbolIndex] <- spec
}
