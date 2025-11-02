import React, { useState, useEffect } from 'react';
import { 
  riskControlV2Api,
  type RiskControlV2SystemStatus,
  type RiskStatus,
  type TimeStatus,
  type AICommitteeStatus,
  type LiquidityStatus,
  type SharpeStatus
} from '../lib/riskControlV2Api';

interface DashboardProps {
  refreshInterval?: number;
}

const RiskControlV2Dashboard: React.FC<DashboardProps> = ({ 
  refreshInterval = 30000 // 30秒刷新间隔
}) => {
  const [systemStatus, setSystemStatus] = useState<RiskControlV2SystemStatus | null>(null);
  const [riskStatus, setRiskStatus] = useState<RiskStatus | null>(null);
  const [timeStatus, setTimeStatus] = useState<TimeStatus | null>(null);
  const [aiStatus, setAiStatus] = useState<AICommitteeStatus | null>(null);
  const [liquidityStatus, setLiquidityStatus] = useState<LiquidityStatus | null>(null);
  const [sharpeStatus, setSharpeStatus] = useState<SharpeStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // 获取所有状态数据
  const fetchAllData = async () => {
    try {
      setError(null);
      const [
        systemData,
        riskData,
        timeData,
        aiData,
        liquidityData,
        sharpeData
      ] = await Promise.all([
        riskControlV2Api.getSystemStatus(),
        riskControlV2Api.getRiskStatus(),
        riskControlV2Api.getTimeStatus(),
        riskControlV2Api.getAICommitteeStatus(),
        riskControlV2Api.getLiquidityStatus(),
        riskControlV2Api.getSharpeStatus()
      ]);

      setSystemStatus(systemData);
      setRiskStatus(riskData);
      setTimeStatus(timeData);
      setAiStatus(aiData);
      setLiquidityStatus(liquidityData);
      setSharpeStatus(sharpeData);
      setLastUpdate(new Date());
    } catch (err) {
      console.error('获取风控v2数据失败:', err);
      setError(err instanceof Error ? err.message : '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAllData();
    const interval = setInterval(fetchAllData, refreshInterval);
    return () => clearInterval(interval);
  }, [refreshInterval]);

  // 状态指示器组件
  const StatusIndicator: React.FC<{ 
    status: 'healthy' | 'warning' | 'critical' | 'unknown';
    label: string;
  }> = ({ status, label }) => {
    const getStatusColor = () => {
      switch (status) {
        case 'healthy': return 'bg-green-500';
        case 'warning': return 'bg-yellow-500';
        case 'critical': return 'bg-red-500';
        default: return 'bg-gray-500';
      }
    };

    return (
      <div className="flex items-center space-x-2">
        <div className={`w-3 h-3 rounded-full ${getStatusColor()}`}></div>
        <span className="text-sm font-medium">{label}</span>
      </div>
    );
  };

  // 模块状态卡片组件
  const ModuleCard: React.FC<{
    title: string;
    status: string;
    data: any;
    icon: string;
  }> = ({ title, status, data, icon }) => (
    <div className="bg-white rounded-lg shadow-md p-6 border-l-4 border-blue-500">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-3">
          <div className="text-2xl">{icon}</div>
          <h3 className="text-lg font-semibold text-gray-800">{title}</h3>
        </div>
        <StatusIndicator 
          status={status as any} 
          label={status.toUpperCase()} 
        />
      </div>
      <div className="space-y-2 text-sm text-gray-600">
        {Object.entries(data).map(([key, value]) => (
          <div key={key} className="flex justify-between">
            <span className="capitalize">{key.replace(/_/g, ' ')}:</span>
            <span className="font-medium">
              {typeof value === 'boolean' ? (value ? '是' : '否') : 
               typeof value === 'number' ? value.toFixed(2) : 
               value?.toString() || 'N/A'}
            </span>
          </div>
        ))}
      </div>
    </div>
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        <span className="ml-3 text-gray-600">加载风控系统数据...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <div className="flex items-center space-x-3">
          <div className="text-red-500 text-xl">⚠️</div>
          <div>
            <h3 className="text-lg font-semibold text-red-800">数据加载失败</h3>
            <p className="text-red-600">{error}</p>
            <button 
              onClick={fetchAllData}
              className="mt-3 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition-colors"
            >
              重试
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 系统概览 */}
      <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg shadow-lg p-6 text-white">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold mb-2">🛡️ 风控优化系统 v2</h2>
            <p className="text-blue-100">
              系统状态: {systemStatus?.running ? '运行中' : '已停止'} | 
              运行时间: {systemStatus?.uptime || 'N/A'}
            </p>
          </div>
          <div className="text-right">
            <div className="text-sm text-blue-100">最后更新</div>
            <div className="text-lg font-semibold">
              {lastUpdate.toLocaleTimeString()}
            </div>
          </div>
        </div>
      </div>

      {/* 核心指标 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">风险等级</p>
              <p className="text-2xl font-bold text-gray-900">
                {riskStatus?.global_risk_level?.toUpperCase() || 'UNKNOWN'}
              </p>
            </div>
            <div className="text-3xl">📊</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">活跃警报</p>
              <p className="text-2xl font-bold text-gray-900">
                {riskStatus?.total_alerts || 0}
              </p>
            </div>
            <div className="text-3xl">🚨</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">夏普比率</p>
              <p className="text-2xl font-bold text-gray-900">
                {sharpeStatus?.current_sharpe_ratio?.toFixed(3) || 'N/A'}
              </p>
            </div>
            <div className="text-3xl">📈</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">AI决策数</p>
              <p className="text-2xl font-bold text-gray-900">
                {aiStatus?.total_decisions || 0}
              </p>
            </div>
            <div className="text-3xl">🤖</div>
          </div>
        </div>
      </div>

      {/* 模块状态 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 时间管理模块 */}
        <ModuleCard
          title="时间管理"
          status="healthy"
          icon="⏰"
          data={{
            reset_count: timeStatus?.reset_count || 0,
            last_reset: timeStatus?.last_daily_reset ? 
              new Date(timeStatus.last_daily_reset).toLocaleDateString() : 'N/A',
            equity_watermark: timeStatus?.equity_high_watermark || 0
          }}
        />

        {/* AI委员会模块 */}
        <ModuleCard
          title="AI委员会"
          status={aiStatus?.active_models?.length ? "healthy" : "warning"}
          icon="🤖"
          data={{
            strategy: aiStatus?.current_strategy || 'N/A',
            market_condition: aiStatus?.market_condition || 'N/A',
            consensus_rate: aiStatus?.total_decisions ? 
              ((aiStatus.consensus_decisions / aiStatus.total_decisions) * 100).toFixed(1) + '%' : 'N/A',
            active_models: aiStatus?.active_models?.length || 0
          }}
        />

        {/* 流动性监控模块 */}
        <ModuleCard
          title="流动性监控"
          status={liquidityStatus?.monitoring_enabled ? "healthy" : "warning"}
          icon="💧"
          data={{
            monitoring: liquidityStatus?.monitoring_enabled ? '启用' : '禁用',
            symbols_monitored: liquidityStatus?.total_symbols_monitored || 0,
            blacklisted: liquidityStatus?.blacklisted_symbols?.length || 0,
            market_health: liquidityStatus?.market_health || 'N/A'
          }}
        />

        {/* 夏普比率模块 */}
        <ModuleCard
          title="夏普比率"
          status="healthy"
          icon="📊"
          data={{
            current_ratio: sharpeStatus?.current_sharpe_ratio?.toFixed(3) || 'N/A',
            avg_ratio: sharpeStatus?.avg_sharpe_ratio?.toFixed(3) || 'N/A',
            trend: sharpeStatus?.sharpe_trend || 'N/A',
            calculations: sharpeStatus?.calculation_count || 0
          }}
        />
      </div>

      {/* 紧急操作 */}
      {riskStatus?.emergency_stop && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex items-center space-x-3">
            <div className="text-red-500 text-2xl">🛑</div>
            <div>
              <h3 className="text-lg font-semibold text-red-800">紧急停止已激活</h3>
              <p className="text-red-600">系统已进入紧急停止状态，所有交易已暂停</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default RiskControlV2Dashboard;