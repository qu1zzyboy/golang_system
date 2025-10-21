package toUpbitListBnSymbol

import (
	"context"
	"fmt"
	"log"

	params "github.com/hhh500/upbitBnServer/api/toUpbit/v1"
	"github.com/hhh500/upbitBnServer/server/grpcAuth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetParam(isMeme bool, cap float64, symbolName string) (gainPct, twapSec float64, err error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithPerRPCCredentials(&grpcAuth.ClientTokenAuth{}))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", "127.0.0.1", "50051"), opts...)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close client connection: %v", err)
		}
	}(conn)
	var client params.ParamsServiceClient = params.NewParamsServiceClient(conn)
	resp, err := client.GetParams(context.Background(), &params.ParamsRequest{
		MarketCapM: cap,
		IsMeme:     isMeme,
		SymbolName: symbolName,
	})
	if err != nil {
		return 0.0, 0.0, err
	}
	for _, f := range resp.Data {
		if f.Key == "gain_pct" {
			gainPct = f.GetNumberValue()
		}
		if f.Key == "twap_sec" {
			twapSec = f.GetNumberValue()
		}
	}
	return gainPct, twapSec, nil
}
