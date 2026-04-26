#!/bin/bash
# PaperBanana 开发工作流脚本
# 支持 DDD + TDD 开发流程

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERR]${NC} $1"; }

# 显示帮助
show_help() {
    cat << EOF
PaperBanana 开发工作流脚本

用法: $0 <命令> [选项]

命令:
  start              一键启动所有服务
  test               运行所有测试
  test-backend       仅运行后端测试
  test-frontend      仅运行前端测试
  test-e2e           运行E2E测试
  
  tdd-backend        TDD循环: 后端
  tdd-frontend       TDD循环: 前端
  
  lint               运行代码检查
  lint-backend       仅检查后端
  lint-frontend      仅检查前端
  
  build              构建项目
  build-backend      仅构建后端
  build-frontend     仅构建前端
  
  clean              清理构建产物
  setup              初始化开发环境
  
  help               显示此帮助

TDD 开发流程:
  1. $0 tdd-backend <package>   # 进入后端TDD模式
  2. 编写测试 -> 编写实现 -> 重构
  3. $0 tdd-frontend <component> # 进入前端TDD模式

示例:
  $0 start                      # 启动所有服务
  $0 test                       # 运行所有测试
  $0 tdd-backend domain/agent   # 对domain/agent进行TDD开发
  $0 tdd-frontend Button        # 对Button组件进行TDD开发

EOF
}

# 启动服务
cmd_start() {
    info "启动 PaperBanana 开发环境..."
    
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        powershell -ExecutionPolicy Bypass -File "$SCRIPT_DIR/start-all.ps1"
    else
        # Unix 系统
        info "检查依赖..."
        command -v go >/dev/null 2>&1 || { error "需要安装 Go"; exit 1; }
        command -v node >/dev/null 2>&1 || { error "需要安装 Node.js"; exit 1; }
        
        # 后端
        info "启动后端..."
        go run ./cmd/server --config ./configs/config.yaml &
        BACKEND_PID=$!
        
        # 前端
        info "启动前端..."
        cd web && npm run dev &
        FRONTEND_PID=$!
        
        # 等待信号
        trap "kill $BACKEND_PID $FRONTEND_PID; exit" INT TERM
        wait
    fi
}

# 测试命令
cmd_test() {
    "$SCRIPT_DIR/test-all.sh" "$@"
}

cmd_test_backend() {
    info "运行后端测试..."
    go test -v ./...
    success "后端测试完成"
}

cmd_test_frontend() {
    info "运行前端测试..."
    cd web
    npm run test:run
    success "前端测试完成"
}

cmd_test_e2e() {
    info "运行E2E测试..."
    cd web
    npm run build
    npx playwright test
    success "E2E测试完成"
}

# TDD 开发模式
cmd_tdd_backend() {
    local package="${1:-.}"
    
    info "进入后端TDD模式: $package"
    info "TDD循环: 编写测试(Red) -> 实现(Green) -> 重构(Refactor)"
    
    while true; do
        echo
        echo "当前包: $package"
        echo "[r] 运行测试"
        echo "[w] 监视模式"
        echo "[c] 覆盖率"
        echo "[q] 退出"
        read -p "选择: " choice
        
        case $choice in
            r|R)
                go test -v ./$package
                ;;
            w|W)
                go test -v -watch ./$package
                ;;
            c|C)
                go test -coverprofile=/tmp/coverage.out ./$package
                go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
                echo "覆盖率报告: /tmp/coverage.html"
                ;;
            q|Q)
                break
                ;;
        esac
    done
}

cmd_tdd_frontend() {
    local component="${1:-.}"
    
    info "进入前端TDD模式: $component"
    info "TDD循环: 编写测试(Red) -> 实现(Green) -> 重构(Refactor)"
    
    cd web
    
    while true; do
        echo
        echo "当前组件: $component"
        echo "[r] 运行测试"
        echo "[w] 监视模式"
        echo "[c] 覆盖率"
        echo "[u] UI模式"
        echo "[q] 退出"
        read -p "选择: " choice
        
        case $choice in
            r|R)
                npx vitest run $component
                ;;
            w|W)
                npx vitest watch $component
                ;;
            c|C)
                npx vitest run --coverage $component
                ;;
            u|U)
                npx vitest --ui
                ;;
            q|Q)
                break
                ;;
        esac
    done
}

# 代码检查
cmd_lint() {
    cmd_lint_backend
    cmd_lint_frontend
}

cmd_lint_backend() {
    info "检查后端代码..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
    else
        go vet ./...
    fi
    success "后端检查完成"
}

cmd_lint_frontend() {
    info "检查前端代码..."
    cd web
    npm run lint
    success "前端检查完成"
}

# 构建
cmd_build() {
    cmd_build_backend
    cmd_build_frontend
}

cmd_build_backend() {
    info "构建后端..."
    go build -o bin/paperbanana-server ./cmd/server
    success "后端构建完成: bin/paperbanana-server"
}

cmd_build_frontend() {
    info "构建前端..."
    cd web
    npm run build
    success "前端构建完成: web/dist/"
}

# 清理
cmd_clean() {
    info "清理构建产物..."
    rm -rf bin/
    rm -rf web/dist/
    rm -rf web/node_modules/.vite
    go clean -cache
    success "清理完成"
}

# 初始化环境
cmd_setup() {
    info "初始化开发环境..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        error "请安装 Go: https://golang.org/dl/"
        exit 1
    fi
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        error "请安装 Node.js: https://nodejs.org/"
        exit 1
    fi
    
    # 创建目录
    mkdir -p data
    mkdir -p bin
    
    # 安装前端依赖
    cd web
    if [ ! -d "node_modules" ]; then
        info "安装前端依赖..."
        npm install
    fi
    
    # 安装 Playwright
    if ! command -v npx playwright &> /dev/null; then
        info "安装 Playwright..."
        npx playwright install
    fi
    
    cd ..
    
    # 创建 .env
    if [ ! -f ".env" ]; then
        cp .env.example .env
        warning "请编辑 .env 文件配置API密钥"
    fi
    
    success "环境初始化完成!"
    info "运行 '$0 start' 启动开发服务器"
}

# 主函数
main() {
    local cmd="${1:-help}"
    shift || true
    
    case $cmd in
        start)
            cmd_start "$@"
            ;;
        test)
            cmd_test "$@"
            ;;
        test-backend)
            cmd_test_backend "$@"
            ;;
        test-frontend)
            cmd_test_frontend "$@"
            ;;
        test-e2e)
            cmd_test_e2e "$@"
            ;;
        tdd-backend)
            cmd_tdd_backend "$@"
            ;;
        tdd-frontend)
            cmd_tdd_frontend "$@"
            ;;
        lint)
            cmd_lint "$@"
            ;;
        lint-backend)
            cmd_lint_backend "$@"
            ;;
        lint-frontend)
            cmd_lint_frontend "$@"
            ;;
        build)
            cmd_build "$@"
            ;;
        build-backend)
            cmd_build_backend "$@"
            ;;
        build-frontend)
            cmd_build_frontend "$@"
            ;;
        clean)
            cmd_clean "$@"
            ;;
        setup)
            cmd_setup "$@"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            error "未知命令: $cmd"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
