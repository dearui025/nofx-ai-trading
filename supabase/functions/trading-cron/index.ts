// 定时交易任务 - 每3分钟扫描市场并执行交易决策
// 这是一个Cron Function，会被Supabase定时调用

Deno.serve(async (req) => {
    const corsHeaders = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
        'Access-Control-Max-Age': '86400',
    };

    if (req.method === 'OPTIONS') {
        return new Response(null, { status: 200, headers: corsHeaders });
    }

    try {
        console.log('🚀 开始定时交易任务...');

        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');
        const deepseekApiKey = Deno.env.get('DEEPSEEK_API_KEY');

        if (!supabaseUrl || !serviceRoleKey) {
            throw new Error('Supabase配置缺失');
        }

        // 1. 获取账户余额
        console.log('📊 获取账户余额...');
        const balanceResponse = await fetch(`${supabaseUrl}/functions/v1/binance-trading`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                action: 'getBalance',
                params: { trader_id: 'binance_testnet' },
            }),
        });

        if (!balanceResponse.ok) {
            throw new Error('获取余额失败');
        }

        const balanceData = await balanceResponse.json();
        const balance = balanceData.data;

        console.log(`✓ 账户余额: ${balance.availableBalance} USDT`);

        // 2. 获取持仓
        console.log('📊 获取当前持仓...');
        const positionsResponse = await fetch(`${supabaseUrl}/functions/v1/binance-trading`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                action: 'getPositions',
                params: {},
            }),
        });

        const positionsData = await positionsResponse.json();
        const positions = positionsData.data || [];

        console.log(`✓ 当前持仓数: ${positions.length}`);

        // 3. 获取市场热门币种
        console.log('📊 获取市场数据...');
        const marketResponse = await fetch(`${supabaseUrl}/functions/v1/market-data`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                action: 'getTopGainers',
                params: {},
            }),
        });

        const marketData = await marketResponse.json();
        const topCoins = marketData.data || [];

        console.log(`✓ 获取到 ${topCoins.length} 个热门币种`);

        // 4. 简单的交易逻辑（示例）
        // 如果有DEEPSEEK_API_KEY，可以调用AI进行决策
        // 这里先实现简单的规则：如果没有持仓且余额充足，开一个多单

        const decisions = [];

        if (positions.length === 0 && parseFloat(balance.availableBalance) > 100) {
            // 选择涨幅最大的币种
            const bestCoin = topCoins[0];
            
            if (bestCoin && bestCoin.priceChangePercent > 2) {
                console.log(`🎯 发现交易机会: ${bestCoin.symbol} 涨幅 ${bestCoin.priceChangePercent}%`);

                // 计算仓位大小（使用10%的可用余额）
                const positionSize = parseFloat(balance.availableBalance) * 0.1;
                const quantity = (positionSize / bestCoin.lastPrice).toFixed(3);

                decisions.push({
                    action: '开多仓',
                    symbol: bestCoin.symbol,
                    reason: `24h涨幅${bestCoin.priceChangePercent}%，动能强劲`,
                    quantity: parseFloat(quantity),
                    leverage: 5,
                });

                // 如果启用真实交易，可以在这里执行
                // const tradeResult = await executeTrade(bestCoin.symbol, quantity, 'BUY', 'LONG', 5);
            }
        }

        // 5. 保存决策到数据库
        if (decisions.length > 0) {
            console.log('💾 保存决策记录...');
            
            for (const decision of decisions) {
                await fetch(`${supabaseUrl}/rest/v1/decisions`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${serviceRoleKey}`,
                        'apikey': serviceRoleKey,
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        trader_id: 'binance_testnet',
                        cycle_number: Math.floor(Date.now() / 180000), // 每3分钟一个周期
                        decision: decision.action,
                        reasoning: decision.reason,
                        market_analysis: JSON.stringify(topCoins.slice(0, 5)),
                        timestamp: new Date().toISOString(),
                    }),
                });
            }
        }

        const result = {
            timestamp: new Date().toISOString(),
            balance: {
                available: balance.availableBalance,
                total: balance.totalWalletBalance,
            },
            positions_count: positions.length,
            decisions_made: decisions.length,
            decisions: decisions,
        };

        console.log('✅ 定时任务完成');

        return new Response(JSON.stringify({ data: result }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

    } catch (error) {
        console.error('❌ 定时任务错误:', error);
        return new Response(JSON.stringify({
            error: {
                code: 'TRADING_CRON_ERROR',
                message: error.message,
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
    }
});
