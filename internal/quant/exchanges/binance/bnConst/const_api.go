package bnConst

const (
	FUTURE_BASE_REST_URL = "https://fapi.binance.com"
	SPOT_BASE_REST_URL   = "https://api.binance.com"
)

const (
	LEVEL_DEPTH_5            = "5"
	LEVEL_DEPTH_10           = "10"
	LEVEL_DEPTH_20           = "20"
	LEVEL_DEPTH_USPEED_100ms = "@100ms"
	LEVEL_DEPTH_USPEED_250ms = ""
	LEVEL_DEPTH_USPEED_500ms = "@500ms"
)

var (
	DefaultLevel  = LEVEL_DEPTH_5
	DefaultUSpeed = LEVEL_DEPTH_USPEED_100ms
)

const (
	FROM_WS        = "ws"
	FROM_WS_STREAM = "ws_stream"
)
