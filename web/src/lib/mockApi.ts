import type {
  SystemStatus,
  AccountInfo,
  Position,
  DecisionRecord,
  Statistics,
  TraderInfo,
  CompetitionData,
} from '../types';

// Mock 数据生成器
const generateMockData = () => {
  const now = new Date();
  
  // Mock TraderInfo
  const mockTraders: TraderInfo[] = [
    {
      trader_id: 'qwen-trader-001',
      trader_name: 'Qwen Trader',
      ai_model: 'qwen',
    },
    {
      trader_id: 'deepseek-trader-001', 
      trader_name: 'DeepSeek Trader',
      ai_model: 'deepseek',
    },
  ];

  // Mock CompetitionData
  const mockCompetition: CompetitionData = {
    traders: mockTraders.map((trader) => ({
      ...trader,
      total_equity: 10000 + Math.random() * 5000,
      daily_pnl: (Math.random() - 0.5) * 500,
      daily_pnl_pct: (Math.random() - 0.5) * 5,
      total_pnl: (Math.random() - 0.5) * 2000,
      total_pnl_pct: (Math.random() - 0.5) * 20,
      win_rate: 50 + Math.random() * 40,
      total_trades: Math.floor(Math.random() * 100),
      avg_trade_duration: `${(Math.random() * 5).toFixed(1)}h`,
      max_drawdown: Math.random() * 15,
      sharpe_ratio: Math.random() * 3,
      last_trade_time: now.toISOString(),
      status: Math.random() > 0.3 ? 'active' : 'inactive',
      position_count: Math.floor(Math.random() * 5),
      margin_used_pct: Math.random() * 30,
      call_count: Math.floor(Math.random() * 100),
      is_running: true,
      unrealized_pnl: (Math.random() - 0.5) * 200,
    })),
    count: mockTraders.length,
  };

  // Mock SystemStatus
  const mockStatus: SystemStatus = {
    trader_id: 'qwen-trader-001',
    trader_name: 'Qwen Trader',
    ai_model: 'qwen',
    is_running: true,
    start_time: now.toISOString(),
    runtime_minutes: Math.floor(Math.random() * 1440), // 24小时内的分钟数
    call_count: Math.floor(Math.random() * 100),
    initial_balance: 10000,
    scan_interval: '15s',
    stop_until: '',
    last_reset_time: now.toISOString(),
    ai_provider: 'qwen',
  };

  // Mock AccountInfo
  const mockAccount: AccountInfo = {
    total_equity: 10000 + Math.random() * 5000,
    wallet_balance: 5000 + Math.random() * 3000,
    unrealized_profit: (Math.random() - 0.5) * 1000,
    available_balance: 5000 + Math.random() * 3000,
    total_pnl: (Math.random() - 0.5) * 2000,
    total_pnl_pct: (Math.random() - 0.5) * 20,
    total_unrealized_pnl: (Math.random() - 0.5) * 1000,
    initial_balance: 10000,
    daily_pnl: (Math.random() - 0.5) * 500,
    position_count: Math.floor(Math.random() * 5),
    margin_used: Math.random() * 2000,
    margin_used_pct: Math.random() * 30,
  };

  // Mock Positions
  const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'ADAUSDT', 'SOLUSDT'];
  const mockPositions: Position[] = Array.from({ length: Math.floor(Math.random() * 3) }, () => ({
    symbol: symbols[Math.floor(Math.random() * symbols.length)],
    side: Math.random() > 0.5 ? 'long' : 'short',
    entry_price: 100 + Math.random() * 90000,
    mark_price: 100 + Math.random() * 90000,
    quantity: 0.1 + Math.random() * 10,
    leverage: 1 + Math.floor(Math.random() * 10),
    unrealized_pnl: (Math.random() - 0.5) * 500,
    unrealized_pnl_pct: (Math.random() - 0.5) * 10,
    liquidation_price: 50 + Math.random() * 95000,
    margin_used: Math.random() * 2000,
  }));

  // Mock DecisionRecord
  const mockDecisions: DecisionRecord[] = Array.from({ length: 5 }, (_, i) => ({
    timestamp: new Date(now.getTime() - i * 300000).toISOString(), // 每5分钟一个决策
    cycle_number: i + 1,
    input_prompt: `分析当前市场趋势，建议${symbols[Math.floor(Math.random() * symbols.length)]}的交易策略...`,
    cot_trace: `基于技术分析，我观察到价格突破关键阻力位，建议${Math.random() > 0.5 ? '做多' : '做空'}...`,
    decision_json: JSON.stringify({ action: 'analyze', result: 'bullish' }),
    account_state: {
      total_balance: 10000 + Math.random() * 5000,
      available_balance: 5000 + Math.random() * 3000,
      total_unrealized_profit: (Math.random() - 0.5) * 1000,
      position_count: Math.floor(Math.random() * 5),
      margin_used_pct: Math.random() * 30,
    },
    positions: [],
    candidate_coins: symbols.slice(0, 3),
    decisions: [
      {
        action: Math.random() > 0.5 ? 'open_long' : 'open_short',
        symbol: symbols[Math.floor(Math.random() * symbols.length)],
        quantity: 0.1 + Math.random() * 2,
        leverage: 1 + Math.floor(Math.random() * 5),
        price: 100 + Math.random() * 90000,
        order_id: Math.floor(Math.random() * 1000000),
        timestamp: new Date(now.getTime() - i * 300000).toISOString(),
        success: Math.random() > 0.2,
        error: Math.random() > 0.8 ? '网络超时' : undefined,
      }
    ],
    execution_log: [
      '✓ 成功连接交易所API',
      '✓ 获取市场数据',
      Math.random() > 0.5 ? '✓ 订单执行成功' : '✗ 订单执行失败：价格变动',
    ],
    success: Math.random() > 0.3,
    error_message: Math.random() > 0.8 ? '网络连接超时，请重试' : undefined,
  }));

  // Mock Statistics
  const mockStats: Statistics = {
    total_cycles: Math.floor(Math.random() * 100),
    successful_cycles: Math.floor(Math.random() * 60),
    failed_cycles: Math.floor(Math.random() * 40),
    total_open_positions: Math.floor(Math.random() * 10),
    total_close_positions: Math.floor(Math.random() * 50),
  };

  return {
    mockTraders,
    mockCompetition,
    mockStatus,
    mockAccount,
    mockPositions,
    mockDecisions,
    mockStats,
  };
};

const { 
  mockTraders, 
  mockCompetition, 
  mockStatus, 
  mockAccount, 
  mockPositions, 
  mockDecisions, 
  mockStats 
} = generateMockData();

// Mock API 实现
export const mockApi = {
  // 竞赛相关接口
  async getCompetition(): Promise<CompetitionData> {
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 1000));
    return mockCompetition;
  },

  async getTraders(): Promise<TraderInfo[]> {
    await new Promise(resolve => setTimeout(resolve, 300 + Math.random() * 700));
    return mockTraders;
  },

  async getStatus(_traderId?: string): Promise<SystemStatus> {
    await new Promise(resolve => setTimeout(resolve, 200 + Math.random() * 500));
    return mockStatus;
  },

  async getAccount(_traderId?: string): Promise<AccountInfo> {
    await new Promise(resolve => setTimeout(resolve, 400 + Math.random() * 800));
    return mockAccount;
  },

  async getPositions(_traderId?: string): Promise<Position[]> {
    await new Promise(resolve => setTimeout(resolve, 300 + Math.random() * 600));
    return mockPositions;
  },

  async getDecisions(_traderId?: string): Promise<DecisionRecord[]> {
    await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 1000));
    return mockDecisions;
  },

  async getLatestDecisions(_traderId?: string): Promise<DecisionRecord[]> {
    await new Promise(resolve => setTimeout(resolve, 400 + Math.random() * 800));
    return mockDecisions.slice(0, 3); // 只返回最新的3个决策
  },

  async getStatistics(_traderId?: string): Promise<Statistics> {
    await new Promise(resolve => setTimeout(resolve, 600 + Math.random() * 1200));
    return mockStats;
  },

  async getEquityHistory(_traderId?: string): Promise<any[]> {
    await new Promise(resolve => setTimeout(resolve, 800 + Math.random() * 1500));
    // 生成历史数据
    const history = [];
    const now = new Date();
    for (let i = 30; i >= 0; i--) {
      history.push({
        timestamp: new Date(now.getTime() - i * 24 * 60 * 60 * 1000).toISOString(),
        equity: 10000 + Math.random() * 5000 + i * 10,
        pnl: (Math.random() - 0.5) * 2000 + i * 5,
        pnl_pct: (Math.random() - 0.5) * 20 + i * 0.1,
      });
    }
    return history;
  },

  async getPerformance(_traderId?: string): Promise<any> {
    await new Promise(resolve => setTimeout(resolve, 700 + Math.random() * 1400));
    return {
      learning_progress: Math.random() * 100,
      adaptation_score: Math.random() * 100,
      risk_adjustment: Math.random() * 100,
      market_awareness: Math.random() * 100,
      recent_improvements: [
        '优化了止损策略',
        '改进了趋势识别算法',
        '增强了风险控制机制',
      ],
      performance_trends: {
        accuracy_improvement: Math.random() * 20,
        risk_reduction: Math.random() * 15,
        profit_optimization: Math.random() * 25,
      },
    };
  },
};