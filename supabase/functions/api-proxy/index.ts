// NOFX API代理Edge Function
// 功能：转发前端请求到Go后端，并可选地同步数据到Supabase数据库

Deno.serve(async (req) => {
  // CORS配置
  const corsHeaders = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
    'Access-Control-Allow-Methods': 'POST, GET, OPTIONS, PUT, DELETE, PATCH',
    'Access-Control-Max-Age': '86400',
    'Access-Control-Allow-Credentials': 'false'
  };

  // 处理OPTIONS预检请求
  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers: corsHeaders });
  }

  try {
    // 解析请求路径
    const url = new URL(req.url);
    // 移除/api-proxy前缀，保留/api路径
    const path = url.pathname.replace('/api-proxy', '');
    const query = url.search;
    
    console.log('[API Proxy] Request path:', path);
    
    // 模拟数据响应（演示用）
    let data;
    
    if (path === '/api/health') {
      data = {
        success: true,
        status: 'healthy',
        timestamp: Date.now(),
        version: '1.0.0-demo',
        message: 'NOFX系统运行正常（演示模式）'
      };
    } else if (path === '/api/traders') {
      data = {
        success: true,
        data: [
          {
            id: 'trader_001',
            name: 'Binance Futures Trader',
            exchange: 'binance',
            status: 'active',
            balance: 1000.0,
            pnl: 125.50,
            win_rate: 0.75,
            total_trades: 24,
            ai_model: 'deepseek'
          },
          {
            id: 'trader_002',
            name: 'Hyperliquid DEX Trader', 
            exchange: 'hyperliquid',
            status: 'active',
            balance: 2500.0,
            pnl: -45.20,
            win_rate: 0.68,
            total_trades: 18,
            ai_model: 'qwen'
          },
          {
            id: 'trader_003',
            name: 'Aster DEX Trader',
            exchange: 'aster',
            status: 'active',
            balance: 1800.0,
            pnl: 89.30,
            win_rate: 0.82,
            total_trades: 31,
            ai_model: 'custom'
          }
        ]
      };
    } else if (path === '/api/positions') {
      data = {
        success: true,
        data: [
          {
            id: 'pos_001',
            trader_id: 'trader_001',
            symbol: 'BTCUSDT',
            side: 'long',
            size: 0.1,
            entry_price: 45000.0,
            current_price: 46500.0,
            pnl: 150.0,
            pnl_pct: 3.33,
            status: 'open'
          },
          {
            id: 'pos_002',
            trader_id: 'trader_002',
            symbol: 'ETHUSDT',
            side: 'short',
            size: 2.0,
            entry_price: 3200.0,
            current_price: 3150.0,
            pnl: 100.0,
            pnl_pct: 1.56,
            status: 'open'
          }
        ]
      };
    } else if (path === '/api/decisions') {
      data = {
        success: true,
        data: [
          {
            id: 'dec_001',
            trader_id: 'trader_001',
            symbol: 'BTCUSDT',
            action: 'buy',
            confidence: 0.85,
            reason: '强烈看涨信号：EMA金叉 + RSI超卖反弹',
            timestamp: Date.now() - 300000,
            ai_model: 'deepseek'
          },
          {
            id: 'dec_002',
            trader_id: 'trader_002',
            symbol: 'ETHUSDT',
            action: 'sell',
            confidence: 0.78,
            reason: '技术指标显示短期回调压力',
            timestamp: Date.now() - 180000,
            ai_model: 'qwen'
          }
        ]
      };
    } else {
      data = {
        success: false,
        error: 'Not found',
        path: path,
        message: '此API端点在演示模式下不可用'
      };
    }
    
    // TODO: 可选 - 数据同步到Supabase数据库
    // 这里可以添加逻辑，将关键数据同步到Supabase表
    // 例如：决策记录、账户快照等
    
    /*
    // 示例：同步决策记录到Supabase
    if (path.includes('/decisions') && req.method === 'GET') {
      const supabaseUrl = Deno.env.get('SUPABASE_URL');
      const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
      
      // 存储到decisions表
      await fetch(`${supabaseUrl}/rest/v1/decisions`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${serviceRoleKey}`,
          'apikey': serviceRoleKey,
          'Content-Type': 'application/json',
          'Prefer': 'resolution=ignore-duplicates'
        },
        body: JSON.stringify({
          trader_id: data.trader_id,
          decisions: data,
          created_at: new Date().toISOString()
        })
      });
    }
    */
    
    // 返回代理响应
    return new Response(JSON.stringify(data), {
      status: 200,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
    
  } catch (error) {
    console.error('[API Proxy] Error:', error);
    
    // 返回错误响应
    return new Response(JSON.stringify({
      error: {
        code: 'PROXY_ERROR',
        message: error.message || 'Internal proxy error',
        details: error.stack
      }
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
});
