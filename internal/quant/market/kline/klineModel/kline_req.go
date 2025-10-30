package klineModel

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/kline/klineEnum"
)

type KlineReq struct {
	SymbolName     string             //交易对
	KlineSize      int                //k线根数
	StartTimeStamp int64              //k线开始时间
	EndTimeStamp   int64              //k线结束时间
	Interval       klineEnum.Interval //k线周期
	AcType         exchangeEnum.AccountType
}

func (s *KlineReq) TypeName() string {
	return "MyKlineArrayReq"
}

func (s *KlineReq) Check() error {
	if s.SymbolName == "" {
		return errDefine.SymbolKeyEmpty
	}
	if s.KlineSize <= 0 {
		return errorx.Newf(errCode.INVALID_VALUE, "INVALID_KLINE_SIZE,[%s]K线根数:%d", s.SymbolName, s.KlineSize)
	}
	if s.StartTimeStamp < 0 {
		return errorx.Newf(errCode.INVALID_VALUE, "INVALID_START_TIME,[%s]开始时间:%d", s.SymbolName, s.StartTimeStamp)
	}
	if s.EndTimeStamp < 0 {
		return errorx.Newf(errCode.INVALID_VALUE, "INVALID_END_TIME,[%s]结束时间:%d", s.SymbolName, s.EndTimeStamp)
	}
	return nil
}
