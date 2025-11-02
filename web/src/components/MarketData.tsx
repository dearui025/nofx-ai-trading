import useSWR from 'swr';
import { api } from '../lib/api';
import { useLanguage } from '../contexts/LanguageContext';
import { t } from '../i18n/translations';
import { useState } from 'react';

interface MarketDataItem {
  symbol: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  volume24h: number;
}

export default function MarketData() {
  const { language } = useLanguage();
  const [isManualRefreshing, setIsManualRefreshing] = useState(false);
  
  const { data: marketData, error, isLoading, mutate } = useSWR(
    'marketData',
    api.getTopGainers,
    {
      refreshInterval: 30000
    }
  );

  // 手动刷新功能
  const handleManualRefresh = async () => {
    setIsManualRefreshing(true);
    try {
      await mutate();
    } catch (error) {
      console.error('手动刷新失败:', error);
    } finally {
      setIsManualRefreshing(false);
    }
  };

  if (error) {
    return (
      <div className="binance-card p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold flex items-center gap-2" style={{ color: '#EAECEF' }}>
            📈 {t('liveMarket', language)}
          </h3>
          <button
            onClick={handleManualRefresh}
            disabled={isManualRefreshing}
            className="text-xs px-3 py-1 rounded transition-all duration-200 hover:scale-105 disabled:opacity-50"
            style={{ 
              background: 'rgba(246, 70, 93, 0.1)', 
              color: '#F6465D',
              border: '1px solid rgba(246, 70, 93, 0.2)'
            }}
          >
            {isManualRefreshing ? '🔄' : '🔄'} {t('retry', language)}
          </button>
        </div>
        <div className="text-center" style={{ color: '#F6465D' }}>
          <div className="text-lg mb-2">⚠️</div>
          <div className="text-sm">{t('marketDataError', language)}</div>
        </div>
      </div>
    );
  }

  if (isLoading && !marketData) {
    return (
      <div className="binance-card p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold flex items-center gap-2" style={{ color: '#EAECEF' }}>
            📈 {t('liveMarket', language)}
          </h3>
          <div className="text-xs px-2 py-1 rounded" style={{ 
            background: 'rgba(132, 142, 156, 0.1)', 
            color: '#848E9C',
            border: '1px solid rgba(132, 142, 156, 0.2)'
          }}>
            🔄 {t('loading', language)}
          </div>
        </div>
        <div className="animate-pulse">
          <div className="space-y-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="flex justify-between items-center p-3 rounded" style={{ 
                background: '#0B0E11',
                border: '1px solid #2B3139'
              }}>
                <div className="skeleton h-4 w-16"></div>
                <div className="skeleton h-4 w-20"></div>
                <div className="skeleton h-4 w-12"></div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="binance-card p-4 animate-slide-in" style={{ animationDelay: '0.2s' }}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold flex items-center gap-2" style={{ color: '#EAECEF' }}>
          📈 {t('liveMarket', language)}
          <span className="text-xs px-2 py-1 rounded" style={{ 
            background: 'rgba(14, 203, 129, 0.2)', 
            color: '#0ECB81',
            border: '1px solid rgba(14, 203, 129, 0.3)'
          }}>
            🔴 LIVE
          </span>
        </h3>
        <div className="flex items-center gap-2">
          {/* 数据状态指示器 */}
          <div className="text-xs px-2 py-1 rounded" style={{ 
            background: marketData && marketData.length > 0 
              ? 'rgba(14, 203, 129, 0.1)' 
              : 'rgba(132, 142, 156, 0.1)', 
            color: marketData && marketData.length > 0 
              ? '#0ECB81' 
              : '#848E9C',
            border: marketData && marketData.length > 0 
              ? '1px solid rgba(14, 203, 129, 0.2)' 
              : '1px solid rgba(132, 142, 156, 0.2)'
          }}>
            ● {marketData && marketData.length > 0 ? t('live', language) : t('cached', language)}
          </div>
          
          {/* 手动刷新按钮 */}
          <button
            onClick={handleManualRefresh}
            disabled={isManualRefreshing || isLoading}
            className="text-xs px-2 py-1 rounded transition-all duration-200 hover:scale-105 disabled:opacity-50"
            style={{ 
              background: 'rgba(14, 203, 129, 0.1)', 
              color: '#0ECB81',
              border: '1px solid rgba(14, 203, 129, 0.2)'
            }}
            title={t('refreshData', language)}
          >
            <span className={`${isManualRefreshing || isLoading ? 'animate-spin' : ''}`}>
              🔄
            </span>
          </button>
        </div>
      </div>
      
      <div className="space-y-2">
        {marketData?.slice(0, 6).map((item) => (
          <div
            key={item.symbol}
            className="flex items-center justify-between p-3 rounded transition-all duration-200 hover:scale-[1.02]"
            style={{ 
              background: '#0B0E11',
              border: '1px solid #2B3139'
            }}
          >
            {/* Symbol & Price */}
            <div className="flex items-center gap-3">
              <div className="text-lg">₿</div>
              <div>
                <div className="font-bold text-sm" style={{ color: '#EAECEF' }}>
                  {item.symbol?.replace('USDT', '') || 'N/A'}
                </div>
                <div className="text-xs mono" style={{ color: '#848E9C' }}>
                  {(item.price || 0).toFixed(4)} USDT
                </div>
              </div>
            </div>

            {/* Change */}
            <div className="text-right">
              <div
                className="text-sm font-bold mono"
                style={{ 
                  color: (item.changePercent24h || 0) >= 0 ? '#0ECB81' : '#F6465D' 
                }}
              >
                {(item.changePercent24h || 0) >= 0 ? '+' : ''}{(item.changePercent24h || 0).toFixed(2)}%
              </div>
              <div
                className="text-xs mono"
                style={{ 
                  color: (item.change24h || 0) >= 0 ? '#0ECB81' : '#F6465D' 
                }}
              >
                {(item.change24h || 0) >= 0 ? '+' : ''}{(item.change24h || 0).toFixed(2)}
              </div>
            </div>

            {/* Volume */}
            <div className="text-right">
              <div className="text-xs" style={{ color: '#848E9C' }}>{t('vol', language)}</div>
              <div className="text-xs mono" style={{ color: '#EAECEF' }}>
                {((item.volume24h || 0) / 1000000).toFixed(1)}M
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-4 pt-3 border-t" style={{ borderColor: '#2B3139' }}>
        <div className="text-xs text-center" style={{ color: '#848E9C' }}>
          {t('dataSource', language)}: Binance Futures API
        </div>
      </div>
    </div>
  );
}
