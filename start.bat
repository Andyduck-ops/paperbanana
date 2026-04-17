@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

REM ============================================================
REM  PaperBanana - Windows Launcher
REM  Double-click to start. Auto-checks dependencies and ports.
REM ============================================================

REM --- Enter project directory ---
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
REM Step 2: Check Node.js installation
REM ============================================================
where node >nul 2>&1
if !errorlevel! neq 0 (
    echo   [!!] Node.js is not installed.
    echo.
    echo       Please install Node.js from: https://nodejs.org/
    echo       Or use: winget install OpenJS.NodeJS
    echo.
    pause
    exit /b 1
)

for /f "tokens=1" %%V in ('node --version 2^>^&1') do (
    echo   [OK] Node.js %%V found
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
REM Step 4: Check ports availability
REM ============================================================
echo   [..] Checking ports...
set BACKEND_PORT=8080
set FRONTEND_PORT=5173

REM Check backend port
netstat -an | findstr ":%BACKEND_PORT% " | findstr "LISTENING" >nul 2>&1
if !errorlevel! equ 0 (
    echo   [!!] Port %BACKEND_PORT% is already in use!
    echo       Please stop the existing service or change the port.
    pause
    exit /b 1
)

REM Check frontend port
netstat -an | findstr ":%FRONTEND_PORT% " | findstr "LISTENING" >nul 2>&1
if !errorlevel! equ 0 (
    echo   [!!] Port %FRONTEND_PORT% is already in use!
    echo       Please stop the existing service or change the port.
    pause
    exit /b 1
)

echo   [OK] Ports %BACKEND_PORT% and %FRONTEND_PORT% are available

REM ============================================================
REM Step 5: Create data directories if needed
REM ============================================================
if not exist "data\PaperBananaBench\diagram" mkdir "data\PaperBananaBench\diagram"
if not exist "data\PaperBananaBench\plot" mkdir "data\PaperBananaBench\plot"
if not exist "data\PaperBananaBench\diagram\images" mkdir "data\PaperBananaBench\diagram\images"
if not exist "data\PaperBananaBench\plot\images" mkdir "data\PaperBananaBench\plot\images"
if not exist ".paperbanana" mkdir ".paperbanana"

REM Create .gitkeep files to preserve directory structure
if not exist "data\.gitkeep" type nul > "data\.gitkeep"
if not exist "data\PaperBananaBench\.gitkeep" type nul > "data\PaperBananaBench\.gitkeep"
if not exist "data\PaperBananaBench\diagram\.gitkeep" type nul > "data\PaperBananaBench\diagram\.gitkeep"
if not exist "data\PaperBananaBench\plot\.gitkeep" type nul > "data\PaperBananaBench\plot\.gitkeep"

echo   [OK] Data directories ready

REM ============================================================
REM Step 6: Check frontend dependencies
REM ============================================================
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
    echo   [OK] Frontend dependencies installed
) else (
    echo   [OK] Frontend dependencies found
)

REM ============================================================
REM Step 7: Check config file
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
REM Step 8: Start backend server
REM ============================================================
echo.
echo   [..] Starting backend server on http://localhost:%BACKEND_PORT% ...
start "PaperBanana Backend" cmd /c "go run ./cmd/server --config ./configs/config.yaml"

REM Wait for backend to start
timeout /t 3 /nobreak >nul

REM ============================================================
REM Step 9: Start frontend dev server
REM ============================================================
echo   [..] Starting frontend dev server on http://localhost:%FRONTEND_PORT% ...
cd web
start "PaperBanana Frontend" cmd /c "npm run dev"
cd ..

REM Wait and open browser
timeout /t 2 /nobreak >nul
start http://localhost:%FRONTEND_PORT%

echo.
echo ==========================================
echo   Services started successfully!
echo   Backend:  http://localhost:%BACKEND_PORT%
echo   Frontend: http://localhost:%FRONTEND_PORT%
echo.
echo   Press any key to STOP all services
echo ==========================================
echo.

pause >nul

REM ============================================================
REM Step 10: Cleanup - Stop all services
REM ============================================================
echo.
echo   [..] Stopping services...

REM Kill backend process
taskkill /FI "WINDOWTITLE eq PaperBanana Backend" /F >nul 2>&1

REM Kill frontend process
taskkill /FI "WINDOWTITLE eq PaperBanana Frontend" /F >nul 2>&1

echo   [OK] All services stopped
timeout /t 2 >nul
