// AI决策引擎 - Supabase Edge Function
// 负责生成AI交易决策

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
        const { trader_id, market_data } = await req.json();

        if (!trader_id) {
            throw new Error('trader_id is required');
        }

        // 获取trader配置
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

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

        // 构建AI决策提示词
        const systemPrompt = `你是一个专业的加密货币期货交易AI。你的目标是最大化夏普比率，同时控制风险。

交易策略：
- 每3分钟扫描一次市场
- 目标风险回报比 ≥ 1:3
- 最多同时持仓3个币种
- 保证金使用率 ≤ 90%

请基于提供的市场数据做出交易决策。`;

        const userPrompt = `市场数据：
${JSON.stringify(market_data, null, 2)}

Trader配置：
- 交易所: ${trader.exchange}
- 杠杆: ${trader.config?.leverage || '1x'}
- 风险偏好: ${trader.config?.risk_level || 'medium'}

请分析并给出：
1. 是否建议开仓 (BUY/SELL/HOLD)
2. 建议的币种和数量
3. 目标价格和止损价格
4. 决策理由
5. 风险评估`;

        // 调用AI模型
        let aiDecision;
        const aiProvider = trader.config?.ai_provider || 'deepseek';

        if (aiProvider === 'deepseek') {
            aiDecision = await callDeepSeekAPI(systemPrompt, userPrompt);
        } else if (aiProvider === 'qwen') {
            aiDecision = await callQwenAPI(systemPrompt, userPrompt);
        } else {
            // 默认使用DeepSeek
            aiDecision = await callDeepSeekAPI(systemPrompt, userPrompt);
        }

        // 解析AI决策
        const decision = parseAIDecision(aiDecision);

        // 保存决策到数据库
        const decisionRecord = {
            trader_id,
            decision_type: decision.action,
            symbol: decision.symbol,
            side: decision.side,
            quantity: decision.quantity,
            entry_price: decision.entry_price,
            stop_loss: decision.stop_loss,
            take_profit: decision.take_profit,
            reasoning: decision.reasoning,
            confidence_score: decision.confidence,
            ai_model: aiProvider,
            created_at: new Date().toISOString()
        };

        const insertResponse = await fetch(`${supabaseUrl}/rest/v1/ai_decisions`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json',
                'Prefer': 'return=representation'
            },
            body: JSON.stringify(decisionRecord)
        });

        const savedDecision = await insertResponse.json();

        return new Response(JSON.stringify({
            data: {
                decision: savedDecision[0],
                reasoning: decision.reasoning,
                confidence: decision.confidence
            }
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });

    } catch (error) {
        console.error('Decision Engine error:', error);
        
        return new Response(JSON.stringify({
            error: {
                code: 'DECISION_ENGINE_ERROR',
                message: error.message
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});

async function callDeepSeekAPI(systemPrompt: string, userPrompt: string) {
    const apiKey = Deno.env.get('DEEPSEEK_API_KEY');
    
    if (!apiKey) {
        throw new Error('DeepSeek API key not configured');
    }

    const response = await fetch('https://api.deepseek.com/v1/chat/completions', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${apiKey}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            model: 'deepseek-chat',
            messages: [
                { role: 'system', content: systemPrompt },
                { role: 'user', content: userPrompt }
            ],
            temperature: 0.3,
            max_tokens: 1000
        })
    });

    if (!response.ok) {
        throw new Error(`DeepSeek API error: ${response.status}`);
    }

    const data = await response.json();
    return data.choices[0].message.content;
}

async function callQwenAPI(systemPrompt: string, userPrompt: string) {
    const apiKey = Deno.env.get('QWEN_API_KEY');
    
    if (!apiKey) {
        throw new Error('Qwen API key not configured');
    }

    const response = await fetch('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${apiKey}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            model: 'qwen-plus',
            input: {
                messages: [
                    { role: 'system', content: systemPrompt },
                    { role: 'user', content: userPrompt }
                ]
            },
            parameters: {
                temperature: 0.3,
                max_tokens: 1000
            }
        })
    });

    if (!response.ok) {
        throw new Error(`Qwen API error: ${response.status}`);
    }

    const data = await response.json();
    return data.output?.text || data.choices[0].message.content;
}

function parseAIDecision(aiResponse: string) {
    // 解析AI响应的简单实现
    // 在实际应用中，这里应该有更复杂的NLP解析逻辑
    
    const lines = aiResponse.toLowerCase().split('\n');
    
    let action = 'HOLD';
    let symbol = 'BTCUSDT';
    let side = 'BUY';
    let quantity = 0.001;
    let entry_price = 0;
    let stop_loss = 0;
    let take_profit = 0;
    let reasoning = aiResponse;
    let confidence = 0.7;

    // 简单的关键词匹配
    for (const line of lines) {
        if (line.includes('buy') || line.includes('买入')) {
            action = 'BUY';
            side = 'BUY';
        } else if (line.includes('sell') || line.includes('卖出')) {
            action = 'SELL';
            side = 'SELL';
        } else if (line.includes('hold') || line.includes('持有')) {
            action = 'HOLD';
        }

        // 提取价格信息
        const priceMatch = line.match(/(\d+\.?\d*)/);
        if (priceMatch && entry_price === 0) {
            entry_price = parseFloat(priceMatch[1]);
        }
    }

    // 如果没有找到明确的价格，使用默认值
    if (entry_price === 0) {
        entry_price = 45000; // 默认BTC价格
    }

    // 计算止损和止盈价格
    if (side === 'BUY') {
        stop_loss = entry_price * 0.95; // 5%止损
        take_profit = entry_price * 1.15; // 15%止盈
    } else {
        stop_loss = entry_price * 1.05; // 5%止损
        take_profit = entry_price * 0.85; // 15%止盈
    }

    return {
        action,
        symbol,
        side,
        quantity,
        entry_price,
        stop_loss,
        take_profit,
        reasoning,
        confidence
    };
}