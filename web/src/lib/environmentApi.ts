// 环境管理API客户端

import {
  EnvironmentStatus,
  EnvironmentSwitchRequest,
  EnvironmentSwitchResponse,
  EnvironmentConfigRequest,
  EnvironmentConfigResponse,
  EnvironmentValidateRequest,
  EnvironmentValidateResponse,
} from '../types/environment';

// 本地后端API配置
const API_BASE_URL = 'http://localhost:8080/api';

class EnvironmentApiClient {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  // 获取环境状态
  async getStatus(): Promise<EnvironmentStatus> {
    return await this.request<EnvironmentStatus>('/environment/status');
  }

  // 切换环境
  async switchEnvironment(request: EnvironmentSwitchRequest): Promise<EnvironmentSwitchResponse> {
    return await this.request<EnvironmentSwitchResponse>('/environment/switch', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // 更新环境配置
  async updateConfig(request: EnvironmentConfigRequest): Promise<EnvironmentConfigResponse> {
    return await this.request<EnvironmentConfigResponse>('/environment/config', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // 验证环境配置
  async validateEnvironment(request: EnvironmentValidateRequest): Promise<EnvironmentValidateResponse> {
    return await this.request<EnvironmentValidateResponse>('/environment/validate', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }
}

export const environmentApi = new EnvironmentApiClient();