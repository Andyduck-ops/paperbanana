# PaperBanana 一键启动脚本 (PowerShell)
# 功能: 检查依赖、启动后端和前端、打开浏览器

param(
    [switch]$SkipBrowser,
    [switch]$BackendOnly,
    [switch]$FrontendOnly,
    [string]$BackendPort = "8080",
    [string]$FrontendPort = "5173"
)

$ErrorActionPreference = "Stop"

# 颜色输出
function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Success($msg) { Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warning($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Error($msg) { Write-Host "[ERR] $msg" -ForegroundColor Red }

# 检查端口是否被占用
function Test-PortInUse($port) {
    $connection = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    return $null -ne $connection
}

# 查找可用端口
function Find-AvailablePort($startPort) {
    $port = $startPort
    while (Test-PortInUse $port) {
        $port++
    }
    return $port
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    # 检查 Go
    try {
        $goVersion = go version
        Write-Success "Go 已安装: $goVersion"
    } catch {
        Write-Error "未找到 Go，请访问 https://golang.org/dl/ 安装"
        exit 1
    }
    
    # 检查 Node.js
    try {
        $nodeVersion = node --version
        Write-Success "Node.js 已安装: $nodeVersion"
    } catch {
        Write-Error "未找到 Node.js，请访问 https://nodejs.org/ 安装"
        exit 1
    }
    
    # 检查 npm
    try {
        $npmVersion = npm --version
        Write-Success "npm 已安装: $npmVersion"
    } catch {
        Write-Error "未找到 npm"
        exit 1
    }
}

# 准备环境
function Initialize-Environment {
    Write-Info "初始化环境..."
    
    # 创建数据目录
    $dataDirs = @("data", "data/PaperBananaBench", "web/dist")
    foreach ($dir in $dataDirs) {
        if (!(Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
            Write-Success "创建目录: $dir"
        }
    }
    
    # 检查 .env 文件
    if (!(Test-Path ".env")) {
        if (Test-Path ".env.example") {
            Copy-Item ".env.example" ".env"
            Write-Warning ".env 文件已创建，请编辑配置API密钥"
        }
    }
    
    # 安装前端依赖
    if (!(Test-Path "web/node_modules")) {
        Write-Info "安装前端依赖..."
        Push-Location web
        npm install
        Pop-Location
        Write-Success "前端依赖安装完成"
    }
}

# 启动后端
function Start-Backend {
    Write-Info "启动后端服务..."
    
    # 查找可用端口
    $actualPort = Find-AvailablePort $BackendPort
    if ($actualPort -ne $BackendPort) {
        Write-Warning "端口 $BackendPort 被占用，使用端口 $actualPort"
    }
    
    # 检查端口占用
    if (Test-PortInUse $actualPort) {
        Write-Error "端口 $actualPort 仍被占用"
        exit 1
    }
    
    # 编译并启动后端
    $env:PORT = $actualPort
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/server", "--config", "./configs/config.yaml" -NoNewWindow -PassThru | Out-Null
    
    # 等待后端启动
    $maxAttempts = 30
    $attempt = 0
    $started = $false
    
    while ($attempt -lt $maxAttempts -and !$started) {
        Start-Sleep -Milliseconds 500
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$actualPort/api/health" -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                $started = $true
            }
        } catch {}
        $attempt++
    }
    
    if ($started) {
        Write-Success "后端服务已启动: http://localhost:$actualPort"
        return $actualPort
    } else {
        Write-Error "后端服务启动失败"
        exit 1
    }
}

# 启动前端
function Start-Frontend($backendPort) {
    Write-Info "启动前端服务..."
    
    # 查找可用端口
    $actualPort = Find-AvailablePort $FrontendPort
    if ($actualPort -ne $FrontendPort) {
        Write-Warning "端口 $FrontendPort 被占用，使用端口 $actualPort"
    }
    
    # 设置后端API地址
    $env:VITE_API_URL = "http://localhost:$backendPort"
    
    Push-Location web
    
    # 启动Vite开发服务器
    Start-Process -FilePath "npm" -ArgumentList "run", "dev", "--", "--port", $actualPort -NoNewWindow -PassThru | Out-Null
    
    # 等待前端启动
    $maxAttempts = 30
    $attempt = 0
    $started = $false
    
    while ($attempt -lt $maxAttempts -and !$started) {
        Start-Sleep -Milliseconds 500
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$actualPort" -UseBasicParsing -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                $started = $true
            }
        } catch {}
        $attempt++
    }
    
    Pop-Location
    
    if ($started) {
        Write-Success "前端服务已启动: http://localhost:$actualPort"
        return $actualPort
    } else {
        Write-Error "前端服务启动失败"
        exit 1
    }
}

# 打开浏览器
function Open-Browser($url) {
    if (!$SkipBrowser) {
        Write-Info "打开浏览器: $url"
        Start-Process $url
    }
}

# 主函数
function Main {
    Write-Host @"
╔══════════════════════════════════════════════════════════╗
║            PaperBanana 一键启动脚本                       ║
║            One-Click Startup Script                      ║
╚══════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan
    
    # 检查依赖
    Test-Dependencies
    
    # 初始化环境
    Initialize-Environment
    
    $backendPort = $null
    $frontendPort = $null
    
    # 启动服务
    if (!$FrontendOnly) {
        $backendPort = Start-Backend
    }
    
    if (!$BackendOnly) {
        $frontendPort = Start-Frontend $backendPort
    }
    
    # 打开浏览器
    if ($frontendPort) {
        Open-Browser "http://localhost:$frontendPort"
    } elseif ($backendPort) {
        Open-Browser "http://localhost:$backendPort"
    }
    
    Write-Host @"

╔══════════════════════════════════════════════════════════╗
║  服务启动成功!                                           ║
║  Services started successfully!                          ║
╠══════════════════════════════════════════════════════════╣
  后端 API: http://localhost:$backendPort
  前端 UI: http://localhost:$frontendPort
╚══════════════════════════════════════════════════════════╝

按 Ctrl+C 停止服务
Press Ctrl+C to stop services
"@ -ForegroundColor Green
    
    # 保持脚本运行
    while ($true) {
        Start-Sleep -Seconds 1
    }
}

# 捕获Ctrl+C
[Console]::TreatControlCAsInput = $true

Main
