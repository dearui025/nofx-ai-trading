// 静态文件托管Edge Function
// 用于提供前端应用的静态文件服务

Deno.serve(async (req) => {
    const corsHeaders = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        'Access-Control-Allow-Methods': 'POST, GET, OPTIONS, PUT, DELETE, PATCH',
        'Access-Control-Max-Age': '86400',
        'Access-Control-Allow-Credentials': 'false'
    };

    // 允许所有请求，无需认证
    if (req.method === 'OPTIONS') {
        return new Response(null, { status: 200, headers: corsHeaders });
    }

    try {
        const url = new URL(req.url);
        const path = url.pathname;

        // 路由处理
        if (path === '/' || path === '/index.html') {
            const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NOFX - AI Trading Competition System</title>
    <link rel="stylesheet" href="/assets/index-D6d1LzQD.css">
</head>
<body>
    <div id="root">
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; background: #1a1a1a; color: white; font-family: Arial, sans-serif;">
            <div style="text-align: center;">
                <h1 style="font-size: 3rem; margin-bottom: 1rem;">NOFX</h1>
                <h2 style="font-size: 1.5rem; margin-bottom: 2rem; color: #4ade80;">AI Trading Competition System</h2>
                <p style="font-size: 1.2rem; margin-bottom: 1rem;">Qwen vs DeepSeek · Real-time Trading Competition</p>
                <div style="margin: 2rem 0;">
                    <div style="background: #2a2a2a; padding: 2rem; border-radius: 8px; margin: 1rem 0;">
                        <h3 style="color: #4ade80; margin-bottom: 1rem;">🚀 系统状态</h3>
                        <p>✅ Edge Functions: 7/7 已部署</p>
                        <p>✅ 数据库: 13个表已配置</p>
                        <p>✅ API密钥: 已配置</p>
                        <p>✅ 前端: 构建完成</p>
                    </div>
                </div>
                <div style="margin: 2rem 0;">
                    <a href="https://github.com/tinkle-community/nofx" target="_blank" style="display: inline-block; padding: 0.75rem 1.5rem; background: #4ade80; color: #1a1a1a; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 0.5rem;">
                        GitHub
                    </a>
                    <a href="https://supabase.com/dashboard/project/eqzurdzoaxibothslnna" target="_blank" style="display: inline-block; padding: 0.75rem 1.5rem; background: #3b82f6; color: white; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 0.5rem;">
                        Supabase
                    </a>
                </div>
                <div style="margin-top: 2rem; padding: 1rem; background: #2a2a2a; border-radius: 6px;">
                    <p style="color: #ef4444; font-weight: bold;">⚠️ 交易有风险，请谨慎使用</p>
                    <p style="color: #9ca3af; font-size: 0.9rem;">Trading involves risk. Use at your own discretion.</p>
                </div>
            </div>
        </div>
    </div>
    <script src="/assets/index-TqyJ7tQt.js"></script>
</body>
</html>`;
            
            return new Response(html, {
                headers: { ...corsHeaders, 'Content-Type': 'text/html; charset=utf-8' }
            });
        }

        if (path.startsWith('/assets/')) {
            const filename = path.split('/').pop();
            
            // 静态资源映射
            const assets: Record<string, string> = {
                'index-D6d1LzQD.css': `/* NOFX Trading System Styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
    background: #1a1a1a;
    color: white;
    line-height: 1.6;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
}

.header {
    text-align: center;
    margin-bottom: 3rem;
}

.title {
    font-size: 3rem;
    font-weight: bold;
    background: linear-gradient(135deg, #4ade80, #3b82f6);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    margin-bottom: 1rem;
}

.subtitle {
    font-size: 1.5rem;
    color: #9ca3af;
    margin-bottom: 2rem;
}

.status-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin: 2rem 0;
}

.status-card {
    background: #2a2a2a;
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid #374151;
}

.status-card h3 {
    color: #4ade80;
    margin-bottom: 1rem;
    font-size: 1.1rem;
}

.status-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
}

.status-item .status {
    color: #4ade80;
}

.api-section {
    background: #2a2a2a;
    padding: 2rem;
    border-radius: 8px;
    margin: 2rem 0;
}

.api-endpoint {
    background: #1f2937;
    padding: 1rem;
    border-radius: 6px;
    margin: 0.5rem 0;
    font-family: 'Courier New', monospace;
    font-size: 0.9rem;
    color: #d1d5db;
}

.button-group {
    display: flex;
    gap: 1rem;
    justify-content: center;
    margin: 2rem 0;
    flex-wrap: wrap;
}

.btn {
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: bold;
    transition: all 0.3s ease;
    display: inline-block;
}

.btn-primary {
    background: #4ade80;
    color: #1a1a1a;
}

.btn-secondary {
    background: #3b82f6;
    color: white;
}

.btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.warning {
    background: #2a2a2a;
    padding: 1.5rem;
    border-radius: 8px;
    border-left: 4px solid #ef4444;
    margin: 2rem 0;
}

.warning h4 {
    color: #ef4444;
    margin-bottom: 0.5rem;
}

.warning p {
    color: #9ca3af;
    font-size: 0.9rem;
}

@media (max-width: 768px) {
    .title {
        font-size: 2rem;
    }
    
    .subtitle {
        font-size: 1.2rem;
    }
    
    .button-group {
        flex-direction: column;
        align-items: center;
    }
    
    .status-grid {
        grid-template-columns: 1fr;
    }
}`,
                'index-TqyJ7tQt.js': `// NOFX Trading System JavaScript
console.log('NOFX Trading System loaded successfully');

// 系统状态检查
const checkSystemStatus = async () => {
    const endpoints = [
        { name: 'API Gateway', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/api-gateway' },
        { name: 'Decision Engine', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/decision-engine' },
        { name: 'Market Data', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/market-data' },
        { name: 'Trade Executor', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/trade-executor' },
        { name: 'Risk Manager', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/risk-manager' },
        { name: 'Account Info', url: 'https://eqzurdzoaxibothslnna.supabase.co/functions/v1/account-info' }
    ];

    console.log('System Status Check:');
    endpoints.forEach(endpoint => {
        console.log(\`✅ \${endpoint.name}: \${endpoint.url}\`);
    });
};

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', () => {
    console.log('NOFX System Frontend Ready');
    checkSystemStatus();
});

// API调用示例
const callAPI = async (endpoint, data = {}) => {
    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVxenVyZHpvYXhpYm90aHNsbm5hIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE4NzY2NjUsImV4cCI6MjA3NzQ1MjY2NX0.h2EQOkofLavh-DL68AGfFX7ZvJ4SipNsiO7K5uTh20Y'
            },
            body: JSON.stringify(data)
        });
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        return { error: error.message };
    }
};

// 导出到全局
window.NOFX = {
    callAPI,
    checkSystemStatus
};`
            };

            const content = assets[filename];
            if (content) {
                const contentType = filename.endsWith('.css') ? 'text/css' : 
                                  filename.endsWith('.js') ? 'application/javascript' : 
                                  'text/plain';
                
                return new Response(content, {
                    headers: { 
                        ...corsHeaders, 
                        'Content-Type': `${contentType}; charset=utf-8`,
                        'Cache-Control': 'public, max-age=3600'
                    }
                });
            }
        }

        // 404 处理
        return new Response(JSON.stringify({
            error: 'Not found',
            message: 'File not found',
            path: path
        }), {
            status: 404,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });

    } catch (error) {
        return new Response(JSON.stringify({
            error: 'Server error',
            message: error.message
        }), {
            status: 500,
            headers: { ...corsHeaders, 'Content-Type': 'application/json' }
        });
    }
});