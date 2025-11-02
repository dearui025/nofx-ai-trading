// 测试币安API密钥配置

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
        const apiKey = Deno.env.get('BINANCE_TESTNET_API_KEY');
        const secretKey = Deno.env.get('BINANCE_TESTNET_SECRET_KEY');

        const diagnostics = {
            api_key_configured: !!apiKey,
            api_key_length: apiKey ? apiKey.length : 0,
            api_key_prefix: apiKey ? apiKey.substring(0, 10) + '...' : 'NOT SET',
            secret_key_configured: !!secretKey,
            secret_key_length: secretKey ? secretKey.length : 0,
            secret_key_prefix: secretKey ? secretKey.substring(0, 10) + '...' : 'NOT SET',
        };

        // 测试简单的签名请求
        if (apiKey && secretKey) {
            try {
                const baseUrl = 'https://testnet.binancefuture.com';
                const timestamp = Date.now();
                
                // 生成签名
                const queryString = `timestamp=${timestamp}`;
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
                const signatureHex = Array.from(new Uint8Array(signature))
                    .map(b => b.toString(16).padStart(2, '0'))
                    .join('');
                
                // 测试API调用
                const testUrl = `${baseUrl}/fapi/v2/account?timestamp=${timestamp}&signature=${signatureHex}`;
                const response = await fetch(testUrl, {
                    method: 'GET',
                    headers: {
                        'X-MBX-APIKEY': apiKey,
                    },
                });

                diagnostics.api_test_status = response.status;
                diagnostics.api_test_ok = response.ok;
                
                if (!response.ok) {
                    const errorText = await response.text();
                    diagnostics.api_test_error = errorText;
                } else {
                    const data = await response.json();
                    diagnostics.api_test_success = true;
                    diagnostics.account_data = {
                        totalWalletBalance: data.totalWalletBalance,
                        availableBalance: data.availableBalance,
                    };
                }
            } catch (testError) {
                diagnostics.api_test_exception = testError.message;
            }
        }

        return new Response(JSON.stringify({ data: diagnostics }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

    } catch (error) {
        console.error('诊断错误:', error);
        return new Response(JSON.stringify({
            error: {
                code: 'DIAGNOSTIC_ERROR',
                message: error.message,
            }
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
    }
});
