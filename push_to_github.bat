@echo off
REM ====================================================
REM 批处理脚本：推送本地 Go 仓库到 GitHub
REM ====================================================

REM 1️⃣ 设置本地仓库路径
SET REPO_PATH=I:\go\neutron-master
SET REMOTE_URL=https://github.com/ruisika/neutron.git
SET BRANCH_NAME=main

REM 2️⃣ 进入仓库
cd /d %REPO_PATH% || (
    echo 仓库路径不存在：%REPO_PATH%
    pause
    exit /b
)

REM 3️⃣ 初始化 Git（如果还没初始化）
git rev-parse --is-inside-work-tree >nul 2>&1
IF ERRORLEVEL 1 (
    echo 初始化 Git 仓库...
    git init
)

REM 4️⃣ 设置远程仓库
git remote remove origin >nul 2>&1
git remote add origin %REMOTE_URL%

REM 5️⃣ 切换或创建本地分支
git checkout %BRANCH_NAME% 2>nul || git checkout -b %BRANCH_NAME%

REM 6️⃣ 添加所有修改
git add .

REM 7️⃣ 提交修改
git commit -m "Update local changes to fork" 2>nul

REM 8️⃣ 拉取远程更新（防止冲突）
git pull origin %BRANCH_NAME% --rebase

REM 9️⃣ 推送到远程
git push -u origin %BRANCH_NAME%

echo.
echo ==============================
echo 本地仓库已成功推送到 GitHub
echo ==============================
pause
