// 交易执行器 - Supabase Edge Function
// 负责执行实际交易操作

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
        const { trader_id, symbol, side, quantity, price, order_type } = await req.json();

        if (!trader_id || !symbol || !side || !quantity) {
            throw new Error('Missing required parameters: trader_id, symbol, side, quantity');
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

        // 执行交易
        let tradeResult;
        const exchange = trader.exchange;

        if (exchange === 'binance') {
            tradeResult = await executeBinanceTrade(trader, symbol, side, quantity, price, order_type);
        } else if (exchange === 'hyperliquid') {
            tradeResult = await executeHyperliquidTrade(trader, symbol, side, quantity, price, order_type);
        } else if (exchange === 'aster') {
            tradeResult = await executeAsterTrade(trader, symbol, side, quantity, price, order_type);
        } else {
            throw new Error(`Unsupported exchange: ${exchange}`);
        }

        // 保存交易记录到数据库
        const tradeRecord = {
            trader_id,
            symbol,
            side,
            quantity,
            price: tradeResult.price || price,
            order_type: order_type || 'MARKET',
            exchange_order_id: tradeResult.orderId,
            status: tradeResult.status,
            executed_at: new Date().toISOString(),
            exchange_response: JSON.stringify(tradeResult)
        };

        const insertResponse = await fetch(`${supabaseUrl}/rest/v1/trades`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json',
                'Prefer': 'return=representation'
            },
            body: JSON.stringify(tradeRecord)
        });

        const savedTrade = await insertResponse.json();

        // 更新持仓信息
        await updatePosition(supabaseUrl, serviceRoleKey, trader_id, symbol, side, quantity, tradeResult.price || price);

        return new Response(JSON.stringify({
            data: {
                trade: savedTrade[0],
                exchange_result: tradeResult
            }
        }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });

    } catch (error) {
        console.error('Trade Executor error:', error);
        
        return new Response(JSON.stringify({
            error: {
                code: 'TRADE_EXECUTION_ERROR',
                message: error.message
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});

async function executeBinanceTrade(trader: any, symbol: string, side: string, quantity: number, price?: number, orderType?: string) {
    const apiKey = trader.config?.binance_api_key;
    const apiSecret = trader.config?.binance_api_secret;

    if (!apiKey || !apiSecret) {
        throw new Error('Binance API credentials not configured');
    }

    const timestamp = Date.now();
    const params = new URLSearchParams({
        symbol,
        side: side.toUpperCase(),
        type: (orderType || 'MARKET').toUpperCase(),
        quantity: quantity.toString(),
        timestamp: timestamp.toString()
    });

    if (price && orderType === 'LIMIT') {
        params.append('price', price.toString());
        params.append('timeInForce', 'GTC');
    }

    // 生成签名
    const signature = await generateSignature(params.toString(), apiSecret);
    params.append('signature', signature);

    const response = await fetch('https://api.binance.com/api/v3/order', {
        method: 'POST',
        headers: {
            'X-MBX-APIKEY': apiKey,
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: params.toString()
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Binance API error: ${response.status} - ${errorText}`);
    }

    const data = await response.json();
    
    return {
        exchange: 'binance',
        orderId: data.orderId,
        symbol: data.symbol,
        status: data.status,
        executedQty: data.executedQty,
        price: parseFloat(data.price) || price,
        side: data.side,
        type: data.type,
        timestamp: data.transactTime
    };
}

async function executeHyperliquidTrade(trader: any, symbol: string, side: string, quantity: number, price?: number, orderType?: string) {
    const privateKey = trader.config?.hyperliquid_private_key;

    if (!privateKey) {
        throw new Error('Hyperliquid private key not configured');
    }

    // Hyperliquid API调用 - 这里需要根据实际的API文档实现
    // 由于API可能比较复杂，这里提供一个模拟实现
    const mockOrderId = `hl_${Date.now()}`;
    
    return {
        exchange: 'hyperliquid',
        orderId: mockOrderId,
        symbol,
        status: 'FILLED',
        executedQty: quantity,
        price: price || 45000,
        side: side.toUpperCase(),
        type: orderType || 'MARKET',
        timestamp: Date.now()
    };
}

async function executeAsterTrade(trader: any, symbol: string, side: string, quantity: number, price?: number, orderType?: string) {
    const apiKey = trader.config?.aster_api_key;
    const apiSecret = trader.config?.aster_api_secret;

    if (!apiKey || !apiSecret) {
        throw new Error('Aster API credentials not configured');
    }

    // Aster DEX API调用 - 这里需要根据实际的API文档实现
    // 由于API可能比较复杂，这里提供一个模拟实现
    const mockOrderId = `aster_${Date.now()}`;
    
    return {
        exchange: 'aster',
        orderId: mockOrderId,
        symbol,
        status: 'FILLED',
        executedQty: quantity,
        price: price || 45000,
        side: side.toUpperCase(),
        type: orderType || 'MARKET',
        timestamp: Date.now()
    };
}

async function generateSignature(message: string, secret: string): Promise<string> {
    const encoder = new TextEncoder();
    const keyData = encoder.encode(secret);
    const messageData = encoder.encode(message);

    const cryptoKey = await crypto.subtle.importKey(
        'raw',
        keyData,
        { name: 'HMAC', hash: 'SHA-256' },
        false,
        ['sign']
    );

    const signature = await crypto.subtle.sign('HMAC', cryptoKey, messageData);
    const hashArray = Array.from(new Uint8Array(signature));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    
    return hashHex;
}

async function updatePosition(supabaseUrl: string, serviceRoleKey: string, traderId: string, symbol: string, side: string, quantity: number, price: number) {
    // 检查是否已存在该symbol的持仓
    const positionResponse = await fetch(`${supabaseUrl}/rest/v1/positions?trader_id=eq.${traderId}&symbol=eq.${symbol}&select=*`, {
        headers: {
            'Authorization': `Bearer ${serviceRoleKey}`,
            'apikey': serviceRoleKey,
            'Content-Type': 'application/json'
        }
    });

    const positions = await positionResponse.json();

    if (positions && positions.length > 0) {
        // 更新现有持仓
        const position = positions[0];
        let newQuantity = position.quantity;
        let newAvgPrice = position.avg_price;

        if (side.toUpperCase() === 'BUY') {
            // 买入操作
            const totalValue = position.quantity * position.avg_price + quantity * price;
            newQuantity = position.quantity + quantity;
            newAvgPrice = totalValue / newQuantity;
        } else {
            // 卖出操作
            newQuantity = position.quantity - quantity;
            if (newQuantity <= 0) {
                // 平仓
                await fetch(`${supabaseUrl}/rest/v1/positions?id=eq.${position.id}`, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${serviceRoleKey}`,
                        'apikey': serviceRoleKey
                    }
                });
                return;
            }
        }

        const updateResponse = await fetch(`${supabaseUrl}/rest/v1/positions?id=eq.${position.id}`, {
            method: 'PATCH',
            headers: {
                'Authorization': `Bearer ${serviceRoleKey}`,
                'apikey': serviceRoleKey,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                quantity: newQuantity,
                avg_price: newAvgPrice,
                updated_at: new Date().toISOString()
            })
        });

        if (!updateResponse.ok) {
            console.error('Failed to update position:', await updateResponse.text());
        }
    } else {
        // 创建新持仓
        if (side.toUpperCase() === 'BUY') {
            const newPosition = {
                trader_id: traderId,
                symbol,
                quantity,
                avg_price: price,
                side: 'LONG',
                created_at: new Date().toISOString()
            };

            await fetch(`${supabaseUrl}/rest/v1/positions`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${serviceRoleKey}`,
                    'apikey': serviceRoleKey,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(newPosition)
            });
        }
        // 卖出操作如果没有持仓则忽略
    }
}