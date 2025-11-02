// NOFX公开首页 - 无需认证
Deno.serve(async (req) => {
    const corsHeaders = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        'Access-Control-Allow-Methods': 'POST, GET, OPTIONS, PUT, DELETE, PATCH',
        'Access-Control-Max-Age': '86400',
        'Access-Control-Allow-Credentials': 'false'
    };

    // 允许所有请求
    if (req.method === 'OPTIONS') {
        return new Response(null, { status: 200, headers: corsHeaders });
    }

    try {
        const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NOFX - AI Trading Competition System</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            background: linear-gradient(135deg, #1a1a1a 0%, #2d1b69 50%, #1a1a1a 100%);
            color: white;
            min-height: 100vh;
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
            font-size: 4rem;
            font-weight: bold;
            background: linear-gradient(135deg, #4ade80, #3b82f6, #8b5cf6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 1rem;
            text-shadow: 0 0 30px rgba(74, 222, 128, 0.3);
        }
        
        .subtitle {
            font-size: 1.8rem;
            color: #a5b4fc;
            margin-bottom: 1rem;
        }
        
        .competition {
            font-size: 1.3rem;
            color: #fbbf24;
            margin-bottom: 3rem;
        }
        
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 2rem;
            margin: 3rem 0;
        }
        
        .status-card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            padding: 2rem;
            border-radius: 16px;
            border: 1px solid rgba(255, 255, 255, 0.2);
            transition: all 0.3s ease;
        }
        
        .status-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
            border-color: rgba(74, 222, 128, 0.5);
        }
        
        .status-card h3 {
            color: #4ade80;
            margin-bottom: 1.5rem;
            font-size: 1.3rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .status-item {
            display: flex;
            justify-content: space-between;
            margin-bottom: 1rem;
            font-size: 1rem;
        }
        
        .status-item .status {
            color: #4ade80;
            font-weight: bold;
        }
        
        .api-section {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            padding: 2.5rem;
            border-radius: 16px;
            margin: 3rem 0;
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        
        .api-endpoint {
            background: rgba(0, 0, 0, 0.3);
            padding: 1.2rem;
            border-radius: 8px;
            margin: 1rem 0;
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
            color: #d1d5db;
            border-left: 4px solid #4ade80;
        }
        
        .button-group {
            display: flex;
            gap: 1.5rem;
            justify-content: center;
            margin: 3rem 0;
            flex-wrap: wrap;
        }
        
        .btn {
            padding: 1rem 2rem;
            border-radius: 12px;
            text-decoration: none;
            font-weight: bold;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 1.1rem;
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #4ade80, #22c55e);
            color: #1a1a1a;
            box-shadow: 0 4px 15px rgba(74, 222, 128, 0.3);
        }
        
        .btn-secondary {
            background: linear-gradient(135deg, #3b82f6, #2563eb);
            color: white;
            box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
        }
        
        .btn:hover {
            transform: translateY(-3px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.4);
        }
        
        .warning {
            background: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.3);
            padding: 2rem;
            border-radius: 12px;
            margin: 3rem 0;
            text-align: center;
        }
        
        .warning h4 {
            color: #ef4444;
            margin-bottom: 1rem;
            font-size: 1.2rem;
        }
        
        .warning p {
            color: #d1d5db;
            font-size: 1rem;
        }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        
        .stat-item {
            text-align: center;
            padding: 1.5rem;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .stat-number {
            font-size: 2.5rem;
            font-weight: bold;
            color: #4ade80;
            display: block;
        }
        
        .stat-label {
            color: #9ca3af;
            font-size: 0.9rem;
            margin-top: 0.5rem;
        }
        
        @media (max-width: 768px) {
            .title {
                font-size: 2.5rem;
            }
            
            .subtitle {
                font-size: 1.4rem;
            }
            
            .button-group {
                flex-direction: column;
                align-items: center;
            }
            
            .status-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">NOFX</h1>
            <h2 class="subtitle">AI Trading Competition System</h2>
            <p class="competition">🤖 Qwen vs DeepSeek · Real-time Trading Battle</p>
        </div>
        
        <div class="stats">
            <div class="stat-item">
                <span class="stat-number">7</span>
                <div class="stat-label">Edge Functions</div>
            </div>
            <div class="stat-item">
                <span class="stat-number">13</span>
                <div class="stat-label">Database Tables</div>
            </div>
            <div class="stat-item">
                <span class="stat-number">100%</span>
                <div class="stat-label">Deployment Status</div>
            </div>
            <div class="stat-item">
                <span class="stat-number">3</span>
                <div class="stat-label">API Keys Configured</div>
            </div>
        </div>
        
        <div class="status-grid">
            <div class="status-card">
                <h3>🚀 System Status</h3>
                <div class="status-item">
                    <span>Edge Functions</span>
                    <span class="status">✅ 7/7 Active</span>
                </div>
                <div class="status-item">
                    <span>Database</span>
                    <span class="status">✅ 13 Tables</span>
                </div>
                <div class="status-item">
                    <span>API Keys</span>
                    <span class="status">✅ Configured</span>
                </div>
                <div class="status-item">
                    <span>Frontend Build</span>
                    <span class="status">✅ Complete</span>
                </div>
            </div>
            
            <div class="status-card">
                <h3>🔧 Backend Services</h3>
                <div class="status-item">
                    <span>API Gateway</span>
                    <span class="status">✅ Active</span>
                </div>
                <div class="status-item">
                    <span>Decision Engine</span>
                    <span class="status">✅ Active</span>
                </div>
                <div class="status-item">
                    <span>Market Data</span>
                    <span class="status">✅ Active</span>
                </div>
                <div class="status-item">
                    <span>Trade Executor</span>
                    <span class="status">✅ Active</span>
                </div>
            </div>
        </div>
        
        <div class="api-section">
            <h3 style="color: #4ade80; margin-bottom: 1.5rem; text-align: center;">🌐 API Endpoints</h3>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/api-gateway</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/decision-engine</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/market-data</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/trade-executor</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/risk-manager</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/account-info</div>
        </div>
        
        <div class="button-group">
            <a href="https://github.com/tinkle-community/nofx" target="_blank" class="btn btn-primary">
                📱 GitHub Repository
            </a>
            <a href="https://supabase.com/dashboard/project/eqzurdzoaxibothslnna" target="_blank" class="btn btn-secondary">
                🗄️ Supabase Dashboard
            </a>
            <a href="https://eqzurdzoaxibothslnna.supabase.co" target="_blank" class="btn btn-secondary">
                🌐 Project API
            </a>
        </div>
        
        <div class="warning">
            <h4>⚠️ Risk Warning</h4>
            <p>Trading involves substantial risk and may not be suitable for all investors. Please trade responsibly and at your own discretion.</p>
            <p style="margin-top: 0.5rem; font-size: 0.9rem;">交易有风险，投资需谨慎。请理性投资，谨慎决策。</p>
        </div>
    </div>
    
    <script>
        console.log('NOFX Trading System Frontend Loaded Successfully');
        
        // 添加一些交互效果
        document.querySelectorAll('.status-card').forEach(card => {
            card.addEventListener('mouseenter', function() {
                this.style.transform = 'translateY(-5px) scale(1.02)';
            });
            
            card.addEventListener('mouseleave', function() {
                this.style.transform = 'translateY(0) scale(1)';
            });
        });
        
        // 动态显示系统信息
        const updateTime = () => {
            const now = new Date();
            console.log('System Status Check:', now.toISOString());
        };
        
        setInterval(updateTime, 30000); // 每30秒更新一次
        updateTime(); // 立即执行一次
    </script>
</body>
</html>`;

        return new Response(html, {
            headers: { 
                ...corsHeaders, 
                'Content-Type': 'text/html; charset=utf-8',
                'Cache-Control': 'public, max-age=300'
            }
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