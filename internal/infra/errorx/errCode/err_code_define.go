package errCode

//按类型或层级划分

const (
	CODE_SUCCESS           uint16 = 1000 + iota // 成功
	CODE_UN_KNOWN                               // 未知错误
	ENUM_NOT_SUPPORT                            // 枚举不支持错误
	MEMORY_NOT_FOUND                            // 内存未找到错误
	MEMORY_EXIST                                // 内存已存在错误
	INVALID_VALUE                               // 无效数值
	EMPTY_VALUE                                 // 空值错误
	PARAM_ERROR                                 // 参数错误
	DO_ERROR                                    // 请求错误
	CONN_ERROR                                  // 连接失败
	JSON_NOT_EXPECTED                           // 不被预期的json
	REDIS_NOT_FOUND                             // Redis客户端未找到
	REDIS_DO_ERROR                              // Redis请求错误
	SFTP_CONN_ERR                               //
	REMOTE_ERROR                                // 远程错误
	FILE_READ_ERROR                             // 文件打开错误
	FILE_WRITE_ERROR                            // 文件写入错误
	CONFIG_LOAD_ERROR                           // 配置加载错误
	JSON_MARSHAL_ERROR                          // JSON序列化错误
	SIGN_HMAC_SHA256_ERROR                      // HMAC SHA256签名错误
	POINTER_NIL                                 // 空指针错误

	SYMBOL_INFO_HTTP_ERROR // 获取交易对信息HTTP错误

	// 动态交易规范
	STATIC_NOT_SUPPORTED
	STATIC_SYMBOL_NOT_FOUND    //
	DYNAMIC_SYMBOL_NOT_FOUND   // 动态交易规范未找到
	INVALID_UP_LIMIT_PERCENT   // 无效的涨停百分比
	INVALID_DOWN_LIMIT_PERCENT // 无效的跌停百分比
	INVALID_MIN_QTY            // 无效的最小下单金额
	INVALID_LOT_SIZE           // 无效的最小交易单位
	INVALID_TICK_SIZE          // 无效的最小价格变动单位
	//http错误码

	HTTP_PARAM_ERROR   // HTTP参数错误
	HTTP_DO_ERROR      // HTTP请求错误
	BN_PING_SEND_ERROR // BN ws ping发送失败
	BN_PONG_SEND_ERROR // BN ws pong发送失败

	CodeWsParamError // ws参数错误
	CodeWsDoError    // ws请求错误
	WS_SEND_ERROR    // ws发送错误

	// redis错误码
	REDIS_CLIENT_NOT_FOUND     // Redis客户端未找到
	REDIS_LUA_LOAD_ERROR       //
	REDIS_LUA_EVAL_ERROR       // Redis eval请求错误
	REDIS_LUA_SCRIPT_NOT_FOUND //

	//SFTP错误码
	CodeSftpParamError // SFTP参数错误
	CodeSftpDoError    // SFTP请求错误
	SFTP_CLIENT_ERR
	OPEN_REMOTE_FILE_ERROR
	GET_REMOTE_FILE_ERROR
	DOWNLOAD_REMOTE_FILE_ERROR
	DEL_REMOTE_FILE_ERR

	//文件错误码

	CodeFileRead      // 文件打开错误
	CodeFileWrite     // 文件写入错误
	CodeDirOpen       // 目录打开错误
	CODE_JSON_MARSHAL // JSON序列化错误

	//对象错误

	CodeNewFailed // 对象创建失败

	//值错误

	INSTANCE_ID_EMPTY          // 实例ID为空
	INSTANCE_EXISTS            //
	INSTANCE_NOT_EXISTS        // 实例ID不存在
	CLIENT_ORDER_ID_EMPTY      // 客户端订单ID为空
	CLIENT_ORDER_ID_EXISTS     //
	CLIENT_ORDER_ID_NOT_EXISTS //
	ORDER_STATUS_EMPTY         // 订单状态为空
	FROM_EMPTY                 // From 不能为空
	SYMBOL_KEY_EMPTY           // 交易对Key为空
	ACCOUNT_KEY_EMPTY          //
	ACCOUNT_KEY_NOT_EXISTS     // 账号Key不存在
	SYMBOL_NAME_NOT_EXISTS     // 账号Key不存在
	CodeEmptyValue             // 空值错误
	CodeNotFound               // 未找到错误
	CodeAlreadyExists          // 已存在错误
	ENUM_DEFINE_ERROR          // 枚举定义错误
	ENUM_NOT_SUPPORTED         // 枚举不支持
	CodeNotSupported           // 不支持的操作
)
