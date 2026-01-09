#!/bin/bash

# 自动从远程分支 feature/order 拉取更新

# TARGET_BRANCH="feature_delete"
TARGET_BRANCH="upbit_bybit"


echo "🚀 正在拉取分支 $TARGET_BRANCH..."

# 确保在 Git 仓库内
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
  echo "❌ 当前目录不是 Git 仓库，请切换到仓库目录再执行。"
  exit 1
fi

# 检查当前分支是否是目标分支
current_branch=$(git symbolic-ref --short HEAD)

if [ "$current_branch" != "$TARGET_BRANCH" ]; then
  echo "⚠️ 当前分支是 '$current_branch'，将自动切换到 '$TARGET_BRANCH'..."
  git checkout $TARGET_BRANCH || {
    echo "❌ 无法切换到分支 '$TARGET_BRANCH'，请确认该分支存在。"
    exit 1
  }
fi

# 拉取远程分支最新代码
git pull origin $TARGET_BRANCH

# 检查是否成功
if [ $? -eq 0 ]; then
  echo "✅ 拉取完成！"
else
  echo "❌ 拉取失败，请检查上面的错误信息。"
fi
