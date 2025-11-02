package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// LiquidityLevel 流动性等级
type LiquidityLevel string

const (
	LiquidityHigh     LiquidityLevel = "high"     // 高流动性 (>50M)
	LiquidityMedium   LiquidityLevel = "medium"   // 中流动性 (15M-50M)
	LiquidityLow      LiquidityLevel = "low"      // 低流动性 (10M-15M)
	LiquidityCritical LiquidityLevel = "critical" // 危机流动性 (<10M)
)

// LiquidityAlert 流动性警报
type LiquidityAlert struct {
	ID           string         `json:"id"`
	Symbol       string         `json:"symbol"`
	AlertType    string         `json:"alert_type"`    // "threshold", "rapid_decline", "blacklist"
	Level        LiquidityLevel `json:"level"`
	OldValue     float64        `json:"old_value"`     // 旧流动性值
	NewValue     float64        `json:"new_value"`     // 新流动性值
	Threshold    float64        `json:"threshold"`     // 触发阈值
	Message      string         `json:"message"`       // 警报消息
	CreatedAt    time.Time      `json:"created_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
	IsResolved   bool           `json:"is_resolved"`
}

// LiquidityData 流动性数据
type LiquidityData struct {
	Symbol        string    `json:"symbol"`
	OpenInterest  float64   `json:"open_interest"`  // 持仓量（USD）
	Volume24h     float64   `json:"volume_24h"`     // 24小时成交量
	Level         LiquidityLevel `json:"level"`
	LastUpdated   time.Time `json:"last_updated"`
	ChangePercent float64   `json:"change_percent"` // 变化百分比
}

// BlacklistEntry 黑名单条目
type BlacklistEntry struct {
	Symbol      string    `json:"symbol"`
	Reason      string    `json:"reason"`
	AddedAt     time.Time `json:"added_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间，nil表示永久
	IsActive    bool      `json:"is_active"`
}

// LiquidityMonitorConfig 流动性监控配置
type LiquidityMonitorConfig struct {
	// 流动性阈值（USD）
	HighLiquidityThreshold     float64 `json:"high_liquidity_threshold"`     // 50M
	MediumLiquidityThreshold   float64 `json:"medium_liquidity_threshold"`   // 15M
	LowLiquidityThreshold      float64 `json:"low_liquidity_threshold"`      // 10M
	
	// 监控参数
	RapidDeclineThreshold      float64 `json:"rapid_decline_threshold"`      // 快速下降阈值 (20%)
	MonitoringIntervalMinutes  int     `json:"monitoring_interval_minutes"`  // 监控间隔
	AlertCooldownMinutes       int     `json:"alert_cooldown_minutes"`       // 警报冷却时间
	
	// 强制平仓参数
	ForceCloseEnabled          bool    `json:"force_close_enabled"`          // 是否启用强制平仓
	ForceCloseThreshold        float64 `json:"force_close_threshold"`        // 强制平仓阈值 (10M)
	
	// 黑名单参数
	BlacklistEnabled           bool    `json:"blacklist_enabled"`            // 是否启用黑名单
	BlacklistDurationHours     int     `json:"blacklist_duration_hours"`     // 黑名单持续时间
	AutoRemoveFromBlacklist    bool    `json:"auto_remove_from_blacklist"`   // 自动移除黑名单
}

// LiquidityMonitorState 流动性监控状态
type LiquidityMonitorState struct {
	LastUpdateTime    time.Time                    `json:"last_update_time"`
	TotalSymbols      int                          `json:"total_symbols"`
	AlertCount        int                          `json:"alert_count"`
	BlacklistCount    int                          `json:"blacklist_count"`
	LiquidityData     map[string]LiquidityData     `json:"liquidity_data"`
	ActiveAlerts      map[string]LiquidityAlert    `json:"active_alerts"`
	Blacklist         map[string]BlacklistEntry    `json:"blacklist"`
}

// LiquidityMonitor 流动性监控器
type LiquidityMonitor struct {
	config       LiquidityMonitorConfig
	state        LiquidityMonitorState
	alertHistory []LiquidityAlert
	mutex        sync.RWMutex
	logger       *log.Logger
}

// NewLiquidityMonitor 创建流动性监控器
func NewLiquidityMonitor(config LiquidityMonitorConfig) *LiquidityMonitor {
	// 设置默认值
	if config.HighLiquidityThreshold <= 0 {
		config.HighLiquidityThreshold = 50000000 // 50M USD
	}
	if config.MediumLiquidityThreshold <= 0 {
		config.MediumLiquidityThreshold = 15000000 // 15M USD
	}
	if config.LowLiquidityThreshold <= 0 {
		config.LowLiquidityThreshold = 10000000 // 10M USD
	}
	if config.RapidDeclineThreshold <= 0 {
		config.RapidDeclineThreshold = 0.2 // 20%
	}
	if config.MonitoringIntervalMinutes <= 0 {
		config.MonitoringIntervalMinutes = 5
	}
	if config.AlertCooldownMinutes <= 0 {
		config.AlertCooldownMinutes = 30
	}
	if config.ForceCloseThreshold <= 0 {
		config.ForceCloseThreshold = config.LowLiquidityThreshold
	}
	if config.BlacklistDurationHours <= 0 {
		config.BlacklistDurationHours = 24
	}

	return &LiquidityMonitor{
		config: config,
		state: LiquidityMonitorState{
			LastUpdateTime: time.Now().UTC(),
			LiquidityData:  make(map[string]LiquidityData),
			ActiveAlerts:   make(map[string]LiquidityAlert),
			Blacklist:      make(map[string]BlacklistEntry),
		},
		alertHistory: make([]LiquidityAlert, 0),
		logger:       log.New(log.Writer(), "[LiquidityMonitor] ", log.LstdFlags),
	}
}

// UpdateLiquidityData 更新流动性数据
func (lm *LiquidityMonitor) UpdateLiquidityData(symbol string, openInterest, volume24h float64) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	now := time.Now().UTC()
	
	// 获取旧数据
	oldData, exists := lm.state.LiquidityData[symbol]
	
	// 计算流动性等级
	level := lm.calculateLiquidityLevel(openInterest)
	
	// 计算变化百分比
	var changePercent float64
	if exists && oldData.OpenInterest > 0 {
		changePercent = (openInterest - oldData.OpenInterest) / oldData.OpenInterest
	}
	
	// 创建新数据
	newData := LiquidityData{
		Symbol:        symbol,
		OpenInterest:  openInterest,
		Volume24h:     volume24h,
		Level:         level,
		LastUpdated:   now,
		ChangePercent: changePercent,
	}
	
	// 检查是否需要触发警报
	if exists {
		lm.checkAndCreateAlerts(oldData, newData)
	}
	
	// 更新数据
	lm.state.LiquidityData[symbol] = newData
	lm.state.LastUpdateTime = now
	lm.state.TotalSymbols = len(lm.state.LiquidityData)
	
	return nil
}

// calculateLiquidityLevel 计算流动性等级
func (lm *LiquidityMonitor) calculateLiquidityLevel(openInterest float64) LiquidityLevel {
	if openInterest >= lm.config.HighLiquidityThreshold {
		return LiquidityHigh
	} else if openInterest >= lm.config.MediumLiquidityThreshold {
		return LiquidityMedium
	} else if openInterest >= lm.config.LowLiquidityThreshold {
		return LiquidityLow
	} else {
		return LiquidityCritical
	}
}

// checkAndCreateAlerts 检查并创建警报
func (lm *LiquidityMonitor) checkAndCreateAlerts(oldData, newData LiquidityData) {
	now := time.Now().UTC()
	
	// 检查阈值警报
	if oldData.Level != newData.Level {
		alert := LiquidityAlert{
			ID:        fmt.Sprintf("threshold_%s_%d", newData.Symbol, now.Unix()),
			Symbol:    newData.Symbol,
			AlertType: "threshold",
			Level:     newData.Level,
			OldValue:  oldData.OpenInterest,
			NewValue:  newData.OpenInterest,
			Message:   fmt.Sprintf("流动性等级变化: %s -> %s (%.2fM -> %.2fM)", 
				oldData.Level, newData.Level, 
				oldData.OpenInterest/1000000, newData.OpenInterest/1000000),
			CreatedAt: now,
		}
		lm.addAlert(alert)
	}
	
	// 检查快速下降警报
	if newData.ChangePercent < -lm.config.RapidDeclineThreshold {
		alert := LiquidityAlert{
			ID:        fmt.Sprintf("decline_%s_%d", newData.Symbol, now.Unix()),
			Symbol:    newData.Symbol,
			AlertType: "rapid_decline",
			Level:     newData.Level,
			OldValue:  oldData.OpenInterest,
			NewValue:  newData.OpenInterest,
			Threshold: lm.config.RapidDeclineThreshold,
			Message:   fmt.Sprintf("流动性快速下降: %.2f%% (%.2fM -> %.2fM)", 
				newData.ChangePercent*100,
				oldData.OpenInterest/1000000, newData.OpenInterest/1000000),
			CreatedAt: now,
		}
		lm.addAlert(alert)
	}
	
	// 检查是否需要强制平仓
	if lm.config.ForceCloseEnabled && newData.OpenInterest < lm.config.ForceCloseThreshold {
		lm.triggerForceClose(newData.Symbol, newData.OpenInterest)
	}
}

// addAlert 添加警报
func (lm *LiquidityMonitor) addAlert(alert LiquidityAlert) {
	// 检查冷却时间
	if lm.isInCooldown(alert.Symbol, alert.AlertType) {
		return
	}
	
	lm.state.ActiveAlerts[alert.ID] = alert
	lm.alertHistory = append(lm.alertHistory, alert)
	lm.state.AlertCount = len(lm.state.ActiveAlerts)
	
	lm.logger.Printf("新警报: %s - %s", alert.Symbol, alert.Message)
}

// isInCooldown 检查是否在冷却时间内
func (lm *LiquidityMonitor) isInCooldown(symbol, alertType string) bool {
	now := time.Now().UTC()
	cooldownDuration := time.Duration(lm.config.AlertCooldownMinutes) * time.Minute
	
	for _, alert := range lm.alertHistory {
		if alert.Symbol == symbol && alert.AlertType == alertType {
			if now.Sub(alert.CreatedAt) < cooldownDuration {
				return true
			}
		}
	}
	return false
}

// triggerForceClose 触发强制平仓
func (lm *LiquidityMonitor) triggerForceClose(symbol string, currentOI float64) {
	now := time.Now().UTC()
	
	alert := LiquidityAlert{
		ID:        fmt.Sprintf("force_close_%s_%d", symbol, now.Unix()),
		Symbol:    symbol,
		AlertType: "force_close",
		Level:     LiquidityCritical,
		NewValue:  currentOI,
		Threshold: lm.config.ForceCloseThreshold,
		Message:   fmt.Sprintf("触发强制平仓: 流动性过低 %.2fM < %.2fM", 
			currentOI/1000000, lm.config.ForceCloseThreshold/1000000),
		CreatedAt: now,
	}
	
	lm.addAlert(alert)
	
	// 添加到黑名单
	if lm.config.BlacklistEnabled {
		lm.addToBlacklist(symbol, "强制平仓后自动加入黑名单")
	}
	
	lm.logger.Printf("强制平仓触发: %s (流动性: %.2fM)", symbol, currentOI/1000000)
}

// addToBlacklist 添加到黑名单
func (lm *LiquidityMonitor) addToBlacklist(symbol, reason string) {
	now := time.Now().UTC()
	var expiresAt *time.Time
	
	if lm.config.AutoRemoveFromBlacklist {
		expiry := now.Add(time.Duration(lm.config.BlacklistDurationHours) * time.Hour)
		expiresAt = &expiry
	}
	
	entry := BlacklistEntry{
		Symbol:    symbol,
		Reason:    reason,
		AddedAt:   now,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}
	
	lm.state.Blacklist[symbol] = entry
	lm.state.BlacklistCount = len(lm.state.Blacklist)
	
	// 创建黑名单警报
	alert := LiquidityAlert{
		ID:        fmt.Sprintf("blacklist_%s_%d", symbol, now.Unix()),
		Symbol:    symbol,
		AlertType: "blacklist",
		Level:     LiquidityCritical,
		Message:   fmt.Sprintf("加入黑名单: %s", reason),
		CreatedAt: now,
	}
	lm.addAlert(alert)
	
	lm.logger.Printf("加入黑名单: %s - %s", symbol, reason)
}

// IsSymbolBlacklisted 检查符号是否在黑名单中
func (lm *LiquidityMonitor) IsSymbolBlacklisted(symbol string) bool {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	entry, exists := lm.state.Blacklist[symbol]
	if !exists || !entry.IsActive {
		return false
	}
	
	// 检查是否过期
	if entry.ExpiresAt != nil && time.Now().UTC().After(*entry.ExpiresAt) {
		// 异步移除过期条目
		go lm.removeFromBlacklist(symbol, "自动过期")
		return false
	}
	
	return true
}

// removeFromBlacklist 从黑名单移除
func (lm *LiquidityMonitor) removeFromBlacklist(symbol, reason string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	if entry, exists := lm.state.Blacklist[symbol]; exists {
		entry.IsActive = false
		lm.state.Blacklist[symbol] = entry
		
		// 重新计算活跃黑名单数量
		activeCount := 0
		for _, e := range lm.state.Blacklist {
			if e.IsActive {
				activeCount++
			}
		}
		lm.state.BlacklistCount = activeCount
		
		lm.logger.Printf("从黑名单移除: %s - %s", symbol, reason)
	}
}

// GetLiquidityLevel 获取符号的流动性等级
func (lm *LiquidityMonitor) GetLiquidityLevel(symbol string) (LiquidityLevel, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	data, exists := lm.state.LiquidityData[symbol]
	if !exists {
		return LiquidityCritical, false
	}
	
	return data.Level, true
}

// GetActiveAlerts 获取活跃警报
func (lm *LiquidityMonitor) GetActiveAlerts() []LiquidityAlert {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	alerts := make([]LiquidityAlert, 0, len(lm.state.ActiveAlerts))
	for _, alert := range lm.state.ActiveAlerts {
		if !alert.IsResolved {
			alerts = append(alerts, alert)
		}
	}
	
	// 按创建时间排序（最新的在前）
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})
	
	return alerts
}

// GetBlacklistedSymbols 获取黑名单符号
func (lm *LiquidityMonitor) GetBlacklistedSymbols() []string {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	symbols := make([]string, 0)
	for symbol, entry := range lm.state.Blacklist {
		if entry.IsActive {
			symbols = append(symbols, symbol)
		}
	}
	
	sort.Strings(symbols)
	return symbols
}

// ResolveAlert 解决警报
func (lm *LiquidityMonitor) ResolveAlert(alertID string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	alert, exists := lm.state.ActiveAlerts[alertID]
	if !exists {
		return fmt.Errorf("警报不存在: %s", alertID)
	}
	
	now := time.Now().UTC()
	alert.IsResolved = true
	alert.ResolvedAt = &now
	
	lm.state.ActiveAlerts[alertID] = alert
	
	// 重新计算活跃警报数量
	activeCount := 0
	for _, a := range lm.state.ActiveAlerts {
		if !a.IsResolved {
			activeCount++
		}
	}
	lm.state.AlertCount = activeCount
	
	lm.logger.Printf("警报已解决: %s", alertID)
	return nil
}

// GetCurrentState 获取当前状态
func (lm *LiquidityMonitor) GetCurrentState() LiquidityMonitorState {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	return lm.state
}

// GetConfig 获取配置
func (lm *LiquidityMonitor) GetConfig() LiquidityMonitorConfig {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	return lm.config
}

// UpdateConfig 更新配置
func (lm *LiquidityMonitor) UpdateConfig(newConfig LiquidityMonitorConfig) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 验证配置
	if newConfig.LowLiquidityThreshold >= newConfig.MediumLiquidityThreshold {
		return fmt.Errorf("低流动性阈值必须小于中流动性阈值")
	}
	if newConfig.MediumLiquidityThreshold >= newConfig.HighLiquidityThreshold {
		return fmt.Errorf("中流动性阈值必须小于高流动性阈值")
	}
	
	lm.config = newConfig
	lm.logger.Printf("配置已更新")
	return nil
}

// ToJSON 序列化为JSON
func (lm *LiquidityMonitor) ToJSON() ([]byte, error) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	data := map[string]interface{}{
		"config":        lm.config,
		"state":         lm.state,
		"alert_history": lm.alertHistory,
	}
	
	return json.MarshalIndent(data, "", "  ")
}