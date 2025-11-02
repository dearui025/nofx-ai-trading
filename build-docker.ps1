# ===========================================
# NOFX AI交易系统 - Docker构建脚本 (PowerShell)
# ===========================================

param(
    [switch]$NoTest,
    [switch]$Help
)

# 显示帮助信息
if ($Help) {
    Write-Host "NOFX AI交易系统 - Docker构建脚本" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法:" -ForegroundColor Yellow
    Write-Host "  .\build-docker.ps1                # 构建并测试镜像"
    Write-Host "  .\build-docker.ps1 -NoTest        # 构建镜像但跳过测试"
    Write-Host "  .\build-docker.ps1 -Help          # 显示此帮助信息"
    Write-Host ""
    exit 0
}

# 日志函数
function Write-Info {
    param($Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param($Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param($Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param($Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查Docker是否安装
function Test-Docker {
    Write-Info "检查Docker环境..."
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker未安装或不在PATH中"
        Write-Info "请安装Docker Desktop: https://www.docker.com/products/docker-desktop"
        exit 1
    }
    
    try {
        docker info | Out-Null
        Write-Success "Docker环境检查通过"
    }
    catch {
        Write-Error "Docker服务未运行"
        Write-Info "请启动Docker Desktop"
        exit 1
    }
}

# 构建后端镜像
function Build-Backend {
    Write-Info "开始构建后端镜像..."
    
    try {
        # 使用标准Dockerfile构建
        docker build -t nofx-backend:latest .
        
        # 也构建Zeabur优化版本
        if (Test-Path "Dockerfile.zeabur") {
            docker build -f Dockerfile.zeabur -t nofx-backend:zeabur .
        }
        
        Write-Success "后端镜像构建完成"
    }
    catch {
        Write-Error "后端镜像构建失败: $_"
        exit 1
    }
}

# 构建前端镜像
function Build-Frontend {
    Write-Info "开始构建前端镜像..."
    
    if (-not (Test-Path "web")) {
        Write-Warning "web目录不存在，跳过前端构建"
        return
    }
    
    try {
        Push-Location "web"
        
        # 检查是否有package.json
        if (-not (Test-Path "package.json")) {
            Write-Error "web/package.json不存在"
            Pop-Location
            return
        }
        
        # 构建前端镜像
        docker build -t nofx-frontend:latest .
        
        Pop-Location
        Write-Success "前端镜像构建完成"
    }
    catch {
        Pop-Location
        Write-Error "前端镜像构建失败: $_"
    }
}

# 运行测试
function Test-Images {
    if ($NoTest) {
        Write-Info "跳过镜像测试"
        return
    }
    
    Write-Info "运行Docker镜像测试..."
    
    try {
        # 测试后端镜像
        Write-Info "测试后端镜像..."
        docker run --rm -d --name nofx-backend-test -p 8080:8080 nofx-backend:latest
        
        # 等待服务启动
        Start-Sleep -Seconds 10
        
        # 检查健康状态
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                Write-Success "后端镜像测试通过"
            } else {
                Write-Warning "后端镜像健康检查失败"
            }
        }
        catch {
            Write-Warning "后端镜像健康检查失败: $_"
        }
        
        # 停止测试容器
        docker stop nofx-backend-test 2>$null
        
        Write-Success "Docker镜像测试完成"
    }
    catch {
        Write-Error "镜像测试失败: $_"
        docker stop nofx-backend-test 2>$null
    }
}

# 清理旧镜像
function Clear-OldImages {
    Write-Info "清理旧镜像..."
    
    try {
        # 删除悬空镜像
        docker image prune -f | Out-Null
        Write-Success "清理完成"
    }
    catch {
        Write-Warning "清理过程中出现错误: $_"
    }
}

# 显示镜像信息
function Show-Images {
    Write-Info "构建的镜像列表:"
    docker images | Select-String "nofx"
}

# 主函数
function Main {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "NOFX AI交易系统 - Docker构建脚本" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    # 检查Docker环境
    Test-Docker
    
    # 构建镜像
    Build-Backend
    Build-Frontend
    
    # 运行测试
    Test-Images
    
    # 清理
    Clear-OldImages
    
    # 显示结果
    Show-Images
    
    Write-Host ""
    Write-Success "Docker构建完成！"
    Write-Host ""
    Write-Host "使用方法:" -ForegroundColor Yellow
    Write-Host "  docker run -d -p 8080:8080 --name nofx-backend nofx-backend:latest"
    Write-Host "  docker run -d -p 3000:80 --name nofx-frontend nofx-frontend:latest"
    Write-Host ""
}

# 运行主函数
try {
    Main
}
catch {
    Write-Error "构建过程中发生错误: $_"
    exit 1
}