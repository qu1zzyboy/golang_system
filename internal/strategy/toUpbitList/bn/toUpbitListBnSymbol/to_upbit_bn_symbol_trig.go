package toUpbitListBnSymbol

import (
	"fmt"

	"github.com/hhh500/quantGoInfra/pkg/container/ring/ringBuf"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

// 写入序号 + 值,用于单调队列(只维护最小值队列)
type pwPairU64 struct {
	seq   int64
	value uint64
}

type trigHotData struct {
	minQ        []pwPairU64           // 单调递增队列,队首为当前窗口最小价
	r           *ringBuf.Ring[uint64] // 环形缓冲区,存储价格
	seq         int64                 // 成功写入次数(严格递增)
	committedTs int64                 // 【已提交】最后一次成功写入的 ts
}

func (w *trigHotData) newPriceMinWindowU64(n ringBuf.Capacity) {
	w.r = ringBuf.NewPow2[uint64](n)
	w.minQ = make([]pwPairU64, 0, int(n))
}

// Commit 阶段 2：解析完价格后尝试提交
func (w *trigHotData) commit(price uint64, threshold float64, ts int64) (riseValue float64, hasTrig, hasWrite bool) {
	// 1) 真正写入 ring
	w.r.Push(price)

	// 2) 维护单调最小队列
	w.seq++
	s := w.seq
	// 维护单调递增队列(去掉 >= price 的队尾元素,因为它永远不会再成为最小值)
	for len(w.minQ) > 0 && w.minQ[len(w.minQ)-1].value >= price {
		w.minQ = w.minQ[:len(w.minQ)-1]
	}
	w.minQ = append(w.minQ, pwPairU64{s, price})

	// 淘汰窗外元素(仅在满容量后触发)
	//当前窗口中,允许的最小序号(小于等于 limit 的都过期)
	limit := s - int64(w.r.Capacity())
	for len(w.minQ) > 0 && w.minQ[0].seq <= limit {
		// 淘汰队首元素
		w.minQ = w.minQ[1:]
	}

	// 3) 计算结果(最新价就是本次 price)
	if w.r.Size() >= 2 && len(w.minQ) > 0 {
		minV := w.minQ[0].value
		if price > minV && minV > 0 {
			riseValue = float64(price-minV) / float64(minV)
			hasTrig = riseValue >= threshold
		}
	}
	// 4) 标记已提交 ts(受 mu 保护,语义比 seenTs 更“落地”)
	w.committedTs = ts
	return riseValue, hasTrig, true
}

func (s *Single) checkMarket(eventTs int64, trigFlag string, priceU64_8 uint64) {
	// 写入价格到环形缓冲区
	riseValue, hasTrig, hasWrite := s.commit(priceU64_8, toUpBitListDataStatic.PriceRiceTrig, eventTs)
	//涨幅触发
	if hasTrig {
		s.IntoExecuteCheck(eventTs, trigFlag, riseValue, priceU64_8)
	}
	//写入成功就更新两分钟之前的价格
	if hasWrite {
		// 涨幅大于0.05并且比上一次递增1%以上
		if riseValue > toUpBitListDataStatic.OrderRiceTrig && riseValue > 0.01+s.lastRiseValue {
			toUpBitListDataStatic.SendToUpBitMsg("发送bn快速上涨消息失败", map[string]string{
				"msg":  trigFlag + "快速上涨",
				"bn品种": s.StMeta.SymbolName,
				"上涨幅度": fmt.Sprintf("%.2f%%", riseValue*100),
			})
		}
		// 保存当前已实现涨幅
		s.lastRiseValue = riseValue
		// 更新两分钟之前的价格
		minuteId := eventTs / (60000)
		if minuteId > s.thisMinTs {
			s.thisMinTs = minuteId
			s.last2MinClose_8 = s.last1MinClose_8
			s.last1MinClose_8 = s.thisMinClose_8
		} else {
			s.thisMinClose_8 = priceU64_8
		}
	}
}

func (s *Single) onOrderPriceCheck(eventTs int64, priceU64_8 uint64, riseValue float64) {
	//涨幅触发
	if riseValue >= toUpBitListDataStatic.OrderRiceTrig {
		toUpBitListDataStatic.SendToUpBitMsg("发送bn快速上涨消息失败", map[string]string{
			"msg":  "orderPrice快速上涨",
			"bn品种": s.StMeta.SymbolName,
			"上涨幅度": fmt.Sprintf("%.2f%%", riseValue*100),
		})
		s.IntoExecuteCheck(eventTs, "preOrder", riseValue, priceU64_8)
	} else {
		toUpBitListDataStatic.SendToUpBitMsg("成交但不满足上市check消息失败", map[string]string{
			"msg":  fmt.Sprintf("成交但不满足上市check,成交价:%d", priceU64_8),
			"bn品种": s.StMeta.SymbolName,
			"上涨幅度": fmt.Sprintf("%.2f%%", riseValue*100),
		})
	}
}
