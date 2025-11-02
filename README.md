# NOFX AI Trading System

NOFX是一个基于人工智能的量化交易系统，支持多个交易平台，提供自动化交易策略和风险管理功能。

## 🚀 快速开始

### 本地开发

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd nofx
   ```

2. **安装依赖**
   ```bash
   # 后端依赖
   go mod download
   
   # 前端依赖
   cd web
   npm install
   cd ..
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置必要的API密钥和数据库连接
   ```

4. **启动服务**
   ```bash
   # 启动后端
   go run cmd/main.go
   
   # 启动前端（新终端）
   cd web
   npm run dev
   ```

### Docker部署

1. **使用Docker Compose（推荐）**
   ```bash
   # 开发环境
   docker-compose up -d
   
   # 生产环境
   docker-compose -f docker-compose.prod.yml up -d
   ```

2. **单独构建镜像**
   ```bash
   # 构建后端镜像
   docker build -t nofx-backend .
   
   # 构建前端镜像
   docker build -t nofx-frontend ./web
   ```

## 🌐 Zeabur部署

### 自动部署（推荐）

使用提供的部署脚本进行一键部署：

**Linux/macOS:**
```bash
chmod +x deploy-zeabur.sh
./deploy-zeabur.sh
```

**Windows PowerShell:**
```powershell
.\deploy-zeabur.ps1
```

### 手动部署

1. **安装Zeabur CLI**
   ```bash
   # Linux/macOS
   curl -sSL https://zeabur.com/install.sh | bash
   
   # Windows
   # 下载并安装 Zeabur CLI
   ```

2. **登录Zeabur**
   ```bash
   zeabur auth login
   ```

3. **创建项目**
   ```bash
   zeabur project create nofx-ai-trading
   zeabur project use nofx-ai-trading
   ```

4. **配置环境变量**
   
   参考 `.env.zeabur` 文件和 `ZEABUR_ENV_GUIDE.md` 配置以下关键环境变量：
   
   - **数据库配置**
     - `DATABASE_URL`
     - `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
   
   - **交易平台API**
     - `BINANCE_API_KEY`, `BINANCE_SECRET_KEY`
     - `HYPERLIQUID_PRIVATE_KEY`, `HYPERLIQUID_WALLET_ADDR`
     - `ASTER_USER`, `ASTER_SIGNER`, `ASTER_PRIVATE_KEY`
   
   - **AI模型API**
     - `QWEN_API_KEY`
     - `DEEPSEEK_API_KEY`
   
   - **安全配置**
     - `JWT_SECRET`

5. **部署服务**
   ```bash
   # 部署后端
   zeabur service create nofx-backend --type=git
   zeabur service deploy nofx-backend
   
   # 部署前端
   zeabur service create nofx-frontend --type=git --path=web
   zeabur service deploy nofx-frontend
   
   # 部署数据库
   zeabur service create postgres --type=prebuilt --image=postgres:15
   ```

6. **配置域名**
   
   在Zeabur控制台配置自定义域名：
   - 后端API: `api.yourdomain.com`
   - 前端: `yourdomain.com`

## 🔧 配置说明

### 环境变量

| 变量名 | 描述 | 必需 | 默认值 |
|--------|------|------|--------|
| `PORT` | 服务端口 | 否 | 8080 |
| `GO_ENV` | 运行环境 | 否 | development |
| `DATABASE_URL` | 数据库连接字符串 | 是 | - |
| `BINANCE_API_KEY` | Binance API密钥 | 是 | - |
| `BINANCE_SECRET_KEY` | Binance Secret密钥 | 是 | - |
| `HYPERLIQUID_PRIVATE_KEY` | Hyperliquid私钥 | 否 | - |
| `QWEN_API_KEY` | 通义千问API密钥 | 否 | - |
| `DEEPSEEK_API_KEY` | DeepSeek API密钥 | 否 | - |
| `JWT_SECRET` | JWT签名密钥 | 是 | - |
| `MAX_DAILY_LOSS` | 最大日损失百分比 | 否 | 5.0 |
| `MAX_DRAWDOWN` | 最大回撤百分比 | 否 | 10.0 |

### 风险控制参数

- `MAX_DAILY_LOSS`: 最大日损失百分比（默认5%）
- `MAX_DRAWDOWN`: 最大回撤百分比（默认10%）
- `STOP_TRADING_MINUTES`: 触发风控后停止交易时间（分钟）
- `BTC_ETH_LEVERAGE`: BTC/ETH杠杆倍数
- `ALTCOIN_LEVERAGE`: 山寨币杠杆倍数

### 币种池配置

- `USE_DEFAULT_COINS`: 是否使用默认币种池
- `DEFAULT_COINS`: 默认交易币种（逗号分隔）

## 📊 监控和日志

### 健康检查

- 后端健康检查: `GET /health`
- 数据库连接检查: `GET /api/health/database`
- 前端健康检查: `GET /health`

### Prometheus指标

系统提供Prometheus格式的监控指标：
- URL: `/metrics`
- 端口: 9090

### 日志配置

- 日志级别: `LOG_LEVEL` (debug, info, warn, error)
- 日志输出: 控制台 + 文件
- 日志轮转: 自动按大小和时间轮转

## 🧪 测试和验证

### 部署验证

使用提供的验证脚本检查部署状态：

```powershell
.\verify-deployment.ps1 -BackendUrl "https://api.yourdomain.com" -FrontendUrl "https://yourdomain.com"
```

验证内容包括：
- 基础连接测试
- API端点测试
- WebSocket连接测试
- 数据库连接测试
- 认证系统测试
- 交易功能测试
- 性能指标测试
- 安全配置测试

### 本地测试

```bash
# 运行后端测试
go test ./...

# 运行前端测试
cd web
npm test
```

## 🔒 安全最佳实践

1. **API密钥管理**
   - 使用环境变量存储敏感信息
   - 定期轮换API密钥
   - 限制API密钥权限

2. **网络安全**
   - 启用HTTPS
   - 配置安全头
   - 使用强JWT密钥

3. **访问控制**
   - 实施用户认证
   - 配置角色权限
   - 监控异常访问

4. **数据保护**
   - 数据库连接加密
   - 敏感数据脱敏
   - 定期备份

## 📚 API文档

### 认证端点

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/logout` - 用户登出
- `GET /api/auth/verify` - 验证JWT令牌

### 交易端点

- `GET /api/trading/status` - 获取交易状态
- `GET /api/trading/positions` - 获取持仓信息
- `GET /api/trading/orders` - 获取订单历史
- `POST /api/trading/order` - 下单
- `DELETE /api/trading/order/:id` - 取消订单

### 市场数据端点

- `GET /api/market/data/:symbol` - 获取市场数据
- `GET /api/market/klines/:symbol` - 获取K线数据
- `GET /api/market/ticker/:symbol` - 获取价格信息

### 风控端点

- `GET /api/risk/status` - 获取风控状态
- `GET /api/risk/metrics` - 获取风控指标
- `POST /api/risk/config` - 更新风控配置

## 🛠️ 故障排除

### 常见问题

1. **编译错误**
   ```bash
   # 清理模块缓存
   go clean -modcache
   go mod download
   ```

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接字符串
   - 确认网络连通性

3. **API密钥错误**
   - 验证密钥格式
   - 检查权限设置
   - 确认测试网/主网配置

4. **前端构建失败**
   ```bash
   cd web
   rm -rf node_modules package-lock.json
   npm install
   npm run build
   ```

### 日志查看

```bash
# Zeabur日志
zeabur service logs nofx-backend
zeabur service logs nofx-frontend

# Docker日志
docker logs nofx-backend
docker logs nofx-frontend

# 本地日志
tail -f logs/app.log
```

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如果您遇到问题或需要帮助：

1. 查看 [故障排除](#故障排除) 部分
2. 搜索现有的 [Issues](../../issues)
3. 创建新的 [Issue](../../issues/new)
4. 联系开发团队

## 🔄 更新日志

### v1.0.0
- 初始版本发布
- 支持Binance、Hyperliquid、Aster交易平台
- 集成AI模型进行策略分析
- 完整的风险管理系统
- Web界面和API接口
- Zeabur一键部署支持