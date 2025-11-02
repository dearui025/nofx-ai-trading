// NOFX数据同步Edge Function (简化版)
// 功能：演示数据同步到Supabase数据库

Deno.serve(async (req) => {
  const corsHeaders = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
    'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
  };

  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers: corsHeaders });
  }

  try {
    console.log('[Data Sync] Starting data synchronization...');
    
    // 获取环境变量
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
    
    if (!supabaseUrl || !serviceRoleKey) {
      throw new Error('Missing Supabase configuration');
    }
    
    // 模拟traders数据（演示用）
    const traders = [
      {
        id: 'trader_001',
        name: 'Binance Futures Trader',
        exchange: 'binance',
        status: 'active',
        ai_model: 'deepseek'
      },
      {
        id: 'trader_002',
        name: 'Hyperliquid DEX Trader',
        exchange: 'hyperliquid', 
        status: 'active',
        ai_model: 'qwen'
      }
    ];
    
    console.log(`[Data Sync] Found ${traders.length} traders (demo data)`);
    
    const syncResults = {
      traders: 0,
      decisions: 0,
      accounts: 0,
      positions: 0,
      errors: []
    };
    
    // 同步每个trader的数据
    for (const trader of traders) {
      try {
        const traderId = trader.id;
        console.log(`[Data Sync] Syncing trader: ${traderId}`);
        
        // 1. 同步Trader信息
        const traderUpsertResponse = await fetch(`${supabaseUrl}/rest/v1/traders`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json',
            'Prefer': 'resolution=merge-duplicates'
          },
          body: JSON.stringify({
            trader_id: traderId,
            name: trader.name || traderId,
            ai_model: trader.ai_model || 'unknown',
            exchange: trader.exchange || 'unknown',
            config: {},
            status: trader.status || 'active',
            updated_at: new Date().toISOString()
          })
        });
        
        if (traderUpsertResponse.ok) {
          syncResults.traders++;
          console.log(`[Data Sync] Trader ${traderId} synced successfully`);
        }
        
        // 2. 模拟账户数据并同步
        const accountData = {
          total_equity: Math.random() * 5000 + 1000,
          available_balance: Math.random() * 3000 + 500,
          total_pnl: Math.random() * 200 - 100,
          total_pnl_pct: Math.random() * 10 - 5,
          position_count: Math.floor(Math.random() * 5) + 1,
          margin_used: Math.random() * 1000,
          margin_used_pct: Math.random() * 50,
          realized_pnl: Math.random() * 150 - 75,
          unrealized_pnl: Math.random() * 100 - 50
        };
        
        const accountInsertResponse = await fetch(`${supabaseUrl}/rest/v1/account_history`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            trader_id: traderId,
            total_equity: accountData.total_equity,
            available_balance: accountData.available_balance,
            total_pnl: accountData.total_pnl,
            total_pnl_pct: accountData.total_pnl_pct,
            position_count: accountData.position_count,
            margin_used: accountData.margin_used,
            margin_used_pct: accountData.margin_used_pct,
            realized_pnl: accountData.realized_pnl,
            unrealized_pnl: accountData.unrealized_pnl,
            created_at: new Date().toISOString()
          })
        });
        
        if (accountInsertResponse.ok) {
          syncResults.accounts++;
          console.log(`[Data Sync] Account data for ${traderId} synced successfully`);
        }
        
        // 3. 模拟决策数据并同步
        const decisionData = {
          cycle_number: Math.floor(Math.random() * 100),
          input_prompt: `分析${trader.name}的当前市场状况`,
          cot_trace: 'Chain of thought reasoning for trading decision',
          decisions: {
            action: Math.random() > 0.5 ? 'buy' : 'sell',
            confidence: Math.random() * 0.3 + 0.7,
            reasoning: '基于技术指标和市场分析的建议'
          },
          execution_log: {
            executed: true,
            timestamp: new Date().toISOString(),
            result: 'success'
          },
          success: true,
          error_message: null,
          account_snapshot: accountData,
          positions_snapshot: [],
          market_data_snapshot: {
            btc_price: 46500 + Math.random() * 1000,
            eth_price: 3200 + Math.random() * 200,
            market_sentiment: 'bullish'
          },
          created_at: new Date().toISOString()
        };
        
        const decisionInsertResponse = await fetch(`${supabaseUrl}/rest/v1/decisions`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json',
            'Prefer': 'resolution=ignore-duplicates'
          },
          body: JSON.stringify({
            trader_id: traderId,
            cycle_number: decisionData.cycle_number,
            input_prompt: decisionData.input_prompt,
            cot_trace: decisionData.cot_trace,
            decisions: decisionData.decisions,
            execution_log: decisionData.execution_log,
            success: decisionData.success,
            error_message: decisionData.error_message,
            account_snapshot: decisionData.account_snapshot,
            positions_snapshot: decisionData.positions_snapshot,
            market_data_snapshot: decisionData.market_data_snapshot,
            created_at: decisionData.created_at
          })
        });
        
        if (decisionInsertResponse.ok) {
          syncResults.decisions++;
          console.log(`[Data Sync] Decision data for ${traderId} synced successfully`);
        }
        
      } catch (traderError) {
        console.error(`[Data Sync] Error syncing trader ${trader.id}:`, traderError);
        syncResults.errors.push({
          trader_id: trader.id,
          error: traderError.message
        });
      }
    }
    
    console.log('[Data Sync] Synchronization complete:', syncResults);
    
    // 返回同步结果
    return new Response(JSON.stringify({
      success: true,
      message: 'Data synchronization completed',
      results: syncResults,
      timestamp: new Date().toISOString(),
      note: 'This is demo data for NOFX Supabase integration'
    }), {
      status: 200,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
    
  } catch (error) {
    console.error('[Data Sync] Fatal error:', error);
    
    return new Response(JSON.stringify({
      success: false,
      error: {
        code: 'SYNC_ERROR',
        message: error.message,
        stack: error.stack
      }
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
});