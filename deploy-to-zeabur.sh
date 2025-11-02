#!/bin/bash

# NOFX AI交易系统 - Zeabur自动部署脚本
# 使用方法: ./deploy-to-zeabur.sh

set -e

echo "🚀 开始部署NOFX AI交易系统到Zeabur..."
echo "================================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 环境变量
ZEABUR_TOKEN="sk-xp4jxe5vwirnzkqgncgaakxqsa4fm"
PROJECT_NAME="nofx-ai-trading"
SERVICE_NAME="nofx-backend"

# API密钥
BINANCE_API_KEY="H2StgimIA1ZlWbOKPxM4WlBdNnBN7kfvQCDTKTFLV0RBnhRbuXmyks9mSu42z3Wd"
BINANCE_SECRET_KEY="5Jw03ZarCQ13eGMV10CJFw2aQe4CJ3NVGXs14jXDWcDNZwe0wvQx9jXsGouVRWIB"
DEEPSEEK_API_KEY="sk-87efaa443e9e4562b2a49ed141db4b2f"

# 检查Zeabur CLI是否安装
check_zeabur_cli() {
    echo -e "${YELLOW}检查Zeabur CLI...${NC}"
    if ! command -v zeabur &> /dev/null; then
        echo -e "${RED}❌ Zeabur CLI未安装${NC}"
        echo "请运行以下命令安装:"
        echo "  curl -fsSL https://zeabur.com/install.sh | bash"
        exit 1
    fi
    echo -e "${GREEN}✅ Zeabur CLI已安装${NC}"
}

# 登录Zeabur
login_zeabur() {
    echo -e "${YELLOW}登录Zeabur...${NC}"
    export ZEABUR_TOKEN="$ZEABUR_TOKEN"
    if zeabur auth login --token "$ZEABUR_TOKEN"; then
        echo -e "${GREEN}✅ Zeabur登录成功${NC}"
    else
        echo -e "${RED}❌ Zeabur登录失败${NC}"
        exit 1
    fi
}

# 创建或选择项目
setup_project() {
    echo -e "${YELLOW}设置项目...${NC}"
    
    # 检查项目是否存在
    if zeabur project list | grep -q "$PROJECT_NAME"; then
        echo -e "${GREEN}✅ 项目已存在: $PROJECT_NAME${NC}"
        zeabur project use "$PROJECT_NAME"
    else
        echo "创建新项目: $PROJECT_NAME"
        if zeabur project create "$PROJECT_NAME"; then
            echo -e "${GREEN}✅ 项目创建成功${NC}"
            zeabur project use "$PROJECT_NAME"
        else
            echo -e "${RED}❌ 项目创建失败${NC}"
            exit 1
        fi
    fi
}

# 初始化Git仓库
setup_git() {
    echo -e "${YELLOW}设置Git仓库...${NC}"
    
    if [ ! -d ".git" ]; then
        git init
        git config user.email "deploy@nofx.ai"
        git config user.name "NOFX Deploy Bot"
        echo -e "${GREEN}✅ Git仓库初始化成功${NC}"
    else
        echo -e "${GREEN}✅ Git仓库已存在${NC}"
    fi
    
    # 添加所有文件
    git add .
    git commit -m "Deploy to Zeabur - $(date '+%Y-%m-%d %H:%M:%S')" || echo "没有新的更改"
}

# 部署服务
deploy_service() {
    echo -e "${YELLOW}部署服务到Zeabur...${NC}"
    
    # 使用Zeabur CLI部署
    if zeabur deploy --service "$SERVICE_NAME"; then
        echo -e "${GREEN}✅ 服务部署成功${NC}"
    else
        echo -e "${RED}❌ 服务部署失败${NC}"
        echo "尝试使用其他方式部署..."
        
        # 如果CLI失败，提供手动部署说明
        echo ""
        echo "请按照以下步骤手动部署:"
        echo "1. 访问 https://zeabur.com"
        echo "2. 创建新项目: $PROJECT_NAME"
        echo "3. 添加服务 > 从Git仓库"
        echo "4. 选择Dockerfile构建方式"
        echo "5. 配置环境变量（见下方）"
        exit 1
    fi
}

# 配置环境变量
configure_env() {
    echo -e "${YELLOW}配置环境变量...${NC}"
    
    # 设置环境变量
    zeabur env set BINANCE_API_KEY="$BINANCE_API_KEY" --service "$SERVICE_NAME"
    zeabur env set BINANCE_SECRET_KEY="$BINANCE_SECRET_KEY" --service "$SERVICE_NAME"
    zeabur env set DEEPSEEK_API_KEY="$DEEPSEEK_API_KEY" --service "$SERVICE_NAME"
    zeabur env set GO_ENV="production" --service "$SERVICE_NAME"
    zeabur env set PORT="8080" --service "$SERVICE_NAME"
    zeabur env set JWT_SECRET="nofx-ai-trading-jwt-secret-2025" --service "$SERVICE_NAME"
    zeabur env set MAX_DAILY_LOSS="10.0" --service "$SERVICE_NAME"
    zeabur env set MAX_DRAWDOWN="20.0" --service "$SERVICE_NAME"
    zeabur env set BTC_ETH_LEVERAGE="5" --service "$SERVICE_NAME"
    zeabur env set ALTCOIN_LEVERAGE="5" --service "$SERVICE_NAME"
    zeabur env set USE_DEFAULT_COINS="true" --service "$SERVICE_NAME"
    zeabur env set DEFAULT_COINS="BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT,XRPUSDT" --service "$SERVICE_NAME"
    
    echo -e "${GREEN}✅ 环境变量配置完成${NC}"
}

# 获取部署URL
get_deployment_url() {
    echo -e "${YELLOW}获取部署URL...${NC}"
    
    # 获取服务列表和URL
    DEPLOYMENT_URL=$(zeabur service list --project "$PROJECT_NAME" | grep "$SERVICE_NAME" | awk '{print $3}')
    
    if [ -n "$DEPLOYMENT_URL" ]; then
        echo -e "${GREEN}✅ 后端部署URL: $DEPLOYMENT_URL${NC}"
        echo "$DEPLOYMENT_URL" > deployment-url.txt
        echo ""
        echo "================================================"
        echo "🎉 部署完成!"
        echo "================================================"
        echo ""
        echo "后端API: $DEPLOYMENT_URL"
        echo "健康检查: $DEPLOYMENT_URL/health"
        echo ""
        echo "请更新前端配置中的API URL:"
        echo "  VITE_API_URL=$DEPLOYMENT_URL"
        echo "  VITE_WS_URL=wss://$(echo $DEPLOYMENT_URL | sed 's/https:\/\///')"
        echo ""
    else
        echo -e "${YELLOW}⚠️  无法自动获取URL，请在Zeabur控制台查看${NC}"
        echo "访问 https://zeabur.com/dashboard"
    fi
}

# 主流程
main() {
    echo "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
    
    check_zeabur_cli
    login_zeabur
    setup_project
    setup_git
    deploy_service
    configure_env
    get_deployment_url
    
    echo ""
    echo "结束时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "================================================"
}

# 运行主流程
main
