#!/bin/bash

# ===========================================
# NOFX AI交易系统 - Docker构建脚本
# ===========================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装或不在PATH中"
        log_info "请安装Docker Desktop: https://www.docker.com/products/docker-desktop"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker服务未运行"
        log_info "请启动Docker Desktop"
        exit 1
    fi
    
    log_success "Docker环境检查通过"
}

# 构建后端镜像
build_backend() {
    log_info "开始构建后端镜像..."
    
    # 使用标准Dockerfile构建
    docker build -t nofx-backend:latest .
    
    # 也构建Zeabur优化版本
    if [ -f "Dockerfile.zeabur" ]; then
        docker build -f Dockerfile.zeabur -t nofx-backend:zeabur .
    fi
    
    log_success "后端镜像构建完成"
}

# 构建前端镜像
build_frontend() {
    log_info "开始构建前端镜像..."
    
    cd web
    
    # 检查是否有package.json
    if [ ! -f "package.json" ]; then
        log_error "web/package.json不存在"
        return 1
    fi
    
    # 构建前端镜像
    docker build -t nofx-frontend:latest .
    
    cd ..
    
    log_success "前端镜像构建完成"
}

# 运行测试
run_tests() {
    log_info "运行Docker镜像测试..."
    
    # 测试后端镜像
    log_info "测试后端镜像..."
    docker run --rm -d --name nofx-backend-test -p 8080:8080 nofx-backend:latest
    
    # 等待服务启动
    sleep 10
    
    # 检查健康状态
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "后端镜像测试通过"
    else
        log_warning "后端镜像健康检查失败"
    fi
    
    # 停止测试容器
    docker stop nofx-backend-test || true
    
    log_success "Docker镜像测试完成"
}

# 清理旧镜像
cleanup() {
    log_info "清理旧镜像..."
    
    # 删除悬空镜像
    docker image prune -f
    
    log_success "清理完成"
}

# 显示镜像信息
show_images() {
    log_info "构建的镜像列表:"
    docker images | grep nofx
}

# 主函数
main() {
    echo "========================================"
    echo "NOFX AI交易系统 - Docker构建脚本"
    echo "========================================"
    
    # 检查Docker环境
    check_docker
    
    # 构建镜像
    build_backend
    
    if [ -d "web" ]; then
        build_frontend
    else
        log_warning "web目录不存在，跳过前端构建"
    fi
    
    # 运行测试
    if [ "$1" != "--no-test" ]; then
        run_tests
    fi
    
    # 清理
    cleanup
    
    # 显示结果
    show_images
    
    echo ""
    log_success "Docker构建完成！"
    echo ""
    echo "使用方法:"
    echo "  docker run -d -p 8080:8080 --name nofx-backend nofx-backend:latest"
    echo "  docker run -d -p 3000:80 --name nofx-frontend nofx-frontend:latest"
    echo ""
}

# 运行主函数
main "$@"