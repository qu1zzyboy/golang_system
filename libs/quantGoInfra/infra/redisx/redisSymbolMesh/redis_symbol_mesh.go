package redisSymbolMesh

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hhh500/quantGoInfra/infra/redisx"
	"github.com/hhh500/quantGoInfra/infra/redisx/redisConfig"
)

func GetAllSymbolMesh(hashKey string) (map[string]uint32, error) {
	redisClient, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return nil, err
	}
	res := redisClient.HGetAll(context.Background(), hashKey)
	if res.Err() != nil {
		return nil, res.Err()
	}
	data, err := res.Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]uint32)
	for k, v := range data {
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("解析币种映射值错误,key:%s,value:%s,err:%w", k, v, err)
		}
		result[k] = uint32(val)
	}
	return result, nil
}
