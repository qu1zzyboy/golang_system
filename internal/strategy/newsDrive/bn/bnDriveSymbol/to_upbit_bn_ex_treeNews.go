package bnDriveSymbol

import (
	"time"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
)

func (s *Single) checkTreeNews() {
	//开启二次校验等待循环
	go func() {
		time.Sleep(2 * time.Second)
		if s.hasTreeNews {
			return
		}
		s.receiveStop(StopByTreeNews)
		driverStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
			"symbol": s.StMeta.SymbolName,
			"op":     "TreeNews未确认",
		})
	}()
}

func (s *Single) ReceiveTreeNews() {
	driverStatic.DyLog.GetLog().Info("--------------------TreeNews确认---------------------")
	s.hasTreeNews = true
	driverStatic.SendToUpBitMsg("TreeNews确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews确认",
	})
}

func (s *Single) ReceiveNoTreeNews() {
	s.hasTreeNews = false
	s.receiveStop(StopByTreeNews)
	driverStatic.SendToUpBitMsg("TreeNews未确认", map[string]string{
		"symbol": s.StMeta.SymbolName,
		"op":     "TreeNews未确认",
	})
}
