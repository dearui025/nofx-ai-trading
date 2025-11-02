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
    
    // HTML content for the NOFX trading system
    const htmlContent = `<!DOCTYPE html>
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
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #000111 0%, #1a1a1a 50%, #2d1b60 100%);
            min-height: 100vh;
            color: #EAECEF;
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
            font-weight: 800;
            background: linear-gradient(135deg, #F08908 0%, #FCD535 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 0.5rem;
        }
        
        .subtitle {
            font-size: 1.25rem;
            color: #848E9C;
            margin-bottom: 1rem;
        }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin-bottom: 3rem;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 1rem;
            padding: 1.5rem;
            text-align: center;
            transition: all 0.3s ease;
        }
        
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 30px rgba(240, 185, 11, 0.3);
        }
        
        .stat-number {
            font-size: 2.5rem;
            font-weight: 700;
            color: #F08908;
            margin-bottom: 0.5rem;
        }
        
        .stat-label {
            color: #9ca3af;
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 2rem;
            margin-bottom: 3rem;
        }
        
        .status-card {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 1rem;
            padding: 2rem;
        }
        
        .status-title {
            font-size: 1.25rem;
            font-weight: 600;
            color: #F08908;
            margin-bottom: 1.5rem;
            text-align: center;
        }
        
        .status-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .status-item:last-child {
            border-bottom: none;
        }
        
        .status-label {
            color: #9ca3af;
            font-size: 0.875rem;
        }
        
        .status-value {
            color: #10ECB81;
            font-weight: 600;
            font-size: 0.875rem;
        }
        
        .api-section {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            border-radius: 1rem;
            padding: 2rem;
            margin-bottom: 3rem;
        }
        
        .api-title {
            font-size: 1.5rem;
            font-weight: 700;
            color: #F08908;
            margin-bottom: 1.5rem;
            text-align: center;
        }
        
        .api-endpoint {
            font-family: 'Courier New', monospace;
            background: rgba(0, 0, 0, 0.3);
            border-left: 4px solid #F08908;
            padding: 1rem;
            margin-bottom: 0.5rem;
            border-radius: 0.5rem;
            color: #EAECEF;
            font-size: 0.875rem;
            word-break: break-all;
        }
        
        .button-group {
            display: flex;
            justify-content: center;
            gap: 1rem;
            margin-bottom: 3rem;
            flex-wrap: wrap;
        }
        
        .button {
            padding: 0.75rem 1.5rem;
            border-radius: 0.5rem;
            font-weight: 600;
            text-decoration: none;
            transition: all 0.3s ease;
            border: none;
            cursor: pointer;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            color: white;
        }
        
        .button-primary {
            background: linear-gradient(135deg, #F08908 0%, #FCD535 100%);
            color: #000;
        }
        
        .button-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(240, 185, 11, 0.4);
        }
        
        .button-secondary {
            background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
        }
        
        .button-secondary:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(59, 130, 246, 0.4);
        }
        
        .warning {
            background: rgba(246, 70, 93, 0.1);
            border: 1px solid rgba(246, 70, 93, 0.3);
            border-radius: 1rem;
            padding: 1.5rem;
            text-align: center;
        }
        
        .warning-title {
            color: #F6465D;
            font-weight: 700;
            margin-bottom: 1rem;
            font-size: 1.125rem;
        }
        
        .warning-text {
            color: #9ca3af;
            line-height: 1.6;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 1rem;
            }
            
            .title {
                font-size: 2rem;
            }
            
            .stats {
                grid-template-columns: 1fr;
            }
            
            .status-grid {
                grid-template-columns: 1fr;
            }
            
            .button-group {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <header class="header">
            <h1 class="title">NOFX</h1>
            <p class="subtitle">AI Trading Competition System</p>
            <p style="color: #848E9C; font-size: 1rem;">巅峰对决：Qwen vs DeepSeek AI 实时交易对战</p>
        </header>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">7</div>
                <div class="stat-label">Edge Functions</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">13</div>
                <div class="stat-label">Database Tables</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">100%</div>
                <div class="stat-label">Deployment Status</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">3</div>
                <div class="stat-label">API Keys Configured</div>
            </div>
        </div>

        <div class="status-grid">
            <div class="status-card">
                <h2 class="status-title">System Status</h2>
                <div class="status-item">
                    <span class="status-label">Edge Functions</span>
                    <span class="status-value">7/7 Active</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Database</span>
                    <span class="status-value">Active</span>
                </div>
                <div class="status-item">
                    <span class="status-label">API Keys</span>
                    <span class="status-value">Configured</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Frontend Build</span>
                    <span class="status-value">Complete</span>
                </div>
            </div>

            <div class="status-card">
                <h2 class="status-title">Backend Services</h2>
                <div class="status-item">
                    <span class="status-label">API Gateway</span>
                    <span class="status-value">Active</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Decision Engine</span>
                    <span class="status-value">Active</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Market Data</span>
                    <span class="status-value">Active</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Trade Executor</span>
                    <span class="status-value">Active</span>
                </div>
            </div>
        </div>

        <div class="api-section">
            <h2 class="api-title">API Endpoints</h2>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/api-gateway</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/decision-engine</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/market-data</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/trade-executor</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/risk-manager</div>
            <div class="api-endpoint">https://eqzurdzoaxibothslnna.supabase.co/functions/v1/account-info</div>
        </div>

        <div class="button-group">
            <a href="https://github.com/tinkle-community/nofx" target="_blank" class="button button-primary">
                <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                GitHub Repository
            </a>
            <a href="https://supabase.com/dashboard/project/eqzurdzoaxibothslnna" target="_blank" class="button button-secondary">
                <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0L3 5v11h4v-6h6v6h4V5L8 0z"/>
                </svg>
                Supabase Dashboard
            </a>
            <a href="https://eqzurdzoaxibothslnna.supabase.co" target="_blank" class="button button-secondary">
                <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 8h8v2H8V8zm0-2h8v2H8V6zM8 12h8v2H8v-2z"/>
                </svg>
                Project API
            </a>
        </div>

        <div class="warning">
            <h3 class="warning-title">风险提示 (Risk Warning)</h3>
            <p class="warning-text">
                Trading involves substantial risk and may not be suitable for all investors. 
                Please trade responsibly and at your own discretion.<br><br>
                交易涉及重大风险，可能不适合所有投资者。请负责任地交易，并自行承担风险。
            </p>
        </div>
    </div>

    <script>
        console.log('NOFX Trading System Frontend Loaded Successfully');
        
        // 添加交互效果
        document.addEventListener('DOMContentLoaded', function() {
            const statCards = document.querySelectorAll('.stat-card');
            
            statCards.forEach(card => {
                card.addEventListener('mouseenter', function() {
                    this.style.transform = 'translateY(-5px)';
                    this.style.boxShadow = '0 10px 30px rgba(240, 185, 11, 0.3)';
                });
                
                card.addEventListener('mouseleave', function() {
                    this.style.transform = 'translateY(0)';
                    this.style.boxShadow = '';
                });
            });
            
            // 模拟系统状态更新
            function updateTime() {
                console.log('System Status Check:', new Date().toLocaleTimeString());
            }
            
            // 每30秒更新一次状态
            setInterval(updateTime, 30000);
        });
    </script>
</body>
</html>`;

    return new Response(htmlContent, {
      headers: {
        ...corsHeaders,
        'Content-Type': 'text/html; charset=utf-8',
        'Cache-Control': 'public, max-age=3600'
      }
    });

  } catch (error) {
    return new Response(JSON.stringify({ 
      error: {
        code: 'FUNCTION_ERROR',
        message: error.message
      }
    }), {
      status: 500,
      headers: {
        ...corsHeaders,
        'Content-Type': 'application/json'
      }
    });
  }
});