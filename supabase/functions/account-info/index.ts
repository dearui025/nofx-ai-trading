// 账户信息 - Supabase Edge Function
// 获取和管理交易账户信息

Deno.serve(async (req) => {
    const corsHeaders = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        'Access-Control-Allow-Methods': 'POST, GET, OPTIONS, PUT, DELETE, PATCH',
        'Access-Control-Max-Age': '86400',
        'Access-Control-Allow-Credentials': 'false'
    };

    if (req.method === 'OPTIONS') {
        return new Response(null, { status: 200, headers: corsHeaders });
    }

    try {
        const { action, trader_id } = await req.json();

        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        switch (action) {
            case 'get_account_summary':
                return await getAccountSummary(corsHeaders, trader_id);
            
            case 'get_balance':
                return await getAccountBalance(corsHeaders, trader_id);
            
            case 'get_performance':
                return await getAccountPerformance(corsHeaders, trader_id);
            
            case 'get_trading_history':
                return await getTradingHistory(corsHeaders, trader_id);
            
            default:
                throw new Error(`Unknown action: ${action}`);
        }

    } catch (error) {
        console.error('Account Info error:', error);
        
        return new Response(JSON.stringify({
            error: {
                code: 'ACCOUNT_INFO_ERROR',
                message: error.message
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});

async function getAccountSummary(corsHeaders: any, traderId?: string) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    let whereClause = '';
    if (traderId) {
        whereClause = `?trader_id=eq.${traderId}`;
    }

    // 获取当前持仓
    const positionsResponse = await fetch(`${supabaseUrl}/rest/v1/positions${whereClause}&select=*`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const positions = await positionsResponse.json();

    // 获取最近交易记录
    const tradesResponse = await fetch(`${supabaseUrl}/rest/v1/trades${whereClause}&select=*&order=executed_at.desc&limit=50`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const trades = await tradesResponse.json();

    // 获取AI决策记录
    const decisionsResponse = await fetch(`${supabaseUrl}/rest/v1/ai_decisions${whereClause}&select=*&order=created_at.desc&limit=20`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const decisions = await decisionsResponse.json();

    // 计算账户统计
    const accountStats = calculateAccountStats(positions, trades);

    return new Response(JSON.stringify({
        data: {
            summary: {
                total_positions: positions.length,
                total_trades_today: trades.filter((t: any) => isToday(t.executed_at)).length,
                active_positions: positions.filter((p: any) => p.quantity > 0).length,
                account_balance: accountStats.totalBalance,
                unrealized_pnl: accountStats.unrealizedPnL,
                realized_pnl: accountStats.realizedPnL,
                win_rate: accountStats.winRate,
                total_return: accountStats.totalReturn
            },
            positions: positions,
            recent_trades: trades.slice(0, 10),
            recent_decisions: decisions.slice(0, 5),
            performance: accountStats.performance
        }
    }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function getAccountBalance(corsHeaders: any, traderId: string) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    // 获取trader配置
    const traderResponse = await fetch(`${supabaseUrl}/rest/v1/traders?id=eq.${traderId}&select=*`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const traders = await traderResponse.json();
    if (!traders || traders.length === 0) {
        throw new Error('Trader not found');
    }

    const trader = traders[0];
    let balanceData = {};

    // 根据交易所获取余额
    if (trader.exchange === 'binance') {
        balanceData = await getBinanceBalance(trader);
    } else if (trader.exchange === 'hyperliquid') {
        balanceData = await getHyperliquidBalance(trader);
    } else if (trader.exchange === 'aster') {
        balanceData = await getAsterBalance(trader);
    }

    return new Response(JSON.stringify({
        data: {
            exchange: trader.exchange,
            balances: balanceData,
            timestamp: new Date().toISOString()
        }
    }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function getAccountPerformance(corsHeaders: any, traderId: string) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    // 获取过去30天的交易记录
    const thirtyDaysAgo = new Date();
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

    const tradesResponse = await fetch(`${supabaseUrl}/rest/v1/trades?trader_id=eq.${traderId}&executed_at=gte.${thirtyDaysAgo.toISOString()}&select=*&order=executed_at.desc`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const trades = await tradesResponse.json();

    const performance = calculatePerformanceMetrics(trades);

    return new Response(JSON.stringify({
        data: performance
    }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function getTradingHistory(corsHeaders: any, traderId: string) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    const tradesResponse = await fetch(`${supabaseUrl}/rest/v1/trades?trader_id=eq.${traderId}&select=*&order=executed_at.desc&limit=100`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const trades = await tradesResponse.json();

    return new Response(JSON.stringify({
        data: trades
    }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

function calculateAccountStats(positions: any[], trades: any[]) {
    let totalBalance = 0;
    let unrealizedPnL = 0;
    let realizedPnL = 0;
    let winningTrades = 0;
    let totalTrades = trades.length;

    // 计算未实现盈亏
    for (const position of positions) {
        if (position.quantity > 0) {
            // 这里应该获取当前市场价格来计算未实现盈亏
            // 暂时使用平均价格作为当前价格
            totalBalance += position.quantity * position.avg_price;
            unrealizedPnL += 0; // 需要实时价格数据
        }
    }

    // 计算已实现盈亏和胜率
    for (const trade of trades) {
        if (trade.pnl) {
            realizedPnL += trade.pnl;
            if (trade.pnl > 0) {
                winningTrades++;
            }
        }
    }

    const winRate = totalTrades > 0 ? (winningTrades / totalTrades) * 100 : 0;
    const totalReturn = totalBalance > 0 ? (realizedPnL / totalBalance) * 100 : 0;

    return {
        totalBalance,
        unrealizedPnL,
        realizedPnL,
        winRate,
        totalReturn,
        performance: {
            daily_return: 0, // 需要计算
            weekly_return: 0,
            monthly_return: totalReturn,
            sharpe_ratio: 0 // 需要更复杂的计算
        }
    };
}

function calculatePerformanceMetrics(trades: any[]) {
    if (trades.length === 0) {
        return {
            total_trades: 0,
            winning_trades: 0,
            losing_trades: 0,
            win_rate: 0,
            total_pnl: 0,
            avg_win: 0,
            avg_loss: 0,
            profit_factor: 0,
            max_drawdown: 0
        };
    }

    let totalPnL = 0;
    let winningTrades = 0;
    let losingTrades = 0;
    let totalWins = 0;
    let totalLosses = 0;
    let peak = 0;
    let maxDrawdown = 0;
    let runningBalance = 10000; // 假设起始余额

    for (const trade of trades) {
        const pnl = trade.pnl || 0;
        totalPnL += pnl;
        runningBalance += pnl;

        if (pnl > 0) {
            winningTrades++;
            totalWins += pnl;
        } else {
            losingTrades++;
            totalLosses += Math.abs(pnl);
        }

        if (runningBalance > peak) {
            peak = runningBalance;
        }

        const drawdown = (peak - runningBalance) / peak;
        if (drawdown > maxDrawdown) {
            maxDrawdown = drawdown;
        }
    }

    const winRate = (winningTrades / trades.length) * 100;
    const avgWin = winningTrades > 0 ? totalWins / winningTrades : 0;
    const avgLoss = losingTrades > 0 ? totalLosses / losingTrades : 0;
    const profitFactor = totalLosses > 0 ? totalWins / totalLosses : 0;

    return {
        total_trades: trades.length,
        winning_trades: winningTrades,
        losing_trades: losingTrades,
        win_rate: winRate,
        total_pnl: totalPnL,
        avg_win: avgWin,
        avg_loss: avgLoss,
        profit_factor: profitFactor,
        max_drawdown: maxDrawdown * 100,
        current_balance: runningBalance
    };
}

function isToday(dateString: string): boolean {
    const date = new Date(dateString);
    const today = new Date();
    return date.toDateString() === today.toDateString();
}

async function getBinanceBalance(trader: any) {
    // 模拟Binance余额获取
    return {
        USDT: 10000,
        BTC: 0.1,
        ETH: 2.5
    };
}

async function getHyperliquidBalance(trader: any) {
    // 模拟Hyperliquid余额获取
    return {
        USD: 10000,
        ETH: 2.5
    };
}

async function getAsterBalance(trader: any) {
    // 模拟Aster余额获取
    return {
        USDC: 10000,
        ETH: 2.5
    };
}