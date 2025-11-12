package autoMarketBybitSub

import (
	"context"
	"time"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/resource/resourceEnum"
)

var (
	symObjArray []*ByBitSymbolSubAuto // 交易对信息
	resourceArr []resourceEnum.ResourceType
)

func InitObjArr(size int, resourceAutoArr []resourceEnum.ResourceType) {
	resourceArr = resourceAutoArr
	symObjArray = make([]*ByBitSymbolSubAuto, size)
	for i := range size {
		symObjArray[i] = newByBitSymbolSubAuto(systemx.SymbolIndex16I(i))
	}
}

func GetSymbolObj(symbolIndex systemx.SymbolIndex16I) *ByBitSymbolSubAuto {
	return symObjArray[symbolIndex]
}

func Register(ctx context.Context, initSymbols []string) {
	go func() {
		for index, symbolName := range initSymbols {
			if err := GetSymbolObj(systemx.SymbolIndex16I(index)).RegisterReadHandler(ctx, symbolName); err != nil {
				errorx.PanicWithCaller(err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
}

func OpenSub(ctx context.Context, stopIndex int) {
	for index, sym := range symObjArray {
		if index >= stopIndex {
			break
		}
		sym.OpenSub(ctx)
	}
}

func CloseSub(ctx context.Context, stopIndex int) {
	for index, sym := range symObjArray {
		if index >= stopIndex {
			break
		}
		sym.CloseSub(ctx)
	}
}
