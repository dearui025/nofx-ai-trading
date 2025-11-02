import React, { useState, useEffect } from 'react';
import { TestTube, DollarSign, AlertCircle, CheckCircle, Clock, Settings, Key } from 'lucide-react';
import { environmentApi } from '../lib/environmentApi';
import { EnvironmentStatus, EnvironmentType } from '../types/environment';
import { EnvironmentConfigForm } from './EnvironmentConfigForm';

interface EnvironmentSwitcherProps {
  onEnvironmentChange?: (environment: string) => void;
}

export const EnvironmentSwitcher: React.FC<EnvironmentSwitcherProps> = ({
  onEnvironmentChange,
}) => {
  const [status, setStatus] = useState<EnvironmentStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [switching, setSwitching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showConfigForm, setShowConfigForm] = useState(false);
  const [configEnvironment, setConfigEnvironment] = useState<EnvironmentType>('mainnet');

  // 获取环境状态
  const fetchStatus = async () => {
    try {
      setLoading(true);
      setError(null);
      const envStatus = await environmentApi.getStatus();
      setStatus(envStatus);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取环境状态失败');
      console.error('获取环境状态失败:', err);
    } finally {
      setLoading(false);
    }
  };

  // 切换环境
  const handleEnvironmentSwitch = async (targetEnv: EnvironmentType) => {
    if (!status || switching) return;

    // 检查目标环境是否已配置
    const targetEnvConfig = status.environments?.[targetEnv];
    if (targetEnv === 'mainnet' && (!targetEnvConfig?.binance_api_key || !targetEnvConfig?.binance_secret_key)) {
      // 如果是真实环境且未配置，显示配置表单
      setConfigEnvironment(targetEnv);
      setShowConfigForm(true);
      return;
    }

    try {
      setSwitching(true);
      setError(null);

      const response = await environmentApi.switchEnvironment({
        target_environment: targetEnv,
      });

      if (response.success) {
        // 更新状态
        await fetchStatus();
        onEnvironmentChange?.(targetEnv);
        
        // 显示成功消息
        console.log('环境切换成功:', response.message);
      } else {
        throw new Error(response.message);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '环境切换失败');
      console.error('环境切换失败:', err);
    } finally {
      setSwitching(false);
    }
  };

  // 显示配置表单
  const handleShowConfig = (env: EnvironmentType) => {
    setConfigEnvironment(env);
    setShowConfigForm(true);
  };

  // 配置更新完成
  const handleConfigUpdate = async () => {
    setShowConfigForm(false);
    await fetchStatus();
  };

  // 取消配置
  const handleConfigCancel = () => {
    setShowConfigForm(false);
  };

  // 获取环境图标
  const getEnvironmentIcon = (env: string) => {
    switch (env) {
      case 'testnet':
        return <TestTube className="w-4 h-4" />;
      case 'mainnet':
        return <DollarSign className="w-4 h-4" />;
      default:
        return <Settings className="w-4 h-4" />;
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'error':
        return <AlertCircle className="w-4 h-4 text-red-500" />;
      case 'inactive':
      default:
        return <Clock className="w-4 h-4 text-yellow-500" />;
    }
  };

  // 获取环境颜色样式
  const getEnvironmentStyle = (env: string, isActive: boolean) => {
    const baseStyle = "flex items-center gap-2 px-4 py-2 rounded-lg border-2 transition-all duration-200 cursor-pointer";
    
    if (env === 'testnet') {
      return isActive
        ? `${baseStyle} bg-orange-100 border-orange-500 text-orange-800`
        : `${baseStyle} bg-gray-50 border-gray-300 text-gray-600 hover:bg-orange-50 hover:border-orange-300`;
    } else {
      return isActive
        ? `${baseStyle} bg-green-100 border-green-500 text-green-800`
        : `${baseStyle} bg-gray-50 border-gray-300 text-gray-600 hover:bg-green-50 hover:border-green-300`;
    }
  };

  useEffect(() => {
    fetchStatus();
    
    // 定期刷新状态
    const interval = setInterval(fetchStatus, 30000); // 30秒刷新一次
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center gap-2 mb-4">
          <Settings className="w-5 h-5 text-gray-600" />
          <h3 className="text-lg font-semibold text-gray-800">环境状态</h3>
        </div>
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-gray-600">加载中...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center gap-2 mb-4">
          <AlertCircle className="w-5 h-5 text-red-500" />
          <h3 className="text-lg font-semibold text-gray-800">环境状态</h3>
        </div>
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-700">{error}</p>
          <button
            onClick={fetchStatus}
            className="mt-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            重试
          </button>
        </div>
      </div>
    );
  }

  if (!status) {
    return null;
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      {/* 标题 */}
      <div className="flex items-center gap-2 mb-6">
        <Settings className="w-5 h-5 text-gray-600" />
        <h3 className="text-lg font-semibold text-gray-800">环境切换</h3>
      </div>

      {/* 当前环境状态 */}
      <div className="mb-6">
        <div className="flex items-center gap-3 mb-2">
          <span className="text-sm font-medium text-gray-600">当前环境:</span>
          <div className="flex items-center gap-2">
            {getEnvironmentIcon(status.current_environment)}
            <span className="font-semibold">
              {status.current_environment === 'testnet' ? '测试网环境' : '真实环境'}
            </span>
            {getStatusIcon(status.status)}
          </div>
        </div>
        
        {/* API状态 */}
        <div className="text-sm text-gray-600 space-y-1">
          <div className="flex items-center gap-2">
            <span>Binance API:</span>
            <span className={status.api_status.binance_configured ? 'text-green-600' : 'text-red-600'}>
              {status.api_status.binance_configured ? '已配置' : '未配置'}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <span>DeepSeek API:</span>
            <span className={status.api_status.deepseek_configured ? 'text-green-600' : 'text-red-600'}>
              {status.api_status.deepseek_configured ? '已配置' : '未配置'}
            </span>
          </div>
          {status.api_status.last_validated && (
            <div className="text-xs text-gray-500">
              最后验证: {new Date(status.api_status.last_validated).toLocaleString()}
            </div>
          )}
        </div>
      </div>

      {/* 环境切换按钮 */}
      <div className="space-y-3">
        <h4 className="text-sm font-medium text-gray-700 mb-3">选择环境:</h4>
        
        {/* 测试网环境 */}
        <div className="flex gap-2">
          <div
            className={`${getEnvironmentStyle('testnet', status.current_environment === 'testnet')} flex-1`}
            onClick={() => !switching && handleEnvironmentSwitch('testnet')}
          >
            <TestTube className="w-5 h-5" />
            <div className="flex-1">
              <div className="font-medium">测试网环境</div>
              <div className="text-xs opacity-75">安全的测试环境，用于验证交易策略</div>
            </div>
            {status.current_environment === 'testnet' && (
              <CheckCircle className="w-5 h-5 text-orange-600" />
            )}
          </div>
          <button
            onClick={() => handleShowConfig('testnet')}
            className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
            title="配置 API 密钥"
          >
            <Key className="w-4 h-4 text-gray-600" />
          </button>
        </div>

        {/* 真实环境 */}
        <div className="flex gap-2">
          <div
            className={`${getEnvironmentStyle('mainnet', status.current_environment === 'mainnet')} flex-1`}
            onClick={() => !switching && handleEnvironmentSwitch('mainnet')}
          >
            <DollarSign className="w-5 h-5" />
            <div className="flex-1">
              <div className="font-medium">真实环境</div>
              <div className="text-xs opacity-75">真实交易环境，请谨慎操作</div>
              {!status.api_status.binance_configured && (
                <div className="text-xs text-red-600 mt-1">需要配置 API 密钥</div>
              )}
            </div>
            {status.current_environment === 'mainnet' && (
              <CheckCircle className="w-5 h-5 text-green-600" />
            )}
          </div>
          <button
            onClick={() => handleShowConfig('mainnet')}
            className="flex items-center justify-center w-12 h-12 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
            title="配置 API 密钥"
          >
            <Key className="w-4 h-4 text-gray-600" />
          </button>
        </div>
      </div>

      {/* 切换状态 */}
      {switching && (
        <div className="mt-4 flex items-center gap-2 text-blue-600">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
          <span className="text-sm">正在切换环境...</span>
        </div>
      )}

      {/* 错误信息 */}
      {error && (
        <div className="mt-4 bg-red-50 border border-red-200 rounded-lg p-3">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-red-500" />
            <span className="text-sm text-red-700">{error}</span>
          </div>
        </div>
      )}

      {/* 最后更新时间 */}
      <div className="mt-4 pt-4 border-t border-gray-200">
        <div className="text-xs text-gray-500">
          最后更新: {new Date(status.last_updated).toLocaleString()}
        </div>
      </div>

      {/* 配置表单模态框 */}
      {showConfigForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="max-w-2xl w-full mx-4">
            <EnvironmentConfigForm
              environment={configEnvironment}
              onConfigUpdate={handleConfigUpdate}
              onCancel={handleConfigCancel}
            />
          </div>
        </div>
      )}
    </div>
  );
};