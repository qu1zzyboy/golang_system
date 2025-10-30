package coinMesh

import (
	"time"

	"upbitBnServer/internal/define/defineTime"
)

//必须先获取cmc信息,因为要做check

func (m *Manager) SetCmcInfo(cmcId, cmcRank uint32, marketCap, supply float64, asset, addStr, describe string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		mesh.CmcRank = cmcRank
		mesh.MarketCap = marketCap
		mesh.SupplyNow = supply
		mesh.CmcDesc = describe
		mesh.CmcAsset = asset
		t, err := time.Parse(defineTime.FormatCmc, addStr)
		if err == nil {
			mesh.CmcDateAdded = t.UnixMilli()
			mesh.CmcDateAddedStr = t.Format(defineTime.FormatHour)
		}
	} else {
		mesh = &CoinMesh{
			CmcId:     cmcId,
			CmcRank:   cmcRank,
			MarketCap: marketCap,
			SupplyNow: supply,
			CmcDesc:   describe,
			CmcAsset:  asset,
		}
		t, err := time.Parse(defineTime.FormatCmc, addStr)
		if err == nil {
			mesh.CmcDateAdded = t.UnixMilli()
			mesh.CmcDateAddedStr = t.Format(defineTime.FormatHour)
		}
		m.coinMap.Store(cmcId, mesh)
	}
}
