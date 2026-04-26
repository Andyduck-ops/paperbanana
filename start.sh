#!/bin/bash

# ============================================================
#  PaperBanana - Linux/Mac One-Click Launcher
#  Run: ./start.sh
# ============================================================

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
# Step 2: Build frontend if needed
# ============================================================
if [ ! -d "web/dist" ]; then
    echo "  [..] Frontend not built. Checking Node.js..."
    if ! command -v node &> /dev/null; then
        echo "  [!!] Node.js is not installed. Cannot build frontend."
        echo "       Please install Node.js from: https://nodejs.org/"
        echo "       Or use: brew install node"
        echo ""
        read -p "Press any key to exit..."
        exit 1
    fi

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
    fi

    echo "  [..] Building frontend ..."
    cd web
    npm run build
    if [ $? -ne 0 ]; then
        echo "  [!!] Failed to build frontend"
        read -p "Press any key to exit..."
        exit 1
    fi
    cd ..
    echo "  [OK] Frontend built"
else
    echo "  [OK] Frontend build found"
fi

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
# Step 4: Check config file
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
# Step 5: Check port availability
# ============================================================
echo "  [..] Checking port..."
BACKEND_PORT=8080

if lsof -Pi :$BACKEND_PORT -sTCP:LISTEN -t >/dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":$BACKEND_PORT "; then
    echo "  [!!] Port $BACKEND_PORT is already in use!"
    echo "       Please stop the existing service or change the port."
    read -p "Press any key to exit..."
    exit 1
fi

echo "  [OK] Port $BACKEND_PORT is available"

# ============================================================
# Step 6: Create data directories if needed
# ============================================================
mkdir -p "data/PaperBananaBench/diagram/images"
mkdir -p "data/PaperBananaBench/plot/images"
mkdir -p ".paperbanana"

echo "  [OK] Data directories ready"

# ============================================================
# Step 7: Start server
# ============================================================
echo ""
echo "  [..] Starting PaperBanana on http://localhost:$BACKEND_PORT ..."
echo ""

# Open browser after a short delay
(
    sleep 3
    if command -v open &> /dev/null; then
        open "http://localhost:$BACKEND_PORT"
    elif command -v xdg-open &> /dev/null; then
        xdg-open "http://localhost:$BACKEND_PORT"
    fi
) &

# Run backend (serves both API and frontend)
go run ./cmd/server --config ./configs/config.yaml

# Server exited
echo ""
echo "  [OK] Server stopped"
