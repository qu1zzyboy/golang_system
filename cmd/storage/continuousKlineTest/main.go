package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/bootx"
	"upbitBnServer/internal/infra/observe/notify"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/strategy/continuousKlineTest"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// 创建一个空的 Notify 实现用于测试
type emptyNotify struct{}

func (e *emptyNotify) SendImportantErrorMsg(payload map[string]string) error { return nil }
func (e *emptyNotify) SendNormalErrorMsg(payload map[string]string) error    { return nil }
func (e *emptyNotify) SendReminderMsg(payload map[string]string) error       { return nil }

func main() {
	fmt.Println("=== 连续 K 线测试服务 ===")

	// 从 .env 文件加载环境变量（如果文件存在）
	// .env 文件应该在项目根目录或当前工作目录
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在时忽略错误（允许使用系统环境变量）
		fmt.Println("未找到 .env 文件，使用系统环境变量")
	}

	// 设置环境变量跳过配置文件（测试程序不需要完整配置）
	os.Setenv("USE_OS_CONFIG", "1")

	// 设置必要的默认配置值（避免配置读取失败）
	viper.SetDefault("cpuHz", "2400000000") // 默认 CPU 频率 2.4GHz
	viper.SetDefault("serverName", "continuous-kline-test")
	viper.SetDefault("msgAble", "false")
	viper.SetDefault("isTestDev", "true")

	// Redis 配置：优先使用环境变量，否则使用默认值
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6379"
	}
	redisPass := os.Getenv("REDIS_PASS")
	viper.SetDefault("redisConfig.hosts", redisHost)
	viper.SetDefault("redisConfig.pass", redisPass)

	// 注册 bootx 服务
	manager := bootx.GetManager()
	manager.Register(conf.NewBoot())                 // 配置加载
	manager.Register(redisConfig.NewBoot())          // Redis 配置
	manager.Register(notify.NewBoot(&emptyNotify{})) // 通知服务（空实现）
	manager.Register(safex.NewBoot())                // 协程安全工具
	manager.Register(continuousKlineTest.NewBoot())  // 连续 K 线测试服务

	// 启动所有服务
	ctx := context.Background()
	fmt.Println("正在启动服务...")
	manager.StartAll(ctx) // StartAll 内部使用 panic 处理错误，无需检查返回值

	fmt.Println("服务启动成功！")
	fmt.Println("正在订阅全市场秒级别连续 K 线...")
	fmt.Println("按 Ctrl+C 退出")
	fmt.Println("----------------------------------------")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务...")
	manager.StopAll(ctx)
	fmt.Println("服务已关闭")
}
