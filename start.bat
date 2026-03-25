@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion

REM ============================================================
REM  PaperBanana - Windows Launcher
REM  Double-click to start. Auto-checks dependencies.
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
REM Step 2: Check environment configuration
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
REM Step 3: Create data directories if needed
REM ============================================================
if not exist "data\PaperBananaBench\diagram" mkdir "data\PaperBananaBench\diagram"
if not exist "data\PaperBananaBench\plot" mkdir "data\PaperBananaBench\plot"
if not exist "data\PaperBananaBench\diagram\images" mkdir "data\PaperBananaBench\diagram\images"
if not exist "data\PaperBananaBench\plot\images" mkdir "data\PaperBananaBench\plot\images"
if not exist ".paperbanana" mkdir ".paperbanana"

REM ============================================================
REM Step 4: Check frontend dependencies
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
)

REM ============================================================
REM Step 5: Start backend server
REM ============================================================
echo.
echo   [..] Starting backend server on http://localhost:8080 ...
start "PaperBanana Backend" cmd /c "go run ./cmd/server"

REM Wait a moment for backend to start
timeout /t 3 /nobreak >nul

REM ============================================================
REM Step 6: Start frontend dev server
REM ============================================================
echo   [..] Starting frontend dev server on http://localhost:5173 ...
cd web
start "PaperBanana Frontend" cmd /c "npm run dev"
cd ..

REM Wait and open browser
timeout /t 2 /nobreak >nul
start http://localhost:5173

echo.
echo ==========================================
echo   Services started!
echo   Backend:  http://localhost:8080
echo   Frontend: http://localhost:5173
echo
echo   Close this window to keep services running
echo   Or press Ctrl+C in the server windows to stop
echo ==========================================
echo.

pause
