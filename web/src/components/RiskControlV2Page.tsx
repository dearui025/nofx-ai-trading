import React, { useState, useEffect } from 'react';
import RiskControlV2Dashboard from './RiskControlV2Dashboard';
import { riskControlV2Api } from '../lib/riskControlV2Api';

const RiskControlV2Page: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'dashboard' | 'config' | 'history' | 'operations'>('dashboard');
  const [configs, setConfigs] = useState<any>(null);
  const [alerts, setAlerts] = useState<any[]>([]);
  const [decisions, setDecisions] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // 获取配置数据
  const fetchConfigs = async () => {
    try {
      setLoading(true);
      const configData = await riskControlV2Api.getAllConfigs();
      setConfigs(configData);
    } catch (error) {
      console.error('获取配置失败:', error);
      setMessage({ type: 'error', text: '获取配置失败' });
    } finally {
      setLoading(false);
    }
  };

  // 获取历史数据
  const fetchHistory = async () => {
    try {
      setLoading(true);
      const [alertData, decisionData] = await Promise.all([
        riskControlV2Api.getRiskAlerts(),
        riskControlV2Api.getRiskDecisions(50)
      ]);
      setAlerts(alertData);
      setDecisions(decisionData);
    } catch (error) {
      console.error('获取历史数据失败:', error);
      setMessage({ type: 'error', text: '获取历史数据失败' });
    } finally {
      setLoading(false);
    }
  };

  // 紧急停止
  const handleEmergencyStop = async () => {
    if (!confirm('确定要执行紧急停止吗？这将暂停所有交易活动。')) {
      return;
    }
    
    try {
      setLoading(true);
      await riskControlV2Api.emergencyStop();
      setMessage({ type: 'success', text: '紧急停止已执行' });
    } catch (error) {
      console.error('紧急停止失败:', error);
      setMessage({ type: 'error', text: '紧急停止失败' });
    } finally {
      setLoading(false);
    }
  };

  // 恢复运行
  const handleResume = async () => {
    if (!confirm('确定要恢复系统运行吗？')) {
      return;
    }
    
    try {
      setLoading(true);
      await riskControlV2Api.resumeRisk();
      setMessage({ type: 'success', text: '系统已恢复运行' });
    } catch (error) {
      console.error('恢复运行失败:', error);
      setMessage({ type: 'error', text: '恢复运行失败' });
    } finally {
      setLoading(false);
    }
  };

  // 手动重置
  const handleManualReset = async () => {
    if (!confirm('确定要执行手动重置吗？这将重置时间管理器状态。')) {
      return;
    }
    
    try {
      setLoading(true);
      await riskControlV2Api.manualReset();
      setMessage({ type: 'success', text: '手动重置已执行' });
    } catch (error) {
      console.error('手动重置失败:', error);
      setMessage({ type: 'error', text: '手动重置失败' });
    } finally {
      setLoading(false);
    }
  };

  // 数据清理
  const handleDataCleanup = async () => {
    const days = prompt('请输入要保留的天数（将删除更早的数据）:', '30');
    if (!days || isNaN(Number(days))) {
      return;
    }
    
    if (!confirm(`确定要删除 ${days} 天前的数据吗？此操作不可撤销。`)) {
      return;
    }
    
    try {
      setLoading(true);
      await riskControlV2Api.cleanupOldData(Number(days));
      setMessage({ type: 'success', text: '数据清理已完成' });
    } catch (error) {
      console.error('数据清理失败:', error);
      setMessage({ type: 'error', text: '数据清理失败' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'config') {
      fetchConfigs();
    } else if (activeTab === 'history') {
      fetchHistory();
    }
  }, [activeTab]);

  // 清除消息
  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => setMessage(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [message]);

  // 标签页组件
  const TabButton: React.FC<{ 
    id: string; 
    label: string; 
    icon: string; 
    active: boolean; 
    onClick: () => void 
  }> = ({ id, label, icon, active, onClick }) => (
    <button
      onClick={onClick}
      className={`flex items-center space-x-2 px-4 py-2 rounded-lg font-medium transition-colors ${
        active 
          ? 'bg-blue-500 text-white' 
          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
      }`}
    >
      <span>{icon}</span>
      <span>{label}</span>
    </button>
  );

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            🛡️ 风控优化系统 v2
          </h1>
          <p className="text-gray-600">
            高级风险控制与智能决策支持系统
          </p>
        </div>

        {/* 消息提示 */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg ${
            message.type === 'success' 
              ? 'bg-green-50 border border-green-200 text-green-800' 
              : 'bg-red-50 border border-red-200 text-red-800'
          }`}>
            <div className="flex items-center space-x-2">
              <span>{message.type === 'success' ? '✅' : '❌'}</span>
              <span>{message.text}</span>
            </div>
          </div>
        )}

        {/* 标签页导航 */}
        <div className="flex space-x-4 mb-8">
          <TabButton
            id="dashboard"
            label="仪表板"
            icon="📊"
            active={activeTab === 'dashboard'}
            onClick={() => setActiveTab('dashboard')}
          />
          <TabButton
            id="config"
            label="配置管理"
            icon="⚙️"
            active={activeTab === 'config'}
            onClick={() => setActiveTab('config')}
          />
          <TabButton
            id="history"
            label="历史记录"
            icon="📋"
            active={activeTab === 'history'}
            onClick={() => setActiveTab('history')}
          />
          <TabButton
            id="operations"
            label="手动操作"
            icon="🔧"
            active={activeTab === 'operations'}
            onClick={() => setActiveTab('operations')}
          />
        </div>

        {/* 标签页内容 */}
        <div className="bg-white rounded-lg shadow-sm">
          {activeTab === 'dashboard' && (
            <div className="p-6">
              <RiskControlV2Dashboard />
            </div>
          )}

          {activeTab === 'config' && (
            <div className="p-6">
              <h2 className="text-xl font-semibold mb-4">系统配置</h2>
              {loading ? (
                <div className="flex items-center justify-center h-32">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                  <span className="ml-3">加载配置...</span>
                </div>
              ) : configs ? (
                <div className="space-y-4">
                  <div className="bg-gray-50 rounded-lg p-4">
                    <h3 className="font-medium mb-2">当前配置</h3>
                    <pre className="text-sm text-gray-700 overflow-auto">
                      {JSON.stringify(configs, null, 2)}
                    </pre>
                  </div>
                  <div className="text-sm text-gray-600">
                    💡 配置修改功能正在开发中，敬请期待
                  </div>
                </div>
              ) : (
                <div className="text-center text-gray-500 py-8">
                  暂无配置数据
                </div>
              )}
            </div>
          )}

          {activeTab === 'history' && (
            <div className="p-6">
              <h2 className="text-xl font-semibold mb-4">历史记录</h2>
              {loading ? (
                <div className="flex items-center justify-center h-32">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                  <span className="ml-3">加载历史数据...</span>
                </div>
              ) : (
                <div className="space-y-6">
                  {/* 警报历史 */}
                  <div>
                    <h3 className="text-lg font-medium mb-3">🚨 风险警报</h3>
                    {alerts.length > 0 ? (
                      <div className="space-y-2">
                        {alerts.slice(0, 10).map((alert, index) => (
                          <div key={index} className="bg-yellow-50 border border-yellow-200 rounded p-3">
                            <div className="flex justify-between items-start">
                              <div>
                                <div className="font-medium">{alert.type || '风险警报'}</div>
                                <div className="text-sm text-gray-600">{alert.message || '警报信息'}</div>
                              </div>
                              <div className="text-xs text-gray-500">
                                {alert.timestamp ? new Date(alert.timestamp).toLocaleString() : '时间未知'}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-gray-500 text-center py-4">暂无警报记录</div>
                    )}
                  </div>

                  {/* 决策历史 */}
                  <div>
                    <h3 className="text-lg font-medium mb-3">🤖 AI决策记录</h3>
                    {decisions.length > 0 ? (
                      <div className="space-y-2">
                        {decisions.slice(0, 10).map((decision, index) => (
                          <div key={index} className="bg-blue-50 border border-blue-200 rounded p-3">
                            <div className="flex justify-between items-start">
                              <div>
                                <div className="font-medium">{decision.type || 'AI决策'}</div>
                                <div className="text-sm text-gray-600">{decision.result || '决策结果'}</div>
                              </div>
                              <div className="text-xs text-gray-500">
                                {decision.timestamp ? new Date(decision.timestamp).toLocaleString() : '时间未知'}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-gray-500 text-center py-4">暂无决策记录</div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'operations' && (
            <div className="p-6">
              <h2 className="text-xl font-semibold mb-4">手动操作</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* 紧急操作 */}
                <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                  <h3 className="text-lg font-semibold text-red-800 mb-4">🛑 紧急操作</h3>
                  <div className="space-y-3">
                    <button
                      onClick={handleEmergencyStop}
                      disabled={loading}
                      className="w-full px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50 transition-colors"
                    >
                      {loading ? '执行中...' : '紧急停止'}
                    </button>
                    <button
                      onClick={handleResume}
                      disabled={loading}
                      className="w-full px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50 transition-colors"
                    >
                      {loading ? '执行中...' : '恢复运行'}
                    </button>
                  </div>
                  <p className="text-sm text-red-600 mt-3">
                    ⚠️ 紧急操作将立即生效，请谨慎使用
                  </p>
                </div>

                {/* 系统维护 */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
                  <h3 className="text-lg font-semibold text-blue-800 mb-4">🔧 系统维护</h3>
                  <div className="space-y-3">
                    <button
                      onClick={handleManualReset}
                      disabled={loading}
                      className="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 transition-colors"
                    >
                      {loading ? '执行中...' : '手动重置'}
                    </button>
                    <button
                      onClick={handleDataCleanup}
                      disabled={loading}
                      className="w-full px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600 disabled:opacity-50 transition-colors"
                    >
                      {loading ? '执行中...' : '数据清理'}
                    </button>
                  </div>
                  <p className="text-sm text-blue-600 mt-3">
                    💡 定期维护有助于系统稳定运行
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default RiskControlV2Page;