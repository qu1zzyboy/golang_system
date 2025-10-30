package coinMesh

import "upbitBnServer/internal/infra/observe/log/dynamicLog"

func (m *Manager) SetBybitFutureUsdtNameCheck(cmcId uint32, tradeAsset, bybitFutureUsdtName string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		if mesh.CmcAsset == tradeAsset {
			mesh.BybitFuUsdtName = bybitFutureUsdtName
		} else {
			if mesh.CmcAsset != "" {
				dynamicLog.Error.GetLog().Errorf("[COIN_MESH] BYBIT_FU 传入的币种与已有币种不符,传入cmcId:%d,tradeAsset:%s,已有cmcAsset:%s",
					cmcId, tradeAsset, mesh.CmcAsset)
			}
		}
	} else {
		mesh = &CoinMesh{
			CmcId:           cmcId,
			BybitFuUsdtName: bybitFutureUsdtName,
		}
		m.coinMap.Store(cmcId, mesh)
	}
}

func (m *Manager) GetAllBybitFuUsdtName() (list []string) {
	m.coinMap.Range(func(_ uint32, value *CoinMesh) bool {
		if value.BybitFuUsdtName != "" {
			list = append(list, value.BybitFuUsdtName)
		}
		return true
	})
	return list
}
