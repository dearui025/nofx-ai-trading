import React, { useState } from 'react';
import { Eye, EyeOff, Key, Shield, AlertCircle, CheckCircle, Loader } from 'lucide-react';
import { environmentApi } from '../lib/environmentApi';
import { EnvironmentType } from '../types/environment';

interface EnvironmentConfigFormProps {
  environment: EnvironmentType;
  onConfigUpdate: () => void;
  onCancel: () => void;
}

interface FormData {
  binance_api_key: string;
  binance_secret_key: string;
  deepseek_api_key: string;
  oi_top_api_url: string;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
  permissions: string[];
}

export const EnvironmentConfigForm: React.FC<EnvironmentConfigFormProps> = ({
  environment,
  onConfigUpdate,
  onCancel,
}) => {
  const [formData, setFormData] = useState<FormData>({
    binance_api_key: '',
    binance_secret_key: '',
    deepseek_api_key: '',
    oi_top_api_url: '',
  });

  const [showKeys, setShowKeys] = useState({
    binance_api_key: false,
    binance_secret_key: false,
    deepseek_api_key: false,
  });

  const [loading, setLoading] = useState(false);
  const [validating, setValidating] = useState(false);
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // 清除之前的验证结果
    if (validationResult) {
      setValidationResult(null);
    }
  };

  const toggleShowKey = (field: keyof typeof showKeys) => {
    setShowKeys(prev => ({ ...prev, [field]: !prev[field] }));
  };

  const validateConfiguration = async () => {
    if (!formData.binance_api_key || !formData.binance_secret_key) {
      setError('Binance API Key 和 Secret Key 是必填项');
      return;
    }

    try {
      setValidating(true);
      setError(null);

      const response = await environmentApi.validateEnvironment({
        environment,
        api_keys: {
          binance_api_key: formData.binance_api_key,
          binance_secret_key: formData.binance_secret_key,
          deepseek_api_key: formData.deepseek_api_key,
        },
      });

      setValidationResult({
        valid: response.valid,
        errors: response.errors || [],
        permissions: response.permissions || [],
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : '验证失败');
    } finally {
      setValidating(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.binance_api_key || !formData.binance_secret_key) {
      setError('Binance API Key 和 Secret Key 是必填项');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await environmentApi.updateConfig({
        environment,
        binance_api_key: formData.binance_api_key,
        binance_secret_key: formData.binance_secret_key,
        deepseek_api_key: formData.deepseek_api_key || undefined,
        oi_top_api_url: formData.oi_top_api_url || undefined,
      });

      if (response.success) {
        onConfigUpdate();
      } else {
        setError(response.message || '配置更新失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '配置更新失败');
    } finally {
      setLoading(false);
    }
  };

  const getEnvironmentTitle = () => {
    return environment === 'mainnet' ? '真实环境配置' : '测试网环境配置';
  };

  const getEnvironmentDescription = () => {
    return environment === 'mainnet' 
      ? '配置真实交易环境的 API 密钥。请确保密钥具有正确的权限。'
      : '配置测试网环境的 API 密钥。这是一个安全的测试环境。';
  };

  return (
    <div className="bg-white rounded-lg shadow-lg p-6 max-w-2xl mx-auto">
      {/* 标题 */}
      <div className="flex items-center gap-3 mb-6">
        <Key className="w-6 h-6 text-blue-600" />
        <div>
          <h2 className="text-xl font-semibold text-gray-800">{getEnvironmentTitle()}</h2>
          <p className="text-sm text-gray-600 mt-1">{getEnvironmentDescription()}</p>
        </div>
      </div>

      {/* 安全提示 */}
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
        <div className="flex items-start gap-3">
          <Shield className="w-5 h-5 text-yellow-600 mt-0.5" />
          <div className="text-sm">
            <p className="font-medium text-yellow-800 mb-1">安全提示</p>
            <ul className="text-yellow-700 space-y-1">
              <li>• 请确保 API 密钥来自可信来源</li>
              <li>• 建议为交易机器人创建专用的 API 密钥</li>
              <li>• 定期检查和更新 API 密钥权限</li>
              {environment === 'mainnet' && (
                <li>• 真实环境将使用真实资金，请谨慎操作</li>
              )}
            </ul>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Binance API Key */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Binance API Key *
          </label>
          <div className="relative">
            <input
              type={showKeys.binance_api_key ? 'text' : 'password'}
              value={formData.binance_api_key}
              onChange={(e) => handleInputChange('binance_api_key', e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pr-12"
              placeholder="输入 Binance API Key"
              required
            />
            <button
              type="button"
              onClick={() => toggleShowKey('binance_api_key')}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
            >
              {showKeys.binance_api_key ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Binance Secret Key */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Binance Secret Key *
          </label>
          <div className="relative">
            <input
              type={showKeys.binance_secret_key ? 'text' : 'password'}
              value={formData.binance_secret_key}
              onChange={(e) => handleInputChange('binance_secret_key', e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pr-12"
              placeholder="输入 Binance Secret Key"
              required
            />
            <button
              type="button"
              onClick={() => toggleShowKey('binance_secret_key')}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
            >
              {showKeys.binance_secret_key ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* DeepSeek API Key */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            DeepSeek API Key (可选)
          </label>
          <div className="relative">
            <input
              type={showKeys.deepseek_api_key ? 'text' : 'password'}
              value={formData.deepseek_api_key}
              onChange={(e) => handleInputChange('deepseek_api_key', e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pr-12"
              placeholder="输入 DeepSeek API Key (可选)"
            />
            <button
              type="button"
              onClick={() => toggleShowKey('deepseek_api_key')}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
            >
              {showKeys.deepseek_api_key ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* OI Top API URL */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            OI Top API URL (可选)
          </label>
          <input
            type="url"
            value={formData.oi_top_api_url}
            onChange={(e) => handleInputChange('oi_top_api_url', e.target.value)}
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="输入 OI Top API URL (例如: https://api.example.com/oi-top)"
          />
          <p className="text-xs text-gray-500 mt-1">
            用于获取持仓量排行数据的API地址。API需要返回包含symbol、rank、current_oi等字段的JSON格式数据。
          </p>
        </div>

        {/* 验证按钮 */}
        <div className="flex gap-3">
          <button
            type="button"
            onClick={validateConfiguration}
            disabled={validating || !formData.binance_api_key || !formData.binance_secret_key}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {validating ? (
              <Loader className="w-4 h-4 animate-spin" />
            ) : (
              <Shield className="w-4 h-4" />
            )}
            {validating ? '验证中...' : '验证配置'}
          </button>
        </div>

        {/* 验证结果 */}
        {validationResult && (
          <div className={`rounded-lg p-4 ${validationResult.valid ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
            <div className="flex items-start gap-3">
              {validationResult.valid ? (
                <CheckCircle className="w-5 h-5 text-green-600 mt-0.5" />
              ) : (
                <AlertCircle className="w-5 h-5 text-red-600 mt-0.5" />
              )}
              <div className="flex-1">
                <p className={`font-medium ${validationResult.valid ? 'text-green-800' : 'text-red-800'}`}>
                  {validationResult.valid ? '配置验证成功' : '配置验证失败'}
                </p>
                
                {validationResult.errors.length > 0 && (
                  <ul className="mt-2 text-sm text-red-700 space-y-1">
                    {validationResult.errors.map((error, index) => (
                      <li key={index}>• {error}</li>
                    ))}
                  </ul>
                )}
                
                {validationResult.permissions.length > 0 && (
                  <div className="mt-2">
                    <p className="text-sm font-medium text-green-700">检测到的权限:</p>
                    <ul className="text-sm text-green-600 space-y-1">
                      {validationResult.permissions.map((permission, index) => (
                        <li key={index}>• {permission}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* 错误信息 */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center gap-3">
              <AlertCircle className="w-5 h-5 text-red-600" />
              <p className="text-red-700">{error}</p>
            </div>
          </div>
        )}

        {/* 操作按钮 */}
        <div className="flex gap-3 pt-4 border-t border-gray-200">
          <button
            type="submit"
            disabled={loading || !formData.binance_api_key || !formData.binance_secret_key}
            className="flex items-center gap-2 px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? (
              <Loader className="w-4 h-4 animate-spin" />
            ) : (
              <CheckCircle className="w-4 h-4" />
            )}
            {loading ? '保存中...' : '保存配置'}
          </button>
          
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            className="px-6 py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors"
          >
            取消
          </button>
        </div>
      </form>
    </div>
  );
};