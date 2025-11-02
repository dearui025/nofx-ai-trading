#!/bin/bash

# NOFX AI交易系统 - 一键部署脚本
# 自动部署后端到Zeabur和前端到Supabase
# 使用方法: ./one-click-deploy.sh

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                                                            ║"
echo "║          NOFX AI交易系统 - 一键部署脚本                    ║"
echo "║                                                            ║"
echo "║  后端: Zeabur                                              ║"
echo "║  前端: Supabase Storage                                    ║"
echo "║                                                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 部署配置
DEPLOYMENT_LOG="deployment-$(date '+%Y%m%d-%H%M%S').log"

# 日志函数
log() {
    echo -e "$1" | tee -a "$DEPLOYMENT_LOG"
}

log_section() {
    log ""
    log "================================================"
    log "$1"
    log "================================================"
}

log_success() {
    log "${GREEN}✅ $1${NC}"
}

log_error() {
    log "${RED}❌ $1${NC}"
}

log_warning() {
    log "${YELLOW}⚠️  $1${NC}"
}

log_info() {
    log "${BLUE}ℹ️  $1${NC}"
}

# 检查必要的工具
check_prerequisites() {
    log_section "步骤 1/5: 检查环境"
    
    local missing_tools=()
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        missing_tools+=("git")
    else
        log_success "Git: $(git --version)"
    fi
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        missing_tools+=("curl")
    else
        log_success "curl: 已安装"
    fi
    
    # 检查Node.js
    if ! command -v node &> /dev/null; then
        missing_tools+=("node")
    else
        log_success "Node.js: $(node --version)"
    fi
    
    # 检查npm
    if ! command -v npm &> /dev/null; then
        missing_tools+=("npm")
    else
        log_success "npm: $(npm --version)"
    fi
    
    # 如果有缺失的工具
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "缺少必要的工具: ${missing_tools[*]}"
        log_info "请先安装缺少的工具，然后重新运行脚本"
        exit 1
    fi
    
    log_success "所有必要工具已安装"
}

# 准备部署环境
prepare_deployment() {
    log_section "步骤 2/5: 准备部署环境"
    
    # 确保脚本有执行权限
    chmod +x deploy-to-zeabur.sh 2>/dev/null || true
    chmod +x deploy-frontend-to-supabase.sh 2>/dev/null || true
    
    log_success "部署脚本权限已设置"
    
    # 检查是否有.env文件
    if [ -f ".env" ]; then
        log_success "环境配置文件已存在"
    else
        log_warning "未找到.env文件，将使用默认配置"
    fi
}

# 部署后端到Zeabur
deploy_backend() {
    log_section "步骤 3/5: 部署后端到Zeabur"
    
    log_info "开始部署Go后端..."
    log_info "这可能需要几分钟时间，请耐心等待..."
    log ""
    
    # 检查是否有deploy-to-zeabur.sh脚本
    if [ -f "deploy-to-zeabur.sh" ]; then
        log_info "使用自动部署脚本..."
        
        if bash deploy-to-zeabur.sh 2>&1 | tee -a "$DEPLOYMENT_LOG"; then
            log_success "后端部署成功"
            
            # 尝试读取部署URL
            if [ -f "deployment-url.txt" ]; then
                BACKEND_URL=$(cat deployment-url.txt)
                log_success "后端URL: $BACKEND_URL"
            else
                log_warning "无法自动获取后端URL"
                read -p "请输入Zeabur部署的后端URL: " BACKEND_URL
            fi
        else
            log_error "后端自动部署失败"
            log_info "请按照以下步骤手动部署:"
            log_info "1. 访问 https://zeabur.com"
            log_info "2. 创建新项目"
            log_info "3. 连接GitHub仓库或上传代码"
            log_info "4. 选择Dockerfile构建"
            log_info "5. 配置环境变量（参见DEPLOYMENT_GUIDE.md）"
            
            read -p "手动部署完成后，请输入后端URL: " BACKEND_URL
        fi
    else
        log_warning "未找到自动部署脚本"
        log_info "请手动部署后端到Zeabur"
        read -p "部署完成后，请输入后端URL: " BACKEND_URL
    fi
    
    # 验证后端URL
    if [ -n "$BACKEND_URL" ]; then
        log_info "正在验证后端连接..."
        if curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health" | grep -q "200"; then
            log_success "后端健康检查通过"
        else
            log_warning "无法访问后端健康检查端点"
            log_warning "请确保后端已正常启动"
        fi
    else
        log_error "未提供后端URL"
        exit 1
    fi
}

# 部署前端到Supabase
deploy_frontend() {
    log_section "步骤 4/5: 部署前端到Supabase"
    
    log_info "后端URL: $BACKEND_URL"
    log_info "开始部署React前端..."
    log_info "这可能需要几分钟时间，请耐心等待..."
    log ""
    
    if [ -f "deploy-frontend-to-supabase.sh" ]; then
        log_info "使用自动部署脚本..."
        
        if bash deploy-frontend-to-supabase.sh "$BACKEND_URL" 2>&1 | tee -a "$DEPLOYMENT_LOG"; then
            log_success "前端部署成功"
        else
            log_error "前端自动部署失败"
            log_info "请参考DEPLOYMENT_GUIDE.md手动部署前端"
        fi
    else
        log_warning "未找到前端部署脚本"
        log_info "请手动部署前端到Supabase Storage"
        log_info "参考文档: DEPLOYMENT_GUIDE.md"
    fi
}

# 部署后验证
post_deployment_verification() {
    log_section "步骤 5/5: 部署验证"
    
    log_info "验证部署状态..."
    
    # 验证后端
    if [ -n "$BACKEND_URL" ]; then
        log_info "测试后端API..."
        
        # 健康检查
        if curl -s "$BACKEND_URL/health" | grep -q "ok"; then
            log_success "后端健康检查: 通过"
        else
            log_warning "后端健康检查: 未通过"
        fi
        
        # API端点测试
        if curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/api/market/data/BTCUSDT" | grep -q "200"; then
            log_success "API端点测试: 通过"
        else
            log_warning "API端点测试: 未通过或需要认证"
        fi
    fi
    
    # 验证前端
    SUPABASE_URL="https://eqzurdzoaxibothslnna.supabase.co"
    FRONTEND_URL="$SUPABASE_URL/storage/v1/object/public/nofx-frontend/index.html"
    
    log_info "测试前端访问..."
    if curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL" | grep -q "200"; then
        log_success "前端访问: 正常"
    else
        log_warning "前端访问: 可能还在部署中或需要配置"
    fi
    
    log_success "部署验证完成"
}

# 生成部署报告
generate_deployment_report() {
    log ""
    log "╔════════════════════════════════════════════════════════════╗"
    log "║                                                            ║"
    log "║                    🎉 部署完成！                           ║"
    log "║                                                            ║"
    log "╚════════════════════════════════════════════════════════════╝"
    log ""
    
    log "📊 部署信息:"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log ""
    log "🔹 后端API:"
    log "   URL: $BACKEND_URL"
    log "   健康检查: $BACKEND_URL/health"
    log "   API文档: $BACKEND_URL/api/docs"
    log ""
    log "🔹 前端界面:"
    log "   URL: $FRONTEND_URL"
    log ""
    log "🔹 部署日志:"
    log "   文件: $DEPLOYMENT_LOG"
    log ""
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log ""
    
    log "📝 下一步操作:"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log ""
    log "1. 访问前端URL进行功能测试"
    log "2. 检查API连接是否正常"
    log "3. 配置自定义域名（可选）"
    log "4. 设置监控和告警（推荐）"
    log "5. 配置备份策略（推荐）"
    log ""
    log "📚 更多信息请参考:"
    log "   - DEPLOYMENT_GUIDE.md - 完整部署指南"
    log "   - README.md - 项目说明"
    log "   - 部署日志: $DEPLOYMENT_LOG"
    log ""
    
    # 保存部署信息到文件
    cat > deployment-info.txt << EOF
NOFX AI交易系统部署信息
========================

部署时间: $(date '+%Y-%m-%d %H:%M:%S')

后端信息:
---------
URL: $BACKEND_URL
平台: Zeabur
状态: 已部署

前端信息:
---------
URL: $FRONTEND_URL
平台: Supabase Storage
状态: 已部署

访问链接:
---------
- 前端界面: $FRONTEND_URL
- 后端API: $BACKEND_URL
- 健康检查: $BACKEND_URL/health

部署日志: $DEPLOYMENT_LOG
EOF
    
    log_success "部署信息已保存到 deployment-info.txt"
}

# 主流程
main() {
    # 记录开始时间
    START_TIME=$(date +%s)
    
    log "部署开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log ""
    
    # 执行部署流程
    check_prerequisites
    prepare_deployment
    deploy_backend
    deploy_frontend
    post_deployment_verification
    generate_deployment_report
    
    # 计算耗时
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    MINUTES=$((DURATION / 60))
    SECONDS=$((DURATION % 60))
    
    log ""
    log "总耗时: ${MINUTES}分${SECONDS}秒"
    log "部署结束时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log ""
    log_success "部署流程全部完成！"
}

# 错误处理
trap 'log_error "部署过程中发生错误，请查看日志: $DEPLOYMENT_LOG"; exit 1' ERR

# 运行主流程
main "$@"
