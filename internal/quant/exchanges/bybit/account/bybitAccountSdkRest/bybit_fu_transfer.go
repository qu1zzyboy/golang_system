package bybitAccountSdkRest

import (
	"fmt"
	"net/http"
	"strconv"
	"upbitBnServer/internal/infra/httpx"
	"upbitBnServer/internal/quant/exchanges/bybit/account/bybitAccountDefine"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitConst"

	"github.com/google/uuid"
)

var fuTransferUrl = fmt.Sprintf("%s/v5/asset/transfer/universal-transfer", bybitConst.BASE_URL)

func (s *FutureRest) DoTransfer(req bybitAccountDefine.TransferReq) ([]byte, error) {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"transferId":"`...)
	buf = append(buf, uuid.New().String()...)

	buf = append(buf, `","coin":"`...)
	buf = append(buf, req.Coin...)

	buf = append(buf, `","amount":"`...)
	buf = append(buf, req.Amount.String()...)

	buf = append(buf, `","fromMemberId":`...)
	buf = strconv.AppendUint(buf, uint64(req.FromMemberId), 10)

	buf = append(buf, `,"toMemberId":`...)
	buf = strconv.AppendUint(buf, uint64(req.ToMemberId), 10)

	buf = append(buf, `,"fromAccountType":"`...)
	buf = append(buf, req.FromAccountType...)

	buf = append(buf, `","toAccountType":"`...)
	buf = append(buf, req.ToAccountType...)

	buf = append(buf, `"}`...)

	r, err := s.addSignPost(fuTransferUrl, buf)
	if err != nil {
		return nil, err
	}
	r.Method = http.MethodPost
	return httpx.DefaultClient.Do(r)
}
