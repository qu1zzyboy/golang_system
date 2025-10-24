请恢复我之前构建的“启动器接口 + 注册机制”架构上下文：

- 我实现了一个名为 Bootable 的接口，定义了组件的 Start、Stop、Name、DependsOn 方法
- 有一个 BootManager 管理组件注册、拓扑排序、依赖检查、逆序关闭
- 每个模块作为组件注册，包括 logger、ws_conn 等
- 系统支持任意注册顺序、自动排序、检测循环依赖与缺失依赖
- 项目结构大致为 internal/infra/boot/{interface.go, manager.go, components/*.go, main.go}
- 我计划扩展健康检查、启动失败重试、配置驱动开关、插件化模块等功能
