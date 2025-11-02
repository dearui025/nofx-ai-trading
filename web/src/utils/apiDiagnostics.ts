// API诊断工具
// 用于测试和诊断NOFX系统的API连接状态

import { SUPABASE_CONFIG, API_ENDPOINTS } from '../lib/supabaseApiClient';
import { supabase } from '../lib/supabase';

export interface DiagnosticResult {
  test: string;
  status: 'success' | 'error' | 'warning';
  message: string;
  details?: any;
  timestamp: string;
}

export class ApiDiagnostics {
  private results: DiagnosticResult[] = [];

  // 记录诊断结果
  private logResult(test: string, status: 'success' | 'error' | 'warning', message: string, details?: any) {
    const result: DiagnosticResult = {
      test,
      status,
      message,
      details,
      timestamp: new Date().toISOString(),
    };
    this.results.push(result);
    console.log(`[${status.toUpperCase()}] ${test}: ${message}`, details || '');
  }

  // 获取所有诊断结果
  getResults(): DiagnosticResult[] {
    return this.results;
  }

  // 清除诊断结果
  clearResults() {
    this.results = [];
  }

  // 测试Supabase连接
  async testSupabaseConnection(): Promise<boolean> {
    try {
      this.logResult('Supabase连接', 'success', `正在测试连接到: ${SUPABASE_CONFIG.url}`);
      
      // 测试基础连接
      const response = await fetch(`${SUPABASE_CONFIG.url}/rest/v1/`, {
        method: 'GET',
        headers: {
          'apikey': SUPABASE_CONFIG.anonKey,
          'Authorization': `Bearer ${SUPABASE_CONFIG.anonKey}`,
        },
      });

      if (response.ok) {
        this.logResult('Supabase连接', 'success', 'Supabase基础连接正常');
        return true;
      } else {
        this.logResult('Supabase连接', 'error', `连接失败: ${response.status} ${response.statusText}`);
        return false;
      }
    } catch (error) {
      this.logResult('Supabase连接', 'error', '连接异常', error);
      return false;
    }
  }

  // 检查数据库表数据
  async checkDatabaseTables(): Promise<void> {
    const tables = ['account_history', 'positions', 'decisions', 'agent_decisions', 'trades'];
    
    for (const table of tables) {
      try {
        const { data, error, count } = await supabase
          .from(table)
          .select('*', { count: 'exact' })
          .limit(5);

        if (error) {
          this.logResult(`表 ${table}`, 'error', `查询失败: ${error.message}`, error);
        } else {
          this.logResult(`表 ${table}`, 'success', `找到 ${count} 条记录`, {
            count,
            sampleData: data?.slice(0, 2) // 只显示前2条记录作为样本
          });
        }
      } catch (error) {
        this.logResult(`表 ${table}`, 'error', '查询异常', error);
      }
    }
  }

  // 检查特定用户的数据
  async checkUserData(userUuid: string): Promise<void> {
    this.logResult('用户数据检查', 'success', `检查用户 ${userUuid} 的数据`);

    // 检查账户历史
    try {
      const { data: accountData, error: accountError } = await supabase
        .from('account_history')
        .select('*')
        .eq('user_id', userUuid)
        .order('timestamp', { ascending: false })
        .limit(5);

      if (accountError) {
        this.logResult('用户账户历史', 'error', `查询失败: ${accountError.message}`, accountError);
      } else {
        this.logResult('用户账户历史', 'success', `找到 ${accountData?.length || 0} 条记录`, accountData);
      }
    } catch (error) {
      this.logResult('用户账户历史', 'error', '查询异常', error);
    }

    // 检查持仓
    try {
      const { data: positionsData, error: positionsError } = await supabase
        .from('positions')
        .select('*')
        .eq('user_id', userUuid);

      if (positionsError) {
        this.logResult('用户持仓', 'error', `查询失败: ${positionsError.message}`, positionsError);
      } else {
        this.logResult('用户持仓', 'success', `找到 ${positionsData?.length || 0} 条记录`, positionsData);
      }
    } catch (error) {
      this.logResult('用户持仓', 'error', '查询异常', error);
    }

    // 检查决策记录
    try {
      const { data: decisionsData, error: decisionsError } = await supabase
        .from('decisions')
        .select('*')
        .eq('trader_id', 'qwen-trader-001') // 使用trader_id而不是user_id
        .order('created_at', { ascending: false })
        .limit(5);

      if (decisionsError) {
        this.logResult('用户决策', 'error', `查询失败: ${decisionsError.message}`, decisionsError);
      } else {
        this.logResult('用户决策', 'success', `找到 ${decisionsData?.length || 0} 条记录`, decisionsData);
      }
    } catch (error) {
      this.logResult('用户决策', 'error', '查询异常', error);
    }
  }

  // 测试前端API调用
  async testFrontendApiCalls(): Promise<void> {
    const { api } = await import('../lib/api');

    // 测试获取账户信息
    try {
      const accountInfo = await api.getAccount('qwen-trader-001');
      this.logResult('前端API-账户信息', 'success', '获取账户信息成功', accountInfo);
    } catch (error) {
      this.logResult('前端API-账户信息', 'error', '获取账户信息失败', error);
    }

    // 测试获取持仓
    try {
      const positions = await api.getPositions('qwen-trader-001');
      this.logResult('前端API-持仓', 'success', `获取持仓成功，共 ${positions.length} 个`, positions);
    } catch (error) {
      this.logResult('前端API-持仓', 'error', '获取持仓失败', error);
    }

    // 测试获取决策
    try {
      const decisions = await api.getDecisions('qwen-trader-001');
      this.logResult('前端API-决策', 'success', `获取决策成功，共 ${decisions.length} 个`, decisions);
    } catch (error) {
      this.logResult('前端API-决策', 'error', '获取决策失败', error);
    }

    // 测试获取账户历史
    try {
      const equityHistory = await api.getEquityHistory('qwen-trader-001');
      this.logResult('前端API-账户历史', 'success', `获取账户历史成功，共 ${equityHistory.length} 个数据点`, equityHistory);
    } catch (error) {
      this.logResult('前端API-账户历史', 'error', '获取账户历史失败', error);
    }
  }

  // 运行完整诊断
  async runFullDiagnostics(): Promise<DiagnosticResult[]> {
    this.clearResults();
    
    this.logResult('诊断开始', 'success', '开始运行完整的API诊断');

    // 1. 测试Supabase连接
    await this.testSupabaseConnection();

    // 2. 检查数据库表
    await this.checkDatabaseTables();

    // 3. 检查特定用户数据
    const userUuid = '550e8400-e29b-41d4-a716-446655440001'; // qwen-trader-001的UUID
    await this.checkUserData(userUuid);

    // 4. 测试前端API调用
    await this.testFrontendApiCalls();

    this.logResult('诊断完成', 'success', '完整诊断已完成');

    return this.getResults();
  }
}

// 导出单例
export const apiDiagnostics = new ApiDiagnostics();

// 便捷函数
export async function runDiagnostics(): Promise<DiagnosticResult[]> {
  return await apiDiagnostics.runFullDiagnostics();
}