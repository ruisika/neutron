@echo off
REM 设置你的本地Go项目路径
REM 设置命令行编码为 UTF-8 (代码页 65001)
chcp 65001 > nul 
SET "PROJECT_DIR=I:\go\neutron-master"
SET "REPO_URL=https://github.com/chainreactors/neutron.git"
SET "MAIN_BRANCH=main" REM 或者 'master'，取决于你的项目主分支名称

echo ----------------------------------------------------
echo 正在进入项目目录: %PROJECT_DIR%
echo ----------------------------------------------------
cd /d "%PROJECT_DIR%"

IF ERRORLEVEL 1 (
    echo 错误: 无法进入项目目录。请检查路径是否正确。
    goto :eof
)

echo ----------------------------------------------------
echo 正在检查 Git 状态...
echo ----------------------------------------------------
git status

REM 检查是否有未提交的更改
git diff --quiet --exit-code
IF NOT ERRORLEVEL 0 (
    echo.
    echo 警告: 存在未提交的更改。请先提交或暂存这些更改。
    echo ----------------------------------------------------
    goto :eof
)

echo ----------------------------------------------------
echo 正在拉取最新的远程更改...
REM 使用 'git pull' 来获取并合并远程更新
echo ----------------------------------------------------
git pull origin %MAIN_BRANCH%

IF ERRORLEVEL 1 (
    echo.
    echo 错误: git pull 失败。请检查网络连接、远程设置或是否存在合并冲突。
    echo ----------------------------------------------------
    goto :eof
)

echo ----------------------------------------------------
echo 更新成功! 正在获取最新的 Commit ID...
echo ----------------------------------------------------
SET "LATEST_COMMIT="

REM 获取最新的 commit ID
FOR /f "delims=" %%a IN ('git rev-parse HEAD') DO (
    SET "LATEST_COMMIT=%%a"
)

echo ----------------------------------------------------
echo 最新 Commit ID (HEAD):
echo %LATEST_COMMIT%
echo ----------------------------------------------------

REM 可选: 获取最新的 commit 消息
echo 正在获取最新的 Commit Message...
SET "COMMIT_MESSAGE="
FOR /f "delims=" %%b IN ('git log -1 --pretty=format:%%s') DO (
    SET "COMMIT_MESSAGE=%%b"
)
echo 最新 Commit Message:
echo %COMMIT_MESSAGE%
echo ----------------------------------------------------

echo 脚本执行完毕。
pause