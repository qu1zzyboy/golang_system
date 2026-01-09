package main

import (
	"context"
	"fmt"
	"log"
	strategyV1 "upbitBnServer/api/strategy/v1"
	"upbitBnServer/internal/conf"
	"upbitBnServer/internal/infra/bootx"
	"upbitBnServer/server/grpcAuth"
	"upbitBnServer/server/grpcEvent"
	"upbitBnServer/server/serverInstanceEnum"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startGrpcClient(client strategyV1.StrategyClient) {
	resp, err := client.StartStrategy(context.Background(), &strategyV1.StrategyReq{
		CommonMeta: &strategyV1.ServerReqBase{
			RequestIp:    conf.ServerIpIn,
			InstanceId:   uint32(serverInstanceEnum.TO_UPBIT_LIST_BN),
			StrategyType: uint32(grpcEvent.TO_UPBIT_TEST),
			StrategyName: "bn上币upbit",
		},
		JsonData: "",
	})
	if err != nil {
		log.Fatalf("failed to start strategyClient: %v", err)
	}
	log.Printf("start strategyClient response: %v", resp)
}

func main() {
	//配置文件相关
	bootx.GetManager().Register(conf.NewBoot())
	bootx.GetManager().StartAll(context.Background())
	// 创建grpc客户端,获取连接
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithPerRPCCredentials(&grpcAuth.ClientTokenAuth{}))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", "127.0.0.1", conf.GrpcCfg.LowLatencyPort), opts...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close client connection: %v", err)
		}
	}(conn)
	// 建立连接
	startGrpcClient(strategyV1.NewStrategyClient(conn))
}
