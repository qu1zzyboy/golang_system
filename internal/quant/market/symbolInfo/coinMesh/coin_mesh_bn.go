package coinMesh

import "github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"

// 要求先获取cmc信息
func (m *Manager) SetBnFuUsdtNameCheck(cmcId uint32, tradeAsset, bnFutureUsdtName string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		mesh.BnFuUsdtName = bnFutureUsdtName
		if mesh.CmcAsset == tradeAsset {
			mesh.BnFuUsdtName = bnFutureUsdtName
		} else {
			if mesh.CmcAsset != "" {
				dynamicLog.Error.GetLog().Errorf("[COIN_MESH] BN_FU_USDT 传入的币种与已有币种不符,传入cmcId:%d,tradeAsset:%s,已有币种:%s", cmcId, tradeAsset, mesh.CmcAsset)
			}
		}
	} else {
		mesh = &CoinMesh{
			CmcId:        cmcId,
			BnFuUsdtName: bnFutureUsdtName,
		}
		m.coinMap.Store(cmcId, mesh)
	}
}

func (m *Manager) SetBnFuUsdcNameCheck(cmcId uint32, tradeAsset, bnFutureUsdcName string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		if mesh.CmcAsset == tradeAsset {
			mesh.BnFuUsdcName = bnFutureUsdcName
		} else {
			if mesh.CmcAsset != "" {
				dynamicLog.Error.GetLog().Errorf("[COIN_MESH] BN_FU_USDC 传入的币种与已有币种不符,传入cmcId:%d,tradeAsset:%s,已有币种:%s", cmcId, tradeAsset, mesh.CmcAsset)
			}
		}
	} else {
		mesh = &CoinMesh{
			CmcId:        cmcId,
			BnFuUsdcName: bnFutureUsdcName,
		}
		m.coinMap.Store(cmcId, mesh)
	}
}

func (m *Manager) SetBnSpotUsdtName(cmcId uint32, bnSpotUsdtName string) {
	if mesh, ok := m.coinMap.Load(cmcId); ok {
		mesh.BnSpUsdtName = bnSpotUsdtName
	} else {
		mesh = &CoinMesh{
			CmcId:        cmcId,
			BnSpUsdtName: bnSpotUsdtName,
		}
		m.coinMap.Store(cmcId, mesh)
	}
}

func (m *Manager) GetAllBnFuUsdtName() (list []string) {
	m.coinMap.Range(func(_ uint32, value *CoinMesh) bool {
		if value.BnFuUsdtName != "" {
			list = append(list, value.BnFuUsdtName)
		}
		return true
	})
	return list
}
