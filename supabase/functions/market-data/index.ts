// 获取币安Testnet市场数据
// 支持：K线数据、24h行情、资金费率等

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
        const baseUrl = 'https://testnet.binancefuture.com';

        let result;

        switch (action) {
            case 'get24hTicker':
                // 获取24小时行情
                const { symbols } = params;
                if (symbols && symbols.length > 0) {
                    // 批量获取
                    const promises = symbols.map((symbol: string) =>
                        fetch(`${baseUrl}/fapi/v1/ticker/24hr?symbol=${symbol}`).then(r => r.json())
                    );
                    result = await Promise.all(promises);
                } else {
                    // 获取所有
                    const response = await fetch(`${baseUrl}/fapi/v1/ticker/24hr`);
                    result = await response.json();
                }
                break;

            case 'getKlines':
                // 获取K线数据
                const { symbol, interval, limit } = params;
                const klinesUrl = `${baseUrl}/fapi/v1/klines?symbol=${symbol}&interval=${interval || '1h'}&limit=${limit || 100}`;
                const klinesResponse = await fetch(klinesUrl);
                const klinesData = await klinesResponse.json();
                
                // 格式化K线数据
                result = klinesData.map((k: any[]) => ({
                    openTime: k[0],
                    open: parseFloat(k[1]),
                    high: parseFloat(k[2]),
                    low: parseFloat(k[3]),
                    close: parseFloat(k[4]),
                    volume: parseFloat(k[5]),
                    closeTime: k[6],
                }));
                break;

            case 'getFundingRate':
                // 获取资金费率
                const { symbol: fundingSymbol } = params;
                const fundingUrl = `${baseUrl}/fapi/v1/fundingRate?symbol=${fundingSymbol}&limit=1`;
                const fundingResponse = await fetch(fundingUrl);
                result = await fundingResponse.json();
                break;

            case 'getTopGainers':
                // 获取涨幅榜前10
                const allTickers = await fetch(`${baseUrl}/fapi/v1/ticker/24hr`);
                const allData = await allTickers.json();
                
                result = allData
                    .filter((t: any) => t.symbol.endsWith('USDT'))
                    .map((t: any) => ({
                        symbol: t.symbol,
                        priceChangePercent: parseFloat(t.priceChangePercent),
                        lastPrice: parseFloat(t.lastPrice),
                        volume: parseFloat(t.volume),
                        quoteVolume: parseFloat(t.quoteVolume),
                    }))
                    .sort((a: any, b: any) => b.priceChangePercent - a.priceChangePercent)
                    .slice(0, 10);
                break;

            case 'getOpenInterest':
                // 获取持仓量
                const { symbol: oiSymbol } = params;
                const oiUrl = `${baseUrl}/fapi/v1/openInterest?symbol=${oiSymbol}`;
                const oiResponse = await fetch(oiUrl);
                result = await oiResponse.json();
                break;

            case 'getMarketDepth':
                // 获取盘口深度
                const { symbol: depthSymbol, limit: depthLimit } = params;
                const depthUrl = `${baseUrl}/fapi/v1/depth?symbol=${depthSymbol}&limit=${depthLimit || 20}`;
                const depthResponse = await fetch(depthUrl);
                result = await depthResponse.json();
                break;

            case 'getMarketPrice':
                // 获取单个币种价格
                const { symbol: priceSymbol } = params;
                const priceUrl = `${baseUrl}/fapi/v1/ticker/price?symbol=${priceSymbol}`;
                const priceResponse = await fetch(priceUrl);
                result = await priceResponse.json();
                break;

            default:
                throw new Error(`未知操作: ${action}`);
        }

        return new Response(JSON.stringify({ data: result }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

    } catch (error) {
        console.error('市场数据错误:', error);
        return new Response(JSON.stringify({
            error: {
                code: 'MARKET_DATA_ERROR',
                message: error.message,
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
    }
});
