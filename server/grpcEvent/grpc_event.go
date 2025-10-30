package grpcEvent

type GrpcEvent uint32

const (
	SYMBOL_ON_LIST        GrpcEvent = iota //上币
	SYMBOL_DOWN_LIST                       //下币,"72339069019489081"
	SYMBOL_DYNAMIC_CHANGE                  //动态交易规范更新
	PRINT_ALL_INSTANCE                     //打印所有实例
	CHECK_HEART_BEAT                       //检查心跳

	TO_UPBIT_ON_LIST      //UpBit待上市
	TO_UPBIT_DOWN_LIST    //UpBit待下市
	TO_UPBIT_LIST_BN      //UpBit上市之bn
	TO_UPBIT_LIST_BYBIT   //UpBit上市之ByBit
	TO_UPBIT_RECEIVE_NEWS //
	TO_UPBIT_CFG          // upbit配置更新
	TO_UPBIT_TEST
	TO_UPBIT_PARAM_TEST
	GET_NOT_REASON //

	//观测服务
	IMPORTANT_ERROR
	NORMAL_ERROR
	REMINDER_MSG
)

type GrpcSend struct {
	Event GrpcEvent `json:"event"`
	Data  any       `json:"data"`
}
