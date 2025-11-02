import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Activity, Settings, TrendingUp, TrendingDown, Clock, AlertTriangle, CheckCircle, Zap } from 'lucide-react';

interface FrequencyStatus {
  enabled: boolean;
  current_mode: 'basic' | 'elastic';
  daily_pnl_percent: number;
  hourly_trade_count: number;
  daily_trade_count: number;
  current_limits: {
    hourly_limit: number;
    daily_limit: number;
  };
  next_mode_threshold: number;
  time_to_hourly_reset: string;
  last_mode_switch: string | null;
  config: {
    basic_mode: { hourly_limit: number; daily_limit: number };
    elastic_mode: { hourly_limit: number; daily_limit: number };
    absolute_hourly_max: number;
    upgrade_threshold: number;
    downgrade_threshold: number;
  };
  error: string | null;
}

interface FrequencyManagerProps {
  traderId?: string;
}

const FrequencyManager: React.FC<FrequencyManagerProps> = ({ traderId }) => {
  const [status, setStatus] = useState<FrequencyStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [configForm, setConfigForm] = useState({
    basic_mode: {
      hourly_limit: 2,
      daily_limit: 10
    },
    elastic_mode: {
      hourly_limit: 5,
      daily_limit: -1
    },
    absolute_limit: {
      hourly_max: 6
    },
    thresholds: {
      upgrade_pnl_percent: 2.0,
      downgrade_pnl_percent: 1.0
    }
  });

  // 获取频率管理器状态
  const fetchStatus = async () => {
    try {
      setLoading(true);
      const data = await api.getFrequencyStatus(traderId);
      setStatus(data);
      setError(null);
      
      // 更新配置表单
      if (data.config) {
        setConfigForm({
          basic_mode: {
            hourly_limit: data.config.basic_mode?.hourly_limit || 2,
            daily_limit: data.config.basic_mode?.daily_limit || 10
          },
          elastic_mode: {
            hourly_limit: data.config.elastic_mode?.hourly_limit || 5,
            daily_limit: data.config.elastic_mode?.daily_limit || -1
          },
          absolute_limit: {
            hourly_max: data.config.absolute_limit?.hourly_max || data.config.absolute_hourly_max || 6
          },
          thresholds: {
            upgrade_pnl_percent: data.config.thresholds?.upgrade_pnl_percent || data.config.upgrade_threshold || 2.0,
            downgrade_pnl_percent: data.config.thresholds?.downgrade_pnl_percent || data.config.downgrade_threshold || 1.0
          }
        });
      }
    } catch (err) {
      setError(`获取频率管理器状态失败: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  // 更新配置
  const updateConfig = async () => {
    try {
      setLoading(true);
      const result = await api.updateFrequencyConfig(configForm, traderId);
      
      if (result.success) {
        setShowConfig(false);
        await fetchStatus(); // 重新获取状态
        alert('配置更新成功！');
      } else {
        alert(`配置更新失败: ${result.message}`);
      }
    } catch (err) {
      alert(`配置更新失败: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    
    // 每30秒自动刷新状态
    const interval = setInterval(fetchStatus, 30000);
    return () => clearInterval(interval);
  }, [traderId]);

  if (loading && !status) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-center h-32">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-gray-600">加载频率管理器状态...</span>
        </div>
      </div>
    );
  }

  if (error || !status) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center text-red-600 mb-4">
          <AlertTriangle className="w-5 h-5 mr-2" />
          <h3 className="text-lg font-semibold">频率管理器状态</h3>
        </div>
        <div className="text-red-600 bg-red-50 p-4 rounded-lg">
          {error || status?.error || '无法获取频率管理器状态'}
        </div>
        <button
          onClick={fetchStatus}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          重试
        </button>
      </div>
    );
  }

  const getModeIcon = (mode: string) => {
    return mode === 'elastic' ? <Zap className="w-5 h-5 text-yellow-500" /> : <Activity className="w-5 h-5 text-blue-500" />;
  };

  const getModeColor = (mode: string) => {
    return mode === 'elastic' ? 'bg-yellow-100 text-yellow-800' : 'bg-blue-100 text-blue-800';
  };

  const getPnLColor = (pnl: number) => {
    if (pnl > 0) return 'text-green-600';
    if (pnl < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  const getProgressPercentage = (current: number, limit: number) => {
    if (limit <= 0) return 0;
    return Math.min((current / limit) * 100, 100);
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      {/* 标题栏 */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <Activity className="w-6 h-6 text-blue-600 mr-2" />
          <h3 className="text-xl font-semibold text-gray-800">绩效驱动弹性频率限制</h3>
          {status.enabled ? (
            <CheckCircle className="w-5 h-5 text-green-500 ml-2" />
          ) : (
            <AlertTriangle className="w-5 h-5 text-red-500 ml-2" />
          )}
        </div>
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setShowConfig(!showConfig)}
            className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors"
            title="配置设置"
          >
            <Settings className="w-5 h-5" />
          </button>
          <button
            onClick={fetchStatus}
            disabled={loading}
            className="px-3 py-1 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
          >
            {loading ? '刷新中...' : '刷新'}
          </button>
        </div>
      </div>

      {/* 当前状态概览 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        {/* 当前模式 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-600">当前模式</span>
            {getModeIcon(status.current_mode)}
          </div>
          <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getModeColor(status.current_mode)}`}>
            {status.current_mode === 'elastic' ? '弹性模式' : '基础模式'}
          </div>
        </div>

        {/* 日收益率 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-600">日收益率</span>
            {status.daily_pnl_percent > 0 ? (
              <TrendingUp className="w-4 h-4 text-green-500" />
            ) : (
              <TrendingDown className="w-4 h-4 text-red-500" />
            )}
          </div>
          <div className={`text-lg font-semibold ${getPnLColor(status.daily_pnl_percent)}`}>
            {status.daily_pnl_percent > 0 ? '+' : ''}{status.daily_pnl_percent.toFixed(2)}%
          </div>
        </div>

        {/* 下次切换阈值 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-600">
              {status.current_mode === 'elastic' ? '降级阈值' : '升级阈值'}
            </span>
            <Clock className="w-4 h-4 text-gray-500" />
          </div>
          <div className="text-lg font-semibold text-gray-800">
            {status.next_mode_threshold.toFixed(1)}%
          </div>
        </div>
      </div>

      {/* 交易频率使用情况 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {/* 小时限制 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-600">小时交易次数</span>
            <span className="text-xs text-gray-500">
              重置时间: {status.time_to_hourly_reset}
            </span>
          </div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-lg font-semibold">
              {status.hourly_trade_count} / {status.current_limits.hourly_limit}
            </span>
            <span className="text-sm text-gray-600">
              {getProgressPercentage(status.hourly_trade_count, status.current_limits.hourly_limit).toFixed(0)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all duration-300 ${
                status.hourly_trade_count >= status.current_limits.hourly_limit
                  ? 'bg-red-500'
                  : status.hourly_trade_count / status.current_limits.hourly_limit > 0.8
                  ? 'bg-yellow-500'
                  : 'bg-green-500'
              }`}
              style={{
                width: `${getProgressPercentage(status.hourly_trade_count, status.current_limits.hourly_limit)}%`
              }}
            ></div>
          </div>
        </div>

        {/* 日限制 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-600">日交易次数</span>
            <span className="text-xs text-gray-500">
              {status.current_limits.daily_limit === -1 ? '无限制' : '每日重置'}
            </span>
          </div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-lg font-semibold">
              {status.daily_trade_count} / {status.current_limits.daily_limit === -1 ? '∞' : status.current_limits.daily_limit}
            </span>
            {status.current_limits.daily_limit !== -1 && (
              <span className="text-sm text-gray-600">
                {getProgressPercentage(status.daily_trade_count, status.current_limits.daily_limit).toFixed(0)}%
              </span>
            )}
          </div>
          {status.current_limits.daily_limit !== -1 ? (
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all duration-300 ${
                  status.daily_trade_count >= status.current_limits.daily_limit
                    ? 'bg-red-500'
                    : status.daily_trade_count / status.current_limits.daily_limit > 0.8
                    ? 'bg-yellow-500'
                    : 'bg-green-500'
                }`}
                style={{
                  width: `${getProgressPercentage(status.daily_trade_count, status.current_limits.daily_limit)}%`
                }}
              ></div>
            </div>
          ) : (
            <div className="w-full bg-green-200 rounded-full h-2">
              <div className="h-2 rounded-full bg-green-500 w-full"></div>
            </div>
          )}
        </div>
      </div>

      {/* 模式切换历史 */}
      {status.last_mode_switch && (
        <div className="bg-blue-50 rounded-lg p-4 mb-6">
          <div className="flex items-center mb-2">
            <Clock className="w-4 h-4 text-blue-600 mr-2" />
            <span className="text-sm font-medium text-blue-800">最近模式切换</span>
          </div>
          <div className="text-sm text-blue-700">
            {new Date(status.last_mode_switch).toLocaleString()}
          </div>
        </div>
      )}

      {/* 配置面板 */}
      {showConfig && (
        <div className="border-t pt-6">
          <h4 className="text-lg font-semibold mb-4">频率限制配置</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* 基础模式配置 */}
            <div className="space-y-4">
              <h5 className="font-medium text-gray-800">基础模式</h5>
              <div>
                <label className="block text-sm text-gray-600 mb-1">小时限制</label>
                <input
                  type="number"
                  value={configForm.basic_mode.hourly_limit}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    basic_mode: {
                      ...configForm.basic_mode,
                      hourly_limit: parseInt(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">日限制</label>
                <input
                  type="number"
                  value={configForm.basic_mode.daily_limit}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    basic_mode: {
                      ...configForm.basic_mode,
                      daily_limit: parseInt(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* 弹性模式配置 */}
            <div className="space-y-4">
              <h5 className="font-medium text-gray-800">弹性模式</h5>
              <div>
                <label className="block text-sm text-gray-600 mb-1">小时限制</label>
                <input
                  type="number"
                  value={configForm.elastic_mode.hourly_limit}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    elastic_mode: {
                      ...configForm.elastic_mode,
                      hourly_limit: parseInt(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">日限制 (-1为无限制)</label>
                <input
                  type="number"
                  value={configForm.elastic_mode.daily_limit}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    elastic_mode: {
                      ...configForm.elastic_mode,
                      daily_limit: parseInt(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* 阈值配置 */}
            <div className="space-y-4">
              <h5 className="font-medium text-gray-800">切换阈值</h5>
              <div>
                <label className="block text-sm text-gray-600 mb-1">升级阈值 (%)</label>
                <input
                  type="number"
                  step="0.1"
                  value={configForm.thresholds.upgrade_pnl_percent}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    thresholds: {
                      ...configForm.thresholds,
                      upgrade_pnl_percent: parseFloat(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">降级阈值 (%)</label>
                <input
                  type="number"
                  step="0.1"
                  value={configForm.thresholds.downgrade_pnl_percent}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    thresholds: {
                      ...configForm.thresholds,
                      downgrade_pnl_percent: parseFloat(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* 安全限制 */}
            <div className="space-y-4">
              <h5 className="font-medium text-gray-800">安全限制</h5>
              <div>
                <label className="block text-sm text-gray-600 mb-1">绝对小时最大值</label>
                <input
                  type="number"
                  value={configForm.absolute_limit.hourly_max}
                  onChange={(e) => setConfigForm({
                    ...configForm, 
                    absolute_limit: {
                      ...configForm.absolute_limit,
                      hourly_max: parseInt(e.target.value)
                    }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
          </div>

          <div className="flex justify-end space-x-3 mt-6">
            <button
              onClick={() => setShowConfig(false)}
              className="px-4 py-2 text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              取消
            </button>
            <button
              onClick={updateConfig}
              disabled={loading}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
            >
              {loading ? '保存中...' : '保存配置'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default FrequencyManager;