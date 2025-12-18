package conf

const (
	MODULE_ID = "conf"
)

var (
	ServerName  string              //服务器名称
	ServerIpIn  string              //服务器内网ip
	ServerIpOut string              //服务器外网ip
	CPU_HZ      uint64 = 2900000000 // CPU频率
	RedisCfg    RedisConfig
	GrpcCfg     GrpcConfig
	ObserveCfg  ObserverConfig
	TreeNewsCfg TreeNewsConfig
	MsgAble     bool
	IsTestDev   bool
)
