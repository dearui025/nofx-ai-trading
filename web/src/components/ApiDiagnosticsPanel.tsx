import React, { useState } from 'react';
import { apiDiagnostics, DiagnosticResult } from '../utils/apiDiagnostics';

export const ApiDiagnosticsPanel: React.FC = () => {
  const [isRunning, setIsRunning] = useState(false);
  const [results, setResults] = useState<DiagnosticResult[]>([]);
  const [showDetails, setShowDetails] = useState<string | null>(null);

  const runDiagnostics = async () => {
    setIsRunning(true);
    setResults([]);
    
    try {
      const diagnosticResults = await apiDiagnostics.runFullDiagnostics();
      setResults(diagnosticResults);
    } catch (error) {
      console.error('诊断运行失败:', error);
    } finally {
      setIsRunning(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success': return '✅';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      default: return '❓';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': return 'text-green-600';
      case 'error': return 'text-red-600';
      case 'warning': return 'text-yellow-600';
      default: return 'text-gray-600';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">API 连接诊断</h2>
        <button
          onClick={runDiagnostics}
          disabled={isRunning}
          className={`px-4 py-2 rounded-lg font-medium ${
            isRunning
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-blue-600 text-white hover:bg-blue-700'
          }`}
        >
          {isRunning ? '诊断中...' : '开始诊断'}
        </button>
      </div>

      {isRunning && (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-3 text-gray-600">正在运行诊断测试...</span>
        </div>
      )}

      {results.length > 0 && (
        <div className="space-y-4">
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="bg-green-50 p-4 rounded-lg">
              <div className="text-2xl font-bold text-green-600">
                {results.filter(r => r.status === 'success').length}
              </div>
              <div className="text-sm text-green-600">成功</div>
            </div>
            <div className="bg-red-50 p-4 rounded-lg">
              <div className="text-2xl font-bold text-red-600">
                {results.filter(r => r.status === 'error').length}
              </div>
              <div className="text-sm text-red-600">错误</div>
            </div>
            <div className="bg-yellow-50 p-4 rounded-lg">
              <div className="text-2xl font-bold text-yellow-600">
                {results.filter(r => r.status === 'warning').length}
              </div>
              <div className="text-sm text-yellow-600">警告</div>
            </div>
            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="text-2xl font-bold text-gray-600">
                {results.length}
              </div>
              <div className="text-sm text-gray-600">总计</div>
            </div>
          </div>

          <div className="space-y-2">
            {results.map((result, index) => (
              <div key={index} className="border rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <span className="text-xl">{getStatusIcon(result.status)}</span>
                    <div>
                      <div className="font-medium text-gray-900">{result.test}</div>
                      <div className={`text-sm ${getStatusColor(result.status)}`}>
                        {result.message}
                      </div>
                    </div>
                  </div>
                  {result.details && (
                    <button
                      onClick={() => setShowDetails(showDetails === `${index}` ? null : `${index}`)}
                      className="text-blue-600 hover:text-blue-800 text-sm"
                    >
                      {showDetails === `${index}` ? '隐藏详情' : '查看详情'}
                    </button>
                  )}
                </div>
                
                {showDetails === `${index}` && result.details && (
                  <div className="mt-3 p-3 bg-gray-50 rounded text-sm">
                    <pre className="whitespace-pre-wrap text-xs overflow-x-auto">
                      {JSON.stringify(result.details, null, 2)}
                    </pre>
                  </div>
                )}
                
                <div className="text-xs text-gray-500 mt-2">
                  {new Date(result.timestamp).toLocaleString()}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {!isRunning && results.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          点击"开始诊断"按钮来测试API连接状态
        </div>
      )}
    </div>
  );
};