package orderSdkBnWsNoSign

import (
	"bytes"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/pkg/utils/time2str"
)

// 测试ok
// len:92
// {"id":"1762410225006qab","method":"v2/account.balance","params":{"timestamp":1762410225006}}

// resp 权重 5
// {"id":"1762410225006qab","status":200,"result":[{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"FDUSD","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"5042.63491298","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"LDUSDT","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"4526.55256413","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"BFUSD","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"5078.90733000","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"BNB","balance":"0.09190630","crossWalletBalance":"0.09190630","crossUnPnl":"0.00000000","availableBalance":"5.07928858","maxWithdrawAmount":"0.09190630","marginAvailable":true,"updateTime":1762263904097},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"ETH","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"1.42474971","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"BTC","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"0.04678856","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"USDT","balance":"5001.25610000","crossWalletBalance":"5001.25610000","crossUnPnl":"0.00000000","availableBalance":"5083.47788998","maxWithdrawAmount":"5001.25610000","marginAvailable":true,"updateTime":1762263904097},{"accountAlias":"SgAuTiAuAuFzAuXq","asset":"USDC","balance":"0.00000000","crossWalletBalance":"0.00000000","crossUnPnl":"0.00000000","availableBalance":"5083.65427713","maxWithdrawAmount":"0.00000000","marginAvailable":true,"updateTime":0}]}

func (s *FutureClient) QueryAccount(reqFrom usageEnum.Type) error {
	reqId := time2str.GetNowTimeStampMillSlice16()
	buf := make([]byte, 0, 128)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `qab","method":"v2/account.balance","params":{"timestamp":`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `}}`...)
	if err := s.conn.WriteAsync(buf); err != nil {
		return err
	}
	reqId[13] = 'q'
	reqId[14] = 'a'
	reqId[15] = 'b'
	wsRequestCache.GetCache().StoreMeta(reqId, wsRequestCache.WsRequestMeta{
		ReqJson:   bytes.Clone(buf),
		ReqType:   wsRequestCache.QUERY_ACCOUNT_BALANCE,
		UsageFrom: reqFrom,
	})
	return nil
}
