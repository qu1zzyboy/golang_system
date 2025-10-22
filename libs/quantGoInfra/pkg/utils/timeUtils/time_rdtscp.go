package timeUtils

import "github.com/hhh500/quantGoInfra/conf"

//cycles=1s*Cpu_HZ

// Rdtscp 读取时间戳计数器(TSC)与 AUX(通常可视为核/NUMA 提示)
// 返回：tsc(cycles),aux(ECX)
func Rdtscp() (tsc uint64, aux uint32)

func CyclesToNs(cycles uint64) uint64 {
	return (cycles * ns_PER_CYCLE_Q32) >> 32
}

func CyclesToUs(cycles uint64) uint64 {
	return (cycles * us_PER_CYCLE_Q32) >> 32
}

var (
	//  每个 cycle 对应多少微秒 (放大 2^32 做定点)
	us_PER_CYCLE_Q32 = (1_000_000 << 32) / conf.CPU_HZ
	ns_PER_CYCLE_Q32 = (1_000_000_000 << 32) / conf.CPU_HZ
)
