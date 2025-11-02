# 🚀 Railway部署详细步骤指南

## 📋 部署概览

您的NOFX AI交易系统已经准备好部署到Railway！以下是详细的部署步骤。

**项目信息：**
- 项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
- 仓库已链接 ✅
- 配置文件已准备 ✅

## 🎯 第一步：配置环境变量

### 1.1 进入Railway控制台
1. 访问 [Railway控制台](https://railway.app/dashboard)
2. 找到项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
3. 点击进入项目

### 1.2 配置必需的环境变量
在Railway控制台中，点击您的服务，然后切换到 **Variables** 标签页：

#### 🔑 必需配置（必须设置）
```bash
# 币安API配置
BINANCE_API_KEY=your_binance_api_key_here
BINANCE_SECRET_KEY=your_binance_secret_key_here
BINANCE_TESTNET=true

# AI模型API（至少配置一个）
QWEN_API_KEY=your_qwen_api_key_here
# 或者
DEEPSEEK_API_KEY=your_deepseek_api_key_here

# JWT安全密钥
JWT_SECRET=your_secure_jwt_secret_at_least_32_characters_long
```

#### ⚙️ 推荐配置（可选但建议）
```bash
# 风控参数
MAX_DAILY_LOSS=5.0
MAX_DRAWDOWN=10.0
STOP_TRADING_MINUTES=60

# 杠杆配置
BTC_ETH_LEVERAGE=3
ALTCOIN_LEVERAGE=3
```

## 🏗️ 第二步：触发部署

### 2.1 自动部署触发
Railway会在以下情况自动触发部署：
- 推送代码到连接的Git分支
- 修改环境变量
- 手动触发部署

### 2.2 手动触发部署
如果需要手动触发：
1. 在Railway控制台中找到您的服务
2. 点击 **Deploy** 按钮
3. 或者点击 **Redeploy** 重新部署

## 📊 第三步：监控部署过程

### 3.1 查看构建日志
1. 在Railway控制台中点击 **Deployments** 标签页
2. 点击最新的部署记录
3. 查看实时构建日志

### 3.2 预期的构建步骤
```
✅ 1. 检测到Dockerfile
✅ 2. 开始Docker构建
✅ 3. 下载Go依赖
✅ 4. 编译Go应用
✅ 5. 创建运行镜像
✅ 6. 启动容器
✅ 7. 健康检查通过
```

### 3.3 构建时间预期
- 首次构建：约3-5分钟
- 后续构建：约2-3分钟（有缓存）

## 🔍 第四步：验证部署

### 4.1 获取应用URL
部署成功后，Railway会提供一个公共URL：
- 格式：`https://your-app-name.railway.app`
- 在控制台的 **Settings** → **Domains** 中查看

### 4.2 健康检查验证
访问健康检查端点：
```
https://your-app.railway.app/health
```

预期响应：
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### 4.3 API端点验证
测试主要API端点：
```bash
# 获取交易员列表
GET https://your-app.railway.app/api/traders

# 获取性能数据
GET https://your-app.railway.app/api/performance?trader_id=binance_qwen_optimized

# 获取最新决策
GET https://your-app.railway.app/api/decisions/latest?trader_id=binance_qwen_optimized
```

## 🎛️ 第五步：配置域名（可选）

### 5.1 添加自定义域名
1. 在Railway控制台中点击 **Settings**
2. 找到 **Domains** 部分
3. 点击 **+ Add Domain**
4. 输入您的域名

### 5.2 配置DNS
在您的域名提供商处添加CNAME记录：
```
CNAME your-domain.com railway.app
```

## 📈 第六步：监控和维护

### 6.1 查看应用日志
```bash
# 在Railway控制台中查看实时日志
# 或使用Railway CLI
railway logs
```

### 6.2 监控指标
在Railway控制台中查看：
- CPU使用率
- 内存使用率
- 网络流量
- 响应时间

### 6.3 设置告警（推荐）
1. 在Railway控制台中设置资源使用告警
2. 配置邮件或Slack通知
3. 监控应用健康状态

## 🔄 第七步：更新部署

### 7.1 代码更新
```bash
# 推送代码更新会自动触发重新部署
git add .
git commit -m "Update application"
git push origin main
```

### 7.2 环境变量更新
1. 在Railway控制台中修改环境变量
2. 保存后会自动重启应用

### 7.3 回滚部署
如果需要回滚：
1. 在 **Deployments** 标签页中找到之前的部署
2. 点击 **Redeploy** 回滚到该版本

## 🚨 常见问题和解决方案

### 构建失败
**问题：** Docker构建失败
**解决：**
1. 检查Dockerfile语法
2. 确保go.mod和go.sum文件存在
3. 查看构建日志中的具体错误

### 启动失败
**问题：** 应用启动后立即崩溃
**解决：**
1. 检查环境变量配置
2. 确保API密钥格式正确
3. 查看应用日志

### 健康检查失败
**问题：** 健康检查端点返回错误
**解决：**
1. 确认端口配置正确
2. 检查防火墙设置
3. 验证健康检查路径

### API连接失败
**问题：** 无法连接到交易所API
**解决：**
1. 验证API密钥正确性
2. 检查网络连接
3. 确认API权限设置

## 📞 获取帮助

### Railway支持
- 文档：https://docs.railway.app/
- 社区：https://discord.gg/railway
- 支持：help@railway.app

### 项目支持
- 查看项目README.md
- 检查常见问题文档
- 联系项目维护者

---

## ✅ 部署成功检查清单

- [ ] 环境变量已配置
- [ ] 构建成功完成
- [ ] 应用启动正常
- [ ] 健康检查通过
- [ ] API端点响应正常
- [ ] 日志无错误信息
- [ ] 监控配置完成

**恭喜！您的NOFX AI交易系统已成功部署到Railway！** 🎉