package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// RiskManagerConfig 风控管理器配置
type RiskManagerConfig struct {
	// 时间管理配置
	TimeManager TimeManagerConfig `json:"time_manager"`
	
	// 流动性监控配置
	LiquidityMonitor LiquidityMonitorConfig `json:"liquidity_monitor"`
	
	// 夏普比率计算配置
	SharpeCalculator SharpeCalculatorConfig `json:"sharpe_calculator"`
	
	// AI委员会配置
	AICommittee AICommitteeConfig `json:"ai_committee"`
	
	// 全局风控参数
	GlobalRiskEnabled    bool    `json:"global_risk_enabled"`    // 是否启用全局风控
	EmergencyStopEnabled bool    `json:"emergency_stop_enabled"` // 是否启用紧急停止
	MaxDrawdownPercent   float64 `json:"max_drawdown_percent"`   // 最大回撤百分比
	MaxDailyLossPercent  float64 `json:"max_daily_loss_percent"` // 最大日损失百分比
	
	// 监控间隔
	MonitoringIntervalSeconds int `json:"monitoring_interval_seconds"` // 监控间隔（秒）
}

// RiskManagerState 风控管理器状态
type RiskManagerState struct {
	IsActive           bool      `json:"is_active"`
	LastUpdateTime     time.Time `json:"last_update_time"`
	EmergencyStop      bool      `json:"emergency_stop"`
	GlobalRiskLevel    string    `json:"global_risk_level"`    // "low", "medium", "high", "critical"
	
	// 各模块状态
	TimeManagerState     TimeManagerState     `json:"time_manager_state"`
	LiquidityState       LiquidityMonitorState `json:"liquidity_state"`
	SharpeState          SharpeCalculatorState `json:"sharpe_state"`
	AICommitteeState     AICommitteeState     `json:"ai_committee_state"`
	
	// 统计信息
	TotalAlerts        int       `json:"total_alerts"`
	CriticalAlerts     int       `json:"critical_alerts"`
	LastAlertTime      time.Time `json:"last_alert_time"`
	SystemUptime       time.Duration `json:"system_uptime"`
	StartTime          time.Time `json:"start_time"`
}

// RiskAlert 风控警报
type RiskAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // "time", "liquidity", "sharpe", "ai", "global"
	Level       string    `json:"level"`       // "info", "warning", "critical"
	Message     string    `json:"message"`
	Details     string    `json:"details"`
	Source      string    `json:"source"`      // 来源模块
	Timestamp   time.Time `json:"timestamp"`
	IsResolved  bool      `json:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// RiskDecision 风控决策
type RiskDecision struct {
	ID                string                 `json:"id"`
	Symbol            string                 `json:"symbol"`
	Action            string                 `json:"action"`            // "allow", "block", "force_close", "reduce_size"
	Reason            string                 `json:"reason"`
	RiskFactors       []string               `json:"risk_factors"`
	Confidence        float64                `json:"confidence"`
	
	// 各模块输入
	TimeCheck         bool                   `json:"time_check"`         // 时间检查通过
	LiquidityCheck    bool                   `json:"liquidity_check"`    // 流动性检查通过
	SharpeCheck       bool                   `json:"sharpe_check"`       // 夏普比率检查通过
	AIDecision        *CommitteeDecision     `json:"ai_decision"`        // AI委员会决策
	
	Timestamp         time.Time              `json:"timestamp"`
	ProcessingTime    time.Duration          `json:"processing_time"`
}

// RiskManager 风控管理器
type RiskManager struct {
	config       RiskManagerConfig
	state        RiskManagerState
	
	// 各模块实例
	timeManager      *TimeManager
	liquidityMonitor *LiquidityMonitor
	sharpeCalculator *SharpeCalculator
	aiCommittee      *AICommittee
	
	// 警报和决策历史
	alerts           []RiskAlert
	decisions        []RiskDecision
	
	// 控制
	stopChan         chan struct{}
	isRunning        bool
	mutex            sync.RWMutex
	logger           *log.Logger
}

// NewRiskManager 创建风控管理器
func NewRiskManager(config RiskManagerConfig) *RiskManager {
	// 设置默认值
	if config.MonitoringIntervalSeconds <= 0 {
		config.MonitoringIntervalSeconds = 60 // 默认1分钟
	}
	if config.MaxDrawdownPercent <= 0 {
		config.MaxDrawdownPercent = 0.2 // 默认20%
	}
	if config.MaxDailyLossPercent <= 0 {
		config.MaxDailyLossPercent = 0.1 // 默认10%
	}

	now := time.Now().UTC()
	
	rm := &RiskManager{
		config: config,
		state: RiskManagerState{
			IsActive:        false,
			LastUpdateTime:  now,
			EmergencyStop:   false,
			GlobalRiskLevel: "low",
			StartTime:       now,
		},
		alerts:    make([]RiskAlert, 0),
		decisions: make([]RiskDecision, 0),
		stopChan:  make(chan struct{}),
		logger:    log.New(log.Writer(), "[RiskManager] ", log.LstdFlags),
	}

	// 初始化各模块
	rm.timeManager = NewTimeManager(config.TimeManager)
	rm.liquidityMonitor = NewLiquidityMonitor(config.LiquidityMonitor)
	rm.sharpeCalculator = NewSharpeCalculator(config.SharpeCalculator)
	rm.aiCommittee = NewAICommittee(config.AICommittee)

	return rm
}

// Start 启动风控管理器
func (rm *RiskManager) Start() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.isRunning {
		return fmt.Errorf("风控管理器已在运行")
	}

	rm.isRunning = true
	rm.state.IsActive = true
	rm.state.StartTime = time.Now().UTC()
	
	// 启动监控协程
	go rm.monitoringLoop()
	
	rm.logger.Printf("风控管理器已启动")
	return nil
}

// Stop 停止风控管理器
func (rm *RiskManager) Stop() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if !rm.isRunning {
		return fmt.Errorf("风控管理器未在运行")
	}

	close(rm.stopChan)
	rm.isRunning = false
	rm.state.IsActive = false
	
	rm.logger.Printf("风控管理器已停止")
	return nil
}

// monitoringLoop 监控循环
func (rm *RiskManager) monitoringLoop() {
	ticker := time.NewTicker(time.Duration(rm.config.MonitoringIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.performMonitoring()
		}
	}
}

// performMonitoring 执行监控
func (rm *RiskManager) performMonitoring() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	now := time.Now().UTC()
	
	// 更新时间管理器
	rm.timeManager.UpdateCurrentTime()
	
	// 检查日重置
	if rm.timeManager.CheckDailyReset() {
		err := rm.timeManager.PerformDailyReset()
		if err != nil {
			rm.addAlert("time", "warning", "日重置失败", err.Error(), "time_manager")
		} else {
			rm.addAlert("time", "info", "执行日重置", "成功执行UTC日重置", "time_manager")
		}
	}
	
	// 更新状态
	rm.updateStates()
	
	// 评估全局风险
	rm.assessGlobalRisk()
	
	rm.state.LastUpdateTime = now
	rm.state.SystemUptime = now.Sub(rm.state.StartTime)
}

// updateStates 更新各模块状态
func (rm *RiskManager) updateStates() {
	rm.state.TimeManagerState = rm.timeManager.GetCurrentState()
	rm.state.LiquidityState = rm.liquidityMonitor.GetCurrentState()
	rm.state.SharpeState = rm.sharpeCalculator.GetCurrentState()
	rm.state.AICommitteeState = rm.aiCommittee.GetCurrentState()
}

// assessGlobalRisk 评估全局风险
func (rm *RiskManager) assessGlobalRisk() {
	riskFactors := make([]string, 0)
	riskScore := 0.0

	// 检查回撤风险
	if rm.state.TimeManagerState.EquityHighWatermark > 0 {
		// 这里需要当前权益值，暂时使用模拟值
		currentEquity := rm.state.TimeManagerState.EquityHighWatermark * 0.95 // 模拟5%回撤
		drawdown := rm.timeManager.CalculateCurrentDrawdown(currentEquity)
		
		if drawdown > rm.config.MaxDrawdownPercent {
			riskFactors = append(riskFactors, fmt.Sprintf("回撤超限: %.2f%% > %.2f%%", 
				drawdown*100, rm.config.MaxDrawdownPercent*100))
			riskScore += 0.4
		} else if drawdown > rm.config.MaxDrawdownPercent*0.8 {
			riskFactors = append(riskFactors, fmt.Sprintf("回撤接近限制: %.2f%%", drawdown*100))
			riskScore += 0.2
		}
	}

	// 检查流动性风险
	blacklistCount := rm.state.LiquidityState.BlacklistCount
	if blacklistCount > 5 {
		riskFactors = append(riskFactors, fmt.Sprintf("黑名单币种过多: %d", blacklistCount))
		riskScore += 0.2
	}

	activeAlerts := len(rm.liquidityMonitor.GetActiveAlerts())
	if activeAlerts > 10 {
		riskFactors = append(riskFactors, fmt.Sprintf("流动性警报过多: %d", activeAlerts))
		riskScore += 0.1
	}

	// 检查夏普比率风险
	if rm.state.SharpeState.CurrentState == SharpeVeryPoor {
		riskFactors = append(riskFactors, "夏普比率极差")
		riskScore += 0.3
	} else if rm.state.SharpeState.CurrentState == SharpePoor {
		riskFactors = append(riskFactors, "夏普比率较差")
		riskScore += 0.1
	}

	// 检查AI委员会风险
	if rm.state.AICommitteeState.AvgConsensusLevel < 0.5 {
		riskFactors = append(riskFactors, "AI模型共识度低")
		riskScore += 0.1
	}

	// 确定风险等级
	var riskLevel string
	if riskScore >= 0.8 {
		riskLevel = "critical"
	} else if riskScore >= 0.5 {
		riskLevel = "high"
	} else if riskScore >= 0.2 {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// 更新风险等级
	oldLevel := rm.state.GlobalRiskLevel
	rm.state.GlobalRiskLevel = riskLevel

	// 如果风险等级变化，发出警报
	if oldLevel != riskLevel {
		message := fmt.Sprintf("全局风险等级变化: %s -> %s", oldLevel, riskLevel)
		details := fmt.Sprintf("风险评分: %.2f, 风险因素: %v", riskScore, riskFactors)
		
		alertLevel := "info"
		if riskLevel == "critical" {
			alertLevel = "critical"
		} else if riskLevel == "high" {
			alertLevel = "warning"
		}
		
		rm.addAlert("global", alertLevel, message, details, "risk_manager")
	}

	// 检查是否需要紧急停止
	if rm.config.EmergencyStopEnabled && riskLevel == "critical" && !rm.state.EmergencyStop {
		rm.triggerEmergencyStop(riskFactors)
	}
}

// triggerEmergencyStop 触发紧急停止
func (rm *RiskManager) triggerEmergencyStop(riskFactors []string) {
	rm.state.EmergencyStop = true
	
	message := "触发紧急停止"
	details := fmt.Sprintf("风险因素: %v", riskFactors)
	rm.addAlert("global", "critical", message, details, "risk_manager")
	
	rm.logger.Printf("紧急停止已触发: %v", riskFactors)
}

// addAlert 添加警报
func (rm *RiskManager) addAlert(alertType, level, message, details, source string) {
	now := time.Now().UTC()
	
	alert := RiskAlert{
		ID:        fmt.Sprintf("%s_%s_%d", alertType, level, now.Unix()),
		Type:      alertType,
		Level:     level,
		Message:   message,
		Details:   details,
		Source:    source,
		Timestamp: now,
		IsResolved: false,
	}
	
	rm.alerts = append(rm.alerts, alert)
	rm.state.TotalAlerts++
	rm.state.LastAlertTime = now
	
	if level == "critical" {
		rm.state.CriticalAlerts++
	}
	
	rm.logger.Printf("新警报 [%s]: %s - %s", level, message, details)
}

// MakeRiskDecision 进行风控决策
func (rm *RiskManager) MakeRiskDecision(symbol string, action string, marketData map[string]interface{}) (*RiskDecision, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	startTime := time.Now()
	decisionID := fmt.Sprintf("risk_%s_%s_%d", symbol, action, startTime.Unix())

	// 检查紧急停止
	if rm.state.EmergencyStop {
		decision := &RiskDecision{
			ID:             decisionID,
			Symbol:         symbol,
			Action:         "block",
			Reason:         "系统紧急停止",
			RiskFactors:    []string{"emergency_stop"},
			Confidence:     1.0,
			Timestamp:      startTime,
			ProcessingTime: time.Since(startTime),
		}
		rm.decisions = append(rm.decisions, *decision)
		return decision, nil
	}

	riskFactors := make([]string, 0)
	
	// 时间检查
	timeCheck := true
	if rm.timeManager.IsTimeForDailyReset() {
		timeCheck = false
		riskFactors = append(riskFactors, "接近日重置时间")
	}

	// 流动性检查
	liquidityCheck := true
	if rm.liquidityMonitor.IsSymbolBlacklisted(symbol) {
		liquidityCheck = false
		riskFactors = append(riskFactors, "币种在黑名单中")
	}
	
	if level, exists := rm.liquidityMonitor.GetLiquidityLevel(symbol); exists {
		if level == LiquidityCritical {
			liquidityCheck = false
			riskFactors = append(riskFactors, "流动性危机")
		}
	}

	// 夏普比率检查
	sharpeCheck := true
	if rm.state.SharpeState.CurrentState == SharpeVeryPoor {
		sharpeCheck = false
		riskFactors = append(riskFactors, "夏普比率极差")
	}

	// AI委员会决策
	aiDecision, err := rm.aiCommittee.MakeDecision(symbol, marketData)
	if err != nil {
		riskFactors = append(riskFactors, "AI委员会决策失败")
	}

	// 综合决策
	finalAction := rm.makeFinalDecision(action, timeCheck, liquidityCheck, sharpeCheck, aiDecision, riskFactors)
	
	// 计算置信度
	confidence := rm.calculateDecisionConfidence(timeCheck, liquidityCheck, sharpeCheck, aiDecision)

	decision := &RiskDecision{
		ID:             decisionID,
		Symbol:         symbol,
		Action:         finalAction,
		Reason:         rm.generateDecisionReason(finalAction, riskFactors),
		RiskFactors:    riskFactors,
		Confidence:     confidence,
		TimeCheck:      timeCheck,
		LiquidityCheck: liquidityCheck,
		SharpeCheck:    sharpeCheck,
		AIDecision:     aiDecision,
		Timestamp:      startTime,
		ProcessingTime: time.Since(startTime),
	}

	rm.decisions = append(rm.decisions, *decision)
	
	rm.logger.Printf("风控决策: %s - %s (%s, 置信度: %.2f)", 
		symbol, finalAction, decision.Reason, confidence)

	return decision, nil
}

// makeFinalDecision 做出最终决策
func (rm *RiskManager) makeFinalDecision(requestedAction string, timeCheck, liquidityCheck, sharpeCheck bool, aiDecision *CommitteeDecision, riskFactors []string) string {
	// 如果任何关键检查失败，阻止操作
	if !liquidityCheck {
		return "block"
	}

	// 如果是开仓操作，需要更严格的检查
	if requestedAction == "open_long" || requestedAction == "open_short" {
		if !timeCheck || !sharpeCheck {
			return "block"
		}
		
		// 检查AI委员会决策
		if aiDecision != nil {
			if aiDecision.FinalDecision == DecisionHold {
				return "block"
			}
			if aiDecision.ConsensusLevel < 0.6 {
				return "block"
			}
		}
	}

	// 如果风险因素过多，降低仓位
	if len(riskFactors) >= 3 {
		return "reduce_size"
	}

	return "allow"
}

// calculateDecisionConfidence 计算决策置信度
func (rm *RiskManager) calculateDecisionConfidence(timeCheck, liquidityCheck, sharpeCheck bool, aiDecision *CommitteeDecision) float64 {
	confidence := 0.0
	
	if timeCheck {
		confidence += 0.2
	}
	if liquidityCheck {
		confidence += 0.3
	}
	if sharpeCheck {
		confidence += 0.2
	}
	
	if aiDecision != nil {
		confidence += 0.3 * aiDecision.Confidence
	}
	
	return confidence
}

// generateDecisionReason 生成决策理由
func (rm *RiskManager) generateDecisionReason(action string, riskFactors []string) string {
	switch action {
	case "allow":
		if len(riskFactors) == 0 {
			return "所有风控检查通过"
		}
		return fmt.Sprintf("风控检查基本通过，存在轻微风险: %v", riskFactors)
	case "block":
		return fmt.Sprintf("风控检查失败，阻止操作: %v", riskFactors)
	case "reduce_size":
		return fmt.Sprintf("存在多项风险因素，建议减少仓位: %v", riskFactors)
	case "force_close":
		return fmt.Sprintf("触发强制平仓条件: %v", riskFactors)
	default:
		return "未知决策类型"
	}
}

// UpdateEquity 更新权益（用于夏普比率计算和回撤监控）
func (rm *RiskManager) UpdateEquity(equity float64) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 更新时间管理器的权益水位
	_, err := rm.timeManager.UpdateEquityWatermark(equity)
	if err != nil {
		return fmt.Errorf("更新权益水位失败: %v", err)
	}

	// 更新夏普比率计算器
	_, err = rm.sharpeCalculator.AddRecord(equity)
	if err != nil {
		return fmt.Errorf("更新夏普比率失败: %v", err)
	}

	return nil
}

// UpdateLiquidity 更新流动性数据
func (rm *RiskManager) UpdateLiquidity(symbol string, openInterest, volume24h float64) error {
	return rm.liquidityMonitor.UpdateLiquidityData(symbol, openInterest, volume24h)
}

// GetCurrentState 获取当前状态
func (rm *RiskManager) GetCurrentState() RiskManagerState {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// 更新状态
	rm.updateStates()
	return rm.state
}

// GetActiveAlerts 获取活跃警报
func (rm *RiskManager) GetActiveAlerts() []RiskAlert {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	alerts := make([]RiskAlert, 0)
	for _, alert := range rm.alerts {
		if !alert.IsResolved {
			alerts = append(alerts, alert)
		}
	}
	
	return alerts
}

// GetRecentDecisions 获取最近的决策
func (rm *RiskManager) GetRecentDecisions(limit int) []RiskDecision {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	if limit <= 0 || limit > len(rm.decisions) {
		limit = len(rm.decisions)
	}
	
	start := len(rm.decisions) - limit
	result := make([]RiskDecision, limit)
	copy(result, rm.decisions[start:])
	
	return result
}

// GetConfig 获取配置
func (rm *RiskManager) GetConfig() RiskManagerConfig {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	return rm.config
}

// UpdateConfig 更新配置
func (rm *RiskManager) UpdateConfig(newConfig RiskManagerConfig) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.config = newConfig
	
	// 更新各模块配置
	rm.timeManager.UpdateConfig(newConfig.TimeManager)
	rm.liquidityMonitor.UpdateConfig(newConfig.LiquidityMonitor)
	rm.sharpeCalculator.UpdateConfig(newConfig.SharpeCalculator)
	rm.aiCommittee.UpdateConfig(newConfig.AICommittee)
	
	rm.logger.Printf("风控管理器配置已更新")
	return nil
}

// ToJSON 序列化为JSON
func (rm *RiskManager) ToJSON() ([]byte, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	data := map[string]interface{}{
		"config":           rm.config,
		"state":            rm.state,
		"active_alerts":    rm.GetActiveAlerts(),
		"recent_decisions": rm.GetRecentDecisions(10),
	}
	
	return json.MarshalIndent(data, "", "  ")
}