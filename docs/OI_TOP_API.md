# OI Top API 配置说明

## 概述

OI Top API 用于获取持仓量增长排行榜数据，帮助系统识别市场热点和资金流向。该API是可选配置，如果未配置，系统将跳过OI Top数据获取。

## API 数据格式要求

### 请求方式
- **方法**: GET
- **超时时间**: 30秒
- **重试次数**: 最多3次

### 响应格式

API需要返回以下JSON格式的数据：

```json
{
  "success": true,
  "data": {
    "positions": [
      {
        "symbol": "BTCUSDT",
        "rank": 1,
        "current_oi": 1000000,
        "oi_delta": 50000,
        "oi_delta_percent": 5.0,
        "oi_delta_value": 2500000,
        "price_delta_percent": 2.1,
        "net_long": 600000,
        "net_short": 400000
      },
      {
        "symbol": "ETHUSDT",
        "rank": 2,
        "current_oi": 800000,
        "oi_delta": 40000,
        "oi_delta_percent": 5.3,
        "oi_delta_value": 2000000,
        "price_delta_percent": 1.8,
        "net_long": 480000,
        "net_short": 320000
      }
    ],
    "count": 20,
    "exchange": "binance",
    "time_range": "24h"
  }
}
```

### 字段说明

#### 根级别字段
- `success` (boolean): 请求是否成功
- `data` (object): 数据对象

#### data 对象字段
- `positions` (array): 持仓量排行数据数组
- `count` (number): 返回的记录数量
- `exchange` (string): 交易所名称
- `time_range` (string): 数据时间范围

#### positions 数组中每个对象的字段
- `symbol` (string): 交易对符号，如 "BTCUSDT"
- `rank` (number): 排名
- `current_oi` (number): 当前持仓量
- `oi_delta` (number): 持仓量变化量
- `oi_delta_percent` (number): 持仓量变化百分比
- `oi_delta_value` (number): 持仓量变化价值（USDT）
- `price_delta_percent` (number): 价格变化百分比
- `net_long` (number): 净多头持仓
- `net_short` (number): 净空头持仓

## 配置方式

### 1. 通过前端界面配置

1. 打开环境配置页面
2. 找到 "OI Top API URL" 配置项
3. 输入API地址，例如：`https://api.example.com/oi-top`
4. 点击保存配置

### 2. 通过配置文件配置

在 `config.json` 文件中设置：

```json
{
  "oi_top_api_url": "https://api.example.com/oi-top"
}
```

## 错误处理

系统具有完善的错误处理机制：

1. **API不可用**: 自动重试3次，每次间隔2秒
2. **网络超时**: 30秒超时后重试
3. **数据格式错误**: 记录错误日志并跳过
4. **API返回失败**: 尝试使用历史缓存数据
5. **缓存不可用**: 跳过OI Top数据，不影响主要功能

## 缓存机制

- **缓存位置**: `coin_pool_cache/oi_top_latest.json`
- **缓存时效**: 24小时
- **自动更新**: 成功获取数据后自动更新缓存
- **降级策略**: API失败时自动使用缓存数据

## 日志示例

### 成功获取数据
```
🔄 正在请求OI Top数据...
✓ 成功获取20个OI Top币种（时间范围: 24h）
```

### 未配置API
```
⚠️  未配置OI Top API URL，跳过OI Top数据获取
```

### API失败使用缓存
```
❌ 第1次请求OI Top失败: 请求超时
⚠️  OI Top API请求全部失败，尝试使用历史缓存数据...
✓ 使用历史OI Top缓存数据（共20个币种）
```

## 注意事项

1. **可选功能**: OI Top API是可选配置，不影响系统核心功能
2. **数据质量**: 确保API返回的数据格式严格符合要求
3. **性能考虑**: API响应时间建议控制在10秒以内
4. **频率限制**: 系统会根据需要调用API，请确保API支持合理的调用频率
5. **安全性**: 如果API需要认证，请在URL中包含必要的参数

## 示例API实现

如果您需要实现自己的OI Top API，可以参考以下Python Flask示例：

```python
from flask import Flask, jsonify
import random

app = Flask(__name__)

@app.route('/oi-top')
def get_oi_top():
    # 模拟数据
    positions = []
    symbols = ["BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"]
    
    for i, symbol in enumerate(symbols):
        positions.append({
            "symbol": symbol,
            "rank": i + 1,
            "current_oi": random.randint(500000, 2000000),
            "oi_delta": random.randint(-100000, 100000),
            "oi_delta_percent": round(random.uniform(-10, 10), 2),
            "oi_delta_value": random.randint(1000000, 5000000),
            "price_delta_percent": round(random.uniform(-5, 5), 2),
            "net_long": random.randint(300000, 1200000),
            "net_short": random.randint(200000, 800000)
        })
    
    return jsonify({
        "success": True,
        "data": {
            "positions": positions,
            "count": len(positions),
            "exchange": "binance",
            "time_range": "24h"
        }
    })

if __name__ == '__main__':
    app.run(debug=True)
```

## 技术支持

如果在配置或使用过程中遇到问题，请检查：

1. API URL是否正确且可访问
2. API返回的数据格式是否符合要求
3. 网络连接是否正常
4. 查看系统日志获取详细错误信息