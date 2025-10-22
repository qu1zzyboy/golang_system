让 Tree News 模块直接读取 config/config_main.yaml 中的 treeNews 配置，并在需要时仍可被环境变量覆盖。主要改动如下：

配置解析

libs/quantGoInfra/conf/model_define.go 新增 TreeNewsConfig 结构体；
define.go 添加 TreeNewsCfg 全局变量；
conf_boot.go 在 initConfig() 中解析 treeNews.* 字段（含 enabled/apiKey/url/...），支持默认值及无效字符串日志提示。
Tree News 模块

internal/strategy/treenews/config.go 先读 conf.TreeNewsCfg，再套默认值、最后读取环境变量，保证配置优先级：“YAML → 默认 → 环境变量”。
internal/strategy/treenews/boot.go 增加对 conf.MODULE_ID 的依赖，确保读取配置后再启动。
配置文件

config/config_main.yaml 已加入 treeNews: 节，包含启停开关、API Key、URL、心跳参数等。