// 环境管理相关类型定义

export interface RiskLimits {
  max_position_size: number;
  max_daily_loss: number;
  daily_loss_limit: number;
  max_drawdown: number;
  max_leverage: number;
}

export interface TradingSettings {
  scan_interval_minutes: number;
  initial_balance: number;
  stop_trading_minutes: number;
  enabled_traders: string[];
}

export interface Environment {
  name: string;
  description: string;
  binance_base_url: string;
  binance_api_key: string;
  binance_secret_key: string;
  deepseek_api_key: string;
  risk_limits: RiskLimits;
  trading_settings: TradingSettings;
  status: 'active' | 'inactive' | 'error';
  last_validated: string | null;
  api_permissions: string[];
}

export interface EnvironmentStatus {
  current_environment: string;
  status: string;
  api_status: {
    binance_configured: boolean;
    deepseek_configured: boolean;
    last_validated: string | null;
    permissions: string[];
  };
  last_updated: string;
  environments: Record<string, Environment>;
  is_healthy: boolean;
}

export interface EnvironmentSwitchRequest {
  target_environment: string;
}

export interface EnvironmentSwitchResponse {
  success: boolean;
  message: string;
  new_environment: string;
}

export interface EnvironmentConfigRequest {
  environment: string;
  binance_api_key?: string;
  binance_secret_key?: string;
  deepseek_api_key?: string;
  oi_top_api_url?: string;
}

export interface EnvironmentConfigResponse {
  success: boolean;
  message: string;
}

export interface EnvironmentValidateRequest {
  environment: string;
  api_keys?: Record<string, any>;
}

export interface EnvironmentValidateResponse {
  valid: boolean;
  permissions: string[];
  errors: string[];
  timestamp: string;
}

export type EnvironmentType = 'testnet' | 'mainnet';