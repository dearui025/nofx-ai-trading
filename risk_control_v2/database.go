package risk_control_v2

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// DatabaseManager 内存数据库管理器
type DatabaseManager struct {
	mu     sync.RWMutex
	logger *log.Logger
	
	// 内存存储
	timeManagerStates      []TimeManagerState
	drawdownResetRecords   []DrawdownResetRecord
	liquidityData          map[string]LiquidityData
	liquidityAlerts        []LiquidityAlert
	blacklistEntries       []BlacklistEntry
	sharpeRecords          []SharpeRecord
	sharpeStateTransitions []StateTransition
	aiCommitteeDecisions   []CommitteeDecision
	riskAlerts             []RiskAlert
	riskDecisions          []RiskDecision
	systemConfigs          map[string]interface{}
}

// NewDatabaseManager 创建内存数据库管理器
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	dm := &DatabaseManager{
		logger:                 log.New(log.Writer(), "[DatabaseManager] ", log.LstdFlags),
		liquidityData:          make(map[string]LiquidityData),
		systemConfigs:          make(map[string]interface{}),
		timeManagerStates:      make([]TimeManagerState, 0),
		drawdownResetRecords:   make([]DrawdownResetRecord, 0),
		liquidityAlerts:        make([]LiquidityAlert, 0),
		blacklistEntries:       make([]BlacklistEntry, 0),
		sharpeRecords:          make([]SharpeRecord, 0),
		sharpeStateTransitions: make([]StateTransition, 0),
		aiCommitteeDecisions:   make([]CommitteeDecision, 0),
		riskAlerts:             make([]RiskAlert, 0),
		riskDecisions:          make([]RiskDecision, 0),
	}

	dm.logger.Printf("内存数据库管理器初始化成功")
	return dm, nil
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.logger.Printf("内存数据库管理器已关闭")
	return nil
}

// SaveTimeManagerState 保存时间管理器状态
func (dm *DatabaseManager) SaveTimeManagerState(state TimeManagerState) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.timeManagerStates = append(dm.timeManagerStates, state)
	return nil
}

// GetTimeManagerState 获取最新的时间管理器状态
func (dm *DatabaseManager) GetTimeManagerState() (*TimeManagerState, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	if len(dm.timeManagerStates) == 0 {
		return nil, fmt.Errorf("未找到时间管理器状态")
	}
	
	latest := dm.timeManagerStates[len(dm.timeManagerStates)-1]
	return &latest, nil
}

// SaveDrawdownResetRecord 保存回撤重置记录
func (dm *DatabaseManager) SaveDrawdownResetRecord(record DrawdownResetRecord) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.drawdownResetRecords = append(dm.drawdownResetRecords, record)
	return nil
}

// GetDrawdownResetHistory 获取回撤重置历史
func (dm *DatabaseManager) GetDrawdownResetHistory(limit int) ([]DrawdownResetRecord, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	records := dm.drawdownResetRecords
	if limit > 0 && len(records) > limit {
		records = records[len(records)-limit:]
	}
	
	return records, nil
}

// SaveLiquidityData 保存流动性数据
func (dm *DatabaseManager) SaveLiquidityData(symbol string, data LiquidityData) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.liquidityData[symbol] = data
	return nil
}

// GetLiquidityData 获取流动性数据
func (dm *DatabaseManager) GetLiquidityData(symbol string) (*LiquidityData, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	if data, exists := dm.liquidityData[symbol]; exists {
		return &data, nil
	}
	
	return nil, fmt.Errorf("未找到符号 %s 的流动性数据", symbol)
}

// GetAllLiquidityData 获取所有流动性数据
func (dm *DatabaseManager) GetAllLiquidityData() (map[string]LiquidityData, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	result := make(map[string]LiquidityData)
	for k, v := range dm.liquidityData {
		result[k] = v
	}
	
	return result, nil
}

// SaveLiquidityAlert 保存流动性警报
func (dm *DatabaseManager) SaveLiquidityAlert(alert LiquidityAlert) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.liquidityAlerts = append(dm.liquidityAlerts, alert)
	return nil
}

// GetActiveLiquidityAlerts 获取活跃的流动性警报
func (dm *DatabaseManager) GetActiveLiquidityAlerts() ([]LiquidityAlert, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	var activeAlerts []LiquidityAlert
	for _, alert := range dm.liquidityAlerts {
		if !alert.IsResolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	
	return activeAlerts, nil
}

// SaveBlacklistEntry 保存黑名单条目
func (dm *DatabaseManager) SaveBlacklistEntry(entry BlacklistEntry) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.blacklistEntries = append(dm.blacklistEntries, entry)
	return nil
}

// GetBlacklistEntries 获取活跃的黑名单条目
func (dm *DatabaseManager) GetBlacklistEntries() ([]BlacklistEntry, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	var activeEntries []BlacklistEntry
	for _, entry := range dm.blacklistEntries {
		if entry.IsActive {
			activeEntries = append(activeEntries, entry)
		}
	}
	
	return activeEntries, nil
}

// SaveSharpeRecord 保存夏普比率记录
func (dm *DatabaseManager) SaveSharpeRecord(record SharpeRecord) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.sharpeRecords = append(dm.sharpeRecords, record)
	return nil
}

// GetSharpeRecords 获取夏普比率记录
func (dm *DatabaseManager) GetSharpeRecords(limit int) ([]SharpeRecord, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	records := dm.sharpeRecords
	if limit > 0 && len(records) > limit {
		records = records[len(records)-limit:]
	}
	
	return records, nil
}

// SaveSharpeStateTransition 保存夏普比率状态转换
func (dm *DatabaseManager) SaveSharpeStateTransition(transition StateTransition) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.sharpeStateTransitions = append(dm.sharpeStateTransitions, transition)
	return nil
}

// GetSharpeStateTransitions 获取夏普比率状态转换历史
func (dm *DatabaseManager) GetSharpeStateTransitions(limit int) ([]StateTransition, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	transitions := dm.sharpeStateTransitions
	if limit > 0 && len(transitions) > limit {
		transitions = transitions[len(transitions)-limit:]
	}
	
	return transitions, nil
}

// SaveAICommitteeDecision 保存AI委员会决策
func (dm *DatabaseManager) SaveAICommitteeDecision(decision CommitteeDecision) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.aiCommitteeDecisions = append(dm.aiCommitteeDecisions, decision)
	return nil
}

// GetAICommitteeDecisions 获取AI委员会决策历史
func (dm *DatabaseManager) GetAICommitteeDecisions(limit int) ([]CommitteeDecision, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	decisions := dm.aiCommitteeDecisions
	if limit > 0 && len(decisions) > limit {
		decisions = decisions[len(decisions)-limit:]
	}
	
	return decisions, nil
}

// SaveRiskAlert 保存风控警报
func (dm *DatabaseManager) SaveRiskAlert(alert RiskAlert) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.riskAlerts = append(dm.riskAlerts, alert)
	return nil
}

// GetActiveRiskAlerts 获取活跃的风控警报
func (dm *DatabaseManager) GetActiveRiskAlerts() ([]RiskAlert, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	var activeAlerts []RiskAlert
	for _, alert := range dm.riskAlerts {
		if !alert.IsResolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	
	return activeAlerts, nil
}

// SaveRiskDecision 保存风控决策
func (dm *DatabaseManager) SaveRiskDecision(decision RiskDecision) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	dm.riskDecisions = append(dm.riskDecisions, decision)
	return nil
}

// GetRiskDecisions 获取风控决策历史
func (dm *DatabaseManager) GetRiskDecisions(limit int) ([]RiskDecision, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	decisions := dm.riskDecisions
	if limit > 0 && len(decisions) > limit {
		decisions = decisions[len(decisions)-limit:]
	}
	
	return decisions, nil
}

// SaveSystemConfig 保存系统配置
func (dm *DatabaseManager) SaveSystemConfig(configType, configName string, configValue interface{}) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	key := fmt.Sprintf("%s.%s", configType, configName)
	dm.systemConfigs[key] = configValue
	return nil
}

// GetSystemConfig 获取系统配置
func (dm *DatabaseManager) GetSystemConfig(configType, configName string) (interface{}, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	key := fmt.Sprintf("%s.%s", configType, configName)
	if value, exists := dm.systemConfigs[key]; exists {
		return value, nil
	}
	
	return nil, fmt.Errorf("未找到配置 %s", key)
}

// CleanOldRecords 清理旧记录
func (dm *DatabaseManager) CleanOldRecords(days int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	cutoff := time.Now().AddDate(0, 0, -days)
	
	// 清理旧的夏普比率记录
	var newSharpeRecords []SharpeRecord
	for _, record := range dm.sharpeRecords {
		if record.Timestamp.After(cutoff) {
			newSharpeRecords = append(newSharpeRecords, record)
		}
	}
	dm.sharpeRecords = newSharpeRecords
	
	// 清理旧的AI决策记录
	var newAIDecisions []CommitteeDecision
	for _, decision := range dm.aiCommitteeDecisions {
		if decision.Timestamp.After(cutoff) {
			newAIDecisions = append(newAIDecisions, decision)
		}
	}
	dm.aiCommitteeDecisions = newAIDecisions
	
	// 清理旧的风控决策记录
	var newRiskDecisions []RiskDecision
	for _, decision := range dm.riskDecisions {
		if decision.Timestamp.After(cutoff) {
			newRiskDecisions = append(newRiskDecisions, decision)
		}
	}
	dm.riskDecisions = newRiskDecisions
	
	dm.logger.Printf("已清理 %d 天前的旧记录", days)
	return nil
}

// GetDatabaseStats 获取数据库统计信息
func (dm *DatabaseManager) GetDatabaseStats() (map[string]interface{}, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	stats := map[string]interface{}{
		"time_manager_states":        len(dm.timeManagerStates),
		"drawdown_reset_records":     len(dm.drawdownResetRecords),
		"liquidity_data_entries":     len(dm.liquidityData),
		"liquidity_alerts":           len(dm.liquidityAlerts),
		"blacklist_entries":          len(dm.blacklistEntries),
		"sharpe_records":             len(dm.sharpeRecords),
		"sharpe_state_transitions":   len(dm.sharpeStateTransitions),
		"ai_committee_decisions":     len(dm.aiCommitteeDecisions),
		"risk_alerts":                len(dm.riskAlerts),
		"risk_decisions":             len(dm.riskDecisions),
		"system_configs":             len(dm.systemConfigs),
	}
	
	return stats, nil
}

// ExportData 导出数据
func (dm *DatabaseManager) ExportData() (map[string]interface{}, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	data := map[string]interface{}{
		"time_manager_states":        dm.timeManagerStates,
		"drawdown_reset_records":     dm.drawdownResetRecords,
		"liquidity_data":             dm.liquidityData,
		"liquidity_alerts":           dm.liquidityAlerts,
		"blacklist_entries":          dm.blacklistEntries,
		"sharpe_records":             dm.sharpeRecords,
		"sharpe_state_transitions":   dm.sharpeStateTransitions,
		"ai_committee_decisions":     dm.aiCommitteeDecisions,
		"risk_alerts":                dm.riskAlerts,
		"risk_decisions":             dm.riskDecisions,
		"system_configs":             dm.systemConfigs,
		"export_timestamp":           time.Now(),
	}
	
	return data, nil
}