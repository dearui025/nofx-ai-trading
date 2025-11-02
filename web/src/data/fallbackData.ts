// 备用数据源 - 当Supabase连接失败时使用
export const fallbackCompetitionData = {
  traders: [
    {
      trader_id: 'ai_trader_1',
      trader_name: 'AlphaBot',
      ai_model: 'gpt-4',
      status: 'active',
      total_equity: 10250.75,
      daily_pnl: 250.75,
      daily_pnl_pct: 2.51,
      total_pnl: 1250.75,
      total_pnl_pct: 12.51,
      win_rate: 75.5,
      total_trades: 45,
      avg_trade_duration: '2.5h',
      max_drawdown: 5.2,
      sharpe_ratio: 1.8,
      last_trade_time: new Date().toISOString(),
      position_count: 3,
      margin_used_pct: 45.2,
      call_count: 156,
      is_running: true,
      unrealized_pnl: 125.30
    },
    {
      trader_id: 'ai_trader_2', 
      trader_name: 'BetaBot',
      ai_model: 'claude-3',
      status: 'active',
      total_equity: 9875.20,
      daily_pnl: -124.80,
      daily_pnl_pct: -1.25,
      total_pnl: -124.80,
      total_pnl_pct: -1.25,
      win_rate: 62.3,
      total_trades: 32,
      avg_trade_duration: '1.8h',
      max_drawdown: 8.1,
      sharpe_ratio: 1.2,
      last_trade_time: new Date().toISOString(),
      position_count: 2,
      margin_used_pct: 32.8,
      call_count: 89,
      is_running: true,
      unrealized_pnl: -75.40
    },
    {
      trader_id: 'ai_trader_3',
      trader_name: 'GammaBot', 
      ai_model: 'gemini-pro',
      status: 'inactive',
      total_equity: 9950.00,
      daily_pnl: -50.00,
      daily_pnl_pct: -0.50,
      total_pnl: -50.00,
      total_pnl_pct: -0.50,
      win_rate: 45.0,
      total_trades: 18,
      avg_trade_duration: '3.2h',
      max_drawdown: 12.5,
      sharpe_ratio: 0.8,
      last_trade_time: new Date().toISOString(),
      position_count: 1,
      margin_used_pct: 15.5,
      call_count: 23,
      is_running: false,
      unrealized_pnl: 25.10
    }
  ],
  count: 3
};

export const fallbackSystemStatus = {
  trader_id: 'system_trader_1',
  trader_name: 'SystemBot',
  ai_model: 'gpt-4',
  is_running: true,
  start_time: new Date(Date.now() - 3600000).toISOString(),
  runtime_minutes: 60,
  call_count: 156,
  initial_balance: 10000,
  scan_interval: '30s',
  stop_until: '',
  last_reset_time: new Date(Date.now() - 86400000).toISOString(),
  ai_provider: 'openai'
};

export const fallbackAccountInfo = {
  total_equity: 10250.75,
  wallet_balance: 10000.00,
  unrealized_profit: 250.75,
  available_balance: 9500.00,
  total_pnl: 250.75,
  total_pnl_pct: 2.51,
  total_unrealized_pnl: 125.30,
  initial_balance: 10000.00,
  daily_pnl: 250.75,
  position_count: 3,
  margin_used: 500.00,
  margin_used_pct: 5.0
};

export const fallbackAccountHistory = [
  {
    id: 'hist_1',
    timestamp: new Date(Date.now() - 3600000).toISOString(), // 1小时前
    equity: 10000,
    pnl: 0,
    pnl_pct: 0
  },
  {
    id: 'hist_2', 
    timestamp: new Date(Date.now() - 1800000).toISOString(), // 30分钟前
    equity: 10125.50,
    pnl: 125.50,
    pnl_pct: 1.26
  },
  {
    id: 'hist_3',
    timestamp: new Date().toISOString(), // 现在
    equity: 10250.75,
    pnl: 250.75,
    pnl_pct: 2.51
  }
];

// 生成模拟的历史数据点
export function generateFallbackEquityHistory(hours: number = 24): Array<{
  timestamp: string;
  equity: number;
  pnl: number;
  pnl_pct: number;
}> {
  const data = [];
  const now = Date.now();
  const baseEquity = 10000;
  
  for (let i = hours; i >= 0; i--) {
    const timestamp = new Date(now - (i * 60 * 60 * 1000)).toISOString();
    // 模拟一些波动
    const volatility = (Math.random() - 0.5) * 0.02; // ±1%的随机波动
    const trend = (hours - i) * 0.001; // 轻微上升趋势
    const pnl_pct = trend + volatility;
    const equity = baseEquity * (1 + pnl_pct);
    const pnl = equity - baseEquity;
    
    data.push({
      timestamp,
      equity: Math.round(equity * 100) / 100,
      pnl: Math.round(pnl * 100) / 100,
      pnl_pct: Math.round(pnl_pct * 10000) / 100
    });
  }
  
  return data;
}