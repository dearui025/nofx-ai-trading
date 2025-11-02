#!/bin/bash

# NOFX 项目状态检查和启动脚本
# 由于当前环境限制，此脚本主要用于状态展示和部署指导

echo "🤖 NOFX AI Trading System - 状态检查"
echo "======================================"

# 检查项目文件
echo ""
echo "📁 项目文件检查:"
if [ -f "config.json" ]; then
    echo "✅ config.json - 配置文件已存在"
else
    echo "❌ config.json - 配置文件不存在"
fi

if [ -f ".env" ]; then
    echo "✅ .env - 环境变量文件已存在"
else
    echo "❌ .env - 环境变量文件不存在"
fi

if [ -f "docker-compose.yml" ]; then
    echo "✅ docker-compose.yml - Docker部署文件已存在"
else
    echo "❌ docker-compose.yml - Docker部署文件不存在"
fi

if [ -f "start.sh" ]; then
    echo "✅ start.sh - Docker启动脚本已存在"
else
    echo "❌ start.sh - Docker启动脚本不存在"
fi

if [ -f "pm2.sh" ]; then
    echo "✅ pm2.sh - PM2启动脚本已存在"
else
    echo "❌ pm2.sh - PM2启动脚本不存在"
fi

# 检查环境工具
echo ""
echo "🔧 环境工具检查:"

if command -v docker &> /dev/null; then
    echo "✅ Docker - 已安装"
else
    echo "❌ Docker - 未安装"
fi

if command -v docker compose &> /dev/null; then
    echo "✅ Docker Compose - 已安装"
elif command -v docker-compose &> /dev/null; then
    echo "✅ Docker Compose - 已安装 (旧版本)"
else
    echo "❌ Docker Compose - 未安装"
fi

if command -v go &> /dev/null; then
    echo "✅ Go - 已安装"
else
    echo "❌ Go - 未安装"
fi

if command -v node &> /dev/null; then
    echo "✅ Node.js - 已安装"
else
    echo "❌ Node.js - 未安装"
fi

if command -v pm2 &> /dev/null; then
    echo "✅ PM2 - 已安装"
else
    echo "❌ PM2 - 未安装"
fi

# 显示部署状态
echo ""
echo "📊 部署状态总结:"
echo "=================="

# 计算已完成的步骤
completed_steps=0
total_steps=5

[ -f "config.json" ] && ((completed_steps++))
[ -f ".env" ] && ((completed_steps++))
[ -f "docker-compose.yml" ] && ((completed_steps++))
[ -f "start.sh" ] && ((completed_steps++))
[ -f "pm2.sh" ] && ((completed_steps++))

echo "项目准备进度: $completed_steps/$total_steps 步骤完成"

if [ $completed_steps -eq $total_steps ]; then
    echo "✅ 项目文件准备完成！"
else
    echo "⚠️  部分项目文件缺失"
fi

# 显示后续步骤
echo ""
echo "🚀 后续部署步骤:"
echo "=================="
echo "1. 安装必要的运行环境："
echo "   - Docker + Docker Compose (推荐)"
echo "   - 或 Go + Node.js + PM2"
echo ""
echo "2. 配置 API 密钥："
echo "   - 编辑 config.json 填入真实 API 密钥"
echo "   - 编辑 .env 填入环境变量"
echo ""
echo "3. 选择部署方式："
echo "   Docker 部署: ./start.sh start --build"
echo "   PM2 部署: ./pm2.sh start"
echo ""
echo "4. 访问服务："
echo "   Web界面: http://localhost:3000"
echo "   API接口: http://localhost:8080"

# 显示项目信息
echo ""
echo "📋 项目信息:"
echo "============="
echo "项目名称: NOFX AI Trading System"
echo "项目类型: Go + React 全栈应用"
echo "主要功能: AI驱动的加密货币期货自动交易"
echo "支持交易所: Binance, Hyperliquid, Aster DEX"
echo "支持AI模型: DeepSeek, Qwen, 自定义API"

echo ""
echo "📖 详细部署指南请查看: DEPLOYMENT_STATUS.md"
echo "============================================="