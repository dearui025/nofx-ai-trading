// 币安Testnet交易核心功能
// 支持：获取余额、持仓、下单、平仓、设置杠杆

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
        const { action, params } = await req.json();

        const apiKey = Deno.env.get('BINANCE_TESTNET_API_KEY');
        const secretKey = Deno.env.get('BINANCE_TESTNET_SECRET_KEY');

        if (!apiKey || !secretKey) {
            throw new Error('币安API密钥未配置');
        }

        const baseUrl = 'https://testnet.binancefuture.com';

        // 生成签名
        const generateSignature = async (queryString: string): Promise<string> => {
            const encoder = new TextEncoder();
            const keyData = encoder.encode(secretKey);
            const messageData = encoder.encode(queryString);
            
            const cryptoKey = await crypto.subtle.importKey(
                'raw',
                keyData,
                { name: 'HMAC', hash: 'SHA-256' },
                false,
                ['sign']
            );
            
            const signature = await crypto.subtle.sign('HMAC', cryptoKey, messageData);
            return Array.from(new Uint8Array(signature))
                .map(b => b.toString(16).padStart(2, '0'))
                .join('');
        };

        // 发送签名请求
        const signedRequest = async (endpoint: string, method: string, params: Record<string, any> = {}) => {
            const timestamp = Date.now();
            const queryParams = new URLSearchParams({
                ...params,
                timestamp: timestamp.toString(),
            });
            
            const signature = await generateSignature(queryParams.toString());
            queryParams.append('signature', signature);
            
            const url = `${baseUrl}${endpoint}?${queryParams.toString()}`;
            
            const response = await fetch(url, {
                method,
                headers: {
                    'X-MBX-APIKEY': apiKey,
                },
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`币安API错误: ${response.status} - ${errorText}`);
            }
            
            return await response.json();
        };

        let result;

        switch (action) {
            case 'getBalance':
                // 获取账户余额
                result = await signedRequest('/fapi/v2/account', 'GET');
                break;

            case 'getPositions':
                // 获取持仓信息
                const positions = await signedRequest('/fapi/v2/positionRisk', 'GET');
                // 只返回有持仓的
                result = positions.filter((pos: any) => parseFloat(pos.positionAmt) !== 0);
                break;

            case 'getMarketPrice':
                // 获取市价
                const { symbol } = params;
                const priceData = await fetch(`${baseUrl}/fapi/v1/ticker/price?symbol=${symbol}`);
                result = await priceData.json();
                break;

            case 'setLeverage':
                // 设置杠杆
                const { symbol: leverageSymbol, leverage } = params;
                result = await signedRequest('/fapi/v1/leverage', 'POST', {
                    symbol: leverageSymbol,
                    leverage: leverage,
                });
                break;

            case 'openLong':
            case 'openShort':
            case 'closeLong':
            case 'closeShort':
                // 下单
                const { symbol: orderSymbol, quantity, side, positionSide } = params;
                
                // 设置持仓模式（如果需要）
                try {
                    await signedRequest('/fapi/v1/positionSide/dual', 'POST', {
                        dualSidePosition: 'true',
                    });
                } catch (e) {
                    // 如果已经设置过，会报错，忽略
                }

                // 下市价单
                result = await signedRequest('/fapi/v1/order', 'POST', {
                    symbol: orderSymbol,
                    side: side, // BUY or SELL
                    positionSide: positionSide, // LONG or SHORT
                    type: 'MARKET',
                    quantity: quantity.toString(),
                });
                break;

            case 'cancelAllOrders':
                // 取消所有挂单
                const { symbol: cancelSymbol } = params;
                result = await signedRequest('/fapi/v1/allOpenOrders', 'DELETE', {
                    symbol: cancelSymbol,
                });
                break;

            case 'setStopLoss':
            case 'setTakeProfit':
                // 设置止损止盈
                const { 
                    symbol: slSymbol, 
                    positionSide: slPositionSide, 
                    quantity: slQuantity, 
                    stopPrice, 
                    takeProfitPrice 
                } = params;
                
                const orderType = action === 'setStopLoss' ? 'STOP_MARKET' : 'TAKE_PROFIT_MARKET';
                const triggerPrice = action === 'setStopLoss' ? stopPrice : takeProfitPrice;
                
                result = await signedRequest('/fapi/v1/order', 'POST', {
                    symbol: slSymbol,
                    side: slPositionSide === 'LONG' ? 'SELL' : 'BUY',
                    positionSide: slPositionSide,
                    type: orderType,
                    stopPrice: triggerPrice.toString(),
                    closePosition: 'true',
                });
                break;

            default:
                throw new Error(`未知操作: ${action}`);
        }

        // 保存到数据库（可选）
        const supabaseUrl = Deno.env.get('SUPABASE_URL');
        const serviceRoleKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY');

        if (action === 'getBalance' && supabaseUrl && serviceRoleKey) {
            // 保存账户历史
            const accountData = {
                trader_id: params?.trader_id || 'binance_testnet',
                total_wallet_balance: parseFloat(result.totalWalletBalance),
                available_balance: parseFloat(result.availableBalance),
                total_unrealized_profit: parseFloat(result.totalUnrealizedProfit),
                timestamp: new Date().toISOString(),
            };

            await fetch(`${supabaseUrl}/rest/v1/account_history`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${serviceRoleKey}`,
                    'apikey': serviceRoleKey,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(accountData),
            });
        }

        return new Response(JSON.stringify({ data: result }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

    } catch (error) {
        console.error('币安交易错误:', error);
        return new Response(JSON.stringify({
            error: {
                code: 'BINANCE_TRADING_ERROR',
                message: error.message,
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
    }
});
