package bnOrderTemplate

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

// -16,-3
// len195
// {"id":"M762412446009244","method":"order.modify","params":{"origClientOrderId":"1762410636008457","price":"3200.00","quantity":"0.01","side":"BUY","symbol":"ETHUSDT","timestamp":"1762412446009"}}

var (
	id_begin  uint16 = 8
	id_end    uint16 = 23
	p_begin   uint16 = 107
	cid_begin uint16 = 80
	cid_end   uint16 = 96
)

type ModifyTemplate struct {
	modify    []byte
	modifyLen uint16
}

func NewModifyTemplate() *ModifyTemplate {
	return &ModifyTemplate{
		modify: make([]byte, 0, 256),
	}
}

func (s *ModifyTemplate) Start(symbolName string, pVal, qVal systemx.ScaledValue, pScale systemx.PScale, qScale systemx.QScale, cId systemx.WsId16B) {
	s.modify = append(s.modify, `{"id":"M762412446009244","method":"order.modify","params":{"origClientOrderId":"`...)
	s.modify = append(s.modify, cId[:]...)
	s.modify = append(s.modify, `","price":"`...)
	s.modify = byteUtils.AppendScaledValue(s.modify, uint64(pVal), int(pScale))
	s.modify = append(s.modify, `","quantity":"`...)
	s.modify = byteUtils.AppendScaledValue(s.modify, uint64(qVal), int(qScale))
	s.modify = append(s.modify, `","side":"SELL","symbol":"`...)
	s.modify = append(s.modify, symbolName...)
	s.modify = append(s.modify, `","timestamp":"1762412446009"}}`...)
	s.modifyLen = uint16(len(s.modify))
}

// 58.71 ns/op	       0 B/op	       0 allocs/op

func (s *ModifyTemplate) RefreshPrice(pVal uint64, pScale systemx.PScale) []byte {
	// 价格替换
	byteUtils.CopyScaledValue(s.modify, p_begin, pVal, int(pScale))
	return s.modify
}

func (s *ModifyTemplate) GetModifyRaw(reqId systemx.WsId16B) []byte {
	// id 替换
	copy(s.modify[id_begin:id_end], reqId[1:])

	// 时间戳替换（13 位）
	copy(s.modify[s.modifyLen-16:s.modifyLen-3], reqId[:13])
	return s.modify
}
