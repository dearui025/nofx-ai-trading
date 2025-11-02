# NOFX AI交易系统 - Railway部署指南

## 📋 部署概述

本指南将帮助您将NOFX AI交易系统部署到Railway平台。

**项目信息：**
- 项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
- 服务名称: `api`
- 部署类型: Docker容器

## 🚀 快速部署步骤

### 1. 准备Railway项目

1. 登录 [Railway控制台](https://railway.app/dashboard)
2. 找到项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
3. 进入项目管理页面

### 2. 连接代码仓库

```bash
# 如果还没有Git仓库，先初始化
git init
git add .
git commit -m "Initial commit for Railway deployment"

# 推送到GitHub/GitLab等代码托管平台
git remote add origin <your-repo-url>
git push -u origin main
```

### 3. 在Railway中配置服务

1. 在Railway项目中点击 "New Service"
2. 选择 "GitHub Repo" 或 "GitLab Repo"
3. 选择您的NOFX项目仓库
4. Railway会自动检测到Dockerfile并开始构建

### 4. 配置环境变量

在Railway控制台的Variables标签页中添加以下环境变量：

#### 🔑 必需的API密钥
```bash
# 币安API（必需）
BINANCE_API_KEY=your_binance_api_key_here
BINANCE_SECRET_KEY=your_binance_secret_key_here

# AI模型API（至少配置一个）
QWEN_API_KEY=your_qwen_api_key_here
DEEPSEEK_API_KEY=your_deepseek_api_key_here

# JWT密钥（必需）
JWT_SECRET=your_secure_jwt_secret_here
```

#### 🏦 交易所配置（可选）
```bash
# Hyperliquid
HYPERLIQUID_PRIVATE_KEY=your_ethereum_private_key_without_0x
HYPERLIQUID_WALLET_ADDR=your_ethereum_wallet_address

# Aster
ASTER_USER=your_main_wallet_address
ASTER_SIGNER=your_api_wallet_address
ASTER_PRIVATE_KEY=your_api_wallet_private_key_without_0x
```

#### 🛡️ 风控参数（可选，有默认值）
```bash
MAX_DAILY_LOSS=10.0
MAX_DRAWDOWN=20.0
STOP_TRADING_MINUTES=60
BTC_ETH_LEVERAGE=5
ALTCOIN_LEVERAGE=5
```

### 5. 添加数据库服务（推荐）

1. 在Railway项目中点击 "New Service"
2. 选择 "PostgreSQL"
3. Railway会自动创建数据库并提供连接信息
4. 数据库URL会自动设置为 `DATABASE_URL` 环境变量

### 6. 部署和验证

1. 保存环境变量后，Railway会自动重新部署
2. 等待构建完成（通常需要2-5分钟）
3. 部署成功后，您会获得一个公共URL

## 🔧 高级配置

### 自定义域名

1. 在Railway控制台的Settings标签页
2. 点击 "Custom Domain"
3. 添加您的域名并配置DNS

### 监控和日志

1. 在Railway控制台查看实时日志
2. 使用 "Metrics" 标签页监控性能
3. 设置告警通知

### 扩展配置

```bash
# Redis缓存（可选）
# 添加Redis服务后会自动提供REDIS_URL

# 自定义API配置
CUSTOM_API_URL=https://api.openai.com/v1
CUSTOM_API_KEY=sk-your-custom-api-key
CUSTOM_MODEL_NAME=gpt-4o

# 外部API配置
COIN_POOL_API_URL=https://your-coin-pool-api.com
OI_TOP_API_URL=https://your-oi-top-api.com
```

## 🏗️ 本地测试Railway配置

在部署前，您可以本地测试Railway配置：

```bash
# 使用Railway环境变量文件
cp .env.railway .env

# 构建Docker镜像
docker build -t nofx-railway .

# 运行容器（使用Railway端口）
docker run -p 8080:8080 --env-file .env nofx-railway
```

## 📊 健康检查和监控

### 健康检查端点
- URL: `https://your-app.railway.app/health`
- 方法: GET
- 预期响应: 200 OK

### 监控指标
- CPU使用率
- 内存使用率
- 响应时间
- 错误率

## 🚨 故障排除

### 常见问题

1. **构建失败**
   - 检查Dockerfile语法
   - 确保所有依赖都在go.mod中

2. **启动失败**
   - 检查环境变量配置
   - 查看Railway日志获取详细错误信息

3. **API连接失败**
   - 验证API密钥是否正确
   - 检查网络连接和防火墙设置

4. **数据库连接失败**
   - 确保PostgreSQL服务已启动
   - 检查DATABASE_URL环境变量

### 日志查看

```bash
# 在Railway控制台查看实时日志
# 或使用Railway CLI
railway logs
```

## 🔄 更新部署

### 自动部署
- 推送代码到主分支会自动触发部署

### 手动部署
```bash
# 使用Railway CLI
railway up

# 或在控制台点击 "Deploy"
```

## 📞 支持和帮助

- Railway文档: https://docs.railway.app/
- NOFX项目文档: 查看项目README.md
- 技术支持: 联系项目维护者

## 🔐 安全注意事项

1. **API密钥安全**
   - 不要在代码中硬编码API密钥
   - 使用Railway的环境变量功能
   - 定期轮换API密钥

2. **网络安全**
   - 启用HTTPS（Railway默认提供）
   - 配置适当的CORS策略
   - 使用强JWT密钥

3. **访问控制**
   - 限制数据库访问权限
   - 使用Railway的团队功能管理访问

## 📈 性能优化

1. **资源配置**
   - 根据需要调整Railway服务规格
   - 监控资源使用情况

2. **缓存策略**
   - 添加Redis服务进行缓存
   - 优化数据库查询

3. **负载均衡**
   - 使用Railway的自动扩展功能
   - 配置健康检查

---

**部署完成后，您的NOFX AI交易系统将在Railway上稳定运行！** 🎉