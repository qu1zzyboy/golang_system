package klineBnRest

import (
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/kline/klineModel"

	"github.com/bitly/go-simplejson"
)

var (
	BnImpl = newRestBn()
)

type restBn struct {
}

func newRestBn() *restBn {
	return &restBn{}
}

func newJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (s *restBn) GetKlineSlice(req *klineModel.KlineReq) ([]*klineModel.Kline, error) {
	switch req.AcType {
	case exchangeEnum.FUTURE:
		return s.getFutureKlineSlice(req)
	default:
		return nil, req.AcType.GetNotSupportError("rest_kline")
	}
}
