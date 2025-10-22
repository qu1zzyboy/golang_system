package redisx

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

var (
	EvalError = errorx.New(errCode.REDIS_LUA_EVAL_ERROR, "Redis Lua脚本执行错误")
)

// LuaScript 封装 Redis Lua 脚本的执行逻辑
// 支持使用 EvalSha 提高性能,并在脚本未缓存 (NOSCRIPT) 时自动恢复
type LuaScript struct {
	script string        // 原始 Lua 脚本内容
	sha    string        // Redis 中缓存的脚本 SHA1
	client *redis.Client // Redis 客户端实例
}

// Eval 尝试通过 EvalSha 执行 Lua 脚本,若 Redis 重启导致缓存失效则自动重新加载
func (l *LuaScript) Eval(ctx context.Context, keys []string, args ...interface{}) (interface{}, error) {
	// 尝试使用 EvalSha(性能更好)
	result, err := l.client.EvalSha(ctx, l.sha, keys, args...).Result()
	if err == nil {
		return result, nil
	}

	// 如果是 NOSCRIPT 错误，则重新加载脚本并执行
	if strings.Contains(err.Error(), "NOSCRIPT") {
		sha, err2 := l.client.ScriptLoad(ctx, l.script).Result()
		if err2 != nil {
			return nil, evalLoadError.WithCause(err2).WithMetadata(map[string]string{defineJson.RawJson: l.script})
		}
		l.sha = sha
		return l.client.EvalSha(ctx, sha, keys, args...).Result()
	}
	// 其他错误直接返回
	return nil, EvalError.WithCause(err).WithMetadata(map[string]string{defineJson.RawJson: l.script})
}
