package treenews

import (
	"os"
	"strconv"
	"time"
)

const defaultAPIKey = "03610598fc45358259ba8c8ebe1e858709ec9a227d38bb87cc66b7c459474985"

// Config aggregates runtime knobs for the Tree News websocket client.
type Config struct {
	Enabled          bool
	URL              string
	APIKey           string
	Workers          int
	PingInterval     time.Duration
	PingTimeout      time.Duration
	RollingReconnect time.Duration
	RollingJitter    time.Duration
	DedupCapacity    int
}

func defaultConfig() Config {
	cfg := Config{
		Enabled:          false,
		URL:              "wss://news.treeofalpha.com/ws",
		APIKey:           defaultAPIKey,
		Workers:          2,
		PingInterval:     15 * time.Second,
		PingTimeout:      2 * time.Second,
		RollingReconnect: 6 * time.Hour,
		RollingJitter:    10 * time.Minute,
		DedupCapacity:    50000,
	}

	if v := os.Getenv("TREE_NEWS_ENABLED"); v != "" {
		cfg.Enabled = v == "1" || v == "true" || v == "TRUE"
	}
	if v := os.Getenv("TREE_NEWS_URL"); v != "" {
		cfg.URL = v
	}
	if v := os.Getenv("TREE_NEWS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("TREE_NEWS_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Workers = n
		}
	}
	if v := os.Getenv("TREE_NEWS_PING_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PingInterval = d
		}
	}
	if v := os.Getenv("TREE_NEWS_PING_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PingTimeout = d
		}
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

	return cfg
}
