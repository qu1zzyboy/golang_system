package coinMesh

import "github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"

func (m *Manager) SetUpbitSpotKrwNameCheck(cmcId uint32, tradeAsset, upbitSpotKrwName string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		if mesh.CmcAsset == tradeAsset {
			mesh.UpbitSpKrwName = upbitSpotKrwName
		} else {
			dynamicLog.Error.GetLog().Errorf("[COIN_MESH] UPBIT_SP 传入的币种与已有币种不符,传入cmcId:%d,tradeAsset:%s,已有币种:%s",
				cmcId, tradeAsset, mesh.CmcAsset)
		}
	} else {
		mesh = &CoinMesh{
			CmcId:          cmcId,
			UpbitSpKrwName: upbitSpotKrwName,
		}
		m.coinMap.Store(cmcId, mesh)
	}
}
