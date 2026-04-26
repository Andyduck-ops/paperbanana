@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

REM ============================================================
REM  PaperBanana - Windows One-Click Launcher
REM  Double-click to start. Auto-checks dependencies.
REM ============================================================

cd /d "%~dp0"

echo.
echo ==========================================
echo   PaperBanana - Academic Figure Generator
echo ==========================================
echo.

REM ============================================================
REM Step 1: Check Go installation
REM ============================================================
where go >nul 2>&1
if !errorlevel! neq 0 (
    echo   [!!] Go is not installed.
    echo.
    echo       Please install Go from: https://go.dev/dl/
    echo       Or use: winget install GoLang.Go
    echo.
    pause
    exit /b 1
)

for /f "tokens=3" %%V in ('go version 2^>^&1') do (
    echo   [OK] Go %%V found
)

REM ============================================================
REM Step 2: Build frontend if needed
REM ============================================================
if not exist "web\dist" (
    echo   [..] Frontend not built. Checking Node.js...
    where node >nul 2>&1
    if !errorlevel! neq 0 (
        echo   [!!] Node.js is not installed. Cannot build frontend.
        echo       Please install Node.js from: https://nodejs.org/
        echo       Or use: winget install OpenJS.NodeJS
        echo.
        pause
        exit /b 1
    )

    if not exist "web\node_modules" (
        echo   [..] Installing frontend dependencies ...
        cd web
        call npm install
        cd ..
        if !errorlevel! neq 0 (
            echo   [!!] Failed to install frontend dependencies
            pause
            exit /b 1
        )
    )

    echo   [..] Building frontend ...
    cd web
    call npm run build
    cd ..
    if !errorlevel! neq 0 (
        echo   [!!] Failed to build frontend
        pause
        exit /b 1
    )
    echo   [OK] Frontend built
) else (
    echo   [OK] Frontend build found
)

REM ============================================================
REM Step 3: Check environment configuration
REM ============================================================
if not exist ".env" (
    if exist ".env.example" (
        echo   [..] Creating .env from .env.example ...
        copy ".env.example" ".env" >nul
        echo   [!!] Please edit .env and add your API keys!
        echo       Required: GEMINI_API_KEY or OPENAI_API_KEY or ANTHROPIC_API_KEY
        echo.
    ) else (
        echo   [!!] No .env file found. Please create one with your API keys.
        echo.
    )
)

REM ============================================================
REM Step 4: Check config file
REM ============================================================
if not exist "configs\config.yaml" (
    if exist "configs\config.yaml.example" (
        echo   [..] Creating config.yaml from example ...
        copy "configs\config.yaml.example" "configs\config.yaml" >nul
        echo   [OK] Config file created
    ) else (
        echo   [!!] No config file found. Please create configs/config.yaml
        pause
        exit /b 1
    )
) else (
    echo   [OK] Config file found
)

REM ============================================================
REM Step 5: Check port availability
REM ============================================================
echo   [..] Checking port...
set BACKEND_PORT=8080

netstat -an | findstr ":%BACKEND_PORT% " | findstr "LISTENING" >nul 2>&1
if !errorlevel! equ 0 (
    echo   [!!] Port %BACKEND_PORT% is already in use!
    echo       Please stop the existing service or change the port.
    pause
    exit /b 1
)

echo   [OK] Port %BACKEND_PORT% is available

REM ============================================================
REM Step 6: Create data directories if needed
REM ============================================================
if not exist "data\PaperBananaBench\diagram" mkdir "data\PaperBananaBench\diagram"
if not exist "data\PaperBananaBench\plot" mkdir "data\PaperBananaBench\plot"
if not exist "data\PaperBananaBench\diagram\images" mkdir "data\PaperBananaBench\diagram\images"
if not exist "data\PaperBananaBench\plot\images" mkdir "data\PaperBananaBench\plot\images"
if not exist ".paperbanana" mkdir ".paperbanana"

echo   [OK] Data directories ready

REM ============================================================
REM Step 7: Start server
REM ============================================================
echo.
echo   [..] Starting PaperBanana on http://localhost:%BACKEND_PORT% ...
echo.

REM Open browser after a short delay
start /b cmd /c "timeout /t 3 /nobreak >nul && start http://localhost:%BACKEND_PORT%"

REM Run backend (serves both API and frontend)
go run ./cmd/server --config ./configs/config.yaml

REM Server exited
echo.
echo   [OK] Server stopped
timeout /t 2 >nul
