package coinMesh

import "fmt"

type CoinMesh struct {
	CmcAsset        string  `json:"cmc_asset"`          //币种标识
	CmcDesc         string  `json:"cmc_desc"`           //币种描述
	BnFuUsdtName    string  `json:"bn_fu_usdt_name"`    //币安usdt本位合约名称
	BnFuUsdcName    string  `json:"bn_fu_usdc_name"`    //币安usdc本位合约名称
	BnSpUsdtName    string  `json:"bn_sp_usdt_name"`    //币安usdt现货名称
	BybitFuUsdtName string  `json:"bybit_fu_usdt_name"` //Bybit usdt本位合约名称
	UpbitSpKrwName  string  `json:"upbit_sp_krw_name"`  //Upbit krw现货名称
	CmcDateAddedStr string  `json:"cmc_date_added_str"` //cmc币种添加时间
	CmcDateAdded    int64   `json:"cmc_date_added"`     //cmc币种添加时间
	MarketCap       float64 `json:"market_cap"`         //cmc市值
	SupplyNow       float64 `json:"supply_now"`         //cmc当前流通量
	CmcId           uint32  `json:"cmc_id"`             //cmc币种id
	CmcRank         uint32  `json:"cmc_rank"`           //cmc币种排名
	IsMeMe          bool    `json:"is_meme"`            // 是否是 Meme 币
}

func (c *CoinMesh) GetTotalInfo() string {
	return fmt.Sprintf("%d CmcAsset=%-10s |cap=%.2f | BnFutureUsdt=%-15s | BnFutureUsdc=%-12s | BnSpotUsdt=%-12s | BybitFutureUsdt=%-14s | UpbitSpotKrw=%-12s  [%s,%d] CmcDesc=%-12s",
		c.CmcRank,
		c.CmcAsset,
		c.MarketCap,
		c.BnFuUsdtName,
		c.BnFuUsdcName,
		c.BnSpUsdtName,
		c.BybitFuUsdtName,
		c.UpbitSpKrwName,
		c.CmcDateAddedStr,
		c.CmcDateAdded,
		c.CmcDesc,
	)
}

func (c *CoinMesh) ToUpBitInfo(flag string) string {
	return fmt.Sprintf("原因:%s\n名称:%s\n 市值:[rank:%d,cap:%.2f]\n [bn:%s,bybit:%s]\n 上架时间[%s,%d]\n 描述:%s",
		flag,
		c.CmcAsset,
		c.CmcRank, c.MarketCap,
		c.BnFuUsdtName, c.BybitFuUsdtName,
		c.CmcDateAddedStr, c.CmcDateAdded,
		c.CmcDesc,
	)
}
