package symbolInfoLoad

import (
	"context"
	"strconv"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolLimit"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"

	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/go-redis/redis/v8"
)

func loadStatic(ctx context.Context, redisClient *redis.Client) error {
	res := redisClient.HGetAll(ctx, symbolStatic.SYMBOL_INFO_STATIC_MANAGER)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	for _, v := range data {
		var static symbolStatic.StaticSave
		if err := jsonUtils.UnmarshalFromString(v, &static); err != nil {
			return err
		}
		symbolStatic.GetTrade().Set(symbolStatic.StaticTrade{
			SymbolName:  static.SymbolName,
			SymbolKeyId: static.SymbolKeyId,
			TradeId:     static.TradeId,
			QuoteId:     static.QuoteId,
			ExType:      static.ExType,
			AcType:      static.AcType,
		})
		symbolStatic.GetSymbol().SetSymbol(static.SymbolName, symbolInfo.MakeSymbolId(static.TradeId, static.QuoteId))
		symbolStatic.GetSymbol().SetSymbolKey(static.SymbolKeyId, static.SymbolKey)
	}
	return nil
}

func loadDynamic(ctx context.Context, redisClient *redis.Client) error {
	res := redisClient.HGetAll(ctx, symbolDynamic.SYMBOL_INFO_DYNAMIC_MANAGER)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	for symbolKeyIdStr, v := range data {
		var dynamic symbolDynamic.DynamicSymbol
		if err := jsonUtils.UnmarshalFromString(v, &dynamic); err != nil {
			return err
		}
		symbolKeyId, err := strconv.ParseUint(symbolKeyIdStr, 10, 64)
		if err != nil {
			return err
		}
		symbolDynamic.GetManager().SetDirect(symbolKeyId, dynamic)
	}
	return nil
}

func loadLimit(ctx context.Context, redisClient *redis.Client) error {
	res := redisClient.HGetAll(ctx, symbolLimit.SYMBOL_INFO_LIMIT_MANAGER)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	for symbolKeyIdStr, v := range data {
		var limit symbolLimit.LimitSymbol
		if err = jsonUtils.UnmarshalFromString(v, &limit); err != nil {
			return err
		}
		symbolKeyId, err := strconv.ParseUint(symbolKeyIdStr, 10, 64)
		if err != nil {
			return err
		}
		symbolLimit.GetManager().SetDirect(symbolKeyId, limit)
	}
	return nil
}

func loadCoinMesh(ctx context.Context, redisClient *redis.Client) error {
	res := redisClient.HGetAll(ctx, coinMesh.REDIS_KEY_COIN_MESH_ALL)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	for _, v := range data {
		var mesh coinMesh.CoinMesh
		if err := jsonUtils.UnmarshalFromString(v, &mesh); err != nil {
			return err
		}
		coinMesh.GetManager().Set(&mesh)
	}
	return nil
}

func loadUpBitCfg(ctx context.Context, redisClient *redis.Client) error {
	res := redisClient.Get(ctx, driverStatic.TO_UPBIT_LIST_CFG)
	if res.Err() != nil {
		return res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return err
	}
	if err := jsonUtils.UnmarshalFromString(data, &driverStatic.GlobalCfg); err != nil {
		return err
	}
	return nil
}
