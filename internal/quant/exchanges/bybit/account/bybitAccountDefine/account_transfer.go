package bybitAccountDefine

import "github.com/shopspring/decimal"

type TransferReq struct {
	Coin            string
	Amount          decimal.Decimal
	FromAccountType string
	ToAccountType   string
	FromMemberId    uint32
	ToMemberId      uint32
}
