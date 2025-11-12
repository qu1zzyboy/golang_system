package byteConvert

// 6.578 ns/op	       0 B/op	       0 allocs/op

func ByteArrToF64(p []byte) float64 {
	var intPart uint64
	var fracPart uint64
	fracScale := uint64(1)

	i := 0
	for ; i < len(p) && p[i] != '.'; i++ {
		intPart = intPart*10 + uint64(p[i]-'0')
	}
	if i < len(p) && p[i] == '.' {
		i++
		for ; i < len(p); i++ {
			fracPart = fracPart*10 + uint64(p[i]-'0')
			fracScale *= 10
		}
	}
	return float64(intPart) + float64(fracPart)/float64(fracScale)
}
