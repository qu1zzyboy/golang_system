package idGen

func BuildName2(front, back string) string {
	return front + "." + back
}

func BuildName3(module, subModule, funcName string) string {
	return module + "." + subModule + "." + funcName
}
