# 🔑 Railway环境变量配置检查清单

## 📋 部署前必须配置的环境变量

### ✅ 核心必需变量（必须配置）

#### 🏦 交易所API配置
- [ ] `BINANCE_API_KEY` - 币安API密钥
- [ ] `BINANCE_SECRET_KEY` - 币安密钥
- [ ] `BINANCE_TESTNET` - 是否使用测试网（建议先设为 `true`）

#### 🤖 AI模型API（至少配置一个）
- [ ] `QWEN_API_KEY` - 通义千问API密钥
- [ ] `DEEPSEEK_API_KEY` - DeepSeek API密钥

#### 🔐 安全配置
- [ ] `JWT_SECRET` - JWT密钥（建议使用强随机字符串）

### 🔧 可选但推荐的变量

#### 🏦 其他交易所（可选）
- [ ] `HYPERLIQUID_PRIVATE_KEY` - Hyperliquid私钥（不含0x前缀）
- [ ] `HYPERLIQUID_WALLET_ADDR` - Hyperliquid钱包地址
- [ ] `HYPERLIQUID_TESTNET` - 是否使用测试网

- [ ] `ASTER_USER` - Aster主钱包地址
- [ ] `ASTER_SIGNER` - Aster API钱包地址
- [ ] `ASTER_PRIVATE_KEY` - Aster API钱包私钥（不含0x前缀）

#### 🛡️ 风控参数（有默认值，可选配置）
- [ ] `MAX_DAILY_LOSS` - 最大日损失百分比（默认：10.0）
- [ ] `MAX_DRAWDOWN` - 最大回撤百分比（默认：20.0）
- [ ] `STOP_TRADING_MINUTES` - 停止交易时间（默认：60分钟）

#### ⚖️ 杠杆配置（有默认值，可选配置）
- [ ] `BTC_ETH_LEVERAGE` - BTC/ETH杠杆倍数（默认：5）
- [ ] `ALTCOIN_LEVERAGE` - 山寨币杠杆倍数（默认：5）

#### 🪙 币种池配置（可选）
- [ ] `USE_DEFAULT_COINS` - 是否使用默认币种（默认：true）
- [ ] `DEFAULT_COINS` - 默认币种列表
- [ ] `COIN_POOL_API_URL` - 自定义币种池API
- [ ] `OI_TOP_API_URL` - 持仓量排行API

#### 🔧 自定义API配置（可选）
- [ ] `CUSTOM_API_URL` - 自定义AI API地址
- [ ] `CUSTOM_API_KEY` - 自定义AI API密钥
- [ ] `CUSTOM_MODEL_NAME` - 自定义模型名称

## 🚀 Railway控制台配置步骤

### 1. 进入Railway项目
1. 访问 [Railway控制台](https://railway.app/dashboard)
2. 找到项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
3. 点击进入项目

### 2. 配置环境变量
1. 在项目页面点击服务名称（通常是仓库名）
2. 切换到 **Variables** 标签页
3. 点击 **+ New Variable** 添加环境变量

### 3. 必需变量配置示例

```bash
# 币安API配置
BINANCE_API_KEY=your_binance_api_key_here
BINANCE_SECRET_KEY=your_binance_secret_key_here
BINANCE_TESTNET=true

# AI模型API（选择一个或多个）
QWEN_API_KEY=your_qwen_api_key_here
# 或者
DEEPSEEK_API_KEY=your_deepseek_api_key_here

# JWT安全密钥（生成一个强随机字符串）
JWT_SECRET=your_very_secure_jwt_secret_here_at_least_32_chars
```

### 4. 可选数据库配置
如果需要数据库：
1. 在Railway项目中点击 **+ New Service**
2. 选择 **PostgreSQL**
3. Railway会自动创建并配置 `DATABASE_URL`

### 5. 可选Redis配置
如果需要缓存：
1. 在Railway项目中点击 **+ New Service**
2. 选择 **Redis**
3. Railway会自动创建并配置 `REDIS_URL`

## ⚠️ 重要安全提醒

### 🔐 API密钥安全
- **绝不要**在代码中硬编码API密钥
- **绝不要**将API密钥提交到Git仓库
- 使用Railway的环境变量功能安全存储
- 定期轮换API密钥

### 🧪 测试网建议
- 首次部署建议使用测试网：
  - `BINANCE_TESTNET=true`
  - `HYPERLIQUID_TESTNET=true`
- 确认系统正常运行后再切换到主网

### 💰 风控设置
- 设置合理的 `MAX_DAILY_LOSS` 和 `MAX_DRAWDOWN`
- 从小额资金开始测试
- 监控交易日志和性能指标

## 📝 配置验证清单

部署前请确认：
- [ ] 所有必需的环境变量已配置
- [ ] API密钥格式正确（无多余空格或字符）
- [ ] 测试网设置正确
- [ ] 风控参数合理
- [ ] JWT密钥足够安全（至少32个字符）

## 🔍 配置后验证

配置完成后，可以通过以下方式验证：
1. 查看Railway部署日志
2. 访问健康检查端点：`https://your-app.railway.app/health`
3. 检查API连接状态
4. 验证交易功能（建议先在测试网）

---

**项目信息:**
- 项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
- 健康检查: `/health`
- 部署区域: `us-west1`