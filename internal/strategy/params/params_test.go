package params
type stubBTC struct{ snap params.BTCSnapshot }
func (s *stubBTC) Start(context.Context) error { return nil }
func (s *stubBTC) Stop(context.Context) error  { return nil }
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

svc := params.NewWithProviders(
    params.Config{}, 
    &stubBTC{snap: params.BTCSnapshot{BTC1D: 1.2, BTC7D: 3.4}},
    &stubFGI{val: 65},
    &stubOI{rec: map[string]params.OIRecord{
        "AAAUSDT": {OpenInterest: ptrFloat(1e6), Timestamp: 12345},
    }},
)
require.NoError(t, svc.Start(context.Background()))
resp, err := svc.Compute(context.Background(), params.ComputeRequest{
    MarketCapM: 50, SymbolName: "AAAUSDT",
})
