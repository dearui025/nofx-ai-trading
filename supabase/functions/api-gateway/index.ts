// NOFX API Gateway - Supabase Edge Function
// 统一入口点，负责路由、认证、限流

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
        const url = new URL(req.url);
        const path = url.pathname;
        const method = req.method;

        // 路由分发
        switch (path) {
            case '/api/health':
                return handleHealthCheck(corsHeaders);
            
            case '/api/traders':
                return handleTraders(method, req, corsHeaders);
            
            case '/api/positions':
                return handlePositions(method, req, corsHeaders);
            
            case '/api/decisions':
                return handleDecisions(method, req, corsHeaders);
            
            case '/api/market-data':
                return handleMarketData(method, req, corsHeaders);
            
            case '/api/trade-execute':
                return handleTradeExecute(method, req, corsHeaders);
            
            case '/api/risk-check':
                return handleRiskCheck(method, req, corsHeaders);
            
            case '/api/account':
                return handleAccount(method, req, corsHeaders);
            
            default:
                return new Response(JSON.stringify({
                    error: {
                        code: 'NOT_FOUND',
                        message: `Endpoint ${path} not found`
                    }
                }), {
                    status: 404,
                    headers: { ...corsHeaders, 'Content-Type': 'application/json' }
                });
        }

    } catch (error) {
        console.error('API Gateway error:', error);
        
        return new Response(JSON.stringify({
            error: {
                code: 'GATEWAY_ERROR',
                message: error.message
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});

async function handleHealthCheck(corsHeaders: any) {
    return new Response(JSON.stringify({
        data: {
            status: 'healthy',
            timestamp: new Date().toISOString(),
            version: '1.0.0',
            service: 'NOFX API Gateway'
        }
    }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleTraders(method: string, req: Request, corsHeaders: any) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    if (method === 'GET') {
        // 获取所有trader信息
        const response = await fetch(`${supabaseUrl}/rest/v1/traders?select=*`, {
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json'
            }
        });

        const traders = await response.json();
        
        return new Response(JSON.stringify({
            data: traders
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    if (method === 'POST') {
        // 创建新的trader
        const body = await req.json();
        const { name, exchange, config } = body;

        const response = await fetch(`${supabaseUrl}/rest/v1/traders`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json',
                'Prefer': 'return=representation'
            },
            body: JSON.stringify({
                name,
                exchange,
                config,
                status: 'active',
                created_at: new Date().toISOString()
            })
        });

        const trader = await response.json();
        
        return new Response(JSON.stringify({
            data: trader[0]
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handlePositions(method: string, req: Request, corsHeaders: any) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    if (method === 'GET') {
        // 获取当前持仓
        const response = await fetch(`${supabaseUrl}/rest/v1/positions?select=*`, {
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json'
            }
        });

        const positions = await response.json();
        
        return new Response(JSON.stringify({
            data: positions
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleDecisions(method: string, req: Request, corsHeaders: any) {
    const supabaseUrl = Deno.env.get('SUPABASE_URL');
    const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

    if (method === 'GET') {
        // 获取AI决策记录
        const response = await fetch(`${supabaseUrl}/rest/v1/ai_decisions?select=*&order=created_at.desc&limit=10`, {
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json'
            }
        });

        const decisions = await response.json();
        
        return new Response(JSON.stringify({
            data: decisions
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    if (method === 'POST') {
        // 生成新的AI决策
        const body = await req.json();
        const { trader_id, market_data } = body;

        // 调用AI决策引擎
        const decisionResponse = await fetch(`${Deno.env.get('SUPABASE_URL')}/functions/v1/decision-engine`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                trader_id,
                market_data
            })
        });

        const decision = await decisionResponse.json();
        
        return new Response(JSON.stringify({
            data: decision
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleMarketData(method: string, req: Request, corsHeaders: any) {
    if (method === 'GET') {
        // 调用市场数据服务
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        const response = await fetch(`${supabaseUrl}/functions/v1/market-data`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ action: 'get_latest' })
        });

        const marketData = await response.json();
        
        return new Response(JSON.stringify({
            data: marketData
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleTradeExecute(method: string, req: Request, corsHeaders: any) {
    if (method === 'POST') {
        const body = await req.json();
        const { trader_id, symbol, side, quantity, price, order_type } = body;

        // 调用交易执行服务
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        const response = await fetch(`${supabaseUrl}/functions/v1/trade-executor`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                trader_id,
                symbol,
                side,
                quantity,
                price,
                order_type
            })
        });

        const result = await response.json();
        
        return new Response(JSON.stringify({
            data: result
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleRiskCheck(method: string, req: Request, corsHeaders: any) {
    if (method === 'POST') {
        const body = await req.json();
        const { trader_id, trade_data } = body;

        // 调用风险管理服务
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        const response = await fetch(`${supabaseUrl}/functions/v1/risk-manager`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                trader_id,
                trade_data
            })
        });

        const result = await response.json();
        
        return new Response(JSON.stringify({
            data: result
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}

async function handleAccount(method: string, req: Request, corsHeaders: any) {
    if (method === 'GET') {
        // 获取账户信息
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        const response = await fetch(`${supabaseUrl}/functions/v1/account-info`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ action: 'get_account_summary' })
        });

        const accountData = await response.json();
        
        return new Response(JSON.stringify({
            data: accountData
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }

    return new Response(JSON.stringify({
        error: { code: 'METHOD_NOT_ALLOWED', message: 'Method not allowed' }
    }), {
        status: 405,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
}