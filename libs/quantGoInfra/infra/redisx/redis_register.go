package redisx

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/infra/observe/log/staticLog"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
)

const (
	poolSize    = 160 //连接池大小
	minIdleConn = 32  //最小空闲连接数
)

var (
	rdbMap            = myMap.NewMySyncMap[string, *redis.Client]()
	luaMap            = myMap.NewMySyncMap[string, *LuaScript]()
	luaScriptNotFound = errorx.New(errCode.REDIS_LUA_SCRIPT_NOT_FOUND, "Redis Lua脚本未找到")
	clientNotFound    = errorx.New(errCode.REDIS_CLIENT_NOT_FOUND, "Redis客户端未找到")
	evalLoadError     = errorx.New(errCode.REDIS_LUA_LOAD_ERROR, "Redis Lua脚本加载错误")
	DoError           = errorx.New(errCode.REDIS_DO_ERROR, "Redis操作错误")
)

func RegisterClient(ctx context.Context, host, pass, dbTarget string, dbIndex int) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     pass,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConn,
		DB:           dbIndex,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return err
	}
	staticLog.Log.Infof("Redis注册成功,host=%s,dbIndex=%d", host, dbIndex)
	rdbMap.Store(dbTarget, rdb)
	return nil
}

func LoadClient(targetKey string) (*redis.Client, error) {
	if conn, ok := rdbMap.Load(targetKey); ok {
		return conn, nil
	}
	return nil, clientNotFound.WithMetadata(map[string]string{"targetKey": targetKey})
}

func RegisterLuaScript(ctx context.Context, client *redis.Client, luaKey, script string) error {
	sha, err := client.ScriptLoad(ctx, script).Result()
	if err != nil {
		return evalLoadError.WithCause(err).WithMetadata(map[string]string{defineJson.RawJson: script, "luaKey": luaKey})
	}
	luaMap.Store(luaKey, &LuaScript{
		sha:    sha,
		script: script,
		client: client,
	})
	return nil
}

func LoadLuaScript(targetKey string) (*LuaScript, error) {
	if conn, ok := luaMap.Load(targetKey); ok {
		return conn, nil
	}
	return nil, luaScriptNotFound.WithMetadata(map[string]string{"targetKey": targetKey})
}
