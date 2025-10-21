package latency

type Latency interface {
	Record(symbolName string, ts float64)
}
