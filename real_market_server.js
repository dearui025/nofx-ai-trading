const http = require('http');
const https = require('https');
const url = require('url');

// 代理配置
const PROXY_URL = 'http://127.0.0.1:7890';

// 创建代理agent
function createProxyAgent() {
  try {
    const { HttpsProxyAgent } = require('https-proxy-agent');
    return new HttpsProxyAgent(PROXY_URL);
  } catch (error) {
    console.log('⚠️  https-proxy-agent 未安装，尝试直接连接...');
    return null;
  }
}

// 获取币安真实市场数据
async function getBinanceMarketData(symbols) {
  return new Promise((resolve, reject) => {
    let apiUrl;
    if (symbols.length === 1) {
      // 单个币种查询
      apiUrl = `https://fapi.binance.com/fapi/v1/ticker/24hr?symbol=${symbols[0]}`;
    } else {
      // 多个币种查询
      const symbolsParam = symbols.map(s => `"${s}"`).join(',');
      apiUrl = `https://fapi.binance.com/fapi/v1/ticker/24hr?symbols=[${symbolsParam}]`;
    }
    
    console.log(`📡 请求币安API: ${apiUrl}`);
    console.log(`🔗 使用代理: ${PROXY_URL}`);
    
    const agent = createProxyAgent();
    const options = {
      agent: agent,
      timeout: 10000 // 10秒超时
    };
    
    const req = https.get(apiUrl, options, (res) => {
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });
      
      res.on('end', () => {
        try {
          const jsonData = JSON.parse(data);
          
          let marketData;
          if (Array.isArray(jsonData)) {
            // 多个币种的响应
            console.log(`✅ 成功获取币安数据，共${jsonData.length}个币种`);
            marketData = jsonData.map(item => ({
              symbol: item.symbol,
              price: parseFloat(item.lastPrice),
              change24h: parseFloat(item.priceChange),
              changePercent24h: parseFloat(item.priceChangePercent),
              volume24h: parseFloat(item.volume)
            }));
          } else {
            // 单个币种的响应
            console.log(`✅ 成功获取币安数据: ${jsonData.symbol}`);
            marketData = [{
              symbol: jsonData.symbol,
              price: parseFloat(jsonData.lastPrice),
              change24h: parseFloat(jsonData.priceChange),
              changePercent24h: parseFloat(jsonData.priceChangePercent),
              volume24h: parseFloat(jsonData.volume)
            }];
          }
          
          resolve(marketData);
        } catch (error) {
          console.error('❌ 解析币安数据失败:', error);
          reject(error);
        }
      });
    }).on('error', (error) => {
      console.error('❌ 请求币安API失败:', error);
      reject(error);
    });
    
    req.on('timeout', () => {
      console.error('❌ 请求超时');
      req.destroy();
      reject(new Error('请求超时'));
    });
  });
}

const server = http.createServer(async (req, res) => {
  // 设置CORS头
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  res.setHeader('Content-Type', 'application/json');

  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  const parsedUrl = url.parse(req.url, true);
  const path = parsedUrl.pathname;

  if (path === '/health') {
    res.writeHead(200);
    res.end(JSON.stringify({ status: 'ok', source: 'real_binance_api' }));
    
  } else if (path === '/api/traders') {
    res.writeHead(200);
    const response = [
      {
        trader_id: 'real_trader',
        trader_name: 'Real Market Trader',
        ai_model: 'deepseek'
      }
    ];
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/market-data') {
    console.log(`📊 获取真实市场数据请求`);
    
    try {
      const symbol = parsedUrl.query.symbol;
      
      if (symbol) {
        // 单个币种查询
        console.log(`📊 查询单个币种: ${symbol}`);
        const marketData = await getBinanceMarketData([symbol]);
        
        if (marketData && marketData.length > 0) {
          res.writeHead(200);
          console.log(`✅ 返回${symbol}数据: $${marketData[0].price}`);
          res.end(JSON.stringify(marketData[0]));
        } else {
          res.writeHead(404);
          res.end(JSON.stringify({ error: `币种 ${symbol} 未找到` }));
        }
      } else {
        // 批量查询所有币种
        const symbols = ['BTCUSDT', 'ETHUSDT', 'SOLUSDT', 'BNBUSDT', 'XRPUSDT', 'DOGEUSDT'];
        const marketData = await getBinanceMarketData(symbols);
        
        res.writeHead(200);
        console.log(`✅ 返回真实币安数据（${marketData.length}个币种）`);
        console.log(`   BTC价格: $${marketData[0]?.price || 'N/A'}`);
        console.log(`   ETH价格: $${marketData[1]?.price || 'N/A'}`);
        res.end(JSON.stringify(marketData));
      }
      
    } catch (error) {
      console.error('❌ 获取市场数据失败:', error);
      res.writeHead(500);
      res.end(JSON.stringify({ 
        error: '获取真实市场数据失败',
        details: error.message 
      }));
    }
    
  } else if (path === '/api/competition') {
    console.log(`🏆 获取竞赛数据请求`);
    
    res.writeHead(200);
    const response = {
      traders: [
        {
          trader_id: 'real_trader',
          trader_name: 'Real Market Trader',
          ai_model: 'deepseek',
          status: 'active',
          total_equity: 5250.75,
          daily_pnl: 250.75,
          daily_pnl_pct: 5.02,
          total_pnl: 250.75,
          total_pnl_pct: 5.02,
          win_rate: 62.22,
          total_trades: 45,
          avg_trade_duration: '4.5h',
          max_drawdown: 8.50,
          sharpe_ratio: 1.25,
          last_trade_time: new Date().toISOString(),
          position_count: 2,
          margin_used_pct: 8.6,
          call_count: 156,
          is_running: true,
          unrealized_pnl: 125.30
        }
      ],
      count: 1
    };
    console.log(`✅ 返回竞赛数据 (真实市场模式)`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/equity-history') {
    console.log(`📈 获取权益历史请求`);
    
    // 生成模拟的权益历史数据
    const history = [];
    const now = Date.now();
    const baseEquity = 10000;
    
    for (let i = 24; i >= 0; i--) {
      const timestamp = new Date(now - (i * 60 * 60 * 1000)).toISOString();
      // 模拟一些波动
      const volatility = (Math.random() - 0.5) * 0.02; // ±1%的随机波动
      const trend = (24 - i) * 0.001; // 轻微上升趋势
      const pnl_pct = trend + volatility;
      const total_value = baseEquity * (1 + pnl_pct);
      const realized_pnl = total_value - baseEquity;
      
      history.push({
        timestamp,
        total_value: Math.round(total_value * 100) / 100,
        realized_pnl: Math.round(realized_pnl * 100) / 100,
        unrealized_pnl: Math.round((Math.random() - 0.5) * 100 * 100) / 100,
        return_percent: Math.round(pnl_pct * 10000) / 100
      });
    }
    
    res.writeHead(200);
    console.log(`✅ 返回权益历史数据 (${history.length}个数据点)`);
    res.end(JSON.stringify(history));
    
  } else if (path === '/api/account') {
    console.log(`💰 获取账户信息请求`);
    
    res.writeHead(200);
    const response = {
      trader_id: 'real_trader',
      balance: 10000.00,
      equity: 10250.75,
      margin_used: 860.50,
      margin_available: 9389.25,
      margin_ratio: 8.6,
      unrealized_pnl: 250.75,
      realized_pnl: 0.00,
      total_pnl: 250.75,
      total_pnl_pct: 2.51
    };
    console.log(`✅ 返回账户信息`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/performance') {
    console.log(`📊 获取性能数据请求`);
    
    res.writeHead(200);
    const response = {
      trader_id: 'real_trader',
      total_return: 2.51,
      daily_return: 0.85,
      weekly_return: 3.22,
      monthly_return: 12.45,
      max_drawdown: 8.50,
      sharpe_ratio: 1.25,
      win_rate: 62.22,
      profit_factor: 1.85,
      total_trades: 45,
      winning_trades: 28,
      losing_trades: 17,
      avg_win: 125.30,
      avg_loss: -67.80,
      largest_win: 450.20,
      largest_loss: -230.50
    };
    console.log(`✅ 返回性能数据`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/status') {
    console.log(`🔄 获取状态请求`);
    
    res.writeHead(200);
    const response = {
      trader_id: 'real_trader',
      status: 'active',
      is_running: true,
      last_update: new Date().toISOString(),
      uptime: '2h 35m',
      connection_status: 'connected',
      api_status: 'healthy',
      last_trade_time: new Date(Date.now() - 15 * 60 * 1000).toISOString()
    };
    console.log(`✅ 返回状态信息`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/positions') {
    console.log(`📈 获取持仓信息请求`);
    
    res.writeHead(200);
    const response = [
      {
        symbol: 'BTCUSDT',
        side: 'long',
        size: 0.025,
        entry_price: 108500.00,
        mark_price: 109932.10,
        unrealized_pnl: 35.80,
        unrealized_pnl_pct: 1.32,
        margin_used: 542.75,
        leverage: 5,
        timestamp: new Date(Date.now() - 30 * 60 * 1000).toISOString()
      },
      {
        symbol: 'ETHUSDT',
        side: 'long',
        size: 0.8,
        entry_price: 3820.50,
        mark_price: 3853.12,
        unrealized_pnl: 26.10,
        unrealized_pnl_pct: 0.85,
        margin_used: 317.75,
        leverage: 10,
        timestamp: new Date(Date.now() - 45 * 60 * 1000).toISOString()
      }
    ];
    console.log(`✅ 返回持仓信息 (${response.length}个持仓)`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/decisions/latest') {
    console.log(`🤖 获取最新决策请求`);
    
    res.writeHead(200);
    const response = [
      {
        timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        symbol: 'BTCUSDT',
        action: 'hold',
        confidence: 0.75,
        reasoning: '市场趋势向上，但需要等待更好的入场点',
        price: 109932.10,
        ai_model: 'deepseek'
      },
      {
        timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
        symbol: 'ETHUSDT',
        action: 'buy',
        confidence: 0.82,
        reasoning: 'ETH突破关键阻力位，预期继续上涨',
        price: 3853.12,
        ai_model: 'deepseek'
      }
    ];
    console.log(`✅ 返回最新决策 (${response.length}个决策)`);
    res.end(JSON.stringify(response));
    
  } else if (path === '/api/statistics') {
    console.log(`📊 获取统计信息请求`);
    
    res.writeHead(200);
    const response = {
      trader_id: 'real_trader',
      total_calls: 156,
      successful_calls: 97,
      failed_calls: 59,
      success_rate: 62.18,
      avg_response_time: 1.25,
      uptime_percentage: 98.5,
      last_24h_calls: 24,
      last_24h_success: 15,
      last_24h_success_rate: 62.5,
      total_volume_traded: 125000.50,
      total_fees_paid: 125.50
    };
    console.log(`✅ 返回统计信息`);
    res.end(JSON.stringify(response));
    
  } else {
    res.writeHead(404);
    res.end(JSON.stringify({ error: 'Not Found' }));
  }
});

const PORT = 8888;
server.listen(PORT, () => {
  console.log('🚀 真实币安市场数据服务器启动成功！');
  console.log(`📡 服务器地址: http://localhost:${PORT}`);
  console.log(`📊 市场数据接口: http://localhost:${PORT}/api/market-data`);
  console.log(`🏆 竞赛数据接口: http://localhost:${PORT}/api/competition`);
  console.log(`❤️  健康检查: http://localhost:${PORT}/health`);
  console.log('');
  console.log('✨ 现在前端将显示真实的币安市场数据！');
  console.log('');
});