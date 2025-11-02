package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// SharpeState 夏普比率状态
type SharpeState string

const (
	SharpeExcellent SharpeState = "excellent" // >2.0
	SharpeGood      SharpeState = "good"      // 1.0-2.0
	SharpeNeutral   SharpeState = "neutral"   // 0.0-1.0
	SharpePoor      SharpeState = "poor"      // -1.0-0.0
	SharpeVeryPoor  SharpeState = "very_poor" // <-1.0
)

// SharpeRecord 夏普比率记录
type SharpeRecord struct {
	ID            string      `json:"id"`
	Timestamp     time.Time   `json:"timestamp"`
	Equity        float64     `json:"equity"`         // 权益值
	Return        float64     `json:"return"`         // 收益率
	SharpeRatio   float64     `json:"sharpe_ratio"`   // 夏普比率
	State         SharpeState `json:"state"`          // 状态
	WindowSize    int         `json:"window_size"`    // 窗口大小
	Confidence    float64     `json:"confidence"`     // 置信度
	IsBuffered    bool        `json:"is_buffered"`    // 是否在缓冲期
}

// StateTransition 状态转换记录
type StateTransition struct {
	ID           string      `json:"id"`
	FromState    SharpeState `json:"from_state"`
	ToState      SharpeState `json:"to_state"`
	Timestamp    time.Time   `json:"timestamp"`
	TriggerValue float64     `json:"trigger_value"`
	BufferCycles int         `json:"buffer_cycles"` // 缓冲周期数
	Reason       string      `json:"reason"`
}

// SharpeCalculatorConfig 夏普比率计算器配置
type SharpeCalculatorConfig struct {
	// 滚动窗口参数
	WindowSize           int     `json:"window_size"`            // 滚动窗口大小（决策周期数）
	MinWindowSize        int     `json:"min_window_size"`        // 最小窗口大小
	
	// 状态缓冲参数
	BufferCycles         int     `json:"buffer_cycles"`          // 状态转换缓冲周期数
	ConfidenceThreshold  float64 `json:"confidence_threshold"`   // 置信度阈值
	
	// 状态阈值
	ExcellentThreshold   float64 `json:"excellent_threshold"`    // 优秀阈值 (2.0)
	GoodThreshold        float64 `json:"good_threshold"`         // 良好阈值 (1.0)
	NeutralThreshold     float64 `json:"neutral_threshold"`      // 中性阈值 (0.0)
	PoorThreshold        float64 `json:"poor_threshold"`         // 较差阈值 (-1.0)
	
	// 计算参数
	RiskFreeRate         float64 `json:"risk_free_rate"`         // 无风险利率
	AnnualizationFactor  float64 `json:"annualization_factor"`  // 年化因子
	OutlierThreshold     float64 `json:"outlier_threshold"`     // 异常值阈值
}

// SharpeCalculatorState 夏普比率计算器状态
type SharpeCalculatorState struct {
	CurrentSharpe      float64     `json:"current_sharpe"`
	CurrentState       SharpeState `json:"current_state"`
	CurrentConfidence  float64     `json:"current_confidence"`
	IsInBuffer         bool        `json:"is_in_buffer"`
	BufferCyclesLeft   int         `json:"buffer_cycles_left"`
	
	WindowRecords      int         `json:"window_records"`      // 窗口内记录数
	TotalRecords       int         `json:"total_records"`       // 总记录数
	LastUpdateTime     time.Time   `json:"last_update_time"`
	
	// 统计信息
	MeanReturn         float64     `json:"mean_return"`
	StdDevReturn       float64     `json:"std_dev_return"`
	MaxSharpe          float64     `json:"max_sharpe"`
	MinSharpe          float64     `json:"min_sharpe"`
	StateTransitions   int         `json:"state_transitions"`   // 状态转换次数
}

// SharpeCalculator 夏普比率计算器
type SharpeCalculator struct {
	config            SharpeCalculatorConfig
	state             SharpeCalculatorState
	records           []SharpeRecord
	transitions       []StateTransition
	bufferTransition  *StateTransition // 待确认的状态转换
	mutex             sync.RWMutex
	logger            *log.Logger
}

// NewSharpeCalculator 创建夏普比率计算器
func NewSharpeCalculator(config SharpeCalculatorConfig) *SharpeCalculator {
	// 设置默认值
	if config.WindowSize <= 0 {
		config.WindowSize = 50 // 默认50个决策周期
	}
	if config.MinWindowSize <= 0 {
		config.MinWindowSize = 10 // 最小10个周期
	}
	if config.BufferCycles <= 0 {
		config.BufferCycles = 2 // 默认2个周期缓冲
	}
	if config.ConfidenceThreshold <= 0 {
		config.ConfidenceThreshold = 0.8 // 80%置信度
	}
	if config.ExcellentThreshold <= 0 {
		config.ExcellentThreshold = 2.0
	}
	if config.GoodThreshold <= 0 {
		config.GoodThreshold = 1.0
	}
	if config.NeutralThreshold == 0 {
		config.NeutralThreshold = 0.0
	}
	if config.PoorThreshold == 0 {
		config.PoorThreshold = -1.0
	}
	if config.AnnualizationFactor <= 0 {
		config.AnnualizationFactor = math.Sqrt(365 * 24 * 12) // 假设每5分钟一个决策周期
	}
	if config.OutlierThreshold <= 0 {
		config.OutlierThreshold = 3.0 // 3倍标准差
	}

	return &SharpeCalculator{
		config:      config,
		state:       SharpeCalculatorState{
			CurrentState:      SharpeNeutral,
			CurrentConfidence: 0.0,
			LastUpdateTime:    time.Now().UTC(),
		},
		records:     make([]SharpeRecord, 0),
		transitions: make([]StateTransition, 0),
		logger:      log.New(log.Writer(), "[SharpeCalculator] ", log.LstdFlags),
	}
}

// AddRecord 添加新的权益记录
func (sc *SharpeCalculator) AddRecord(equity float64) (*SharpeRecord, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if equity <= 0 {
		return nil, fmt.Errorf("权益值必须大于0: %.8f", equity)
	}

	now := time.Now().UTC()
	recordID := fmt.Sprintf("sharpe_%d", now.Unix())

	// 计算收益率
	var returnRate float64
	if len(sc.records) > 0 {
		lastEquity := sc.records[len(sc.records)-1].Equity
		if lastEquity > 0 {
			returnRate = (equity - lastEquity) / lastEquity
		}
	}

	// 检查异常值
	if sc.isOutlier(returnRate) {
		sc.logger.Printf("检测到异常收益率: %.6f%%, 将进行平滑处理", returnRate*100)
		returnRate = sc.smoothOutlier(returnRate)
	}

	// 添加记录
	record := SharpeRecord{
		ID:        recordID,
		Timestamp: now,
		Equity:    equity,
		Return:    returnRate,
	}

	sc.records = append(sc.records, record)

	// 维护滚动窗口
	if len(sc.records) > sc.config.WindowSize {
		sc.records = sc.records[1:]
	}

	// 计算夏普比率
	sharpeRatio, confidence := sc.calculateSharpeRatio()
	record.SharpeRatio = sharpeRatio
	record.Confidence = confidence
	record.WindowSize = len(sc.records)

	// 确定状态
	newState := sc.determineState(sharpeRatio)
	record.State = newState

	// 检查状态转换
	sc.checkStateTransition(newState, sharpeRatio)

	// 更新记录
	sc.records[len(sc.records)-1] = record

	// 更新状态
	sc.updateState(sharpeRatio, newState, confidence)

	sc.logger.Printf("新记录: 权益=%.8f, 收益率=%.4f%%, 夏普=%.4f, 状态=%s, 置信度=%.2f", 
		equity, returnRate*100, sharpeRatio, newState, confidence)

	return &record, nil
}

// isOutlier 检查是否为异常值
func (sc *SharpeCalculator) isOutlier(returnRate float64) bool {
	if len(sc.records) < sc.config.MinWindowSize {
		return false
	}

	// 计算历史收益率的标准差
	returns := make([]float64, 0, len(sc.records))
	for _, record := range sc.records {
		returns = append(returns, record.Return)
	}

	mean, stdDev := sc.calculateMeanAndStdDev(returns)
	threshold := sc.config.OutlierThreshold * stdDev

	return math.Abs(returnRate-mean) > threshold
}

// smoothOutlier 平滑异常值
func (sc *SharpeCalculator) smoothOutlier(returnRate float64) float64 {
	if len(sc.records) < 3 {
		return returnRate
	}

	// 使用最近3个记录的中位数
	recentReturns := make([]float64, 0, 3)
	start := len(sc.records) - 3
	if start < 0 {
		start = 0
	}

	for i := start; i < len(sc.records); i++ {
		recentReturns = append(recentReturns, sc.records[i].Return)
	}

	sort.Float64s(recentReturns)
	median := recentReturns[len(recentReturns)/2]

	// 使用加权平均：70%中位数 + 30%原值
	return 0.7*median + 0.3*returnRate
}

// calculateSharpeRatio 计算夏普比率
func (sc *SharpeCalculator) calculateSharpeRatio() (float64, float64) {
	if len(sc.records) < sc.config.MinWindowSize {
		return 0.0, 0.0
	}

	// 提取收益率
	returns := make([]float64, 0, len(sc.records))
	for _, record := range sc.records {
		returns = append(returns, record.Return)
	}

	// 计算均值和标准差
	meanReturn, stdDev := sc.calculateMeanAndStdDev(returns)

	if stdDev == 0 {
		if meanReturn > 0 {
			return 999.0, 1.0
		} else if meanReturn < 0 {
			return -999.0, 1.0
		}
		return 0.0, 0.0
	}

	// 计算夏普比率
	excessReturn := meanReturn - sc.config.RiskFreeRate
	sharpeRatio := (excessReturn / stdDev) * sc.config.AnnualizationFactor

	// 计算置信度
	confidence := sc.calculateConfidence(len(returns), stdDev)

	return sharpeRatio, confidence
}

// calculateMeanAndStdDev 计算均值和标准差
func (sc *SharpeCalculator) calculateMeanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0.0, 0.0
	}

	// 计算均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 计算标准差
	if len(values) == 1 {
		return mean, 0.0
	}

	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values)-1)
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}

// calculateConfidence 计算置信度
func (sc *SharpeCalculator) calculateConfidence(sampleSize int, stdDev float64) float64 {
	// 基于样本大小的置信度
	sizeConfidence := math.Min(float64(sampleSize)/float64(sc.config.WindowSize), 1.0)
	
	// 基于标准差稳定性的置信度
	stabilityConfidence := 1.0 / (1.0 + stdDev*10) // 标准差越小，置信度越高
	
	// 综合置信度
	confidence := (sizeConfidence + stabilityConfidence) / 2.0
	
	return math.Min(confidence, 1.0)
}

// determineState 确定夏普比率状态
func (sc *SharpeCalculator) determineState(sharpeRatio float64) SharpeState {
	if sharpeRatio >= sc.config.ExcellentThreshold {
		return SharpeExcellent
	} else if sharpeRatio >= sc.config.GoodThreshold {
		return SharpeGood
	} else if sharpeRatio >= sc.config.NeutralThreshold {
		return SharpeNeutral
	} else if sharpeRatio >= sc.config.PoorThreshold {
		return SharpePoor
	} else {
		return SharpeVeryPoor
	}
}

// checkStateTransition 检查状态转换
func (sc *SharpeCalculator) checkStateTransition(newState SharpeState, sharpeRatio float64) {
	currentState := sc.state.CurrentState

	// 如果状态没有变化，重置缓冲
	if newState == currentState {
		if sc.bufferTransition != nil {
			sc.bufferTransition = nil
			sc.state.IsInBuffer = false
			sc.state.BufferCyclesLeft = 0
		}
		return
	}

	// 如果已经在缓冲期
	if sc.state.IsInBuffer && sc.bufferTransition != nil {
		// 检查是否仍然是同一个目标状态
		if sc.bufferTransition.ToState == newState {
			sc.state.BufferCyclesLeft--
			if sc.state.BufferCyclesLeft <= 0 {
				// 缓冲期结束，确认状态转换
				sc.confirmStateTransition(sharpeRatio)
			}
		} else {
			// 目标状态改变，重新开始缓冲
			sc.startStateTransition(currentState, newState, sharpeRatio)
		}
	} else {
		// 开始新的状态转换缓冲
		sc.startStateTransition(currentState, newState, sharpeRatio)
	}
}

// startStateTransition 开始状态转换
func (sc *SharpeCalculator) startStateTransition(fromState, toState SharpeState, sharpeRatio float64) {
	now := time.Now().UTC()
	
	sc.bufferTransition = &StateTransition{
		ID:           fmt.Sprintf("transition_%d", now.Unix()),
		FromState:    fromState,
		ToState:      toState,
		Timestamp:    now,
		TriggerValue: sharpeRatio,
		BufferCycles: sc.config.BufferCycles,
		Reason:       fmt.Sprintf("夏普比率从 %.4f (%s) 变化到 %.4f (%s)", 
			sc.state.CurrentSharpe, fromState, sharpeRatio, toState),
	}
	
	sc.state.IsInBuffer = true
	sc.state.BufferCyclesLeft = sc.config.BufferCycles
	
	sc.logger.Printf("开始状态转换缓冲: %s -> %s (缓冲周期: %d)", 
		fromState, toState, sc.config.BufferCycles)
}

// confirmStateTransition 确认状态转换
func (sc *SharpeCalculator) confirmStateTransition(sharpeRatio float64) {
	if sc.bufferTransition == nil {
		return
	}

	// 记录状态转换
	sc.transitions = append(sc.transitions, *sc.bufferTransition)
	
	// 更新当前状态
	oldState := sc.state.CurrentState
	sc.state.CurrentState = sc.bufferTransition.ToState
	sc.state.StateTransitions++
	
	// 重置缓冲
	sc.bufferTransition = nil
	sc.state.IsInBuffer = false
	sc.state.BufferCyclesLeft = 0
	
	sc.logger.Printf("状态转换确认: %s -> %s (夏普比率: %.4f)", 
		oldState, sc.state.CurrentState, sharpeRatio)
}

// updateState 更新状态
func (sc *SharpeCalculator) updateState(sharpeRatio float64, state SharpeState, confidence float64) {
	sc.state.CurrentSharpe = sharpeRatio
	sc.state.CurrentConfidence = confidence
	sc.state.WindowRecords = len(sc.records)
	sc.state.TotalRecords++
	sc.state.LastUpdateTime = time.Now().UTC()

	// 更新统计信息
	if len(sc.records) > 0 {
		returns := make([]float64, 0, len(sc.records))
		sharpes := make([]float64, 0, len(sc.records))
		
		for _, record := range sc.records {
			returns = append(returns, record.Return)
			sharpes = append(sharpes, record.SharpeRatio)
		}
		
		sc.state.MeanReturn, sc.state.StdDevReturn = sc.calculateMeanAndStdDev(returns)
		
		if len(sharpes) > 0 {
			sort.Float64s(sharpes)
			sc.state.MinSharpe = sharpes[0]
			sc.state.MaxSharpe = sharpes[len(sharpes)-1]
		}
	}
}

// GetCurrentState 获取当前状态
func (sc *SharpeCalculator) GetCurrentState() SharpeCalculatorState {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	return sc.state
}

// GetRecentRecords 获取最近的记录
func (sc *SharpeCalculator) GetRecentRecords(limit int) []SharpeRecord {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	if limit <= 0 || limit > len(sc.records) {
		limit = len(sc.records)
	}
	
	start := len(sc.records) - limit
	result := make([]SharpeRecord, limit)
	copy(result, sc.records[start:])
	
	return result
}

// GetStateTransitions 获取状态转换历史
func (sc *SharpeCalculator) GetStateTransitions(limit int) []StateTransition {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	if limit <= 0 || limit > len(sc.transitions) {
		limit = len(sc.transitions)
	}
	
	start := len(sc.transitions) - limit
	result := make([]StateTransition, limit)
	copy(result, sc.transitions[start:])
	
	// 按时间倒序排列
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	return result
}

// GetConfig 获取配置
func (sc *SharpeCalculator) GetConfig() SharpeCalculatorConfig {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	return sc.config
}

// UpdateConfig 更新配置
func (sc *SharpeCalculator) UpdateConfig(newConfig SharpeCalculatorConfig) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	// 验证配置
	if newConfig.WindowSize < newConfig.MinWindowSize {
		return fmt.Errorf("窗口大小不能小于最小窗口大小")
	}
	if newConfig.BufferCycles < 1 {
		return fmt.Errorf("缓冲周期数必须至少为1")
	}
	
	sc.config = newConfig
	sc.logger.Printf("配置已更新")
	return nil
}

// ToJSON 序列化为JSON
func (sc *SharpeCalculator) ToJSON() ([]byte, error) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	data := map[string]interface{}{
		"config":            sc.config,
		"state":             sc.state,
		"recent_records":    sc.GetRecentRecords(20),
		"recent_transitions": sc.GetStateTransitions(10),
		"buffer_transition": sc.bufferTransition,
	}
	
	return json.MarshalIndent(data, "", "  ")
}