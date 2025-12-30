package toUpbitListBnSymbol

import (
	"sync/atomic"

	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

var (
	trans = [11]int32{
		3,  //账户0-->账户3
		4,  //账户1-->账户4
		5,  //账户2-->账户5
		6,  //账户3-->账户6
		7,  //账户4-->账户7
		8,  //账户5-->账户8
		9,  //账户6-->账户9
		10, //账户7-->账户10
		1,  //账户8-->账户1
		2,  //账户9-->账户2
		3,  //账户10-->账户3
	}
)

type secondPerInfo struct {
	maxNotional          atomic.Value
	hasInToSecondPerLoop atomic.Bool
	stopThisSecondPer    atomic.Bool
}

func (s *secondPerInfo) clear() {
	if s.maxNotional.Load() != nil {
		s.maxNotional.Store(decimal.Zero)
	}
	s.hasInToSecondPerLoop.Store(false)
	s.stopThisSecondPer.Store(false)
}

func (s *secondPerInfo) start() {
	s.stopThisSecondPer.Store(false)   // 开启本轮抽奖信号
	s.hasInToSecondPerLoop.Store(true) // 确认进入了每秒抽奖循环
}

func (s *secondPerInfo) receiveStop(accountKeyId uint8) {
	if s.hasInToSecondPerLoop.Load() {
		if !s.stopThisSecondPer.Load() {
			s.stopThisSecondPer.Store(true)
			toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d]没钱,停止这一秒抽奖", accountKeyId)
		}
	}
}

func (s *secondPerInfo) loadStop() bool {
	return s.stopThisSecondPer.Load()
}

// 给定一个整数 i,算出它在 1–10 的“循环序列”里的前2个下标
func getPreIndex(i int32) int32 {
	res := i%10 + 8
	if res > 10 {
		res = res - 10
	}
	return res
}

// 把任意整数 i 映射到 1–10 之间
func getCurIndex(i int32) int32 {
	res := i % 10
	if res == 0 {
		res = 10
	}
	return res
}
