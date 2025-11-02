#!/bin/bash

# NOFX AI交易系统 - Supabase前端部署脚本
# 使用方法: ./deploy-frontend-to-supabase.sh [BACKEND_URL]

set -e

echo "🚀 开始部署前端到Supabase..."
echo "================================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Supabase配置
SUPABASE_PROJECT_ID="eqzurdzoaxibothslnna"
SUPABASE_URL="https://eqzurdzoaxibothslnna.supabase.co"
SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVxenVyZHpvYXhpYm90aHNsbm5hIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE4NzY2NjUsImV4cCI6MjA3NzQ1MjY2NX0.h2EQOkofLavh-DL68AGfFX7ZvJ4SipNsiO7K5uTh20Y"
SUPABASE_ACCESS_TOKEN="sbp_cb3f3a6f373315e288f532e1ede5442ef4fbf311"
BUCKET_NAME="nofx-frontend"

# 后端URL（从参数获取或使用默认值）
BACKEND_URL="${1:-https://nofx-backend.zeabur.app}"

# 检查Node.js和npm
check_nodejs() {
    echo -e "${YELLOW}检查Node.js环境...${NC}"
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js未安装${NC}"
        echo "请安装Node.js: https://nodejs.org/"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}❌ npm未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Node.js $(node --version)${NC}"
    echo -e "${GREEN}✅ npm $(npm --version)${NC}"
}

# 检查Supabase CLI
check_supabase_cli() {
    echo -e "${YELLOW}检查Supabase CLI...${NC}"
    if ! command -v supabase &> /dev/null; then
        echo -e "${YELLOW}⚠️  Supabase CLI未安装，正在安装...${NC}"
        npm install -g supabase
    fi
    echo -e "${GREEN}✅ Supabase CLI已安装${NC}"
}

# 更新前端配置
update_frontend_config() {
    echo -e "${YELLOW}更新前端配置...${NC}"
    
    cd web
    
    # 创建生产环境配置
    cat > .env.production << EOF
# Supabase配置
VITE_SUPABASE_URL=$SUPABASE_URL
VITE_SUPABASE_ANON_KEY=$SUPABASE_ANON_KEY

# 后端API配置
VITE_API_URL=$BACKEND_URL
VITE_WS_URL=wss://$(echo $BACKEND_URL | sed 's/https:\/\///')

# 应用配置
VITE_APP_NAME=NOFX AI Trading System
VITE_APP_VERSION=1.0.0
EOF
    
    echo -e "${GREEN}✅ 前端配置更新完成${NC}"
    echo "后端URL: $BACKEND_URL"
    cd ..
}

# 安装依赖
install_dependencies() {
    echo -e "${YELLOW}安装前端依赖...${NC}"
    cd web
    
    if [ -f "package-lock.json" ]; then
        npm ci
    else
        npm install
    fi
    
    echo -e "${GREEN}✅ 依赖安装完成${NC}"
    cd ..
}

# 构建前端
build_frontend() {
    echo -e "${YELLOW}构建前端应用...${NC}"
    cd web
    
    # 运行构建
    npm run build
    
    if [ ! -d "dist" ]; then
        echo -e "${RED}❌ 构建失败：dist目录不存在${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 前端构建完成${NC}"
    cd ..
}

# 登录Supabase
login_supabase() {
    echo -e "${YELLOW}登录Supabase...${NC}"
    
    export SUPABASE_ACCESS_TOKEN="$SUPABASE_ACCESS_TOKEN"
    
    if supabase login --token "$SUPABASE_ACCESS_TOKEN" 2>/dev/null; then
        echo -e "${GREEN}✅ Supabase登录成功${NC}"
    else
        echo -e "${YELLOW}⚠️  使用access token进行身份验证${NC}"
    fi
}

# 链接项目
link_project() {
    echo -e "${YELLOW}链接Supabase项目...${NC}"
    
    if supabase link --project-ref "$SUPABASE_PROJECT_ID" 2>/dev/null; then
        echo -e "${GREEN}✅ 项目链接成功${NC}"
    else
        echo -e "${YELLOW}⚠️  项目可能已链接${NC}"
    fi
}

# 创建存储桶
create_bucket() {
    echo -e "${YELLOW}创建Storage桶...${NC}"
    
    # 使用Supabase API创建桶
    curl -X POST "$SUPABASE_URL/storage/v1/bucket" \
        -H "Authorization: Bearer $SUPABASE_ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"$BUCKET_NAME\",\"name\":\"$BUCKET_NAME\",\"public\":true}" \
        2>/dev/null || echo "桶可能已存在"
    
    echo -e "${GREEN}✅ Storage桶准备就绪${NC}"
}

# 上传文件到Supabase Storage
upload_files() {
    echo -e "${YELLOW}上传文件到Supabase Storage...${NC}"
    
    cd web/dist
    
    # 遍历所有文件并上传
    find . -type f | while read file; do
        # 移除前导 ./
        clean_path="${file#./}"
        
        echo "上传: $clean_path"
        
        # 读取文件内容并上传
        curl -X POST "$SUPABASE_URL/storage/v1/object/$BUCKET_NAME/$clean_path" \
            -H "Authorization: Bearer $SUPABASE_ACCESS_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            --data-binary "@$file" \
            2>/dev/null || echo "  ⚠️  上传失败或文件已存在: $clean_path"
    done
    
    cd ../..
    echo -e "${GREEN}✅ 文件上传完成${NC}"
}

# 生成访问URL
generate_urls() {
    echo ""
    echo "================================================"
    echo "🎉 前端部署完成!"
    echo "================================================"
    echo ""
    echo "前端访问URL:"
    echo "  $SUPABASE_URL/storage/v1/object/public/$BUCKET_NAME/index.html"
    echo ""
    echo "所有文件URL前缀:"
    echo "  $SUPABASE_URL/storage/v1/object/public/$BUCKET_NAME/"
    echo ""
    echo "提示："
    echo "  1. 如需自定义域名，请在Supabase控制台配置"
    echo "  2. 建议使用CDN加速访问"
    echo "  3. 可以配置Cloudflare等服务进行域名托管"
    echo ""
}

# 使用Edge Function托管（可选方案）
deploy_edge_function() {
    echo -e "${YELLOW}部署Edge Function作为静态托管...${NC}"
    
    # 创建Edge Function目录
    mkdir -p supabase/functions/static-host
    
    # 创建托管函数
    cat > supabase/functions/static-host/index.ts << 'EOF'
import { serve } from "https://deno.land/std@0.168.0/http/server.ts"

const BUCKET_NAME = "nofx-frontend"
const SUPABASE_URL = Deno.env.get("SUPABASE_URL") || ""

serve(async (req) => {
  const url = new URL(req.url)
  let path = url.pathname.replace("/static-host", "") || "/index.html"
  
  // 移除前导斜杠
  path = path.startsWith("/") ? path.slice(1) : path
  
  // 如果是目录，添加index.html
  if (path.endsWith("/")) {
    path += "index.html"
  }
  
  // 从Storage获取文件
  const fileUrl = `${SUPABASE_URL}/storage/v1/object/public/${BUCKET_NAME}/${path}`
  
  try {
    const response = await fetch(fileUrl)
    
    if (!response.ok) {
      return new Response("File not found", { status: 404 })
    }
    
    return new Response(response.body, {
      headers: response.headers,
    })
  } catch (error) {
    return new Response("Internal Server Error", { status: 500 })
  }
})
EOF
    
    # 部署Edge Function
    if supabase functions deploy static-host 2>/dev/null; then
        echo -e "${GREEN}✅ Edge Function部署成功${NC}"
        echo "Edge Function URL: $SUPABASE_URL/functions/v1/static-host"
    else
        echo -e "${YELLOW}⚠️  Edge Function部署失败，使用Storage直接访问${NC}"
    fi
}

# 主流程
main() {
    echo "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "后端URL: $BACKEND_URL"
    echo ""
    
    check_nodejs
    check_supabase_cli
    update_frontend_config
    install_dependencies
    build_frontend
    login_supabase
    link_project
    create_bucket
    upload_files
    generate_urls
    
    echo ""
    echo "结束时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "================================================"
}

# 运行主流程
main
