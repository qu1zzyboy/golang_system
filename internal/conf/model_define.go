package conf

import "time"

type RedisConfig struct {
	Hosts string `yaml:"hosts"`
	Pass  string `yaml:"pass"`
}

type GrpcConfig struct {
	LowLatencyPort string `yaml:"lowLatencyPort"`
	AppId          string `yaml:"appId"`
	AppKey         string `yaml:"appKey"`
}

type ObserverConfig struct {
	Host string `yaml:"port"`
	Port string `yaml:"host"`
}

type SymbolMesh struct {
	PathBnFu       string `yaml:"pathBnFu"`
	PathBnToCmc    string `yaml:"pathBnToCmc"`
	PathByBitToCmc string `yaml:"pathByBitToCmc"`
	PathUpBitToCmc string `yaml:"pathUpBitToCmc"`
	PathBybitFu    string `yaml:"pathBybitFu"`
	PathUpbitSp    string `yaml:"pathUpbitSp"`
}

type TreeNewsConfig struct {
	Enabled          bool
	APIKey           string
	URL              string
	Workers          int
	PingInterval     time.Duration
	PingTimeout      time.Duration
	RollingReconnect time.Duration
	RollingJitter    time.Duration
	DedupCapacity    int
	QueueCapacity    int
	LatencyWarnMS    int
	LatencyWarnCount int
	RTTWarnMS        int
	RTTWarnCount     int
}
