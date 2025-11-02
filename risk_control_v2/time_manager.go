package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// DrawdownResetCondition 回撤重置条件
type DrawdownResetCondition struct {
	Type      string  `json:"type"`      // "daily", "new_high", "manual"
	Threshold float64 `json:"threshold"` // 阈值（如新高确认缓冲）
	Enabled   bool    `json:"enabled"`   // 是否启用
}

// TimeManagerConfig 时间管理器配置
type TimeManagerConfig struct {
	DailyResetHour          int                      `json:"daily_reset_hour"`           // UTC每日重置小时
	Timezone                string                   `json:"timezone"`                   // 时区（固定为UTC）
	EquityBufferPercent     float64                  `json:"equity_buffer_percent"`      // 权益缓冲百分比
	DrawdownResetConditions []DrawdownResetCondition `json:"drawdown_reset_conditions"`  // 回撤重置条件
}

// TimeManagerState 时间管理器状态
type TimeManagerState struct {
	CurrentUTCTime       time.Time `json:"current_utc_time"`       // 当前UTC时间
	LastDailyReset       time.Time `json:"last_daily_reset"`       // 上次日重置时间
	EquityHighWatermark  float64   `json:"equity_high_watermark"`  // 权益最高水位
	LastWatermarkUpdate  time.Time `json:"last_watermark_update"`  // 上次水位更新时间
	ResetCount           int       `json:"reset_count"`            // 重置次数
	LastResetReason      string    `json:"last_reset_reason"`      // 上次重置原因
}

// DrawdownResetRecord 回撤重置记录
type DrawdownResetRecord struct {
	ID           string    `json:"id"`
	ResetType    string    `json:"reset_type"`    // "daily", "new_high", "manual"
	OldWatermark float64   `json:"old_watermark"` // 旧水位
	NewWatermark float64   `json:"new_watermark"` // 新水位
	Reason       string    `json:"reason"`        // 重置原因
	ResetAt      time.Time `json:"reset_at"`      // 重置时间
}

// TimeManager 时间管理器
type TimeManager struct {
	config      TimeManagerConfig
	state       TimeManagerState
	resetHistory []DrawdownResetRecord
	mutex       sync.RWMutex
	logger      *log.Logger
}

// NewTimeManager 创建时间管理器
func NewTimeManager(config TimeManagerConfig) *TimeManager {
	if config.Timezone == "" {
		config.Timezone = "UTC"
	}
	if config.DailyResetHour < 0 || config.DailyResetHour > 23 {
		config.DailyResetHour = 0 // 默认UTC 00:00
	}
	if config.EquityBufferPercent <= 0 {
		config.EquityBufferPercent = 0.001 // 默认0.1%
	}

	// 设置默认重置条件
	if len(config.DrawdownResetConditions) == 0 {
		config.DrawdownResetConditions = []DrawdownResetCondition{
			{Type: "daily", Threshold: 0, Enabled: true},
			{Type: "new_high", Threshold: 0.001, Enabled: true},
		}
	}

	now := time.Now().UTC()
	return &TimeManager{
		config: config,
		state: TimeManagerState{
			CurrentUTCTime:      now,
			LastDailyReset:      now.Truncate(24 * time.Hour), // 当天00:00
			EquityHighWatermark: 0,
			LastWatermarkUpdate: now,
			ResetCount:          0,
			LastResetReason:     "初始化",
		},
		resetHistory: make([]DrawdownResetRecord, 0),
		logger:       log.New(log.Writer(), "[TimeManager] ", log.LstdFlags),
	}
}

// UpdateCurrentTime 更新当前时间
func (tm *TimeManager) UpdateCurrentTime() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.state.CurrentUTCTime = time.Now().UTC()
}

// CheckDailyReset 检查是否需要进行日重置
func (tm *TimeManager) CheckDailyReset() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// 检查是否启用日重置
	dailyResetEnabled := false
	for _, condition := range tm.config.DrawdownResetConditions {
		if condition.Type == "daily" && condition.Enabled {
			dailyResetEnabled = true
			break
		}
	}

	if !dailyResetEnabled {
		return false
	}

	now := tm.state.CurrentUTCTime
	lastReset := tm.state.LastDailyReset
	
	// 计算今天的重置时间点
	todayReset := time.Date(now.Year(), now.Month(), now.Day(), 
		tm.config.DailyResetHour, 0, 0, 0, time.UTC)
	
	// 如果当前时间已过今天的重置点，且上次重置不是今天
	if now.After(todayReset) && lastReset.Before(todayReset) {
		return true
	}
	
	return false
}

// PerformDailyReset 执行日重置
func (tm *TimeManager) PerformDailyReset() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := tm.state.CurrentUTCTime
	oldWatermark := tm.state.EquityHighWatermark
	
	// 记录重置
	resetRecord := DrawdownResetRecord{
		ID:           fmt.Sprintf("daily_%d", now.Unix()),
		ResetType:    "daily",
		OldWatermark: oldWatermark,
		NewWatermark: oldWatermark, // 日重置不改变水位，只重置时间
		Reason:       fmt.Sprintf("UTC %02d:00 日重置", tm.config.DailyResetHour),
		ResetAt:      now,
	}
	
	tm.resetHistory = append(tm.resetHistory, resetRecord)
	tm.state.LastDailyReset = now
	tm.state.ResetCount++
	tm.state.LastResetReason = resetRecord.Reason
	
	tm.logger.Printf("执行日重置: %s", resetRecord.Reason)
	return nil
}

// UpdateEquityWatermark 更新权益高水位
func (tm *TimeManager) UpdateEquityWatermark(currentEquity float64) (bool, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if currentEquity <= 0 {
		return false, fmt.Errorf("权益值必须大于0: %.8f", currentEquity)
	}

	// 检查是否启用新高重置
	newHighResetEnabled := false
	var newHighThreshold float64
	for _, condition := range tm.config.DrawdownResetConditions {
		if condition.Type == "new_high" && condition.Enabled {
			newHighResetEnabled = true
			newHighThreshold = condition.Threshold
			break
		}
	}

	oldWatermark := tm.state.EquityHighWatermark
	
	// 如果是第一次设置或当前权益创新高
	if oldWatermark == 0 || currentEquity > oldWatermark*(1+newHighThreshold) {
		now := tm.state.CurrentUTCTime
		
		// 记录水位更新
		if newHighResetEnabled && oldWatermark > 0 {
			resetRecord := DrawdownResetRecord{
				ID:           fmt.Sprintf("new_high_%d", now.Unix()),
				ResetType:    "new_high",
				OldWatermark: oldWatermark,
				NewWatermark: currentEquity,
				Reason:       fmt.Sprintf("权益创新高: %.8f -> %.8f (涨幅: %.2f%%)", 
					oldWatermark, currentEquity, 
					(currentEquity-oldWatermark)/oldWatermark*100),
				ResetAt:      now,
			}
			tm.resetHistory = append(tm.resetHistory, resetRecord)
			tm.state.ResetCount++
			tm.state.LastResetReason = resetRecord.Reason
		}
		
		tm.state.EquityHighWatermark = currentEquity
		tm.state.LastWatermarkUpdate = now
		
		tm.logger.Printf("更新权益高水位: %.8f", currentEquity)
		return true, nil
	}
	
	return false, nil
}

// CalculateCurrentDrawdown 计算当前回撤
func (tm *TimeManager) CalculateCurrentDrawdown(currentEquity float64) float64 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.state.EquityHighWatermark <= 0 || currentEquity <= 0 {
		return 0.0
	}
	
	if currentEquity >= tm.state.EquityHighWatermark {
		return 0.0 // 无回撤
	}
	
	// 计算回撤百分比
	drawdown := (tm.state.EquityHighWatermark - currentEquity) / tm.state.EquityHighWatermark
	return drawdown
}

// ManualReset 手动重置回撤
func (tm *TimeManager) ManualReset(reason string, newWatermark float64) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if newWatermark <= 0 {
		return fmt.Errorf("新水位必须大于0: %.8f", newWatermark)
	}

	now := tm.state.CurrentUTCTime
	oldWatermark := tm.state.EquityHighWatermark
	
	resetRecord := DrawdownResetRecord{
		ID:           fmt.Sprintf("manual_%d", now.Unix()),
		ResetType:    "manual",
		OldWatermark: oldWatermark,
		NewWatermark: newWatermark,
		Reason:       reason,
		ResetAt:      now,
	}
	
	tm.resetHistory = append(tm.resetHistory, resetRecord)
	tm.state.EquityHighWatermark = newWatermark
	tm.state.LastWatermarkUpdate = now
	tm.state.ResetCount++
	tm.state.LastResetReason = reason
	
	tm.logger.Printf("手动重置回撤: %s, 新水位: %.8f", reason, newWatermark)
	return nil
}

// GetCurrentState 获取当前状态
func (tm *TimeManager) GetCurrentState() TimeManagerState {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	// 更新当前时间
	state := tm.state
	state.CurrentUTCTime = time.Now().UTC()
	return state
}

// GetResetHistory 获取重置历史（最近N条）
func (tm *TimeManager) GetResetHistory(limit int) []DrawdownResetRecord {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if limit <= 0 || limit > len(tm.resetHistory) {
		limit = len(tm.resetHistory)
	}
	
	// 返回最近的记录（倒序）
	start := len(tm.resetHistory) - limit
	result := make([]DrawdownResetRecord, limit)
	for i := 0; i < limit; i++ {
		result[i] = tm.resetHistory[start+limit-1-i]
	}
	
	return result
}

// GetConfig 获取配置
func (tm *TimeManager) GetConfig() TimeManagerConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.config
}

// UpdateConfig 更新配置
func (tm *TimeManager) UpdateConfig(newConfig TimeManagerConfig) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 验证配置
	if newConfig.DailyResetHour < 0 || newConfig.DailyResetHour > 23 {
		return fmt.Errorf("日重置小时必须在0-23之间: %d", newConfig.DailyResetHour)
	}
	
	if newConfig.EquityBufferPercent < 0 {
		return fmt.Errorf("权益缓冲百分比不能为负: %.6f", newConfig.EquityBufferPercent)
	}

	tm.config = newConfig
	tm.logger.Printf("配置已更新")
	return nil
}

// ToJSON 序列化为JSON
func (tm *TimeManager) ToJSON() ([]byte, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	data := map[string]interface{}{
		"config":        tm.config,
		"state":         tm.state,
		"reset_history": tm.resetHistory,
	}
	
	return json.MarshalIndent(data, "", "  ")
}

// IsTimeForDailyReset 检查是否到了日重置时间
func (tm *TimeManager) IsTimeForDailyReset() bool {
	return tm.CheckDailyReset()
}

// GetTimeSinceLastReset 获取距离上次重置的时间
func (tm *TimeManager) GetTimeSinceLastReset() time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	return tm.state.CurrentUTCTime.Sub(tm.state.LastDailyReset)
}

// GetDrawdownDuration 获取当前回撤持续时间
func (tm *TimeManager) GetDrawdownDuration() time.Duration {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	return tm.state.CurrentUTCTime.Sub(tm.state.LastWatermarkUpdate)
}