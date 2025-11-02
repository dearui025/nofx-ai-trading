import { goApiClient } from './goApiClient';
import { 
  fallbackCompetitionData, 
  fallbackSystemStatus, 
  fallbackAccountInfo,
  generateFallbackEquityHistory 
} from '../data/fallbackData';
import type {
  SystemStatus,
  AccountInfo,
  Position,
  DecisionRecord,
  Statistics,
  TraderInfo,
  CompetitionData,
} from '../types';

// 实时市场数据类型
interface MarketData {
  symbol: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  volume24h: number;
}

// 市场数据缓存
const marketDataCache = {
  data: null as MarketData[] | null,
  timestamp: 0,
  ttl: 30000, // 30秒缓存
};

// 缓存市场数据
function cacheMarketData(data: MarketData[]): void {
  marketDataCache.data = data;
  marketDataCache.timestamp = Date.now();
}

// 获取缓存的市场数据
function getCachedMarketData(): MarketData[] | null {
  const now = Date.now();
  if (marketDataCache.data && (now - marketDataCache.timestamp) < marketDataCache.ttl) {
    return marketDataCache.data;
  }
  return null;
}

// 通用缓存系统
const cache = new Map<string, { data: any; timestamp: number; ttl: number }>();

function getCachedData(key: string): any | null {
  const cached = cache.get(key);
  if (cached && (Date.now() - cached.timestamp) < cached.ttl) {
    return cached.data;
  }
  return null;
}

function setCachedData(key: string, data: any, ttl: number): void {
  cache.set(key, {
    data,
    timestamp: Date.now(),
    ttl
  });
}

// 生成增强的模拟市场数据
function generateEnhancedMockMarketData(): MarketData[] {
  const symbols = ['BTCUSDT', 'ETHUSDT', 'SOLUSDT', 'BNBUSDT', 'XRPUSDT', 'DOGEUSDT', 'ADAUSDT', 'HYPEUSDT'];
  const basePrices: Record<string, number> = {
    'BTCUSDT': 69000,
    'ETHUSDT': 2500,
    'SOLUSDT': 180,
    'BNBUSDT': 600,
    'XRPUSDT': 0.55,
    'DOGEUSDT': 0.15,
    'ADAUSDT': 0.45,
    'HYPEUSDT': 25,
  };

  return symbols.map(symbol => {
    const basePrice = basePrices[symbol] || 100;
    const changePercent = (Math.random() - 0.5) * 20;
    const price = basePrice * (1 + changePercent / 100);
    
    return {
      symbol,
      price: parseFloat(price.toFixed(symbol.includes('USDT') && price < 1 ? 6 : 2)),
      change24h: parseFloat((price * changePercent / 100).toFixed(6)),
      changePercent24h: parseFloat(changePercent.toFixed(2)),
      volume24h: Math.floor(Math.random() * 1000000000),
    };
  });
}

// Go后端API实现
export const api = {
  // 竞赛相关接口
  async getCompetition(): Promise<CompetitionData> {
    try {
      const competitionData = await goApiClient.getCompetition();
      return competitionData;
    } catch (error) {
      console.error('获取竞赛数据失败:', error);
      return fallbackCompetitionData;
    }
  },

  async getTraders(): Promise<TraderInfo[]> {
    try {
      const traders = await goApiClient.getTraders();
      return traders.map((trader: any) => ({
        trader_id: trader.trader_id,
        name: trader.trader_name,
        ai_model: trader.ai_model,
        status: 'active',
        created_at: new Date().toISOString(),
      }));
    } catch (error) {
      console.error('获取trader列表失败:', error);
      return [
        {
          trader_id: 'binance_qwen',
          name: 'Binance Qwen Trader (Testnet)',
          ai_model: 'deepseek',
          status: 'active',
          created_at: new Date().toISOString(),
        }
      ];
    }
  },

  async getStatus(traderId?: string): Promise<SystemStatus> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const statusData = await goApiClient.getStatus(actualTraderId);
      
      return {
        trader_id: actualTraderId,
        status: statusData.status || 'active',
        uptime: statusData.uptime || '00:00:00',
        last_decision_time: statusData.last_decision_time || new Date().toISOString(),
        total_cycles: statusData.total_cycles || 0,
        successful_cycles: statusData.successful_cycles || 0,
        success_rate: statusData.success_rate || 0,
        ai_model: statusData.ai_model || 'deepseek',
        current_strategy: statusData.current_strategy || 'AI自主决策',
        risk_level: statusData.risk_level || 'medium',
        next_scan_time: statusData.next_scan_time || new Date(Date.now() + 180000).toISOString(),
      };
    } catch (error) {
      console.error('获取状态失败:', error);
      return fallbackSystemStatus;
    }
  },

  async getAccount(traderId?: string): Promise<AccountInfo> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const accountData = await goApiClient.getAccount(actualTraderId);
      
      return {
        total_equity: accountData.total_equity || 0,
        available_balance: accountData.available_balance || 0,
        margin_used: accountData.margin_used || 0,
        margin_used_pct: accountData.margin_used_pct || 0,
        total_pnl: accountData.total_pnl || 0,
        total_pnl_pct: accountData.total_pnl_pct || 0,
        daily_pnl: accountData.daily_pnl || 0,
        total_unrealized_pnl: accountData.total_unrealized_pnl || 0,
        unrealized_pnl_pct: accountData.unrealized_pnl_pct || 0,
        position_count: accountData.position_count || 0,
        initial_balance: accountData.initial_balance || 5000,
      };
    } catch (error) {
      console.error('获取账户信息失败:', error);
      return fallbackAccountInfo;
    }
  },

  async getEquityHistory(traderId?: string): Promise<any[]> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const historyData = await goApiClient.getEquityHistory(actualTraderId);
      
      if (!historyData || historyData.length === 0) {
        return generateFallbackEquityHistory();
      }
      
      return historyData.map((item: any) => ({
        timestamp: item.timestamp,
        total_value: item.total_value || item.total_equity,
        realized_pnl: item.realized_pnl || 0,
        unrealized_pnl: item.unrealized_pnl || 0,
        return_percent: item.return_percent || 0,
      }));
    } catch (error) {
      console.error('获取权益历史失败:', error);
      return generateFallbackEquityHistory();
    }
  },

  async getPositions(traderId?: string): Promise<Position[]> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const positionsData = await goApiClient.getPositions(actualTraderId);
      
      if (!positionsData || positionsData.length === 0) {
        return [];
      }
      
      return positionsData.map((pos: any) => ({
        symbol: pos.symbol,
        side: pos.side,
        quantity: parseFloat(pos.quantity || pos.size || '0'),
        entry_price: parseFloat(pos.entry_price || pos.avgPrice || '0'),
        mark_price: parseFloat(pos.mark_price || pos.markPrice || '0'),
        unrealized_pnl: parseFloat(pos.unrealized_pnl || pos.pnl || '0'),
        unrealized_pnl_pct: parseFloat(pos.unrealized_pnl_pct || pos.pnl_pct || '0'),
        liquidation_price: parseFloat(pos.liquidation_price || '0'),
        margin: parseFloat(pos.margin || '0'),
        leverage: parseFloat(pos.leverage || '1'),
        timestamp: pos.timestamp || new Date().toISOString(),
      }));
    } catch (error) {
      console.error('获取持仓信息失败:', error);
      return [];
    }
  },

  async getDecisions(traderId?: string): Promise<DecisionRecord[]> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const decisionsData = await goApiClient.getDecisions(actualTraderId);
      
      if (!decisionsData || decisionsData.length === 0) {
        return [];
      }
      
      return decisionsData.map((decision: any) => {
        // 解析决策JSON数据
        let parsedDecisions = [];
        try {
          if (typeof decision.decision_json === 'string') {
            parsedDecisions = JSON.parse(decision.decision_json);
          } else if (Array.isArray(decision.decision_json)) {
            parsedDecisions = decision.decision_json;
          }
        } catch (e) {
          console.warn('解析决策JSON失败:', e);
          parsedDecisions = [];
        }

        // 映射决策动作数据
        const mappedDecisions = (decision.decisions || parsedDecisions || []).map((d: any) => ({
          action: d.action || 'hold',
          symbol: d.symbol || '',
          quantity: d.quantity || 0,
          leverage: d.leverage || 0,
          price: d.price || 0,
          order_id: d.order_id || 0,
          timestamp: d.timestamp || decision.timestamp,
          success: d.success !== undefined ? d.success : true,
          error: d.error || '',
        }));

        return {
          timestamp: decision.timestamp,
          cycle_number: decision.cycle_number || 0,
          input_prompt: decision.input_prompt || '',
          cot_trace: decision.cot_trace || '',
          decision_json: decision.decision_json || JSON.stringify(parsedDecisions),
          account_state: decision.account_state || {
            total_balance: 0,
            available_balance: 0,
            total_unrealized_profit: 0,
            position_count: 0,
            margin_used_pct: 0,
          },
          positions: decision.positions || [],
          candidate_coins: decision.candidate_coins || [],
          decisions: mappedDecisions,
          execution_log: decision.execution_log || [],
          success: decision.success !== undefined ? decision.success : true,
          error_message: decision.error_message || '',
        };
      });
    } catch (error) {
      console.error('获取决策记录失败:', error);
      return [];
    }
  },

  async getLatestDecisions(traderId?: string): Promise<DecisionRecord[]> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const decisionsData = await goApiClient.getLatestDecisions(actualTraderId);
      
      if (!decisionsData || decisionsData.length === 0) {
        return [];
      }
      
      return decisionsData.slice(0, 3).map((decision: any) => {
        // 解析决策JSON数据
        let parsedDecisions = [];
        try {
          if (typeof decision.decision_json === 'string') {
            parsedDecisions = JSON.parse(decision.decision_json);
          } else if (Array.isArray(decision.decision_json)) {
            parsedDecisions = decision.decision_json;
          }
        } catch (e) {
          console.warn('解析决策JSON失败:', e);
          parsedDecisions = [];
        }

        // 映射决策动作数据
        const mappedDecisions = (decision.decisions || parsedDecisions || []).map((d: any) => ({
          action: d.action || 'hold',
          symbol: d.symbol || '',
          quantity: d.quantity || 0,
          leverage: d.leverage || 0,
          price: d.price || 0,
          order_id: d.order_id || 0,
          timestamp: d.timestamp || decision.timestamp,
          success: d.success !== undefined ? d.success : true,
          error: d.error || '',
        }));

        return {
          timestamp: decision.timestamp,
          cycle_number: decision.cycle_number || 0,
          input_prompt: decision.input_prompt || '',
          cot_trace: decision.cot_trace || '',
          decision_json: decision.decision_json || JSON.stringify(parsedDecisions),
          account_state: decision.account_state || {
            total_balance: 0,
            available_balance: 0,
            total_unrealized_profit: 0,
            position_count: 0,
            margin_used_pct: 0,
          },
          positions: decision.positions || [],
          candidate_coins: decision.candidate_coins || [],
          decisions: mappedDecisions,
          execution_log: decision.execution_log || [],
          success: decision.success !== undefined ? decision.success : true,
          error_message: decision.error_message || '',
        };
      });
    } catch (error) {
      console.error('获取最新决策失败:', error);
      return [];
    }
  },

  async getStatistics(traderId?: string): Promise<Statistics> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const statsData = await goApiClient.getStatistics(actualTraderId);
      
      return {
        total_cycles: statsData.total_cycles || 0,
        successful_cycles: statsData.successful_cycles || 0,
        success_rate: statsData.success_rate || 0,
        total_open_positions: statsData.total_open_positions || 0,
        total_close_positions: statsData.total_close_positions || 0,
        avg_holding_time: statsData.avg_holding_time || '0h 0m',
        best_trade: statsData.best_trade || 0,
        worst_trade: statsData.worst_trade || 0,
        win_rate: statsData.win_rate || 0,
        profit_factor: statsData.profit_factor || 0,
        sharpe_ratio: statsData.sharpe_ratio || 0,
        max_drawdown: statsData.max_drawdown || 0,
      };
    } catch (error) {
      console.error('获取统计信息失败:', error);
      return {
        total_cycles: 0,
        successful_cycles: 0,
        success_rate: 0,
        total_open_positions: 0,
        total_close_positions: 0,
        avg_holding_time: '0h 0m',
        best_trade: 0,
        worst_trade: 0,
        win_rate: 0,
        profit_factor: 0,
        sharpe_ratio: 0,
        max_drawdown: 0,
      };
    }
  },

  async getPerformance(traderId?: string): Promise<any> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const performanceData = await goApiClient.getPerformance(actualTraderId);
      
      return {
        total_return: performanceData.total_return || 0,
        daily_return: performanceData.daily_return || 0,
        weekly_return: performanceData.weekly_return || 0,
        monthly_return: performanceData.monthly_return || 0,
        volatility: performanceData.volatility || 0,
        sharpe_ratio: performanceData.sharpe_ratio || 0,
        max_drawdown: performanceData.max_drawdown || 0,
        win_rate: performanceData.win_rate || 0,
        profit_factor: performanceData.profit_factor || 0,
        total_trades: performanceData.total_trades || 0,
        winning_trades: performanceData.winning_trades || 0,
        losing_trades: performanceData.losing_trades || 0,
        avg_win: performanceData.avg_win || 0,
        avg_loss: performanceData.avg_loss || 0,
        largest_win: performanceData.largest_win || 0,
        largest_loss: performanceData.largest_loss || 0,
        consecutive_wins: performanceData.consecutive_wins || 0,
        consecutive_losses: performanceData.consecutive_losses || 0,
        market_awareness: Math.min((performanceData.total_trades || 0) * 5, 100),
        risk_management: performanceData.risk_management || 75,
        adaptability: performanceData.adaptability || 80,
        learning_progress: performanceData.total_trades > 10 ? '增强了市场分析能力' : '正在学习市场模式',
        // 添加历史成交记录和币种统计数据
        recent_trades: performanceData.recent_trades || [],
        symbol_stats: performanceData.symbol_stats || {},
        best_symbol: performanceData.best_symbol || '',
        worst_symbol: performanceData.worst_symbol || '',
      };
    } catch (error) {
      console.error('获取性能数据失败:', error);
      return {
        total_return: 0,
        daily_return: 0,
        weekly_return: 0,
        monthly_return: 0,
        volatility: 0,
        sharpe_ratio: 0,
        max_drawdown: 0,
        win_rate: 0,
        profit_factor: 0,
        total_trades: 0,
        winning_trades: 0,
        losing_trades: 0,
        avg_win: 0,
        avg_loss: 0,
        largest_win: 0,
        largest_loss: 0,
        consecutive_wins: 0,
        consecutive_losses: 0,
        market_awareness: 0,
        risk_management: 75,
        adaptability: 80,
        learning_progress: '正在学习市场模式',
        // 添加历史成交记录和币种统计数据的默认值
        recent_trades: [],
        symbol_stats: {},
        best_symbol: '',
        worst_symbol: '',
      };
    }
  },

  // 市场数据相关（保持原有的mock实现）
  async getTopGainers(): Promise<MarketDataResponse[]> {
    const cacheKey = 'topGainers';
    const cached = getCachedData(cacheKey);
    
    if (cached) {
      return cached;
    }
  
    try {
      const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'ADAUSDT', 'SOLUSDT', 'XRPUSDT'];
      const marketDataPromises = symbols.map(symbol => 
        goApiClient.getMarketData(symbol)
      );
      
      const results = await Promise.allSettled(marketDataPromises);
      const marketData = results
        .filter((result): result is PromiseFulfilledResult<any> => result.status === 'fulfilled')
        .map(result => ({
          symbol: result.value.symbol,
          price: result.value.price,
          change24h: result.value.change24h,
          changePercent24h: result.value.changePercent24h,
          volume24h: result.value.volume24h
        }));
  
      setCachedData(cacheKey, marketData, 30000); // Cache for 30 seconds
      return marketData;
    } catch (error) {
      console.error('[API] Error fetching top gainers:', error);
      return [];
    }
  },

  async getMarketPrice(symbol: string): Promise<number> {
    try {
      const marketData = await this.getTopGainers();
      const symbolData = marketData.find(item => item.symbol === symbol);
      return symbolData?.price || 0;
    } catch (error) {
      console.error(`获取${symbol}价格失败:`, error);
      return 0;
    }
  },

  // 频率管理相关API
  async getFrequencyStatus(traderId?: string): Promise<any> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const frequencyData = await goApiClient.getFrequencyStatus(actualTraderId);
      
      return {
        enabled: frequencyData.enabled || false,
        current_mode: frequencyData.current_mode || 'basic',
        daily_pnl_percent: frequencyData.daily_pnl_percent || 0,
        hourly_trade_count: frequencyData.hourly_trade_count || 0,
        daily_trade_count: frequencyData.daily_trade_count || 0,
        current_limits: frequencyData.current_limits || {
          hourly_limit: 2,
          daily_limit: 10,
        },
        next_mode_threshold: frequencyData.next_mode_threshold || 2.0,
        time_to_hourly_reset: frequencyData.time_to_hourly_reset || '0h 0m',
        last_mode_switch: frequencyData.last_mode_switch || null,
        config: frequencyData.config || {
          basic_mode: { hourly_limit: 2, daily_limit: 10 },
          elastic_mode: { hourly_limit: 5, daily_limit: -1 },
          absolute_hourly_max: 6,
          upgrade_threshold: 2.0,
          downgrade_threshold: 1.0,
        },
        error: frequencyData.error || null,
      };
    } catch (error) {
      console.error('获取频率管理器状态失败:', error);
      return {
        enabled: false,
        error: '无法连接到频率管理器',
        current_mode: 'basic',
        daily_pnl_percent: 0,
        hourly_trade_count: 0,
        daily_trade_count: 0,
        current_limits: { hourly_limit: 2, daily_limit: 10 },
        next_mode_threshold: 2.0,
        time_to_hourly_reset: '0h 0m',
        last_mode_switch: null,
        config: {
          basic_mode: { hourly_limit: 2, daily_limit: 10 },
          elastic_mode: { hourly_limit: 5, daily_limit: -1 },
          absolute_hourly_max: 6,
          upgrade_threshold: 2.0,
          downgrade_threshold: 1.0,
        },
      };
    }
  },

  async updateFrequencyConfig(config: any, traderId?: string): Promise<any> {
    try {
      const actualTraderId = traderId || 'binance_qwen';
      const result = await goApiClient.updateFrequencyConfig(config, actualTraderId);
      
      return {
        success: result.success || false,
        message: result.message || '配置更新成功',
        config: result.config || config,
      };
    } catch (error) {
      console.error('更新频率管理器配置失败:', error);
      return {
        success: false,
        message: `更新配置失败: ${error}`,
        config: null,
      };
    }
  },

  // AI优化数据相关API
  async getAIOptimizationData(): Promise<any> {
    try {
      // 尝试从后端获取真实数据
      const response = await goApiClient.getAIOptimizationData();
      return response;
    } catch (error) {
      console.error('获取AI优化数据失败:', error);
      // 返回模拟数据
      return {
        market_regime: {
          regime: 'sideways',
          confidence: 0.75,
          last_updated: new Date().toISOString()
        },
        signal_strengths: [
          { indicator: 'RSI', strength: 0.65, direction: 'bullish', confidence: 0.8 },
          { indicator: 'MACD', strength: 0.45, direction: 'bearish', confidence: 0.7 },
          { indicator: 'Bollinger Bands', strength: 0.55, direction: 'neutral', confidence: 0.6 },
          { indicator: 'Moving Average', strength: 0.75, direction: 'bullish', confidence: 0.85 }
        ],
        correlation_risks: [
          { symbol1: 'BTCUSDT', symbol2: 'ETHUSDT', correlation: 0.85, risk_level: 'high' },
          { symbol1: 'BTCUSDT', symbol2: 'ADAUSDT', correlation: 0.65, risk_level: 'medium' },
          { symbol1: 'ETHUSDT', symbol2: 'ADAUSDT', correlation: 0.55, risk_level: 'medium' }
        ],
        system_health: {
          status: 'healthy',
          uptime: 99.8,
          last_backup: new Date(Date.now() - 3600000).toISOString(),
          recovery_ready: true
        }
      };
    }
  },
};
