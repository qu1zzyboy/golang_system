package klineModel

import (
	"github.com/sirupsen/logrus"
)

type Kline struct {
	OpenTimeStr     string  //字符串形式的开盘时间 2021-01-01 00:00:00
	OpenTimeStamp   int64   //开盘毫秒级时间戳
	EndTimeStamp    int64   //结束毫秒级时间戳
	TradeNumber     int64   //成交笔数
	OpenPrice       float64 //开盘价
	HighPrice       float64 //最高价
	LowPrice        float64 //最低价
	ClosePrice      float64 //收盘价
	Volume          float64 //这根K线期间成交量
	Qty             float64 //这根K线期间成交额
	TakerBuyVolume  float64 //主动买入的成交量
	TakerBuyQty     float64 //主动买入的成交额
	TakerSellVolume float64 //主动卖出的成交量
	TakerSellQty    float64 //主动卖出的成交额
}

func (s *Kline) PrintMe(log *logrus.Logger) {
	log.Debugf("时间:[%s,%d] O:%.8f H:%.8f L:%.8f C:%.8f N:%d V:%.2f Q:%.2f ABV:%.2f ABQ:%.2f\n",
		s.OpenTimeStr, s.OpenTimeStamp, s.OpenPrice, s.HighPrice, s.LowPrice, s.ClosePrice, s.TradeNumber, s.Volume, s.Qty, s.TakerBuyVolume, s.TakerBuyQty)
}
