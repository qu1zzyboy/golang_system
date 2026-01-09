#!/bin/bash

# 自动检测当前 Git 分支
# ./git_commit_push.sh "修复订单逻辑"

# ✅ 用法提示
if [ $# -lt 1 ]; then
  echo "❗ 用法: $0 <提交信息>"
  echo "示例: $0 '修复订单逻辑'"
  exit 1
fi

COMMIT_MSG="$1"

# ✅ 确保在 Git 仓库中
REPO_DIR=$(git rev-parse --show-toplevel 2>/dev/null)
if [ $? -ne 0 ]; then
  echo "❌ 当前目录不是一个 Git 仓库"
  exit 1
fi

cd "$REPO_DIR"

CURRENT_BRANCH=$(git branch --show-current)

echo "📍 当前分支: $CURRENT_BRANCH"

# ✅ 添加所有更改（新增、修改、删除）
echo "➕ 添加所有变更..."
git add -A

# ✅ 提交
echo "💬 提交内容: $COMMIT_MSG"
git commit -m "$COMMIT_MSG"
if [ $? -ne 0 ]; then
  echo "⚠️ 没有可以提交的变更，或提交失败"
  exit 0
fi

echo "✅ 提交成功 ✅"

# ✅ 自动推送
echo "🚀 推送中 → origin/$CURRENT_BRANCH ..."
git push origin "$CURRENT_BRANCH"

if [ $? -eq 0 ]; then
  echo "🎉 推送成功"
else
  echo "❌ 推送失败，请检查远程分支或网络"
fi
