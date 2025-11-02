// 风险管理器 - Supabase Edge Function
// 负责交易风险评估和控制

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
        const { trader_id, trade_data } = await req.json();

        if (!trader_id) {
            throw new Error('trader_id is required');
        }

        // 获取trader配置和当前风险状态
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        // 获取trader配置
        const traderResponse = await fetch(`${supabaseUrl}/rest/v1/traders?id=eq.${trader_id}&select=*`, {
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

        // 获取当前持仓
        const positionsResponse = await fetch(`${supabaseUrl}/rest/v1/positions?trader_id=eq.${trader_id}&select=*`, {
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json'
            }
        });

        const positions = await positionsResponse.json();

        // 获取风险管理规则
        const riskRules = trader.config?.risk_rules || {
            max_position_size: 0.1, // 最大仓位10%
            max_leverage: 10,
            max_positions: 3,
            stop_loss_percentage: 0.05,
            max_daily_loss: 0.02 // 最大日损失2%
        };

        // 执行风险检查
        const riskAssessment = await performRiskCheck(trader, positions, trade_data, riskRules);

        // 保存风险检查结果
        const riskRecord = {
            trader_id,
            trade_data: JSON.stringify(trade_data),
            risk_score: riskAssessment.risk_score,
            risk_level: riskAssessment.risk_level,
            recommendations: JSON.stringify(riskAssessment.recommendations),
            approved: riskAssessment.approved,
            created_at: new Date().toISOString()
        };

        const insertResponse = await fetch(`${supabaseUrl}/rest/v1/risk_assessments`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json',
                'Prefer': 'return=representation'
            },
            body: JSON.stringify(riskRecord)
        });

        const savedRisk = await insertResponse.json();

        return new Response(JSON.stringify({
            data: {
                risk_assessment: savedRisk[0],
                approved: riskAssessment.approved,
                risk_score: riskAssessment.risk_score,
                recommendations: riskAssessment.recommendations
            }
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });

    } catch (error) {
        console.error('Risk Manager error:', error);
        
        return new Response(JSON.stringify({
            error: {
                code: 'RISK_MANAGER_ERROR',
                message: error.message
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});

async function performRiskCheck(trader: any, positions: any[], tradeData: any, riskRules: any) {
    let riskScore = 0;
    const recommendations = [];

    // 1. 检查仓位数量限制
    const currentPositionCount = positions.length;
    if (currentPositionCount >= riskRules.max_positions) {
        riskScore += 30;
        recommendations.push(`当前持仓数量(${currentPositionCount})已达到最大值(${riskRules.max_positions})`);
    }

    // 2. 检查杠杆限制
    const leverage = tradeData.leverage || 1;
    if (leverage > riskRules.max_leverage) {
        riskScore += 25;
        recommendations.push(`杠杆倍数(${leverage})超过限制(${riskRules.max_leverage})`);
    }

    // 3. 检查单笔交易大小
    const tradeValue = tradeData.quantity * (tradeData.price || 0);
    const maxPositionValue = 10000 * riskRules.max_position_size; // 假设账户余额10000
    if (tradeValue > maxPositionValue) {
        riskScore += 20;
        recommendations.push(`单笔交易金额(${tradeValue})超过最大仓位限制(${maxPositionValue})`);
    }

    // 4. 检查止损设置
    if (!tradeData.stop_loss) {
        riskScore += 15;
        recommendations.push('建议设置止损价格以控制风险');
    }

    // 5. 检查风险回报比
    if (tradeData.stop_loss && tradeData.take_profit) {
        const risk = Math.abs(tradeData.price - tradeData.stop_loss);
        const reward = Math.abs(tradeData.take_profit - tradeData.price);
        const riskRewardRatio = reward / risk;
        
        if (riskRewardRatio < 1.5) {
            riskScore += 10;
            recommendations.push(`风险回报比(${riskRewardRatio.toFixed(2)})建议不低于1:3`);
        }
    }

    // 6. 检查市场波动性
    if (tradeData.market_volatility && tradeData.market_volatility > 0.1) {
        riskScore += 5;
        recommendations.push('市场波动性较高，建议谨慎操作');
    }

    // 确定风险等级
    let riskLevel = 'LOW';
    if (riskScore >= 70) {
        riskLevel = 'HIGH';
    } else if (riskScore >= 40) {
        riskLevel = 'MEDIUM';
    }

    // 决定是否批准交易
    const approved = riskScore < 60 && riskLevel !== 'HIGH';

    return {
        risk_score: riskScore,
        risk_level: riskLevel,
        approved,
        recommendations,
        current_position_count: currentPositionCount,
        max_positions: riskRules.max_positions,
        leverage_used: leverage,
        max_leverage: riskRules.max_leverage
    };
}