package toUpbitListChan

import (
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

type Sig uint8

const (
	ReceiveTreeNews Sig = iota
	ReceiveNoTreeNews
	CancelOrderReturn
	FailureOrder
	QUERY_ACCOUNT_RETURN
)

type Special struct {
	Amount       decimal.Decimal
	ErrCode      int64
	SigType      Sig
	AccountKeyId uint8
}

type Job struct {
	Buf *[]byte
	Len int
}

var (
	bookTickChanArr    []chan Job
	aggTradeChanArr    []chan Job
	markPriceChanArr   []chan Job
	chanTradeLiteArr   []chan []byte
	chanOrderUpdateArr []chan []byte
	chanMonitorArr     []chan []byte
	chanSuOrderArr     []chan toUpBitListDataAfter.OnSuccessEvt
	chanSpecialArr     []chan Special
)

func InitUpBit(size int) {
	bookTickChanArr = make([]chan Job, size)
	aggTradeChanArr = make([]chan Job, size)
	markPriceChanArr = make([]chan Job, size)
	chanTradeLiteArr = make([]chan []byte, size)
	chanOrderUpdateArr = make([]chan []byte, size)
	chanMonitorArr = make([]chan []byte, size)
	chanSuOrderArr = make([]chan toUpBitListDataAfter.OnSuccessEvt, size)
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
	chanSuOrder chan toUpBitListDataAfter.OnSuccessEvt) {
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

func SendSuOrder(symbolIndex int, evt toUpBitListDataAfter.OnSuccessEvt) {
	chanSuOrderArr[symbolIndex] <- evt
}

func SendSpecial(symbolIndex int, amount decimal.Decimal, errCode int64, sigType Sig, accountKeyId uint8) {
	chanSpecialArr[symbolIndex] <- Special{Amount: amount, ErrCode: errCode, SigType: sigType, AccountKeyId: accountKeyId}
}
