#!/bin/bash

# 连续 K 线服务停止脚本

set -e

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

PID_FILE=".kline.pid"
BINARY_NAME="continuous-kline-test"

echo "=== 停止连续 K 线服务 ==="

# 方法1: 从 PID 文件停止
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "找到进程 (PID: $PID)，正在停止..."
        kill "$PID"
        
        # 等待进程结束
        for i in {1..10}; do
            if ! ps -p "$PID" > /dev/null 2>&1; then
                echo "✅ 服务已停止"
                rm -f "$PID_FILE"
                exit 0
            fi
            sleep 1
        done
        
        # 如果还没停止，强制杀死
        if ps -p "$PID" > /dev/null 2>&1; then
            echo "强制停止进程..."
            kill -9 "$PID"
            sleep 1
        fi
        
        rm -f "$PID_FILE"
        echo "✅ 服务已停止"
        exit 0
    else
        echo "PID 文件存在但进程不存在，清理 PID 文件"
        rm -f "$PID_FILE"
    fi
fi

# 方法2: 通过进程名查找并停止
PIDS=$(pgrep -f "$BINARY_NAME" || true)
if [ -n "$PIDS" ]; then
    echo "找到运行中的进程: $PIDS"
    for PID in $PIDS; do
        echo "停止进程 (PID: $PID)..."
        kill "$PID" 2>/dev/null || kill -9 "$PID" 2>/dev/null
    done
    sleep 1
    echo "✅ 服务已停止"
else
    echo "未找到运行中的服务"
fi

# 清理 PID 文件
rm -f "$PID_FILE"

