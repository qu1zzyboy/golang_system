git checkout main

git merge dev

git add .

git commit -m "从dev合并"

git push origin main


// 压缩版
git checkout main
git merge --squash dev
git commit -m "全局止损之前的合并"
git push origin main