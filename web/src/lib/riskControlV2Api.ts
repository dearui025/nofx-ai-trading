import { goApiClient } from './goApiClient';

// 风控v2 API 基础URL
const RISK_CONTROL_V2_BASE_URL = 'http://localhost:8081/api/v2';

// 风控v2数据类型定义
export interface RiskControlV2SystemStatus {
  initialized: boolean;
  running: boolean;
  start_time: string;
  uptime: string;
}

export interface RiskControlV2HealthStatus {
  overall_status: string;
  last_check_time: string;
  checks: Record<string, HealthCheck>;
}

export interface HealthCheck {
  name: string;
  status: string;
  message: string;
  last_check: string;
  check_count: number;
  fail_count: number;
  response_time: string;
}

export interface RiskStatus {
  is_active: boolean;
  last_update_time: string;
  emergency_stop: boolean;
  global_risk_level: string;
  time_manager_state: any;
  liquidity_state: any;
  sharpe_state: any;
  ai_committee_state: any;
  total_alerts: number;
  critical_alerts: number;
  last_alert_time: string;
  system_health: string;
}

export interface TimeStatus {
  current_utc_time: string;
  last_daily_reset: string;
  equity_high_watermark: number;
  last_watermark_update: string;
  reset_count: number;
  last_reset_reason: string;
}

export interface AICommitteeStatus {
  current_strategy: string;
  market_condition: string;
  last_decision_time: string;
  total_decisions: number;
  consensus_decisions: number;
  conflict_decisions: number;
  avg_consensus_level: number;
  model_performances: Record<string, any>;
  active_models: string[];
}

export interface LiquidityStatus {
  monitoring_enabled: boolean;
  last_update_time: string;
  total_symbols_monitored: number;
  blacklisted_symbols: string[];
  active_alerts: number;
  avg_liquidity_score: number;
  market_health: string;
}

export interface SharpeStatus {
  current_sharpe_ratio: number;
  last_calculation_time: string;
  calculation_count: number;
  avg_sharpe_ratio: number;
  sharpe_trend: string;
  risk_adjusted_return: number;
}

// API响应包装类型
interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

// 风控v2 API客户端类
export class RiskControlV2ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = RISK_CONTROL_V2_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  // 系统管理API
  async getSystemStatus(): Promise<RiskControlV2SystemStatus> {
    const response = await goApiClient.get<ApiResponse<RiskControlV2SystemStatus>>(
      `${this.baseUrl}/system/status`
    );
    return response.data;
  }

  async getHealthStatus(): Promise<RiskControlV2HealthStatus> {
    const response = await goApiClient.get<ApiResponse<RiskControlV2HealthStatus>>(
      `${this.baseUrl}/system/health`
    );
    return response.data;
  }

  async getSystemStats(): Promise<any> {
    const response = await goApiClient.get<ApiResponse<any>>(
      `${this.baseUrl}/system/stats`
    );
    return response.data;
  }

  // 风控管理API
  async getRiskStatus(): Promise<RiskStatus> {
    const response = await goApiClient.get<ApiResponse<RiskStatus>>(
      `${this.baseUrl}/risk-control/risk/status`
    );
    return response.data;
  }

  async getRiskDecisions(limit?: number): Promise<any[]> {
    const params = limit ? `?limit=${limit}` : '';
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/risk/decisions${params}`
    );
    return response.data;
  }

  async getRiskAlerts(): Promise<any[]> {
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/risk/alerts`
    );
    return response.data;
  }

  async emergencyStop(): Promise<void> {
    await goApiClient.post(`${this.baseUrl}/risk-control/risk/emergency-stop`);
  }

  async resumeRisk(): Promise<void> {
    await goApiClient.post(`${this.baseUrl}/risk-control/risk/resume`);
  }

  // 时间管理API
  async getTimeStatus(): Promise<TimeStatus> {
    const response = await goApiClient.get<ApiResponse<TimeStatus>>(
      `${this.baseUrl}/risk-control/time/status`
    );
    return response.data;
  }

  async manualReset(): Promise<void> {
    await goApiClient.post(`${this.baseUrl}/risk-control/time/reset`);
  }

  async getResetHistory(): Promise<any[]> {
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/time/reset-history`
    );
    return response.data;
  }

  // AI委员会API
  async getAICommitteeStatus(): Promise<AICommitteeStatus> {
    const response = await goApiClient.get<ApiResponse<AICommitteeStatus>>(
      `${this.baseUrl}/risk-control/ai-committee/status`
    );
    return response.data;
  }

  async getAIDecisions(limit?: number): Promise<any[]> {
    const params = limit ? `?limit=${limit}` : '';
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/ai-committee/decisions${params}`
    );
    return response.data;
  }

  async getModelPerformance(): Promise<any> {
    const response = await goApiClient.get<ApiResponse<any>>(
      `${this.baseUrl}/risk-control/ai-committee/performance`
    );
    return response.data;
  }

  // 流动性监控API
  async getLiquidityStatus(): Promise<LiquidityStatus> {
    const response = await goApiClient.get<ApiResponse<LiquidityStatus>>(
      `${this.baseUrl}/risk-control/liquidity/status`
    );
    return response.data;
  }

  async getLiquidityAlerts(): Promise<any[]> {
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/liquidity/alerts`
    );
    return response.data;
  }

  async getBlacklist(): Promise<string[]> {
    const response = await goApiClient.get<ApiResponse<string[]>>(
      `${this.baseUrl}/risk-control/liquidity/blacklist`
    );
    return response.data;
  }

  // 夏普比率API
  async getSharpeStatus(): Promise<SharpeStatus> {
    const response = await goApiClient.get<ApiResponse<SharpeStatus>>(
      `${this.baseUrl}/risk-control/sharpe/status`
    );
    return response.data;
  }

  async getSharpeRecords(limit?: number): Promise<any[]> {
    const params = limit ? `?limit=${limit}` : '';
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/sharpe/records${params}`
    );
    return response.data;
  }

  async getSharpeTransitions(): Promise<any[]> {
    const response = await goApiClient.get<ApiResponse<any[]>>(
      `${this.baseUrl}/risk-control/sharpe/transitions`
    );
    return response.data;
  }

  // 配置管理API
  async getConfig(type: string, name: string): Promise<any> {
    const response = await goApiClient.get<ApiResponse<any>>(
      `${this.baseUrl}/risk-control/config/${type}/${name}`
    );
    return response.data;
  }

  async setConfig(type: string, name: string, config: any): Promise<void> {
    await goApiClient.put(`${this.baseUrl}/risk-control/config/${type}/${name}`, config);
  }

  async getAllConfigs(): Promise<any> {
    const response = await goApiClient.get<ApiResponse<any>>(
      `${this.baseUrl}/risk-control/config/all`
    );
    return response.data;
  }

  // 数据管理API
  async getDataStats(): Promise<any> {
    const response = await goApiClient.get<ApiResponse<any>>(
      `${this.baseUrl}/risk-control/data/stats`
    );
    return response.data;
  }

  async cleanupOldData(days: number): Promise<void> {
    await goApiClient.post(`${this.baseUrl}/risk-control/data/cleanup`, { days });
  }
}

// 创建全局实例
export const riskControlV2Api = new RiskControlV2ApiClient();

// 导出便捷函数
export const getRiskControlV2SystemStatus = () => riskControlV2Api.getSystemStatus();
export const getRiskControlV2HealthStatus = () => riskControlV2Api.getHealthStatus();
export const getRiskStatus = () => riskControlV2Api.getRiskStatus();
export const getTimeStatus = () => riskControlV2Api.getTimeStatus();
export const getAICommitteeStatus = () => riskControlV2Api.getAICommitteeStatus();
export const getLiquidityStatus = () => riskControlV2Api.getLiquidityStatus();
export const getSharpeStatus = () => riskControlV2Api.getSharpeStatus();