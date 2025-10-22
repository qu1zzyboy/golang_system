package main

import (
	"context"
	"fmt"

	// ⚠️ 把下面的模块路径替换成你自己的 module 名称
	// 比如你的 go.mod 里第一行是: module github.com/you/myproject
	// 那么这里就写: "github.com/you/myproject/params"
	"github.com/hhh500/upbitBnServer/internal/strategy/params"
)

type stubBTC struct{ snap params.BTCSnapshot }

func (s *stubBTC) Start(context.Context) error  { return nil }
func (s *stubBTC) Stop(context.Context) error   { return nil }
func (s *stubBTC) Snapshot() params.BTCSnapshot { return s.snap }

type stubFGI struct{ val int }

func (s *stubFGI) Start(context.Context) error { return nil }
func (s *stubFGI) Stop(context.Context) error  { return nil }
func (s *stubFGI) GetValue() int               { return s.val }

type stubOI struct{ rec map[string]params.OIRecord }

func (s *stubOI) Start(context.Context) error { return nil }
func (s *stubOI) Stop(context.Context) error  { return nil }
func (s *stubOI) Get(sym string) (params.OIRecord, bool) {
	v, ok := s.rec[sym]
	return v, ok
}

func ptrFloat(v float64) *float64 { return &v }

func main() {
	svc := params.NewWithProviders(
		params.Config{},
		&stubBTC{snap: params.BTCSnapshot{BTC1D: 1.2, BTC7D: 3.4}},
		&stubFGI{val: 65},
		&stubOI{rec: map[string]params.OIRecord{
			"AAAUSDT": {OpenInterest: ptrFloat(1e6), Timestamp: 12345},
		}},
	)

	if err := svc.Start(context.Background()); err != nil {
		panic(err)
	}

	resp, err := svc.Compute(context.Background(), params.ComputeRequest{
		MarketCapM: 50,
		SymbolName: "AAAUSDT",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Compute result: %+v\n", resp)
}
