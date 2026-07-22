@echo off
chcp 65001 >nul
title BBS-GO 构建与运行

:start
cls
echo ========================================
echo          BBS-GO 一键构建启动
echo ========================================
echo.

echo [1/3] 构建前端 SPA...
cd /d D:\bbs-go\web 2>nul
if errorlevel 1 (
    echo 错误：找不到 D:\bbs-go\web 目录
    goto error
)
corepack pnpm build:spa
if errorlevel 1 goto error

echo.
echo [2/3] 构建 Go 后端...
cd /d D:\bbs-go 2>nul
if errorlevel 1 (
    echo 错误：找不到 D:\bbs-go 目录
    goto error
)
go build -trimpath -ldflags="-s -w" -o bbs-go.exe ./main.go
if errorlevel 1 goto error

echo.
echo [3/3] 启动服务...
echo.
echo ========================================
echo  服务正在运行，按 Ctrl+C 可停止
echo  关闭此窗口将终止服务
echo ========================================
echo.
.\bbs-go.exe

echo.
echo ========================================
echo  服务已停止
echo ========================================
pause
goto end

:error
echo.
echo ========================================
echo  构建失败！请检查上方错误信息
echo ========================================
pause

:end