// 前端API客户端配置
// 用于与Supabase Edge Functions通信

export const SUPABASE_CONFIG = {
  url: 'https://eqzurdzoaxibothslnna.supabase.co',
  anonKey: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVxenVyZHpvYXhpYm90aHNsbm5hIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE4NzY2NjUsImV4cCI6MjA3NzQ1MjY2NX0.h2EQOkofLavh-DL68AGfFX7ZvJ4SipNsiO7K5uTh20Y',
};

// Edge Functions端点
export const API_ENDPOINTS = {
  trading: `${SUPABASE_CONFIG.url}/functions/v1/binance-trading`,
  marketData: `${SUPABASE_CONFIG.url}/functions/v1/market-data`,
  tradingCron: `${SUPABASE_CONFIG.url}/functions/v1/trading-cron`,
};

// API客户端类
export class NOFXApiClient {
  constructor() {
    // 初始化Supabase配置
  }

  // 通用请求方法
  private async request(endpoint: string, data: any) {
    console.log(`🔍 API请求: ${endpoint}`, data);
    
    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${SUPABASE_CONFIG.anonKey}`,
          'apikey': SUPABASE_CONFIG.anonKey,
        },
        body: JSON.stringify(data),
      });

      console.log(`📡 API响应状态: ${response.status} ${response.statusText}`);

      if (!response.ok) {
        const errorText = await response.text();
        console.error(`❌ API请求失败: ${response.status}`, errorText);
        
        try {
          const error = JSON.parse(errorText);
          throw new Error(error.error?.message || error.message || '请求失败');
        } catch (parseError) {
          throw new Error(`请求失败: ${response.status} ${response.statusText} - ${errorText}`);
        }
      }

      const result = await response.json();
      console.log(`✅ API响应成功:`, result);
      return result;
    } catch (error) {
      console.error(`❌ API请求异常:`, error);
      throw error;
    }
  }

  // 市场数据API
  async getTopGainers() {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getTopGainers',
      params: {},
    });
  }

  async getMarketPrice(symbol: string) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getMarketPrice',
      params: { symbol },
    });
  }

  async getKlines(symbol: string, interval: string = '1h', limit: number = 100) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getKlines',
      params: { symbol, interval, limit },
    });
  }

  async get24hTicker(symbols: string[]) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'get24hTicker',
      params: { symbols },
    });
  }

  async getFundingRate(symbol: string) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getFundingRate',
      params: { symbol },
    });
  }

  async getOpenInterest(symbol: string) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getOpenInterest',
      params: { symbol },
    });
  }

  async getMarketDepth(symbol: string, limit: number = 20) {
    return this.request(API_ENDPOINTS.marketData, {
      action: 'getMarketDepth',
      params: { symbol, limit },
    });
  }

  // 交易API（需要正确的Testnet密钥）
  async getBalance(traderId: string = 'binance_testnet') {
    return this.request(API_ENDPOINTS.trading, {
      action: 'getBalance',
      params: { trader_id: traderId },
    });
  }

  async getPositions() {
    return this.request(API_ENDPOINTS.trading, {
      action: 'getPositions',
      params: {},
    });
  }

  async setLeverage(symbol: string, leverage: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'setLeverage',
      params: { symbol, leverage },
    });
  }

  async openLong(symbol: string, quantity: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'openLong',
      params: {
        symbol,
        quantity,
        side: 'BUY',
        positionSide: 'LONG',
      },
    });
  }

  async openShort(symbol: string, quantity: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'openShort',
      params: {
        symbol,
        quantity,
        side: 'SELL',
        positionSide: 'SHORT',
      },
    });
  }

  async closeLong(symbol: string, quantity: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'closeLong',
      params: {
        symbol,
        quantity,
        side: 'SELL',
        positionSide: 'LONG',
      },
    });
  }

  async closeShort(symbol: string, quantity: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'closeShort',
      params: {
        symbol,
        quantity,
        side: 'BUY',
        positionSide: 'SHORT',
      },
    });
  }

  async setStopLoss(symbol: string, positionSide: 'LONG' | 'SHORT', stopPrice: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'setStopLoss',
      params: {
        symbol,
        positionSide,
        stopPrice,
      },
    });
  }

  async setTakeProfit(symbol: string, positionSide: 'LONG' | 'SHORT', takeProfitPrice: number) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'setTakeProfit',
      params: {
        symbol,
        positionSide,
        takeProfitPrice,
      },
    });
  }

  async cancelAllOrders(symbol: string) {
    return this.request(API_ENDPOINTS.trading, {
      action: 'cancelAllOrders',
      params: { symbol },
    });
  }

  // 触发定时任务（手动测试用）
  async triggerTradingCron() {
    return this.request(API_ENDPOINTS.tradingCron, {});
  }
}

// 导出单例
export const nofxApi = new NOFXApiClient();

// 使用示例：
// import { nofxApi } from './supabaseApiClient';
// 
// // 获取涨幅榜
// const gainers = await nofxApi.getTopGainers();
// console.log('涨幅榜:', gainers.data);
// 
// // 获取BTC价格
// const btcPrice = await nofxApi.getMarketPrice('BTCUSDT');
// console.log('BTC价格:', btcPrice.data);
// 
// // 获取账户余额（需要正确的Testnet密钥）
// const balance = await nofxApi.getBalance();
// console.log('账户余额:', balance.data);
