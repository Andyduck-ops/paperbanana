#!/bin/bash
# PaperBanana 一键测试脚本
# 功能: 运行所有测试并生成覆盖率报告

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info() { echo -e "${CYAN}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERR]${NC} $1"; }

echo "╔══════════════════════════════════════════════════════════╗"
echo "║            PaperBanana 一键测试脚本                       ║"
echo "║            One-Click Test Script                         ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo

# 后端测试
run_backend_tests() {
    info "运行后端测试..."
    
    # 单元测试
    info "运行后端单元测试..."
    go test -v -short ./... 2>&1 | tee test-results/backend-test.log
    
    # 覆盖率
    info "生成后端覆盖率报告..."
    go test -coverprofile=test-results/backend-coverage.out ./...
    go tool cover -func=test-results/backend-coverage.out > test-results/backend-coverage.txt
    
    # 提取总覆盖率
    COVERAGE=$(go tool cover -func=test-results/backend-coverage.out | grep total | awk '{print $3}')
    success "后端测试完成，覆盖率: $COVERAGE"
}

# 前端测试
run_frontend_tests() {
    info "运行前端测试..."
    
    cd web
    
    # 单元测试
    info "运行前端单元测试..."
    npm run test:run -- --coverage 2>&1 | tee ../test-results/frontend-test.log
    
    cd ..
    success "前端测试完成"
}

# E2E测试
run_e2e_tests() {
    info "运行E2E测试..."
    
    cd web
    
    # 确保构建成功
    info "构建前端..."
    npm run build
    
    # 运行Playwright测试
    info "运行Playwright测试..."
    npx playwright test 2>&1 | tee ../test-results/e2e-test.log
    
    cd ..
    success "E2E测试完成"
}

# 契约测试
run_contract_tests() {
    info "运行契约测试..."
    
    # 验证OpenAPI规范
    if command -v swagger-cli &> /dev/null; then
        swagger-cli validate docs/api/openapi.yaml 2>&1 | tee test-results/contract-test.log
        success "契约测试完成"
    else
        warning "swagger-cli 未安装，跳过契约测试"
    fi
}

# 生成测试报告
generate_report() {
    info "生成测试报告..."
    
    cat > test-results/TEST_REPORT.md << EOF
# PaperBanana 测试报告

生成时间: $(date)

## 后端测试

- 日志: [backend-test.log](backend-test.log)
- 覆盖率: [backend-coverage.txt](backend-coverage.txt)

$(cat test-results/backend-coverage.txt | grep total)

## 前端测试

- 日志: [frontend-test.log](frontend-test.log)
- 覆盖率报告: web/coverage/

## E2E测试

- 日志: [e2e-test.log](e2e-test.log)
- Playwright报告: web/playwright-report/

## 契约测试

- 日志: [contract-test.log](contract-test.log)
EOF
    
    success "测试报告生成: test-results/TEST_REPORT.md"
}

# 主函数
main() {
    # 创建结果目录
    mkdir -p test-results
    
    # 解析参数
    RUN_BACKEND=true
    RUN_FRONTEND=true
    RUN_E2E=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backend-only)
                RUN_FRONTEND=false
                shift
                ;;
            --frontend-only)
                RUN_BACKEND=false
                shift
                ;;
            --e2e)
                RUN_E2E=true
                shift
                ;;
            --all)
                RUN_E2E=true
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    # 运行测试
    if [ "$RUN_BACKEND" = true ]; then
        run_backend_tests
    fi
    
    if [ "$RUN_FRONTEND" = true ]; then
        run_frontend_tests
    fi
    
    if [ "$RUN_E2E" = true ]; then
        run_e2e_tests
    fi
    
    run_contract_tests
    generate_report
    
    echo
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║  所有测试完成!                                           ║"
    echo "║  All tests completed!                                    ║"
    echo "╚══════════════════════════════════════════════════════════╝"
}

main "$@"
