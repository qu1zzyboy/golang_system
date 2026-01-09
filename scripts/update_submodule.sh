#!/bin/bash
# ================================================
# 自动更新 quantGoInfra 子模块并整理依赖
# 适用于: upbitBnServer + quantGoInfra 架构
# 作者: 你的名字
# ================================================

set -e  # 遇到错误立即退出

# ---------- 1️⃣ 更新子模块 ----------
echo "🚀 正在更新子模块 quantGoInfra ..."
git submodule update --init --recursive --remote libs/quantGoInfra

# ---------- 2️⃣ 整理子模块依赖 ----------
echo "🧩 整理子模块依赖 ..."
cd libs/quantGoInfra

if [ -f "go.mod" ]; then
  echo "执行 go mod tidy ..."
  go mod tidy
else
  echo "⚠️ 未找到 go.mod，是否忘记初始化 quantGoInfra？"
  exit 1
fi

# ---------- 3️⃣ 返回主仓 ----------
cd ../..

# 检查主仓 go.mod 是否包含 replace 语句
echo "🔍 检查主仓 go.mod ..."
if ! grep -q "replace github.com/hhh500/quantGoInfra => ./libs/quantGoInfra" go.mod; then
  echo "⚠️ 主仓 go.mod 未找到 replace 语句，请添加以下内容："
  echo 'replace github.com/hhh500/quantGoInfra => ./libs/quantGoInfra'
  echo 'require github.com/hhh500/quantGoInfra v0.0.0'
  exit 1
fi

# ---------- 4️⃣ 清理 & 整理主仓依赖 ----------
echo "🧹 清理缓存并重新整理主仓依赖 ..."
go clean -modcache
go mod tidy

# ---------- 5️⃣ 成功提示 ----------
echo ""
echo "✅ 子模块 quantGoInfra 更新成功并同步依赖完成！"
