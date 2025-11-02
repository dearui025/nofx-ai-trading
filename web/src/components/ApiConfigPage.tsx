import React, { useState, useEffect } from 'react';
import { Key, Eye, EyeOff, Save, TestTube, DollarSign, CheckCircle, AlertCircle } from 'lucide-react';
import { environmentApi } from '../lib/environmentApi';
import { EnvironmentStatus, EnvironmentType } from '../types/environment';

interface ApiConfigFormData {
  binance_api_key: string;
  binance_secret_key: string;
  deepseek_api_key: string;
}

interface ApiConfigPageProps {
  onConfigUpdate?: () => void;
}

export const ApiConfigPage: React.FC<ApiConfigPageProps> = ({ onConfigUpdate }) => {
  const [status, setStatus] = useState<EnvironmentStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [validating, setValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [selectedEnvironment, setSelectedEnvironment] = useState<EnvironmentType>('testnet');
  
  // 表单数据
  const [testnetConfig, setTestnetConfig] = useState<ApiConfigFormData>({
    binance_api_key: '',
    binance_secret_key: '',
    deepseek_api_key: '',
  });
  
  const [mainnetConfig, setMainnetConfig] = useState<ApiConfigFormData>({
    binance_api_key: '',
    binance_secret_key: '',
    deepseek_api_key: '',
  });

  // 密码显示状态
  const [showSecrets, setShowSecrets] = useState({
    testnet_binance_secret: false,
    testnet_deepseek: false,
    mainnet_binance_secret: false,
    mainnet_deepseek: false,
  });

  // 获取环境状态
  const fetchStatus = async () => {
    try {
      setLoading(true);
      setError(null);
      const envStatus = await environmentApi.getStatus();
      setStatus(envStatus);
      setSelectedEnvironment(envStatus.current_environment as EnvironmentType);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取环境状态失败');
    } finally {
      setLoading(false);
    }
  };

  // 保存配置
  const handleSaveConfig = async (environment: EnvironmentType) => {
    const config = environment === 'testnet' ? testnetConfig : mainnetConfig;
    
    try {
      setSaving(true);
      setError(null);
      setSuccess(null);

      const response = await environmentApi.updateConfig({
        environment,
        binance_api_key: config.binance_api_key,
        binance_secret_key: config.binance_secret_key,
        deepseek_api_key: config.deepseek_api_key,
      });

      if (response.success) {
        setSuccess(`${environment === 'testnet' ? '测试网' : '真实环境'}配置保存成功`);
        await fetchStatus();
        onConfigUpdate?.();
      } else {
        throw new Error(response.message);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存配置失败');
    } finally {
      setSaving(false);
    }
  };

  // 验证配置
  const handleValidateConfig = async (environment: EnvironmentType) => {
    try {
      setValidating(true);
      setError(null);
      setSuccess(null);

      const response = await environmentApi.validateEnvironment({
        environment,
      });

      if (response.valid) {
        setSuccess(`${environment === 'testnet' ? '测试网' : '真实环境'}配置验证成功`);
        await fetchStatus();
      } else {
        setError(`验证失败: ${response.errors.join(', ')}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '验证配置失败');
    } finally {
      setValidating(false);
    }
  };

  // 切换密码显示
  const toggleSecretVisibility = (key: string) => {
    setShowSecrets(prev => ({
      ...prev,
      [key]: !prev[key as keyof typeof prev],
    }));
  };

  // 获取当前配置
  const getCurrentConfig = () => {
    return selectedEnvironment === 'testnet' ? testnetConfig : mainnetConfig;
  };

  // 设置当前配置
  const setCurrentConfig = (config: Partial<ApiConfigFormData>) => {
    if (selectedEnvironment === 'testnet') {
      setTestnetConfig(prev => ({ ...prev, ...config }));
    } else {
      setMainnetConfig(prev => ({ ...prev, ...config }));
    }
  };

  useEffect(() => {
    fetchStatus();
  }, []);

  useEffect(() => {
    // 清除消息
    const timer = setTimeout(() => {
      setSuccess(null);
      setError(null);
    }, 5000);
    return () => clearTimeout(timer);
  }, [success, error]);

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span className="ml-2 text-gray-600">加载中...</span>
          </div>
        </div>
      </div>
    );
  }

  const currentConfig = getCurrentConfig();

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center gap-3 mb-4">
          <Key className="w-6 h-6 text-blue-600" />
          <h1 className="text-2xl font-bold text-gray-800">API配置管理</h1>
        </div>
        <p className="text-gray-600">
          配置测试网和真实环境的API密钥，确保交易系统正常运行。
        </p>
      </div>

      {/* 环境选择 */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">选择环境</h2>
        <div className="flex gap-4">
          <button
            onClick={() => setSelectedEnvironment('testnet')}
            className={`flex items-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
              selectedEnvironment === 'testnet'
                ? 'bg-orange-100 border-orange-500 text-orange-800'
                : 'bg-gray-50 border-gray-300 text-gray-600 hover:bg-orange-50 hover:border-orange-300'
            }`}
          >
            <TestTube className="w-5 h-5" />
            <span className="font-medium">测试网环境</span>
            {status?.current_environment === 'testnet' && (
              <CheckCircle className="w-4 h-4 text-orange-600" />
            )}
          </button>
          
          <button
            onClick={() => setSelectedEnvironment('mainnet')}
            className={`flex items-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
              selectedEnvironment === 'mainnet'
                ? 'bg-green-100 border-green-500 text-green-800'
                : 'bg-gray-50 border-gray-300 text-gray-600 hover:bg-green-50 hover:border-green-300'
            }`}
          >
            <DollarSign className="w-5 h-5" />
            <span className="font-medium">真实环境</span>
            {status?.current_environment === 'mainnet' && (
              <CheckCircle className="w-4 h-4 text-green-600" />
            )}
          </button>
        </div>
      </div>

      {/* API配置表单 */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-gray-800">
            {selectedEnvironment === 'testnet' ? '测试网' : '真实环境'} API配置
          </h2>
          <div className="flex gap-2">
            <button
              onClick={() => handleValidateConfig(selectedEnvironment)}
              disabled={validating}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {validating ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              ) : (
                <CheckCircle className="w-4 h-4" />
              )}
              验证配置
            </button>
            <button
              onClick={() => handleSaveConfig(selectedEnvironment)}
              disabled={saving}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {saving ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              ) : (
                <Save className="w-4 h-4" />
              )}
              保存配置
            </button>
          </div>
        </div>

        <div className="space-y-6">
          {/* Binance API配置 */}
          <div className="border border-gray-200 rounded-lg p-4">
            <h3 className="text-md font-medium text-gray-800 mb-4">Binance API配置</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  API Key
                </label>
                <input
                  type="text"
                  value={currentConfig.binance_api_key}
                  onChange={(e) => setCurrentConfig({ binance_api_key: e.target.value })}
                  placeholder="输入Binance API Key"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Secret Key
                </label>
                <div className="relative">
                  <input
                    type={showSecrets[`${selectedEnvironment}_binance_secret` as keyof typeof showSecrets] ? 'text' : 'password'}
                    value={currentConfig.binance_secret_key}
                    onChange={(e) => setCurrentConfig({ binance_secret_key: e.target.value })}
                    placeholder="输入Binance Secret Key"
                    className="w-full px-3 py-2 pr-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => toggleSecretVisibility(`${selectedEnvironment}_binance_secret`)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  >
                    {showSecrets[`${selectedEnvironment}_binance_secret` as keyof typeof showSecrets] ? (
                      <EyeOff className="w-4 h-4 text-gray-400" />
                    ) : (
                      <Eye className="w-4 h-4 text-gray-400" />
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* DeepSeek API配置 */}
          <div className="border border-gray-200 rounded-lg p-4">
            <h3 className="text-md font-medium text-gray-800 mb-4">DeepSeek API配置</h3>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                API Key
              </label>
              <div className="relative">
                <input
                  type={showSecrets[`${selectedEnvironment}_deepseek` as keyof typeof showSecrets] ? 'text' : 'password'}
                  value={currentConfig.deepseek_api_key}
                  onChange={(e) => setCurrentConfig({ deepseek_api_key: e.target.value })}
                  placeholder="输入DeepSeek API Key"
                  className="w-full px-3 py-2 pr-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
                <button
                  type="button"
                  onClick={() => toggleSecretVisibility(`${selectedEnvironment}_deepseek`)}
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                >
                  {showSecrets[`${selectedEnvironment}_deepseek` as keyof typeof showSecrets] ? (
                    <EyeOff className="w-4 h-4 text-gray-400" />
                  ) : (
                    <Eye className="w-4 h-4 text-gray-400" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* 状态消息 */}
        {success && (
          <div className="mt-4 bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="w-5 h-5 text-green-500" />
              <span className="text-green-700">{success}</span>
            </div>
          </div>
        )}

        {error && (
          <div className="mt-4 bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-5 h-5 text-red-500" />
              <span className="text-red-700">{error}</span>
            </div>
          </div>
        )}
      </div>

      {/* 配置说明 */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-lg font-medium text-blue-800 mb-3">配置说明</h3>
        <div className="space-y-2 text-sm text-blue-700">
          <p>• <strong>测试网环境</strong>: 使用Binance测试网API，资金为虚拟资金，用于策略验证</p>
          <p>• <strong>真实环境</strong>: 使用Binance真实API，涉及真实资金，请谨慎操作</p>
          <p>• <strong>DeepSeek API</strong>: 用于AI决策分析，两个环境可以共享同一个API密钥</p>
          <p>• 配置保存后，系统将自动验证API密钥的有效性和权限</p>
          <p>• 建议先在测试网环境验证策略，确认无误后再切换到真实环境</p>
        </div>
      </div>
    </div>
  );
};