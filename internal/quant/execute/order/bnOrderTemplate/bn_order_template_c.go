package bnOrderTemplate

import (
	"upbitBnServer/internal/infra/systemx"
)

// -16,-3
// len 146
// {"id":"C762412447010163","method":"order.cancel","params":{"origClientOrderId":"1762410636008457","symbol":"ETHUSDT","timestamp":"1762412447010"}}

type CancelTemplate struct {
	cancel    []byte
	cancelLen uint16
}

func NewCancelTemplate() *CancelTemplate {
	return &CancelTemplate{
		cancel: make([]byte, 0, 256),
	}
}

func (s *CancelTemplate) Start(symbolName string) {
	s.cancel = append(s.cancel, `{"id":"C762412447010163","method":"order.cancel","params":{"origClientOrderId":"1762410636008457","symbol":"`...)
	s.cancel = append(s.cancel, symbolName...)
	s.cancel = append(s.cancel, `","timestamp":"1762412446009"}}`...)
	s.cancelLen = uint16(len(s.cancel))
}

// 58.71 ns/op	       0 B/op	       0 allocs/op

func (s *CancelTemplate) RefreshClientOrderId(cId systemx.WsId16B) {
	// clientOrderId替换
	copy(s.cancel[cid_begin:cid_end], cId[:])
}

func (s *CancelTemplate) GetCancelRaw(reqId systemx.WsId16B) []byte {
	// id 替换
	copy(s.cancel[id_begin:id_end], reqId[1:])

	// 时间戳替换（13 位）
	copy(s.cancel[s.cancelLen-16:s.cancelLen-3], reqId[:13])
	return s.cancel
}
