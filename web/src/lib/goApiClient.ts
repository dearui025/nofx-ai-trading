// Go后端API客户端
// 连接到本地Go服务器API

const GO_API_BASE_URL = 'http://localhost:8080/api';

// 市场数据响应类型
interface MarketDataResponse {
  symbol: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  volume24h: number;
}

// API客户端类
class GoApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = GO_API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  // 通用请求方法
  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        ...options,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error(`API request failed for ${endpoint}:`, error);
      throw error;
    }
  }

  // 通用GET方法
  async get<T>(url: string): Promise<T> {
    // 如果URL是完整的URL，直接使用；否则拼接baseUrl
    const fullUrl = url.startsWith('http') ? url : `${this.baseUrl}${url}`;
    
    try {
      const response = await fetch(fullUrl, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error(`GET request failed for ${fullUrl}:`, error);
      throw error;
    }
  }

  // 通用POST方法
  async post<T>(url: string, data?: any): Promise<T> {
    // 如果URL是完整的URL，直接使用；否则拼接baseUrl
    const fullUrl = url.startsWith('http') ? url : `${this.baseUrl}${url}`;
    
    try {
      const response = await fetch(fullUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: data ? JSON.stringify(data) : undefined,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const responseData = await response.json();
      return responseData;
    } catch (error) {
      console.error(`POST request failed for ${fullUrl}:`, error);
      throw error;
    }
  }

  // 获取trader列表
  async getTraders() {
    return this.request('/traders');
  }

  // 获取账户信息
  async getAccount(traderId: string) {
    return this.request(`/account?trader_id=${traderId}`);
  }

  // 获取持仓信息
  async getPositions(traderId: string) {
    return this.request(`/positions?trader_id=${traderId}`);
  }

  // 获取决策记录
  async getDecisions(traderId: string) {
    return this.request(`/decisions?trader_id=${traderId}`);
  }

  // 获取最新决策记录
  async getLatestDecisions(traderId: string) {
    return this.request(`/decisions/latest?trader_id=${traderId}`);
  }

  // 获取状态信息
  async getStatus(traderId: string) {
    return this.request(`/status?trader_id=${traderId}`);
  }

  // 获取统计信息
  async getStatistics(traderId: string) {
    return this.request(`/statistics?trader_id=${traderId}`);
  }

  // 获取权益历史
  async getEquityHistory(traderId: string) {
    return this.request(`/equity-history?trader_id=${traderId}`);
  }

  // 获取性能数据
  async getPerformance(traderId: string) {
    return this.request(`/performance?trader_id=${traderId}`);
  }

  // 获取竞赛数据
  async getCompetition() {
    return this.request('/competition');
  }

  // 获取市场数据
  async getMarketData(symbol: string): Promise<MarketDataResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/market-data?symbol=${symbol}`);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      
      // 转换后端数据格式为前端期望的格式
      const priceChangePercent = data.PriceChange24h || data.PriceChange4h || 0;
      const currentPrice = data.CurrentPrice || 0;
      
      const marketData: MarketDataResponse = {
        symbol: data.Symbol || symbol,
        price: currentPrice,
        change24h: currentPrice * priceChangePercent / 100, // 计算绝对价格变化
        changePercent24h: priceChangePercent, // 百分比变化
        volume24h: data.Volume24h || data.LongerTermContext?.CurrentVolume || 0
      };
      
      return marketData;
    } catch (error) {
      console.error(`Error fetching market data for ${symbol}:`, error);
      throw error;
    }
  }

  // 健康检查
  async healthCheck() {
    return this.request('/health');
  }

  // 获取频率管理器状态
  async getFrequencyStatus(traderId: string) {
    return this.request(`/frequency-status?trader_id=${traderId}`);
  }

  // 更新频率管理器配置
  async updateFrequencyConfig(config: any, traderId: string) {
    return this.request(`/frequency-config?trader_id=${traderId}`, {
      method: 'POST',
      body: JSON.stringify(config),
    });
  }

  // 获取AI优化数据
  async getAIOptimizationData() {
    return this.request('/ai-optimization');
  }
}

// 导出单例实例
export const goApiClient = new GoApiClient();
export default goApiClient;