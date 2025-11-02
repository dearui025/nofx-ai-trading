# 币安测试网API配置指南

## 当前问题
系统显示API权限错误：`Invalid API-key, IP, or permissions for action`

## 解决步骤

### 1. 重新创建币安测试网API密钥

1. 访问币安测试网：https://testnet.binance.vision/
2. 使用GitHub账号登录
3. 进入API管理页面
4. 删除现有的API密钥（如果有）
5. 创建新的API密钥

### 2. 配置API权限

创建API密钥时，确保启用以下权限：
- ✅ **Enable Reading** (读取权限)
- ✅ **Enable Spot & Margin Trading** (现货和保证金交易)
- ✅ **Enable Futures** (期货交易) - **这个最重要！**
- ❌ Enable Withdrawals (提现权限 - 测试网不需要)

### 3. IP白名单设置

- 可以设置为 `0.0.0.0/0` (允许所有IP)
- 或者添加您当前的公网IP地址

### 4. 更新配置文件

将新的API密钥更新到 `config.json` 文件中：

```json
{
  "binance_api_key": "您的新API密钥",
  "binance_secret_key": "您的新密钥",
  "binance_testnet": true
}
```

### 5. 测试网资金

确保您的测试网账户有足够的USDT余额：
1. 登录测试网
2. 进入钱包页面
3. 申请测试网USDT（通常可以免费获取）

## 常见问题

### Q: 为什么会出现权限错误？
A: 通常是因为API密钥没有启用期货交易权限，或者IP白名单配置不正确。

### Q: 测试网和主网的区别？
A: 测试网使用虚拟资金，不会产生真实的盈亏，适合测试交易策略。

### Q: 如何验证API配置是否正确？
A: 重启系统后，查看日志中是否还有API权限错误。

## 下一步

配置完成后：
1. 重启交易系统
2. 观察日志输出
3. 确认AI开始做出交易决策
4. 验证交易是否成功执行