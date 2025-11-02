#!/bin/bash

# NOFX AI交易系统 - Railway部署脚本
# 项目ID: d9845ff4-c4a3-4c5d-8e9f-db95151d21bc

echo "🚀 开始部署NOFX AI交易系统到Railway..."

# 检查Railway CLI是否安装
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI未安装，请先安装："
    echo "npm install -g @railway/cli"
    exit 1
fi

# 登录Railway（如果未登录）
echo "🔐 检查Railway登录状态..."
if ! railway whoami &> /dev/null; then
    echo "请先登录Railway："
    railway login
fi

# 连接到指定项目
echo "🔗 连接到Railway项目..."
railway link d9845ff4-c4a3-4c5d-8e9f-db95151d21bc

# 设置环境变量（如果.env.railway存在）
if [ -f ".env.railway" ]; then
    echo "⚙️ 上传环境变量..."
    railway variables --environment production < .env.railway
else
    echo "⚠️ 未找到.env.railway文件，请手动在Railway控制台配置环境变量"
fi

# 部署应用
echo "📦 开始部署..."
railway up --detach

echo "✅ 部署完成！"
echo "🌐 您可以在Railway控制台查看部署状态："
echo "   https://railway.app/project/d9845ff4-c4a3-4c5d-8e9f-db95151d21bc"

# 显示服务URL（如果可用）
echo "🔍 获取服务URL..."
SERVICE_URL=$(railway domain 2>/dev/null)
if [ ! -z "$SERVICE_URL" ]; then
    echo "🎉 您的应用已部署到: $SERVICE_URL"
    echo "🏥 健康检查: $SERVICE_URL/health"
else
    echo "ℹ️ 服务URL将在部署完成后可用"
fi

echo ""
echo "📋 下一步："
echo "1. 在Railway控制台配置必需的环境变量（API密钥等）"
echo "2. 等待构建完成（通常需要2-5分钟）"
echo "3. 访问您的应用URL进行测试"
echo ""
echo "🆘 如需帮助，请查看 railway-deploy.md 文档"