package toUpbitListChan

type Job struct {
	Buf *[]byte
	Len int
}

var (
	bookTickChanArr   []chan Job
	aggTradeChanArr   []chan Job
	markPriceChanArr  []chan Job
	chanTradeLiteArr  []chan []byte
	chanDeltaOrderArr []chan []byte
	chanMonitorArr    []chan []byte
)

func InitUpBit(size int) {
	bookTickChanArr = make([]chan Job, size)
	aggTradeChanArr = make([]chan Job, size)
	markPriceChanArr = make([]chan Job, size)
	chanTradeLiteArr = make([]chan []byte, size)
	chanDeltaOrderArr = make([]chan []byte, size)
	chanMonitorArr = make([]chan []byte, size)
}

func RegisterMarket(symbolIndex int, bookTickChan chan Job, aggTradeChan chan Job, markPriceChan chan Job) {
	bookTickChanArr[symbolIndex] = bookTickChan
	aggTradeChanArr[symbolIndex] = aggTradeChan
	markPriceChanArr[symbolIndex] = markPriceChan
}

func RegisterPrivate(symbolIndex int, tradeLiteChan chan []byte, deltaOrderChan chan []byte, markPriceChan chan []byte) {
	chanTradeLiteArr[symbolIndex] = tradeLiteChan
	chanDeltaOrderArr[symbolIndex] = deltaOrderChan
	chanMonitorArr[symbolIndex] = markPriceChan
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
	chanDeltaOrderArr[symbolIndex] <- buf
}

func SendMonitorData(symbolIndex int, data []byte) {
	chanMonitorArr[symbolIndex] <- data
}
