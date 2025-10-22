package conf

const (
	LogDirBase        = "logs/"
	MODULE_ID         = "conf"
	USDT       uint16 = 825
	USDC       uint16 = 3408
)

var (
	ServerName   string              //服务器名称
	ServerIpIn   string              //服务器内网ip
	ServerIpOut  string              //服务器外网ip
	DES_KEY      string              //DES加密key
	ToUpBitPath  string              //上upbit配置文件
	DataSavePath string              //数据保存路径
	CPU_HZ       uint64 = 2900000000 // CPU频率
	RedisCfg     RedisConfig
	GrpcCfg      GrpcConfig
	ObserveCfg   ObserverConfig
	TreeNewsCfg  TreeNewsConfig
	MsgAble      bool
)
