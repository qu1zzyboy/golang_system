package bnOrderDedup

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/execute"
)

type OrderUnique struct {
	T  int64  //毫秒时间戳
	ID uint64 //从 16 字节压缩得来，包含状态+下单微秒时间
}

// HashOrderKey hashOrderKey 是为 OrderKey 优化过的哈希函数
// 特点：
// - 无分支
// - 无内存访问
// - 无循环
// - 无乘法
// - 仅一条 XOR CPU 指令
// - 分布性足够好（ID 与 T 都是高熵）
func HashOrderKey(k OrderUnique) uint64 {
	return k.ID ^ uint64(k.T)
}

func GetId(clientOrderId systemx.WsId16B, orderStatus execute.OrderStatus) uint64 {
	n := uint64(orderStatus)
	for i := 1; i < 16; i++ {
		n = n*10 + uint64(clientOrderId[i]-'0')
	}
	return n
}
