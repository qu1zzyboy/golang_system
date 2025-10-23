package conf

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/observe/log/staticLog"
	"github.com/spf13/viper"
)

var (
	log        = staticLog.Log // 静态日志
	configFile string
)

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	initConfig()
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}

func initConfig() {
	// flag.StringVar 用于定义一个命令行标志(flag)并将其解析为一个字符串变量
	// &configFile: 变量的指针：用于存储命令行解析后得到的值
	// "config": 这是命令行标志的名称,使用 -config 来指定配置文件的路径.
	// ""：默认值：如果命令行没有传 -config，则 configFile 默认是空字符串 ""
	// "configFile"： 帮助信息：当你运行 ./main -h 时会显示这段说明文字
	flag.StringVar(&configFile, "config", "config_main.yaml", "指定配置文件解析")
	flag.Parse() // 解析命令行参数
	useOs := os.Getenv("USE_OS_CONFIG")
	if useOs == "" {
		readConfig() // 读取配置文件
	}

	ServerName = getConfig("serverName")
	ServerIpIn = getConfig("serverIpIn")
	ServerIpOut = getConfig("serverIpOut")
	DES_KEY = getConfig("desKey")
	DataSavePath = getConfig("dataSavePath")
	ToUpBitPath = getConfig("toUpbitPath")

	var err error
	CPU_HZ, err = strconv.ParseUint(getConfig("cpuHz"), 10, 64)
	if err != nil {
		panic(err)
	}
	MsgAble = getConfig("msgAble") == "true"
	log.Info("服务器名称:", ServerName)
	log.Info("服务器内网ip:", ServerIpIn)
	log.Info("服务器外网ip:", ServerIpOut)
	log.Info("cpu频率:", CPU_HZ)
	log.Infof("消息服务开关:%v", MsgAble)

	RedisCfg.Hosts = getConfig("redisConfig.hosts")
	RedisCfg.Pass = getConfig("redisConfig.pass")

	GrpcCfg.ObservePort = getConfig("grpc.notifyPort")
	GrpcCfg.DownLoadPort = getConfig("grpc.dynamicPort")
	GrpcCfg.StrategyPort = getConfig("grpc.strategyPort")
	GrpcCfg.CrossPort = getConfig("grpc.crossPort")
	GrpcCfg.ExecutePort = getConfig("grpc.executePort")
	GrpcCfg.LowLatencyPort = getConfig("grpc.lowLatencyPort")
	GrpcCfg.AppId = getConfig("grpc.appId")
	GrpcCfg.AppKey = getConfig("grpc.appKey")
	log.Info("grpc配置:观测服务端口:", GrpcCfg.ObservePort)
	log.Info("grpc配置:动态下载端口:", GrpcCfg.DownLoadPort)
	log.Info("grpc配置:策略服务端口:", GrpcCfg.StrategyPort)
	log.Info("grpc配置:截面服务端口:", GrpcCfg.CrossPort)
	log.Info("grpc配置:执行服务端口:", GrpcCfg.ExecutePort)
	log.Info("grpc配置:低延时服务端口:", GrpcCfg.LowLatencyPort)
	log.Info("grpc配置:appId:", GrpcCfg.AppId)
	log.Info("grpc配置:appKey:", GrpcCfg.AppKey)

	ObserveCfg.Host = getConfig("observe.host")
	ObserveCfg.Port = getConfig("observe.port")

	TreeNewsCfg.Enabled = viper.GetBool("treeNews.enabled")
	TreeNewsCfg.APIKey = getConfig("treeNews.apiKey")
	TreeNewsCfg.URL = getConfig("treeNews.url")
	TreeNewsCfg.Workers = viper.GetInt("treeNews.workers")
	TreeNewsCfg.PingInterval = parseDuration(getConfig("treeNews.pingInterval"), 15*time.Second)
	TreeNewsCfg.PingTimeout = parseDuration(getConfig("treeNews.pingTimeout"), 2*time.Second)
	TreeNewsCfg.RollingReconnect = parseDuration(getConfig("treeNews.rollingReconnect"), 6*time.Hour)
	TreeNewsCfg.RollingJitter = parseDuration(getConfig("treeNews.rollingJitter"), 10*time.Minute)
	TreeNewsCfg.DedupCapacity = viper.GetInt("treeNews.dedupCapacity")
	TreeNewsCfg.QueueCapacity = viper.GetInt("treeNews.queueCapacity")
	TreeNewsCfg.LatencyWarnMS = viper.GetInt("treeNews.latencyWarnMs")
	TreeNewsCfg.LatencyWarnCount = viper.GetInt("treeNews.latencyWarnCount")
	TreeNewsCfg.RTTWarnMS = viper.GetInt("treeNews.rttWarnMs")
	TreeNewsCfg.RTTWarnCount = viper.GetInt("treeNews.rttWarnCount")

	log.Info("===============结束读取配置信息==================")
}

func parseDuration(value string, def time.Duration) time.Duration {
	if value == "" {
		return def
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		log.Warnf("parse duration [%s] failed: %v, use default %s", value, err, def)
		return def
	}
	return d
}

func getConfig(name string) string {
	return viper.GetString(name)
}

func readConfig() bool {
	box := packr.New("test", "./config") //使用packr创建一个新的盒子box,用于从指定路径(./config)中加载配置文件。这允许你在构建时将文件打包到可执行文件中。
	configType := "yaml"                 //配置文件类型为yaml
	v := viper.New()                     //创建一个新的Viper实例,
	v.SetConfigType(configType)          //并设置配置类型为 yaml
	env := configFile                    // 根据配置的env读取相应的配置信息
	log.Infof("读取配置文件: %s", env)
	if env != "" {
		envConfig, err := box.Find(env)
		if err != nil {
			errorx.PanicWithCaller(fmt.Sprintf("Fatal error config: %s", err.Error()))
		}
		viper.SetConfigType(configType)
		err = viper.ReadConfig(bytes.NewReader(envConfig))
		if err != nil {
			errorx.PanicWithCaller(fmt.Sprintf("Fatal error config: %s", err.Error()))
		}
	}
	return true
}
