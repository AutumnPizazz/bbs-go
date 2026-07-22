@echo off
setlocal EnableExtensions EnableDelayedExpansion
chcp 65001 >nul
title BBS-GO 构建与运行

set "ROOT=%~dp0"
if "%ROOT:~-1%"=="\" set "ROOT=%ROOT:~0,-1%"
set "WEB_DIR=%ROOT%\web"
set "APP=%ROOT%\bbs-go.exe"

cls
echo ========================================
echo          BBS-GO 一键构建启动
echo ========================================
echo.

if not exist "%ROOT%\go.mod" (
    echo 错误：脚本所在目录不是 bbs-go 项目根目录
    goto error
)
if not exist "%WEB_DIR%\package.json" (
    echo 错误：找不到前端项目目录：%WEB_DIR%
    goto error
)

where go >nul 2>&1
if errorlevel 1 (
    echo 错误：未找到 Go，请先安装 Go 并加入 PATH
    goto error
)
where corepack >nul 2>&1
if errorlevel 1 (
    echo 错误：未找到 Corepack，请先安装 Node.js
    goto error
)

tasklist /FI "IMAGENAME eq bbs-go.exe" 2>nul | find /I "bbs-go.exe" >nul
if not errorlevel 1 (
    echo 错误：bbs-go.exe 当前正在运行，请先关闭服务后再构建
    goto error
)

if not exist "%WEB_DIR%\node_modules" (
    echo [准备] 安装前端依赖...
    pushd "%WEB_DIR%"
    call corepack pnpm install --frozen-lockfile
    set "RESULT=!ERRORLEVEL!"
    popd
    if not "%RESULT%"=="0" goto error
)

echo [1/3] 构建前端 SPA...
pushd "%WEB_DIR%"
call corepack pnpm build:spa
set "RESULT=%ERRORLEVEL%"
popd
if not "%RESULT%"=="0" goto error

echo.
echo [2/3] 构建 Go 后端...
pushd "%ROOT%"
go build -trimpath -ldflags "-s -w" -o "%APP%" .
set "RESULT=%ERRORLEVEL%"
popd
if not "%RESULT%"=="0" goto error

echo.
echo [3/3] 启动服务...
echo.
echo ========================================
echo  服务正在运行，按 Ctrl+C 可停止
echo  关闭此窗口将终止服务
echo ========================================
echo.
pushd "%ROOT%"
"%APP%"
set "RESULT=%ERRORLEVEL%"
popd

echo.
echo ========================================
if "%RESULT%"=="0" (
    echo  服务已停止
) else (
    echo  服务异常退出，退出码：%RESULT%
)
echo ========================================
pause
exit /b %RESULT%

:error
echo.
echo ========================================
echo  构建失败！请检查上方错误信息
echo ========================================
pause
exit /b 1
