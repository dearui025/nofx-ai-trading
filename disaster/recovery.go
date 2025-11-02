// disaster/recovery.go
package disaster

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// SOSStatus SOS状态枚举
type SOSStatus int

const (
	SOSInactive SOSStatus = iota // SOS未激活
	SOSActive                    // SOS已激活
	SOSResolving                 // SOS解决中
	SOSResolved                  // SOS已解决
)

// String 返回SOS状态的字符串表示
func (s SOSStatus) String() string {
	switch s {
	case SOSInactive:
		return "Inactive"
	case SOSActive:
		return "Active"
	case SOSResolving:
		return "Resolving"
	case SOSResolved:
		return "Resolved"
	default:
		return "Unknown"
	}
}

// TriggerCondition 触发条件
type TriggerCondition struct {
	Type         string  `json:"type"`          // 条件类型
	Threshold    float64 `json:"threshold"`     // 阈值
	CurrentValue float64 `json:"current_value"` // 当前值
	Status       string  `json:"status"`        // 状态 (safe/warning/danger)
}

// SOSEvent SOS事件
type SOSEvent struct {
	ID              string             `json:"id"`
	TraderID        string             `json:"trader_id"`
	Reason          string             `json:"reason"`
	TriggerCondition string            `json:"trigger_condition"`
	Actions         []string           `json:"actions"`
	AccountEquity   float64            `json:"account_equity"`
	TotalPnL        float64            `json:"total_pnl"`
	MarginUsedPct   float64            `json:"margin_used_pct"`
	ActivatedAt     time.Time          `json:"activated_at"`
	ResolvedAt      *time.Time         `json:"resolved_at,omitempty"`
	Status          SOSStatus          `json:"status"`
}

// HedgeRecord 对冲记录
type HedgeRecord struct {
	ID               string    `json:"id"`
	TraderID         string    `json:"trader_id"`
	PrimaryExchange  string    `json:"primary_exchange"`
	HedgeExchange    string    `json:"hedge_exchange"`
	Symbol           string    `json:"symbol"`
	OriginalSide     string    `json:"original_side"`
	HedgeSide        string    `json:"hedge_side"`
	OriginalQuantity float64   `json:"original_quantity"`
	HedgeQuantity    float64   `json:"hedge_quantity"`
	HedgeRatio       float64   `json:"hedge_ratio"`
	ExecutedAt       time.Time `json:"executed_at"`
}

// DisasterRecoveryManager 灾难恢复管理器
type DisasterRecoveryManager struct {
	mu                sync.RWMutex
	sosEvents         map[string]*SOSEvent
	hedgeRecords      []HedgeRecord
	isSOSActive       bool
	currentSOSEvent   *SOSEvent
	
	// 配置参数
	MaxDrawdownPct    float64 // 最大回撤百分比
	MaxMarginUsedPct  float64 // 最大保证金使用百分比
	MinEquityThreshold float64 // 最小权益阈值
	
	// 回调函数
	OnSOSTriggered    func(*SOSEvent) error
	OnHedgeExecuted   func(*HedgeRecord) error
	OnPositionClosed  func(string, string) error // symbol, reason
}

// NewDisasterRecoveryManager 创建灾难恢复管理器
func NewDisasterRecoveryManager() *DisasterRecoveryManager {
	return &DisasterRecoveryManager{
		sosEvents:         make(map[string]*SOSEvent),
		hedgeRecords:      []HedgeRecord{},
		isSOSActive:       false,
		currentSOSEvent:   nil,
		MaxDrawdownPct:    0.15, // 15%最大回撤
		MaxMarginUsedPct:  95.0, // 95%最大保证金使用（百分比格式）
		MinEquityThreshold: 1000, // 最小1000 USDT权益
	}
}

// CheckSOSConditions 检查SOS触发条件
func (drm *DisasterRecoveryManager) CheckSOSConditions(
	accountEquity, totalPnL, marginUsedPct float64,
	traderID string,
) (*SOSEvent, error) {
	drm.mu.Lock()
	defer drm.mu.Unlock()

	// 如果SOS已经激活，不重复触发
	if drm.isSOSActive {
		return drm.currentSOSEvent, nil
	}

	// 检查各种触发条件
	conditions := drm.evaluateConditions(accountEquity, totalPnL, marginUsedPct)
	
	for _, condition := range conditions {
		if condition.Status == "danger" {
			// 触发SOS
			sosEvent := &SOSEvent{
				ID:               fmt.Sprintf("sos_%d", time.Now().Unix()),
				TraderID:         traderID,
				Reason:           fmt.Sprintf("%s超过危险阈值", condition.Type),
				TriggerCondition: fmt.Sprintf("%s: %.2f > %.2f", condition.Type, condition.CurrentValue, condition.Threshold),
				Actions:          drm.generateSOSActions(condition.Type),
				AccountEquity:    accountEquity,
				TotalPnL:         totalPnL,
				MarginUsedPct:    marginUsedPct,
				ActivatedAt:      time.Now(),
				Status:           SOSActive,
			}

			drm.sosEvents[sosEvent.ID] = sosEvent
			drm.currentSOSEvent = sosEvent
			drm.isSOSActive = true

			// 执行SOS回调
			if drm.OnSOSTriggered != nil {
				if err := drm.OnSOSTriggered(sosEvent); err != nil {
					log.Printf("SOS回调执行失败: %v", err)
				}
			}

			return sosEvent, nil
		}
	}

	return nil, nil
}

// evaluateConditions 评估触发条件
func (drm *DisasterRecoveryManager) evaluateConditions(
	accountEquity, totalPnL, marginUsedPct float64,
) []TriggerCondition {
	conditions := []TriggerCondition{
		{
			Type:         "drawdown",
			Threshold:    drm.MaxDrawdownPct,
			CurrentValue: math.Abs(totalPnL) / accountEquity,
			Status:       "safe",
		},
		{
			Type:         "margin",
			Threshold:    drm.MaxMarginUsedPct,
			CurrentValue: marginUsedPct,
			Status:       "safe",
		},
		{
			Type:         "equity",
			Threshold:    drm.MinEquityThreshold,
			CurrentValue: accountEquity,
			Status:       "safe",
		},
	}

	// 评估每个条件的状态
	for i := range conditions {
		condition := &conditions[i]
		
		switch condition.Type {
		case "drawdown":
			if condition.CurrentValue > condition.Threshold {
				condition.Status = "danger"
			} else if condition.CurrentValue > condition.Threshold*0.8 {
				condition.Status = "warning"
			}
		case "margin":
			if condition.CurrentValue > condition.Threshold {
				condition.Status = "danger"
			} else if condition.CurrentValue > condition.Threshold*0.9 {
				condition.Status = "warning"
			}
		case "equity":
			if condition.CurrentValue < condition.Threshold {
				condition.Status = "danger"
			} else if condition.CurrentValue < condition.Threshold*1.2 {
				condition.Status = "warning"
			}
		}
	}

	return conditions
}

// generateSOSActions 生成SOS行动计划
func (drm *DisasterRecoveryManager) generateSOSActions(triggerType string) []string {
	baseActions := []string{
		"立即停止新开仓",
		"评估当前持仓风险",
		"准备紧急平仓",
	}

	switch triggerType {
	case "drawdown":
		return append(baseActions, []string{
			"平仓亏损最大的持仓",
			"降低整体仓位规模",
			"启动跨市场对冲",
		}...)
	case "margin":
		return append(baseActions, []string{
			"立即平仓部分持仓释放保证金",
			"优先平仓保证金占用最高的持仓",
			"暂停所有高杠杆交易",
		}...)
	case "equity":
		return append(baseActions, []string{
			"全部平仓保护剩余资金",
			"暂停所有交易活动",
			"等待人工干预",
		}...)
	default:
		return baseActions
	}
}

// ExecuteEmergencyHedge 执行紧急对冲
func (drm *DisasterRecoveryManager) ExecuteEmergencyHedge(
	traderID, symbol, originalSide string,
	originalQuantity float64,
	hedgeExchange string,
) (*HedgeRecord, error) {
	drm.mu.Lock()
	defer drm.mu.Unlock()

	// 计算对冲参数
	hedgeSide := "sell"
	if originalSide == "short" {
		hedgeSide = "buy"
	}

	hedgeRatio := 0.8 // 80%对冲比例
	hedgeQuantity := originalQuantity * hedgeRatio

	hedgeRecord := &HedgeRecord{
		ID:               fmt.Sprintf("hedge_%d", time.Now().Unix()),
		TraderID:         traderID,
		PrimaryExchange:  "binance", // 假设主交易所是币安
		HedgeExchange:    hedgeExchange,
		Symbol:           symbol,
		OriginalSide:     originalSide,
		HedgeSide:        hedgeSide,
		OriginalQuantity: originalQuantity,
		HedgeQuantity:    hedgeQuantity,
		HedgeRatio:       hedgeRatio,
		ExecutedAt:       time.Now(),
	}

	drm.hedgeRecords = append(drm.hedgeRecords, *hedgeRecord)

	// 执行对冲回调
	if drm.OnHedgeExecuted != nil {
		if err := drm.OnHedgeExecuted(hedgeRecord); err != nil {
			return nil, fmt.Errorf("对冲执行回调失败: %w", err)
		}
	}

	log.Printf("紧急对冲已执行: %s %s %.4f -> %s %s %.4f",
		originalSide, symbol, originalQuantity,
		hedgeSide, symbol, hedgeQuantity)

	return hedgeRecord, nil
}

// ResolveSOSEvent 解决SOS事件
func (drm *DisasterRecoveryManager) ResolveSOSEvent(sosID string, resolution string) error {
	drm.mu.Lock()
	defer drm.mu.Unlock()

	sosEvent, exists := drm.sosEvents[sosID]
	if !exists {
		return fmt.Errorf("SOS事件不存在: %s", sosID)
	}

	now := time.Now()
	sosEvent.ResolvedAt = &now
	sosEvent.Status = SOSResolved
	sosEvent.Actions = append(sosEvent.Actions, fmt.Sprintf("解决方案: %s", resolution))

	// 如果这是当前活跃的SOS事件，则停用SOS
	if drm.currentSOSEvent != nil && drm.currentSOSEvent.ID == sosID {
		drm.isSOSActive = false
		drm.currentSOSEvent = nil
	}

	log.Printf("SOS事件已解决: %s - %s", sosID, resolution)
	return nil
}

// GetSOSStatus 获取SOS状态
func (drm *DisasterRecoveryManager) GetSOSStatus() map[string]interface{} {
	drm.mu.RLock()
	defer drm.mu.RUnlock()

	status := map[string]interface{}{
		"is_active":    drm.isSOSActive,
		"last_triggered": nil,
		"trigger_conditions": drm.getTriggerConditionsStatus(),
	}

	if drm.currentSOSEvent != nil {
		status["current_event"] = drm.currentSOSEvent
		status["last_triggered"] = drm.currentSOSEvent.ActivatedAt
	} else if len(drm.sosEvents) > 0 {
		// 找到最近的SOS事件
		var lastEvent *SOSEvent
		for _, event := range drm.sosEvents {
			if lastEvent == nil || event.ActivatedAt.After(lastEvent.ActivatedAt) {
				lastEvent = event
			}
		}
		if lastEvent != nil {
			status["last_triggered"] = lastEvent.ActivatedAt
		}
	}

	return status
}

// getTriggerConditionsStatus 获取触发条件状态
func (drm *DisasterRecoveryManager) getTriggerConditionsStatus() []TriggerCondition {
	return []TriggerCondition{
		{
			Type:      "drawdown",
			Threshold: drm.MaxDrawdownPct,
			Status:    "safe", // 这里应该根据实际情况更新
		},
		{
			Type:      "margin",
			Threshold: drm.MaxMarginUsedPct,
			Status:    "safe",
		},
		{
			Type:      "equity",
			Threshold: drm.MinEquityThreshold,
			Status:    "safe",
		},
	}
}

// GetHedgeRecords 获取对冲记录
func (drm *DisasterRecoveryManager) GetHedgeRecords(traderID string) []HedgeRecord {
	drm.mu.RLock()
	defer drm.mu.RUnlock()

	var records []HedgeRecord
	for _, record := range drm.hedgeRecords {
		if record.TraderID == traderID {
			records = append(records, record)
		}
	}

	return records
}

// GetSOSEvents 获取SOS事件列表
func (drm *DisasterRecoveryManager) GetSOSEvents(traderID string) []*SOSEvent {
	drm.mu.RLock()
	defer drm.mu.RUnlock()

	var events []*SOSEvent
	for _, event := range drm.sosEvents {
		if event.TraderID == traderID {
			events = append(events, event)
		}
	}

	return events
}

// TriggerManualSOS 手动触发SOS
func (drm *DisasterRecoveryManager) TriggerManualSOS(traderID, reason string) (*SOSEvent, error) {
	drm.mu.Lock()
	defer drm.mu.Unlock()

	if drm.isSOSActive {
		return drm.currentSOSEvent, fmt.Errorf("SOS已经激活，无法重复触发")
	}

	sosEvent := &SOSEvent{
		ID:               fmt.Sprintf("manual_sos_%d", time.Now().Unix()),
		TraderID:         traderID,
		Reason:           fmt.Sprintf("手动触发: %s", reason),
		TriggerCondition: "manual_trigger",
		Actions:          []string{"等待人工处理", "评估当前风险", "制定应对策略"},
		ActivatedAt:      time.Now(),
		Status:           SOSActive,
	}

	drm.sosEvents[sosEvent.ID] = sosEvent
	drm.currentSOSEvent = sosEvent
	drm.isSOSActive = true

	// 执行SOS回调
	if drm.OnSOSTriggered != nil {
		if err := drm.OnSOSTriggered(sosEvent); err != nil {
			log.Printf("手动SOS回调执行失败: %v", err)
		}
	}

	log.Printf("手动SOS已触发: %s - %s", traderID, reason)
	return sosEvent, nil
}

// SetThresholds 设置阈值
// maxDrawdown: 最大回撤百分比（小数形式，如0.15表示15%）
// maxMargin: 最大保证金使用率（百分比形式，如95.0表示95%）
// minEquity: 最小权益阈值（USDT）
func (drm *DisasterRecoveryManager) SetThresholds(maxDrawdown, maxMargin, minEquity float64) {
	drm.mu.Lock()
	defer drm.mu.Unlock()

	if maxDrawdown > 0 && maxDrawdown < 1 {
		drm.MaxDrawdownPct = maxDrawdown
	}
	if maxMargin > 0 && maxMargin <= 100 {
		drm.MaxMarginUsedPct = maxMargin
	}
	if minEquity > 0 {
		drm.MinEquityThreshold = minEquity
	}
}

// IsSOSActive 检查SOS是否激活
func (drm *DisasterRecoveryManager) IsSOSActive() bool {
	drm.mu.RLock()
	defer drm.mu.RUnlock()
	return drm.isSOSActive
}

// GetCurrentSOSEvent 获取当前SOS事件
func (drm *DisasterRecoveryManager) GetCurrentSOSEvent() *SOSEvent {
	drm.mu.RLock()
	defer drm.mu.RUnlock()
	return drm.currentSOSEvent
}