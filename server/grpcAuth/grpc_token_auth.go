package grpcAuth

import (
	"context"

	"upbitBnServer/internal/conf"
)

type ClientTokenAuth struct {
}

func (c *ClientTokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"app_id":  conf.GrpcCfg.AppId,
		"app_key": conf.GrpcCfg.AppKey,
	}, nil
}

func (c *ClientTokenAuth) RequireTransportSecurity() bool {
	return false
}
