package toUpbitListBybitSymbol

import (
	"upbitBnServer/pkg/container/ring/ringBuf"
)

// 写入序号 + 值,用于单调队列(只维护最小值队列)
type pwPairU64 struct {
	seq   int64
	value uint64
}

func (s *Single) newPriceMinWindowU64(n ringBuf.Capacity) {
}

// Commit 阶段 2：解析完价格后尝试提交
func (s *Single) commit(price uint64, threshold float64, ts int64) (riseValue float64, hasTrig, hasWrite bool) {
	// 1) 真正写入 ring
	s.r.Push(price)

	// 2) 维护单调最小队列
	s.seq++
	seq := s.seq
	// 维护单调递增队列(去掉 >= price 的队尾元素,因为它永远不会再成为最小值)
	for len(s.minQ) > 0 && s.minQ[len(s.minQ)-1].value >= price {
		s.minQ = s.minQ[:len(s.minQ)-1]
	}
	s.minQ = append(s.minQ, pwPairU64{seq, price})

	// 淘汰窗外元素(仅在满容量后触发)
	//当前窗口中,允许的最小序号(小于等于 limit 的都过期)
	limit := seq - int64(s.r.Capacity())
	for len(s.minQ) > 0 && s.minQ[0].seq <= limit {
		// 淘汰队首元素
		s.minQ = s.minQ[1:]
	}

	// 3) 计算结果(最新价就是本次 price)
	if s.r.Size() >= 2 && len(s.minQ) > 0 {
		minV := s.minQ[0].value
		if price > minV && minV > 0 {
			riseValue = float64(price-minV) / float64(minV)
			hasTrig = riseValue >= threshold
		}
	}
	// 4) 标记已提交 ts(受 mu 保护,语义比 seenTs 更“落地”)
	s.committedTs = ts
	return riseValue, hasTrig, true
}

func (s *Single) checkMarket(eventTs int64, priceU8 uint64) {
	// 更新两分钟之前的价格
	minuteId := eventTs / (60000)
	if minuteId > s.thisMinTs {
		s.thisMinTs = minuteId
		s.last2MinClose_8 = s.last1MinClose_8
		s.last1MinClose_8 = s.thisMinClose_8
	} else {
		s.thisMinClose_8 = priceU8
	}
}
