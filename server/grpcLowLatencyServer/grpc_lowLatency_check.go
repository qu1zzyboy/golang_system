package grpcLowLatencyServer

import (
	"context"

	"github.com/hhh500/quantGoInfra/conf"
	"github.com/hhh500/quantGoInfra/pkg/utils/jsonUtils"
	strategyV1 "github.com/hhh500/upbitBnServer/api/strategy/v1"
	"google.golang.org/grpc/metadata"
)

func success(msg string, data interface{}) (*strategyV1.CommonReplay, error) {
	result, err := jsonUtils.MarshalStructToString(data)
	if err != nil {
		logError.GetLog().Error(err)
		return nil, err
	}
	m := strategyV1.CommonReplay{
		Code:     strategyV1.ErrorCode_OK,
		Msg:      msg,
		JsonData: result,
	}
	return &m, err
}

func failure(code strategyV1.ErrorCode, msg string, data interface{}) (*strategyV1.CommonReplay, error) {
	result, err := jsonUtils.MarshalStructToString(data)
	if err != nil {
		logError.GetLog().Error(err)
		return nil, err
	}
	m := strategyV1.CommonReplay{
		Code:     code,
		Msg:      msg,
		JsonData: result,
	}
	return &m, err
}

func isAuthOk(ctx context.Context) bool {
	metaDate, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logError.GetLog().Error("元数据获取失败")
		return false
	}
	log.GetLog().Debug("metaDate:", metaDate)
	var appId, appKey string
	if v, ok := metaDate["app_id"]; ok {
		appId = v[0]
	}
	if v, ok := metaDate["app_key"]; ok {
		appKey = v[0]
	}
	if appId == conf.GrpcCfg.AppId && appKey == conf.GrpcCfg.AppKey {
		return true
	}
	logError.GetLog().Error("appId或appKey不匹配,appId:", appId, "appKey:", appKey)
	return false
}
