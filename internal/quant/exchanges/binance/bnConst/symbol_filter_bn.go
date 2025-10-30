package bnConst

var bn_spot_unneed map[string]bool   //剔除不需要的现货交易对
var bn_future_unneed map[string]bool //剔除不需要的合约交易对
var bn_spot_asset_unused map[string]bool

func initFutureUnneed() {
	bn_future_unneed = make(map[string]bool)
	bn_future_unneed["BTCDOMUSDT"] = true
	bn_future_unneed["DEFIUSDT"] = true
}

func initSpotUnneed() {
	bn_spot_unneed = make(map[string]bool)
	//meme
	bn_spot_unneed["SHIBUSDT"] = true
	bn_spot_unneed["XECUSDT"] = true
	bn_spot_unneed["LUNCUSDT"] = true
	bn_spot_unneed["PEPEUSDT"] = true
	bn_spot_unneed["FLOKIUSDT"] = true
	bn_spot_unneed["BONKUSDT"] = true
	bn_spot_unneed["1000SATSUSDT"] = true
	bn_spot_unneed["RATSUSDT"] = true
	//封装
	bn_spot_unneed["WBETHUSDT"] = true
	bn_spot_unneed["WBTCUSDT"] = true
	bn_spot_unneed["FTTUSDT"] = true
	//重组
	bn_spot_unneed["LUNAUSDT"] = true
	bn_spot_unneed["DODOUSDT"] = true
	//稳定币
	bn_spot_unneed["TUSDUSDT"] = true
	bn_spot_unneed["AEURUSDT"] = true
	bn_spot_unneed["USDPUSDT"] = true
	bn_spot_unneed["FDUSDUSDT"] = true
	bn_spot_unneed["USDCUSDT"] = true
	bn_spot_unneed["EURUSDT"] = true
	bn_spot_unneed["PAXGUSDT"] = true
}

func initSpotAssetUnused() {
	bn_spot_asset_unused = make(map[string]bool)
	bn_spot_asset_unused["USDT"] = true
	bn_spot_asset_unused["FDUSD"] = true
	bn_spot_asset_unused["LDUSDT"] = true
	bn_spot_asset_unused["MITH"] = true
	bn_spot_asset_unused["ETHW"] = true
	bn_spot_asset_unused["LDFUN"] = true
	bn_spot_asset_unused["LDSHIB2"] = true
	bn_spot_asset_unused["LDAVAX"] = true
	bn_spot_asset_unused["LDBTTC"] = true
	bn_spot_asset_unused["LDLUNC"] = true
}

func init() {
	initFutureUnneed()
	initSpotUnneed()
	initSpotAssetUnused()
}

func IsBnSpotSymbolNameOk(symbol string) bool {
	return !bn_spot_unneed[symbol]
}

func IsBnSpotAssetUnused(asset string) bool {
	return bn_spot_asset_unused[asset]
}
