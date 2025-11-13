package byBitOrderAppManager

import "upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

func (s *TradeManager) SetByBitLeverage(leverage uint8, symbolName string) error {
	for k, app := range s.appArray {
		res, err := app.rest.DoLeverage(leverage, symbolName)
		if err != nil {
			toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]设置bybit[%s]杠杆失败,leverage:%d,err:%v", k, symbolName, leverage, err)
			return err
		}
		toUpBitDataStatic.DyLog.GetLog().Infof("[%d]设置bybit[%s]杠杆成功,leverage:%d,res:%v", k, symbolName, leverage, string(res))
	}
	return nil
}
