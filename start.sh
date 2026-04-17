#!/bin/bash

# ============================================================
#  PaperBanana - Linux/Mac Launcher
#  Run: ./start.sh
# ============================================================

# --- Enter project directory ---
cd "$(dirname "$0")"

echo ""
echo "=========================================="
echo "  PaperBanana - Academic Figure Generator"
echo "=========================================="
echo ""

# ============================================================
# Step 1: Check Go installation
# ============================================================
if ! command -v go &> /dev/null; then
    echo "  [!!] Go is not installed."
    echo ""
    echo "       Please install Go from: https://go.dev/dl/"
    echo "       Or use: brew install go"
    echo ""
    read -p "Press any key to exit..."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "  [OK] Go $GO_VERSION found"

# ============================================================
# Step 2: Check Node.js installation
# ============================================================
if ! command -v node &> /dev/null; then
    echo "  [!!] Node.js is not installed."
    echo ""
    echo "       Please install Node.js from: https://nodejs.org/"
    echo "       Or use: brew install node"
    echo ""
    read -p "Press any key to exit..."
    exit 1
fi

NODE_VERSION=$(node --version)
echo "  [OK] Node.js $NODE_VERSION found"

# ============================================================
# Step 3: Check environment configuration
# ============================================================
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        echo "  [..] Creating .env from .env.example ..."
        cp ".env.example" ".env"
        echo "  [!!] Please edit .env and add your API keys!"
        echo "       Required: GEMINI_API_KEY or OPENAI_API_KEY or ANTHROPIC_API_KEY"
        echo ""
    else
        echo "  [!!] No .env file found. Please create one with your API keys."
        echo ""
    fi
fi

# ============================================================
# Step 4: Check ports availability
# ============================================================
echo "  [..] Checking ports..."
BACKEND_PORT=8080
FRONTEND_PORT=5173

# Check backend port
if lsof -Pi :$BACKEND_PORT -sTCP:LISTEN -t >/dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":$BACKEND_PORT "; then
    echo "  [!!] Port $BACKEND_PORT is already in use!"
    echo "       Please stop the existing service or change the port."
    read -p "Press any key to exit..."
    exit 1
fi

# Check frontend port
if lsof -Pi :$FRONTEND_PORT -sTCP:LISTEN -t >/dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":$FRONTEND_PORT "; then
    echo "  [!!] Port $FRONTEND_PORT is already in use!"
    echo "       Please stop the existing service or change the port."
    read -p "Press any key to exit..."
    exit 1
fi

echo "  [OK] Ports $BACKEND_PORT and $FRONTEND_PORT are available"

# ============================================================
# Step 5: Create data directories if needed
# ============================================================
mkdir -p "data/PaperBananaBench/diagram/images"
mkdir -p "data/PaperBananaBench/plot/images"
mkdir -p ".paperbanana"

# Create .gitkeep files to preserve directory structure
touch "data/.gitkeep"
touch "data/PaperBananaBench/.gitkeep"
touch "data/PaperBananaBench/diagram/.gitkeep"
touch "data/PaperBananaBench/plot/.gitkeep"

echo "  [OK] Data directories ready"

# ============================================================
# Step 6: Check frontend dependencies
# ============================================================
if [ ! -d "web/node_modules" ]; then
    echo "  [..] Installing frontend dependencies ..."
    cd web
    npm install
    if [ $? -ne 0 ]; then
        echo "  [!!] Failed to install frontend dependencies"
        read -p "Press any key to exit..."
        exit 1
    fi
    cd ..
    echo "  [OK] Frontend dependencies installed"
else
    echo "  [OK] Frontend dependencies found"
fi

# ============================================================
# Step 7: Check config file
# ============================================================
if [ ! -f "configs/config.yaml" ]; then
    if [ -f "configs/config.yaml.example" ]; then
        echo "  [..] Creating config.yaml from example ..."
        cp "configs/config.yaml.example" "configs/config.yaml"
        echo "  [OK] Config file created"
    else
        echo "  [!!] No config file found. Please create configs/config.yaml"
        read -p "Press any key to exit..."
        exit 1
    fi
else
    echo "  [OK] Config file found"
fi

# ============================================================
# Step 8: Start backend server
# ============================================================
echo ""
echo "  [..] Starting backend server on http://localhost:$BACKEND_PORT ..."

go run ./cmd/server --config ./configs/config.yaml &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

# Check if backend started successfully
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo "  [!!] Backend failed to start"
    exit 1
fi

# ============================================================
# Step 9: Start frontend dev server
# ============================================================
echo "  [..] Starting frontend dev server on http://localhost:$FRONTEND_PORT ..."

cd web
npm run dev &
FRONTEND_PID=$!
cd ..

# Wait for frontend to start
sleep 2

# Check if frontend started successfully
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo "  [!!] Frontend failed to start"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

# Open browser
if command -v open &> /dev/null; then
    open "http://localhost:$FRONTEND_PORT"
elif command -v xdg-open &> /dev/null; then
    xdg-open "http://localhost:$FRONTEND_PORT"
fi

echo ""
echo "=========================================="
echo "  Services started successfully!"
echo "  Backend:  http://localhost:$BACKEND_PORT"
echo "  Frontend: http://localhost:$FRONTEND_PORT"
echo ""
echo "  Press Ctrl+C to STOP all services"
echo "=========================================="
echo ""

# ============================================================
# Cleanup function
# ============================================================
cleanup() {
    echo ""
    echo "  [..] Stopping services..."
    kill $FRONTEND_PID 2>/dev/null
    kill $BACKEND_PID 2>/dev/null
    echo "  [OK] All services stopped"
    exit 0
}

# Trap signals
trap cleanup SIGINT SIGTERM

# Wait for processes
wait
