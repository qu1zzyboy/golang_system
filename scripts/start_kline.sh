#!/bin/bash

# 连续 K 线服务启动脚本

set -e

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# 程序名称
APP_NAME="continuous-kline-test"
BINARY_NAME="continuous-kline-test"
MAIN_PATH="cmd/storage/continuousKlineTest/main.go"
LOG_DIR="logs"
PID_FILE=".kline.pid"

echo "=== 启动连续 K 线服务 ==="

# 检查是否已经运行
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p "$OLD_PID" > /dev/null 2>&1; then
        echo "服务已经在运行中 (PID: $OLD_PID)"
        echo "如需重启，请先运行: ./scripts/stop_kline.sh"
        exit 1
    else
        echo "清理旧的 PID 文件"
        rm -f "$PID_FILE"
    fi
fi

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo "警告: 未找到 .env 文件，将使用系统环境变量"
fi

# 编译程序
echo "正在编译程序..."
go build -o "$BINARY_NAME" "$MAIN_PATH"
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

# 创建日志目录
mkdir -p "$LOG_DIR"

# 启动程序
echo "正在启动服务..."
nohup ./"$BINARY_NAME" > /dev/null 2>&1 &
NEW_PID=$!

# 保存 PID
echo "$NEW_PID" > "$PID_FILE"

# 等待一下，检查进程是否还在运行
sleep 2
if ps -p "$NEW_PID" > /dev/null 2>&1; then
    echo "✅ 服务启动成功！"
    echo "PID: $NEW_PID"
    echo "日志文件: $LOG_DIR/$(date +%Y-%m-%d)/normal.log"
    echo ""
    echo "查看日志: tail -f $LOG_DIR/$(date +%Y-%m-%d)/normal.log"
    echo "停止服务: ./scripts/stop_kline.sh"
else
    echo "❌ 服务启动失败！"
    rm -f "$PID_FILE"
    exit 1
fi

