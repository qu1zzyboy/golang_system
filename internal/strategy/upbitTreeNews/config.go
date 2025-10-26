package treenews

import (
	"os"
	"strconv"
	"time"

	"upbitBnServer/internal/conf"
)

const url = "wss://news.treeofalpha.com/ws"

// {
//   "time": 1761390911451,
//   "user": {
//     "id": "1100035037564510218",
//     "username": "joshqing",
//     "highestRole": 0,
//     "isSub": false,
//     "addons": [],
//     "isDiscordMember": true
//   }
// }

// Config 汇总 Tree News WebSocket 客户端相关的运行时参数。
type Config struct {
	APIKey           string
	RollingReconnect time.Duration
	RollingJitter    time.Duration
	DedupCapacity    int
	LatencyWarnMS    int
	LatencyWarnCount int
	RTTWarnMS        int
	RTTWarnCount     int
}

func defaultConfig() Config {
	cfg := Config{
		APIKey:           conf.TreeNewsCfg.APIKey,
		RollingReconnect: conf.TreeNewsCfg.RollingReconnect,
		RollingJitter:    conf.TreeNewsCfg.RollingJitter,
		DedupCapacity:    conf.TreeNewsCfg.DedupCapacity,
		LatencyWarnMS:    conf.TreeNewsCfg.LatencyWarnMS,
		LatencyWarnCount: conf.TreeNewsCfg.LatencyWarnCount,
		RTTWarnMS:        conf.TreeNewsCfg.RTTWarnMS,
		RTTWarnCount:     conf.TreeNewsCfg.RTTWarnCount,
	}

	if cfg.APIKey == "" {
		cfg.APIKey = "03610598fc45358259ba8c8ebe1e858709ec9a227d38bb87cc66b7c459474985"
	}
	if cfg.RollingReconnect <= 0 {
		cfg.RollingReconnect = 6 * time.Hour
	}
	if cfg.RollingJitter < 0 {
		cfg.RollingJitter = 10 * time.Minute
	}
	if cfg.DedupCapacity <= 0 {
		cfg.DedupCapacity = 50000
	}
	if cfg.LatencyWarnMS <= 0 {
		cfg.LatencyWarnMS = 500
	}
	if cfg.LatencyWarnCount <= 0 {
		cfg.LatencyWarnCount = 3
	}
	if cfg.RTTWarnMS <= 0 {
		cfg.RTTWarnMS = 400
	}
	if cfg.RTTWarnCount <= 0 {
		cfg.RTTWarnCount = 3
	}

	if v := os.Getenv("TREE_NEWS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("TREE_NEWS_ROLLING_RECONNECT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RollingReconnect = d
		}
	}
	if v := os.Getenv("TREE_NEWS_ROLLING_JITTER"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RollingJitter = d
		}
	}
	if v := os.Getenv("TREE_NEWS_DEDUP_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.DedupCapacity = n
		}
	}
	if v := os.Getenv("TREE_NEWS_LATENCY_WARN_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.LatencyWarnMS = n
		}
	}
	if v := os.Getenv("TREE_NEWS_LATENCY_WARN_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.LatencyWarnCount = n
		}
	}
	if v := os.Getenv("TREE_NEWS_RTT_WARN_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RTTWarnMS = n
		}
	}
	if v := os.Getenv("TREE_NEWS_RTT_WARN_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RTTWarnCount = n
		}
	}

	return cfg
}
