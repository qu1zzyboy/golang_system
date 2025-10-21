package toUpbitMesh

type Save struct {
	SymbolName string `json:"symbolName"`
	Reason     string `json:"reason"`
	IsList     bool   `json:"isList"`
}

const REDIS_KEY_TO_UPBIT_LIST_COIN_BN = "TO_UPBIT_LIST_COIN_BN"
