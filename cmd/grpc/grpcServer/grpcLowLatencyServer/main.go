package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	strategyV1 "upbitBnServer/api/strategy/v1"
	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/bootx"
	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/infra/global/globalTask"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/metric/metricIndex"
	"upbitBnServer/internal/infra/observe/notify"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/account/bnPayloadManager"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolInfoLoad"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitParam"
	"upbitBnServer/internal/strategy/treenews"
	"upbitBnServer/pkg/container/pool/antPool"
	"upbitBnServer/server/grpcLowLatencyServer"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	reg := prometheus.NewRegistry()
	metricIndex.RegisterWSMetrics(reg)
	metricIndex.RegisterRuntimeMetricsWith(reg)

	go func() {
		// 启动 pprof 服务
		http.ListenAndServe("localhost:6060", nil)
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Printf("Prometheus metrics server listening at :2112")
		log.Fatal(http.ListenAndServe(":2112", mux))
	}()
	//配置文件相关
	bootx.GetManager().Register(conf.NewBoot())

	bootx.GetManager().Register(safex.NewBoot())
	bootx.GetManager().Register(antPool.NewBoot(18, 48))
	bootx.GetManager().Register(redisConfig.NewBoot())

	// 账户相关
	bootx.GetManager().Register(accountConfig.NewBoot())
	bootx.GetManager().Register(bnPayloadManager.NewBoot())
	bootx.GetManager().Register(bnOrderAppManager.NewBoot())

	//定时任务相关
	bootx.GetManager().Register(globalCron.NewBoot())
	bootx.GetManager().Register(globalTask.NewBoot())
	bootx.GetManager().Register(dynamicLog.NewBoot())
	bootx.GetManager().Register(symbolInfoLoad.NewBoot())
	//观测相关
	bootx.GetManager().Register(notify.NewBoot(notifyTg.GetTg()))

	//启动服务
	bootx.GetManager().Register(toUpbitBnMode.NewBoot(toUpbitBnMode.LiveMode{}))
	bootx.GetManager().Register(treenews.NewBoot())
	bootx.GetManager().Register(toUpbitParam.NewBoot())
	bootx.GetManager().Register(grpcLowLatencyServer.NewBoot())
	bootx.GetManager().StartAll(context.Background())
	// 开启端口监听
	listen, err := net.Listen("tcp", ":"+conf.GrpcCfg.LowLatencyPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 创建grpc服务
	grpcServer := grpc.NewServer()
	// 注册执行服务
	strategyV1.RegisterStrategyServer(grpcServer, &grpcLowLatencyServer.Server{})
	log.Printf("server listening at %v", listen.Addr())
	// 启动服务
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
