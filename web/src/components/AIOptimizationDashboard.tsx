import React, { useState, useEffect } from 'react';
import useSWR from 'swr';
import { api } from '../lib/api';

interface MarketRegime {
  regime: 'bull' | 'bear' | 'sideways';
  confidence: number;
  last_updated: string;
}

interface SignalStrength {
  indicator: string;
  strength: number;
  direction: 'bullish' | 'bearish' | 'neutral';
  confidence: number;
}

interface CorrelationRisk {
  symbol1: string;
  symbol2: string;
  correlation: number;
  risk_level: 'low' | 'medium' | 'high';
}

interface SystemHealth {
  status: 'healthy' | 'warning' | 'critical';
  uptime: number;
  last_backup: string;
  recovery_ready: boolean;
}

interface AIOptimizationData {
  market_regime: MarketRegime;
  signal_strengths: SignalStrength[];
  correlation_risks: CorrelationRisk[];
  system_health: SystemHealth;
}

const AIOptimizationDashboard: React.FC = () => {
  const [lastUpdate, setLastUpdate] = useState<string>('--:--:--');

  // 获取AI优化数据
  const { data: optimizationData, error } = useSWR<AIOptimizationData>(
    'ai-optimization',
    () => api.getAIOptimizationData?.() || Promise.resolve({
      market_regime: {
        regime: 'sideways',
        confidence: 0.75,
        last_updated: new Date().toISOString()
      },
      signal_strengths: [
        { indicator: 'RSI', strength: 0.65, direction: 'bullish', confidence: 0.8 },
        { indicator: 'MACD', strength: 0.45, direction: 'bearish', confidence: 0.7 },
        { indicator: 'Bollinger Bands', strength: 0.55, direction: 'neutral', confidence: 0.6 },
        { indicator: 'Moving Average', strength: 0.75, direction: 'bullish', confidence: 0.85 }
      ],
      correlation_risks: [
        { symbol1: 'BTCUSDT', symbol2: 'ETHUSDT', correlation: 0.85, risk_level: 'high' },
        { symbol1: 'BTCUSDT', symbol2: 'ADAUSDT', correlation: 0.65, risk_level: 'medium' },
        { symbol1: 'ETHUSDT', symbol2: 'ADAUSDT', correlation: 0.55, risk_level: 'medium' }
      ],
      system_health: {
        status: 'healthy',
        uptime: 99.8,
        last_backup: new Date(Date.now() - 3600000).toISOString(),
        recovery_ready: true
      }
    }),
    {
      refreshInterval: 15000,
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    }
  );

  useEffect(() => {
    if (optimizationData) {
      const now = new Date().toLocaleTimeString();
      setLastUpdate(now);
    }
  }, [optimizationData]);

  const getRegimeColor = (regime: string) => {
    switch (regime) {
      case 'bull': return '#0ECB81';
      case 'bear': return '#F6465D';
      case 'sideways': return '#F0B90B';
      default: return '#848E9C';
    }
  };

  const getRegimeIcon = (regime: string) => {
    switch (regime) {
      case 'bull': return '📈';
      case 'bear': return '📉';
      case 'sideways': return '↔️';
      default: return '❓';
    }
  };

  const getDirectionColor = (direction: string) => {
    switch (direction) {
      case 'bullish': return '#0ECB81';
      case 'bearish': return '#F6465D';
      case 'neutral': return '#F0B90B';
      default: return '#848E9C';
    }
  };

  const getRiskColor = (level: string) => {
    switch (level) {
      case 'low': return '#0ECB81';
      case 'medium': return '#F0B90B';
      case 'high': return '#F6465D';
      default: return '#848E9C';
    }
  };

  const getHealthColor = (status: string) => {
    switch (status) {
      case 'healthy': return '#0ECB81';
      case 'warning': return '#F0B90B';
      case 'critical': return '#F6465D';
      default: return '#848E9C';
    }
  };

  if (error) {
    return (
      <div className="p-6">
        <div className="rounded-lg p-4" style={{ background: '#1E2329', border: '1px solid #F6465D' }}>
          <p style={{ color: '#F6465D' }}>加载AI优化数据时出错</p>
        </div>
      </div>
    );
  }

  if (!optimizationData) {
    return (
      <div className="p-6">
        <div className="rounded-lg p-4" style={{ background: '#1E2329' }}>
          <p style={{ color: '#848E9C' }}>加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#EAECEF' }}>
            AI决策系统优化监控
          </h1>
          <p className="text-sm" style={{ color: '#848E9C' }}>
            最后更新: {lastUpdate}
          </p>
        </div>
        <div className="flex items-center gap-2 px-3 py-2 rounded" style={{ background: '#1E2329' }}>
          <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
          <span className="text-sm font-medium" style={{ color: '#0ECB81' }}>实时监控中</span>
        </div>
      </div>

      {/* Market Regime Panel */}
      <div className="rounded-lg p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <h2 className="text-lg font-semibold mb-4" style={{ color: '#EAECEF' }}>
          市场状态监控
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-3xl mb-2">
              {getRegimeIcon(optimizationData.market_regime.regime)}
            </div>
            <div className="text-xl font-bold mb-1" style={{ color: getRegimeColor(optimizationData.market_regime.regime) }}>
              {optimizationData.market_regime.regime === 'bull' ? '牛市' : 
               optimizationData.market_regime.regime === 'bear' ? '熊市' : '震荡'}
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              当前状态
            </div>
          </div>
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-2xl font-bold mb-1" style={{ color: '#F0B90B' }}>
              {(optimizationData.market_regime.confidence * 100).toFixed(1)}%
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              置信度
            </div>
          </div>
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
              {new Date(optimizationData.market_regime.last_updated).toLocaleTimeString()}
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              最后更新
            </div>
          </div>
        </div>
      </div>

      {/* Signal Strength Panel */}
      <div className="rounded-lg p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <h2 className="text-lg font-semibold mb-4" style={{ color: '#EAECEF' }}>
          信号强度量化
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {optimizationData.signal_strengths.map((signal, index) => (
            <div key={index} className="p-4 rounded" style={{ background: '#0B0E11' }}>
              <div className="flex justify-between items-center mb-2">
                <span className="font-medium" style={{ color: '#EAECEF' }}>
                  {signal.indicator}
                </span>
                <span className="text-sm px-2 py-1 rounded" style={{ 
                  background: getDirectionColor(signal.direction) + '20',
                  color: getDirectionColor(signal.direction)
                }}>
                  {signal.direction === 'bullish' ? '看涨' : 
                   signal.direction === 'bearish' ? '看跌' : '中性'}
                </span>
              </div>
              <div className="mb-2">
                <div className="flex justify-between text-sm mb-1">
                  <span style={{ color: '#848E9C' }}>强度</span>
                  <span style={{ color: '#EAECEF' }}>{(signal.strength * 100).toFixed(0)}%</span>
                </div>
                <div className="w-full bg-gray-700 rounded-full h-2">
                  <div 
                    className="h-2 rounded-full transition-all duration-300"
                    style={{ 
                      width: `${signal.strength * 100}%`,
                      background: getDirectionColor(signal.direction)
                    }}
                  />
                </div>
              </div>
              <div className="text-xs" style={{ color: '#848E9C' }}>
                置信度: {(signal.confidence * 100).toFixed(0)}%
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Correlation Risk Panel */}
      <div className="rounded-lg p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <h2 className="text-lg font-semibold mb-4" style={{ color: '#EAECEF' }}>
          相关性风险监控
        </h2>
        <div className="space-y-3">
          {optimizationData.correlation_risks.map((risk, index) => (
            <div key={index} className="flex items-center justify-between p-3 rounded" style={{ background: '#0B0E11' }}>
              <div className="flex items-center gap-3">
                <span className="font-medium" style={{ color: '#EAECEF' }}>
                  {risk.symbol1} ↔ {risk.symbol2}
                </span>
                <span className="text-sm px-2 py-1 rounded" style={{
                  background: getRiskColor(risk.risk_level) + '20',
                  color: getRiskColor(risk.risk_level)
                }}>
                  {risk.risk_level === 'low' ? '低风险' : 
                   risk.risk_level === 'medium' ? '中风险' : '高风险'}
                </span>
              </div>
              <div className="text-right">
                <div className="font-bold" style={{ color: '#EAECEF' }}>
                  {(risk.correlation * 100).toFixed(0)}%
                </div>
                <div className="text-xs" style={{ color: '#848E9C' }}>
                  相关性
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* System Health Panel */}
      <div className="rounded-lg p-6" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <h2 className="text-lg font-semibold mb-4" style={{ color: '#EAECEF' }}>
          系统健康状态
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-2xl mb-2">
              {optimizationData.system_health.status === 'healthy' ? '✅' : 
               optimizationData.system_health.status === 'warning' ? '⚠️' : '❌'}
            </div>
            <div className="font-bold mb-1" style={{ color: getHealthColor(optimizationData.system_health.status) }}>
              {optimizationData.system_health.status === 'healthy' ? '健康' : 
               optimizationData.system_health.status === 'warning' ? '警告' : '严重'}
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              系统状态
            </div>
          </div>
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-xl font-bold mb-1" style={{ color: '#0ECB81' }}>
              {optimizationData.system_health.uptime.toFixed(1)}%
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              运行时间
            </div>
          </div>
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
              {new Date(optimizationData.system_health.last_backup).toLocaleTimeString()}
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              最后备份
            </div>
          </div>
          <div className="text-center p-4 rounded" style={{ background: '#0B0E11' }}>
            <div className="text-2xl mb-2">
              {optimizationData.system_health.recovery_ready ? '🛡️' : '⚠️'}
            </div>
            <div className="font-bold mb-1" style={{ 
              color: optimizationData.system_health.recovery_ready ? '#0ECB81' : '#F0B90B' 
            }}>
              {optimizationData.system_health.recovery_ready ? '就绪' : '未就绪'}
            </div>
            <div className="text-sm" style={{ color: '#848E9C' }}>
              灾难恢复
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AIOptimizationDashboard;