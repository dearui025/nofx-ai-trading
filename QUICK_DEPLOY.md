# NOFX 快速部署指南

## 🚀 5分钟快速部署到Zeabur

### 前置要求

- Git
- Node.js (可选，用于本地测试)
- Go (可选，用于本地测试)

### 步骤1: 准备代码

```bash
# 克隆或下载项目代码
git clone <your-repository-url>
cd nofx
```

### 步骤2: 一键部署

**Windows用户:**
```powershell
.\deploy-zeabur.ps1
```

**Linux/macOS用户:**
```bash
chmod +x deploy-zeabur.sh
./deploy-zeabur.sh
```

### 步骤3: 配置环境变量

部署脚本会提示您配置以下关键环境变量：

#### 必需配置
```env
# 数据库
DATABASE_URL=postgresql://username:password@host:5432/database

# Binance API (必需)
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key

# JWT安全
JWT_SECRET=your_super_secret_jwt_key_min_32_chars
```

#### 可选配置
```env
# Hyperliquid (可选)
HYPERLIQUID_PRIVATE_KEY=your_hyperliquid_private_key
HYPERLIQUID_WALLET_ADDR=your_wallet_address

# AI模型 (可选)
QWEN_API_KEY=your_qwen_api_key
DEEPSEEK_API_KEY=your_deepseek_api_key

# 风控参数
MAX_DAILY_LOSS=5.0
MAX_DRAWDOWN=10.0
```

### 步骤4: 验证部署

```powershell
.\verify-deployment.ps1 -BackendUrl "https://your-backend-url" -FrontendUrl "https://your-frontend-url"
```

## 🎯 手动部署（详细步骤）

### 1. 安装Zeabur CLI

**Windows:**
```powershell
# 下载并安装Zeabur CLI
Invoke-WebRequest -Uri "https://zeabur.com/install.ps1" -UseBasicParsing | Invoke-Expression
```

**Linux/macOS:**
```bash
curl -sSL https://zeabur.com/install.sh | bash
```

### 2. 登录Zeabur

```bash
zeabur auth login
```

### 3. 创建项目

```bash
zeabur project create nofx-ai-trading
zeabur project use nofx-ai-trading
```

### 4. 部署数据库

```bash
zeabur service create postgres --type=prebuilt --image=postgres:15
```

在Zeabur控制台配置数据库环境变量：
- `POSTGRES_DB`: nofx
- `POSTGRES_USER`: nofx
- `POSTGRES_PASSWORD`: your_secure_password

### 5. 部署后端

```bash
zeabur service create nofx-backend --type=git
```

配置后端环境变量（在Zeabur控制台）：
```env
PORT=8080
GO_ENV=production
DATABASE_URL=postgresql://nofx:your_secure_password@postgres:5432/nofx?sslmode=disable
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key
JWT_SECRET=your_super_secret_jwt_key_min_32_chars
MAX_DAILY_LOSS=5.0
MAX_DRAWDOWN=10.0
```

部署后端：
```bash
zeabur service deploy nofx-backend
```

### 6. 部署前端

```bash
zeabur service create nofx-frontend --type=git --path=web
```

配置前端环境变量：
```env
VITE_API_URL=https://your-backend-domain.zeabur.app
VITE_WS_URL=wss://your-backend-domain.zeabur.app
VITE_APP_NAME=NOFX AI Trading System
```

部署前端：
```bash
zeabur service deploy nofx-frontend
```

### 7. 配置域名（可选）

在Zeabur控制台为每个服务配置自定义域名：
- 后端: `api.yourdomain.com`
- 前端: `yourdomain.com`

## 🔧 环境变量快速参考

### 核心配置
| 变量 | 描述 | 示例 |
|------|------|------|
| `DATABASE_URL` | 数据库连接 | `postgresql://user:pass@host:5432/db` |
| `BINANCE_API_KEY` | Binance API密钥 | `your_api_key` |
| `BINANCE_SECRET_KEY` | Binance Secret | `your_secret_key` |
| `JWT_SECRET` | JWT签名密钥 | `min_32_chars_secret_key` |

### 风控配置
| 变量 | 描述 | 默认值 |
|------|------|--------|
| `MAX_DAILY_LOSS` | 最大日损失% | 5.0 |
| `MAX_DRAWDOWN` | 最大回撤% | 10.0 |
| `STOP_TRADING_MINUTES` | 停止交易时间(分钟) | 120 |

### 交易配置
| 变量 | 描述 | 默认值 |
|------|------|--------|
| `BTC_ETH_LEVERAGE` | BTC/ETH杠杆 | 3 |
| `ALTCOIN_LEVERAGE` | 山寨币杠杆 | 2 |
| `DEFAULT_COINS` | 默认币种 | `BTCUSDT,ETHUSDT,SOLUSDT` |

## ✅ 部署检查清单

- [ ] Zeabur CLI已安装并登录
- [ ] 项目代码已上传到Git仓库
- [ ] 数据库服务已创建并配置
- [ ] 后端服务已部署并运行
- [ ] 前端服务已部署并运行
- [ ] 环境变量已正确配置
- [ ] API密钥已设置且有效
- [ ] 健康检查通过
- [ ] 域名已配置（可选）
- [ ] SSL证书已配置（可选）

## 🚨 常见问题快速解决

### 1. 部署失败
```bash
# 查看服务日志
zeabur service logs nofx-backend
zeabur service logs nofx-frontend

# 重新部署
zeabur service deploy nofx-backend
```

### 2. 数据库连接失败
- 检查 `DATABASE_URL` 格式
- 确认数据库服务状态
- 验证用户名密码

### 3. API密钥错误
- 确认密钥格式正确
- 检查权限设置
- 验证测试网/主网配置

### 4. 前端无法访问后端
- 检查 `VITE_API_URL` 配置
- 确认后端服务运行状态
- 验证CORS设置

## 📞 获取帮助

1. **查看日志**: `zeabur service logs <service-name>`
2. **检查状态**: `zeabur service list`
3. **重新部署**: `zeabur service deploy <service-name>`
4. **验证部署**: 运行 `verify-deployment.ps1` 脚本

## 🎉 部署成功！

部署完成后，您可以：

1. **访问前端**: `https://your-frontend-domain`
2. **测试API**: `https://your-backend-domain/health`
3. **查看指标**: `https://your-backend-domain/metrics`
4. **监控日志**: Zeabur控制台

恭喜！您的NOFX AI交易系统已成功部署到Zeabur平台！