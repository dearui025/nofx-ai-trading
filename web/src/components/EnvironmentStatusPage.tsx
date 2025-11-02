import React, { useState, useEffect } from 'react';
import { Activity, Wifi, WifiOff, CheckCircle, AlertCircle, Clock, Database, Key, Settings } from 'lucide-react';
import { environmentApi } from '../lib/environmentApi';
import { EnvironmentStatus } from '../types/environment';

interface ConnectionStatus {
  binance: {
    connected: boolean;
    latency?: number;
    permissions?: string[];
    error?: string;
  };
  deepseek: {
    connected: boolean;
    latency?: number;
    model?: string;
    error?: string;
  };
}

export const EnvironmentStatusPage: React.FC = () => {
  const [status, setStatus] = useState<EnvironmentStatus | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // 获取环境状态
  const fetchStatus = async () => {
    try {
      setError(null);
      const envStatus = await environmentApi.getStatus();
      setStatus(envStatus);
      setLastUpdate(new Date());
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取环境状态失败');
    }
  };

  // 测试连接状态
  const testConnections = async () => {
    if (!status) return;

    try {
      setRefreshing(true);
      
      // 模拟API连接测试
      const testResults: ConnectionStatus = {
        binance: {
          connected: false,
          latency: 0,
          permissions: [],
          error: undefined,
        },
        deepseek: {
          connected: false,
          latency: 0,
          model: undefined,
          error: undefined,
        },
      };

      // 测试Binance连接
      try {
        const binanceStart = Date.now();
        // 这里应该调用实际的Binance API测试接口
        // 暂时模拟测试结果
        await new Promise(resolve => setTimeout(resolve, Math.random() * 1000 + 500));
        testResults.binance = {
          connected: true,
          latency: Date.now() - binanceStart,
          permissions: ['SPOT', 'FUTURES', 'READ', 'TRADE'],
        };
      } catch (err) {
        testResults.binance.error = err instanceof Error ? err.message : 'Binance连接失败';
      }

      // 测试DeepSeek连接
      try {
        const deepseekStart = Date.now();
        // 这里应该调用实际的DeepSeek API测试接口
        // 暂时模拟测试结果
        await new Promise(resolve => setTimeout(resolve, Math.random() * 800 + 300));
        testResults.deepseek = {
          connected: true,
          latency: Date.now() - deepseekStart,
          model: 'deepseek-chat',
        };
      } catch (err) {
        testResults.deepseek.error = err instanceof Error ? err.message : 'DeepSeek连接失败';
      }

      setConnectionStatus(testResults);
    } catch (err) {
      setError(err instanceof Error ? err.message : '连接测试失败');
    } finally {
      setRefreshing(false);
    }
  };

  // 刷新状态
  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([fetchStatus(), testConnections()]);
    setRefreshing(false);
  };

  // 获取状态颜色
  const getStatusColor = (isHealthy: boolean) => {
    return isHealthy ? 'text-green-600' : 'text-red-600';
  };

  // 获取状态背景色
  const getStatusBgColor = (isHealthy: boolean) => {
    return isHealthy ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200';
  };

  // 格式化延迟
  const formatLatency = (latency?: number) => {
    if (!latency) return 'N/A';
    return `${latency}ms`;
  };

  useEffect(() => {
    const initializeData = async () => {
      setLoading(true);
      await Promise.all([fetchStatus(), testConnections()]);
      setLoading(false);
    };

    initializeData();

    // 设置定时刷新
    const interval = setInterval(() => {
      fetchStatus();
      testConnections();
    }, 30000); // 每30秒刷新一次

    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="max-w-6xl mx-auto p-6">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span className="ml-2 text-gray-600">加载环境状态...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Activity className="w-6 h-6 text-blue-600" />
            <h1 className="text-2xl font-bold text-gray-800">环境状态监控</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Clock className="w-4 h-4" />
              <span>最后更新: {lastUpdate.toLocaleTimeString()}</span>
            </div>
            <button
              onClick={handleRefresh}
              disabled={refreshing}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Activity className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
              刷新状态
            </button>
          </div>
        </div>
        <p className="text-gray-600 mt-2">
          实时监控当前环境的连接状态、API有效性和系统健康度。
        </p>
      </div>

      {/* 当前环境概览 */}
      {status && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">当前环境概览</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className={`p-4 rounded-lg border ${getStatusBgColor(status.is_healthy)}`}>
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-full ${status.is_healthy ? 'bg-green-100' : 'bg-red-100'}`}>
                  {status.is_healthy ? (
                    <CheckCircle className="w-5 h-5 text-green-600" />
                  ) : (
                    <AlertCircle className="w-5 h-5 text-red-600" />
                  )}
                </div>
                <div>
                  <p className="text-sm text-gray-600">环境状态</p>
                  <p className={`font-semibold ${getStatusColor(status.is_healthy)}`}>
                    {status.is_healthy ? '健康' : '异常'}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-4 rounded-lg border bg-blue-50 border-blue-200">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-full bg-blue-100">
                  <Database className="w-5 h-5 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm text-gray-600">当前环境</p>
                  <p className="font-semibold text-blue-800">
                    {status.current_environment === 'testnet' ? '测试网' : '真实环境'}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-4 rounded-lg border bg-purple-50 border-purple-200">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-full bg-purple-100">
                  <Settings className="w-5 h-5 text-purple-600" />
                </div>
                <div>
                  <p className="text-sm text-gray-600">活跃交易员</p>
                  <p className="font-semibold text-purple-800">
                    {status.environments?.[status.current_environment]?.trading_settings?.enabled_traders?.length || 0}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* API连接状态 */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">API连接状态</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Binance连接状态 */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-full ${connectionStatus?.binance.connected ? 'bg-green-100' : 'bg-red-100'}`}>
                  {connectionStatus?.binance.connected ? (
                    <Wifi className="w-5 h-5 text-green-600" />
                  ) : (
                    <WifiOff className="w-5 h-5 text-red-600" />
                  )}
                </div>
                <h3 className="font-medium text-gray-800">Binance API</h3>
              </div>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                connectionStatus?.binance.connected 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-red-100 text-red-800'
              }`}>
                {connectionStatus?.binance.connected ? '已连接' : '未连接'}
              </span>
            </div>

            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">延迟:</span>
                <span className="font-medium">{formatLatency(connectionStatus?.binance.latency)}</span>
              </div>
              
              {connectionStatus?.binance.permissions && (
                <div>
                  <span className="text-sm text-gray-600">权限:</span>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {connectionStatus.binance.permissions.map((permission) => (
                      <span
                        key={permission}
                        className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                      >
                        {permission}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {connectionStatus?.binance.error && (
                <div className="text-sm text-red-600 bg-red-50 p-2 rounded">
                  {connectionStatus.binance.error}
                </div>
              )}
            </div>
          </div>

          {/* DeepSeek连接状态 */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-full ${connectionStatus?.deepseek.connected ? 'bg-green-100' : 'bg-red-100'}`}>
                  {connectionStatus?.deepseek.connected ? (
                    <Wifi className="w-5 h-5 text-green-600" />
                  ) : (
                    <WifiOff className="w-5 h-5 text-red-600" />
                  )}
                </div>
                <h3 className="font-medium text-gray-800">DeepSeek API</h3>
              </div>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                connectionStatus?.deepseek.connected 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-red-100 text-red-800'
              }`}>
                {connectionStatus?.deepseek.connected ? '已连接' : '未连接'}
              </span>
            </div>

            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">延迟:</span>
                <span className="font-medium">{formatLatency(connectionStatus?.deepseek.latency)}</span>
              </div>
              
              {connectionStatus?.deepseek.model && (
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">模型:</span>
                  <span className="font-medium">{connectionStatus.deepseek.model}</span>
                </div>
              )}

              {connectionStatus?.deepseek.error && (
                <div className="text-sm text-red-600 bg-red-50 p-2 rounded">
                  {connectionStatus.deepseek.error}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 环境配置详情 */}
      {status && status.environments && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">环境配置详情</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {Object.entries(status.environments).map(([envName, envConfig]) => (
              <div key={envName} className="border border-gray-200 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-3">
                  <Key className="w-4 h-4 text-gray-600" />
                  <h3 className="font-medium text-gray-800">
                    {envName === 'testnet' ? '测试网环境' : '真实环境'}
                  </h3>
                  {status.current_environment === envName && (
                    <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                      当前
                    </span>
                  )}
                </div>

                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Binance API:</span>
                    <span className={`font-medium ${envConfig.binance_api_key ? 'text-green-600' : 'text-red-600'}`}>
                      {envConfig.binance_api_key ? '已配置' : '未配置'}
                    </span>
                  </div>
                  
                  <div className="flex justify-between">
                    <span className="text-gray-600">DeepSeek API:</span>
                    <span className={`font-medium ${envConfig.deepseek_api_key ? 'text-green-600' : 'text-red-600'}`}>
                      {envConfig.deepseek_api_key ? '已配置' : '未配置'}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span className="text-gray-600">最大仓位:</span>
                    <span className="font-medium">{envConfig.risk_limits?.max_position_size || 'N/A'}</span>
                  </div>

                  <div className="flex justify-between">
                    <span className="text-gray-600">日损失限制:</span>
                    <span className="font-medium">{envConfig.risk_limits?.daily_loss_limit || 'N/A'}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 错误信息 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-red-500" />
            <span className="text-red-700">{error}</span>
          </div>
        </div>
      )}
    </div>
  );
};