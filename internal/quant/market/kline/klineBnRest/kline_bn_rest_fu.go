package klineBnRest

import (
	"fmt"
	"strconv"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/market/kline/klineModel"
	"upbitBnServer/pkg/utils/timeUtils"
)

var fullFutureUrl = fmt.Sprintf("%s/fapi/v1/klines", bnConst.FUTURE_BASE_REST_URL)

func (s *restBn) getFutureKlineSlice(req *klineModel.KlineReq) ([]*klineModel.Kline, error) {
	trueUrl := fmt.Sprintf("%s?symbol=%s&interval=%s&limit=%d", fullFutureUrl, req.SymbolName, req.Interval, req.KlineSize)
	data, err := httpx.Get(trueUrl)
	if err != nil {
		return nil, err
	}
	j, err := newJSON(data)
	if err != nil {
		return nil, err
	}
	num := len(j.MustArray())
	klineList := make([]*klineModel.Kline, num)
	for i := range num {
		item := j.GetIndex(i)
		if len(item.MustArray()) < 11 {
			return nil, errorx.New(errCode.HTTP_DO_ERROR, "KLINE_HTTP_ERROR,K线数据格式错误,期望长度为11").WithMetadata(map[string]string{
				defineJson.RawJson: string(data),
			})
		}
		open, _ := strconv.ParseFloat(item.GetIndex(1).MustString(), 64)
		high, _ := strconv.ParseFloat(item.GetIndex(2).MustString(), 64)
		low, _ := strconv.ParseFloat(item.GetIndex(3).MustString(), 64)
		close, _ := strconv.ParseFloat(item.GetIndex(4).MustString(), 64)
		vol, _ := strconv.ParseFloat(item.GetIndex(5).MustString(), 64)
		qty, _ := strconv.ParseFloat(item.GetIndex(7).MustString(), 64)
		abv, _ := strconv.ParseFloat(item.GetIndex(9).MustString(), 64)
		abq, _ := strconv.ParseFloat(item.GetIndex(10).MustString(), 64)
		// [[1741968300000,"0.2792","0.2792","0.2787","0.2788","31442",1741968359999,"8772.1616",105,"8334","2324.6937","0"]]
		klineList[i] = &klineModel.Kline{
			OpenTimeStr:    timeUtils.GetSecTimeStrBy(item.GetIndex(0).MustInt64()),
			OpenTimeStamp:  item.GetIndex(0).MustInt64(),
			OpenPrice:      open,
			HighPrice:      high,
			LowPrice:       low,
			ClosePrice:     close,
			Volume:         vol,
			EndTimeStamp:   item.GetIndex(6).MustInt64(),
			Qty:            qty,
			TradeNumber:    item.GetIndex(8).MustInt64(),
			TakerBuyVolume: abv,
			TakerBuyQty:    abq,
		}
	}
	return klineList, nil
}
