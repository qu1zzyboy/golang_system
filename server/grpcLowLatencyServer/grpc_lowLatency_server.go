package grpcLowLatencyServer

import (
	"context"
	"fmt"
	"time"

	strategyV1 "upbitBnServer/api/strategy/v1"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/market/symbolInfo"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpBitListBn"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitMesh"
	"upbitBnServer/internal/strategy/toUpbitParam"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/server/grpcEvent"
	"upbitBnServer/server/instance"
	"upbitBnServer/server/instance/instanceCenter"
	"upbitBnServer/server/serverInstanceEnum"

	"github.com/tidwall/gjson"
)

var (
	log      = dynamicLog.Log
	logError = dynamicLog.Error
)

type Server struct {
	strategyV1.UnimplementedStrategyServer
}

func (s *Server) StartStrategy(ctx context.Context, in *strategyV1.StrategyReq) (*strategyV1.CommonReplay, error) {
	grpcEventType := grpcEvent.GrpcEvent(in.CommonMeta.StrategyType)
	if grpcEventType == grpcEvent.CHECK_HEART_BEAT {
		return success("heart beat is ok", nil)
	}
	if !isAuthOk(ctx) {
		return failure(strategyV1.ErrorCode_AUTH_FAILED, "auth is not ok", nil)
	}
	log.GetLog().Debug("收到启动请求==>通用参数:", in.CommonMeta)
	log.GetLog().Debug("收到启动请求==>特有参数", in.JsonData)

	var err error
	switch grpcEventType {

	case grpcEvent.TO_UPBIT_LIST_BN:
		{
			req := toUpBitListBn.Req{}
			if err = jsonUtils.UnmarshalFromString(in.JsonData, &req); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			if err = req.Check(); err != nil {
				logError.GetLog().Error("请求参数校验失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			err = toUpBitListBn.Start(ctx, in.CommonMeta, &req)
		}

		// 上币静态信息更新
	case grpcEvent.SYMBOL_ON_LIST:
		{
			var staticSave symbolStatic.StaticSave
			var dynamic symbolDynamic.DynamicSymbol
			if err = jsonUtils.UnmarshalFromString(gjson.Get(in.JsonData, "data.static").String(), &staticSave); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			if err = jsonUtils.UnmarshalFromString(gjson.Get(in.JsonData, "data.dynamic").String(), &dynamic); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			var static symbolStatic.StaticTrade
			static.SymbolName = staticSave.SymbolName
			static.SymbolKeyId = staticSave.SymbolKeyId
			static.TradeId = staticSave.TradeId
			static.QuoteId = staticSave.QuoteId
			static.ExType = staticSave.ExType
			static.AcType = staticSave.AcType

			symbolStatic.GetTrade().Set(static)
			symbolStatic.GetSymbol().SetSymbol(static.SymbolName, symbolInfo.MakeSymbolId(static.TradeId, static.QuoteId))
			symbolStatic.GetSymbol().SetSymbolKey(static.SymbolKeyId, staticSave.SymbolKey)
			symbolDynamic.GetManager().SetDirect(static.SymbolKeyId, dynamic)
			symbolStatic.GetHandle().OnSymbolList(ctx, static)
			// 设置BN_FUTURE杠杆
			if static.ExType == exchangeEnum.BINANCE && static.AcType == exchangeEnum.FUTURE {
				err = bnOrderAppManager.GetTradeManager().SetBnLeverage(5, static.SymbolName)
			}
		}
	case grpcEvent.SYMBOL_DOWN_LIST:
		{
			symbolKeyId := gjson.Get(in.JsonData, "data").Uint()
			static, err := symbolStatic.GetTrade().Get(symbolKeyId)
			if err != nil {
				logError.GetLog().Error("删除交易对失败,获取静态信息失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			symbolStatic.GetTrade().Delete(symbolKeyId)
			symbolDynamic.GetManager().Delete(symbolKeyId)
			symbolStatic.GetHandle().OnSymbolDel(ctx, static)
		}
	case grpcEvent.SYMBOL_DYNAMIC_CHANGE:
		{

		}
	case grpcEvent.TO_UPBIT_ON_LIST:
		{
			var mesh coinMesh.CoinMesh
			if err = jsonUtils.UnmarshalFromString(gjson.Get(in.JsonData, "data").String(), &mesh); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			coinMesh.GetManager().Set(&mesh)
			toUpbitMesh.GetHandle().OnSymbolList(ctx, &mesh)
		}
	case grpcEvent.TO_UPBIT_DOWN_LIST:
		{
			var mesh coinMesh.CoinMesh
			if err = jsonUtils.UnmarshalFromString(gjson.Get(in.JsonData, "data").String(), &mesh); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			toUpbitMesh.GetHandle().OnSymbolDel(ctx, &mesh)
		}
	case grpcEvent.TO_UPBIT_RECEIVE_NEWS:
		{
			Asset := gjson.Get(in.JsonData, "events.0.symbols.0").String()
			symbolName := Asset + "USDT"
			symbolIndexTrue, ok := toUpBitListDataStatic.SymbolIndex.Load(symbolName)
			if !ok {
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, "TreeNews品种不在品种池内", nil)
			}

			// 触发品种和TreeNews品种一致
			if symbolIndexTrue == toUpBitListDataAfter.TrigSymbolIndex {
				toUpbitListBnSymbolArr.GetSymbolObj(symbolIndexTrue).ReceiveTreeNews()
			} else {
				toUpbitListBnSymbolArr.GetSymbolObj(toUpBitListDataAfter.TrigSymbolIndex).ReceiveNoTreeNews()
			}
		}
	case grpcEvent.TO_UPBIT_TEST:
		{
			symbolIndex, _ := toUpBitListDataStatic.SymbolIndex.Load("XPINUSDT")
			obj := toUpbitListBnSymbolArr.GetSymbolObj(symbolIndex)
			go obj.ReceiveTreeNews()
			obj.IntoExecuteNoCheck(time.Now().UnixMilli(), "test", 200000000)
		}
	case grpcEvent.TO_UPBIT_PARAM_TEST:
		{
			var req toUpbitParam.ComputeRequest
			if err = jsonUtils.UnmarshalFromString(in.JsonData, &req); err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			res, err := toUpbitParam.GetService().Compute(ctx, req)
			fmt.Println(err)
			res.PrintMe()
		}
	case grpcEvent.TO_UPBIT_CFG:
		{
			var cfg toUpBitListDataStatic.ConfigVir
			err = jsonUtils.UnmarshalFromString(in.JsonData, &cfg)
			if err != nil {
				logError.GetLog().Error("特有参数json解析失败:", err)
				return failure(strategyV1.ErrorCode_INVALID_ARGUMENT, err.Error(), nil)
			}
			toUpBitListDataStatic.UpdateParam(cfg.PriceRiceTrig)
		}
	default:
	}
	if err != nil {
		logError.GetLog().Error("策略启动失败:", err)
		return failure(strategyV1.ErrorCode_INTERNAL_ERROR, err.Error(), nil)
	}
	return success("start success", 123)
}

func (s *Server) UpdateStrategy(ctx context.Context, in *strategyV1.StrategyReq) (*strategyV1.CommonReplay, error) {
	if !isAuthOk(ctx) {
		return failure(strategyV1.ErrorCode_AUTH_FAILED, "auth is not ok", nil)
	}
	log.GetLog().Debug("收到更新请求==>通用参数:", in.CommonMeta)
	log.GetLog().Debug("收到更新请求==>特有参数", in.JsonData)
	instanceId := in.CommonMeta.InstanceId
	if instanceId == 0 {
		logError.GetLog().Error("UpdateStrategy instanceId is null")
		return failure(strategyV1.ErrorCode_INSTANCE_ID_EMPTY, "instanceId is empty", nil)
	}
	if err := instanceCenter.GetManager().UpdateInstance(ctx, serverInstanceEnum.Type(instanceId), instance.InstanceUpdate{JsonData: in.JsonData}); err != nil {
		return failure(strategyV1.ErrorCode_INTERNAL_ERROR, err.Error(), instanceId)
	}
	return success("update success", "")
}

func (s *Server) StopStrategy(ctx context.Context, in *strategyV1.StrategyReq) (*strategyV1.CommonReplay, error) {
	if !isAuthOk(ctx) {
		return failure(strategyV1.ErrorCode_AUTH_FAILED, "auth is not ok", nil)
	}
	log.GetLog().Debug("收到停止请求==>通用参数:", in.CommonMeta)
	if in.CommonMeta.InstanceId == 0 {
		logError.GetLog().Error("StopStrategy instanceId is null")
		return failure(strategyV1.ErrorCode_INSTANCE_ID_EMPTY, "instanceId is empty", nil)
	}
	if err := instanceCenter.GetManager().StopInstance(ctx, serverInstanceEnum.Type(in.CommonMeta.InstanceId)); err != nil {
		return failure(strategyV1.ErrorCode_INTERNAL_ERROR, err.Error(), in.CommonMeta.InstanceId)
	}
	return success("stop success", "")
}
