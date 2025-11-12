package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"
)

func (s *Single) checkTreeNews() {
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if s.hasTreeNews {
			return
		}
		s.receiveStop(toUpbitDefine.StopByTreeNews)
		toUpBitDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{"symbol": s.symbolName, "op": "TreeNews未确认"})
	}()
}

func (s *Single) ReceiveTreeNews() {
	toUpBitDataStatic.DyLog.GetLog().Info("--------------------TreeNews确认---------------------")
	s.hasTreeNews = true
	toUpBitDataStatic.SendToUpBitMsg("TreeNews确认", map[string]string{
		"symbol": s.symbolName,
		"op":     "TreeNews确认",
	})
}

func (s *Single) ReceiveNoTreeNews() {
	s.hasTreeNews = false
	s.receiveStop(toUpbitDefine.StopByTreeNews)
	toUpBitDataStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{"symbol": s.symbolName, "op": "TreeNews未确认"})
}
