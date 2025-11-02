import { useState } from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { t } from '../i18n/translations';

interface ConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (config: ConfigData) => void;
}

interface ConfigData {
  qwenApiKey: string;
  hyperliquidPrivateKey: string;
  asterApiKey: string;
}

export function ConfigModal({ isOpen, onClose, onSave }: ConfigModalProps) {
  const { language } = useLanguage();
  const [config, setConfig] = useState<ConfigData>({
    qwenApiKey: '',
    hyperliquidPrivateKey: '',
    asterApiKey: ''
  });

  const handleSave = () => {
    onSave(config);
    onClose();
  };

  const handleInputChange = (field: keyof ConfigData, value: string) => {
    setConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* 背景遮罩 */}
      <div 
        className="absolute inset-0 bg-black bg-opacity-50 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* 弹窗内容 */}
      <div className="relative w-full max-w-md mx-4 bg-gray-900 rounded-lg shadow-2xl border border-gray-700">
        {/* 标题栏 */}
        <div className="flex items-center justify-between p-6 border-b border-gray-700">
          <h2 className="text-xl font-bold text-white">
            {t('getKeys', language) || '获取密钥'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* 表单内容 */}
        <div className="p-6 space-y-4">
          {/* QWEN API Key */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              QWEN_API_KEY
            </label>
            <input
              type="password"
              value={config.qwenApiKey}
              onChange={(e) => handleInputChange('qwenApiKey', e.target.value)}
              placeholder={t('enterKey', language, { key: 'QWEN_API_KEY' }) || '请输入密钥 QWEN_API_KEY'}
              className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:border-transparent"
            />
          </div>

          {/* HYPERLIQUID PRIVATE KEY */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              HYPERLIQUID_PRIVATE_KEY
            </label>
            <input
              type="password"
              value={config.hyperliquidPrivateKey}
              onChange={(e) => handleInputChange('hyperliquidPrivateKey', e.target.value)}
              placeholder={t('enterKey', language, { key: 'HYPERLIQUID_PRIVATE_KEY' }) || '请输入密钥 HYPERLIQUID_PRIVATE_KEY'}
              className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:border-transparent"
            />
          </div>

          {/* ASTER API KEY */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              ASTER_API_KEY
            </label>
            <input
              type="password"
              value={config.asterApiKey}
              onChange={(e) => handleInputChange('asterApiKey', e.target.value)}
              placeholder={t('enterKey', language, { key: 'ASTER_API_KEY' }) || '请输入密钥 ASTER_API_KEY'}
              className="w-full px-3 py-2 bg-gray-800 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* 底部按钮 */}
        <div className="flex justify-end p-6 border-t border-gray-700">
          <button
            onClick={handleSave}
            className="px-6 py-2 bg-yellow-500 hover:bg-yellow-600 text-black font-semibold rounded-md transition-colors"
          >
            {t('submit', language) || '提交'}
          </button>
        </div>
      </div>
    </div>
  );
}