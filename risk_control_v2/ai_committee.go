package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// ModelType AI模型类型
type ModelType string

const (
	ModelQwen     ModelType = "qwen"
	ModelDeepSeek ModelType = "deepseek"
	ModelClaude   ModelType = "claude"
	ModelGPT      ModelType = "gpt"
)

// DecisionType 决策类型
type DecisionType string

const (
	DecisionLong  DecisionType = "long"
	DecisionShort DecisionType = "short"
	DecisionHold  DecisionType = "hold"
	DecisionClose DecisionType = "close"
)

// MarketCondition 市场状态
type MarketCondition string

const (
	MarketBullish MarketCondition = "bullish" // 牛市
	MarketBearish MarketCondition = "bearish" // 熊市
	MarketSideways MarketCondition = "sideways" // 震荡
	MarketVolatile MarketCondition = "volatile" // 高波动
)

// ModelDecision 单个模型的决策
type ModelDecision struct {
	ModelType    ModelType       `json:"model_type"`
	Symbol       string          `json:"symbol"`
	Decision     DecisionType    `json:"decision"`
	Confidence   float64         `json:"confidence"`   // 0.0-1.0
	Reasoning    string          `json:"reasoning"`    // 决策理由
	RiskScore    float64         `json:"risk_score"`   // 风险评分 0.0-1.0
	Timestamp    time.Time       `json:"timestamp"`
	ResponseTime time.Duration   `json:"response_time"` // 响应时间
}

// CommitteeDecision 委员会决策
type CommitteeDecision struct {
	ID               string                   `json:"id"`
	Symbol           string                   `json:"symbol"`
	FinalDecision    DecisionType             `json:"final_decision"`
	Confidence       float64                  `json:"confidence"`
	ConsensusLevel   float64                  `json:"consensus_level"`   // 共识度 0.0-1.0
	ModelDecisions   map[ModelType]ModelDecision `json:"model_decisions"`
	VotingResults    map[DecisionType]int     `json:"voting_results"`
	MarketCondition  MarketCondition          `json:"market_condition"`
	Strategy         string                   `json:"strategy"`          // 使用的策略
	Reasoning        string                   `json:"reasoning"`         // 最终决策理由
	RiskAssessment   string                   `json:"risk_assessment"`   // 风险评估
	Timestamp        time.Time                `json:"timestamp"`
	ProcessingTime   time.Duration            `json:"processing_time"`
}

// ModelPerformance 模型性能统计
type ModelPerformance struct {
	ModelType        ModelType `json:"model_type"`
	TotalDecisions   int       `json:"total_decisions"`
	CorrectDecisions int       `json:"correct_decisions"`
	Accuracy         float64   `json:"accuracy"`
	AvgConfidence    float64   `json:"avg_confidence"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	LastActive       time.Time `json:"last_active"`
	IsActive         bool      `json:"is_active"`
}

// AICommitteeConfig AI委员会配置
type AICommitteeConfig struct {
	// 模型配置
	EnabledModels        []ModelType `json:"enabled_models"`
	PrimaryModel         ModelType   `json:"primary_model"`         // 主模型
	FallbackModel        ModelType   `json:"fallback_model"`        // 备用模型
	
	// 决策参数
	MinConsensusLevel    float64     `json:"min_consensus_level"`   // 最小共识度
	ConservativeMode     bool        `json:"conservative_mode"`     // 保守模式
	RequireUnanimity     bool        `json:"require_unanimity"`     // 是否要求一致同意
	
	// 超时设置
	ModelTimeoutSeconds  int         `json:"model_timeout_seconds"` // 单模型超时
	TotalTimeoutSeconds  int         `json:"total_timeout_seconds"` // 总超时
	
	// 市场状态检测
	VolatilityThreshold  float64     `json:"volatility_threshold"`  // 波动率阈值
	TrendThreshold       float64     `json:"trend_threshold"`       // 趋势阈值
	
	// 风险控制
	MaxRiskScore         float64     `json:"max_risk_score"`        // 最大风险评分
	RiskWeightEnabled    bool        `json:"risk_weight_enabled"`   // 是否启用风险加权
}

// AICommitteeState AI委员会状态
type AICommitteeState struct {
	CurrentStrategy      string                          `json:"current_strategy"`
	MarketCondition      MarketCondition                 `json:"market_condition"`
	LastDecisionTime     time.Time                       `json:"last_decision_time"`
	TotalDecisions       int                             `json:"total_decisions"`
	ConsensusDecisions   int                             `json:"consensus_decisions"`
	ConflictDecisions    int                             `json:"conflict_decisions"`
	AvgConsensusLevel    float64                         `json:"avg_consensus_level"`
	ModelPerformances    map[ModelType]ModelPerformance  `json:"model_performances"`
	ActiveModels         []ModelType                     `json:"active_models"`
}

// AICommittee AI决策委员会
type AICommittee struct {
	config           AICommitteeConfig
	state            AICommitteeState
	decisionHistory  []CommitteeDecision
	mutex            sync.RWMutex
	logger           *log.Logger
}

// NewAICommittee 创建AI委员会
func NewAICommittee(config AICommitteeConfig) *AICommittee {
	// 设置默认值
	if len(config.EnabledModels) == 0 {
		config.EnabledModels = []ModelType{ModelQwen, ModelDeepSeek}
	}
	if config.PrimaryModel == "" {
		config.PrimaryModel = ModelQwen
	}
	if config.FallbackModel == "" {
		config.FallbackModel = ModelDeepSeek
	}
	if config.MinConsensusLevel <= 0 {
		config.MinConsensusLevel = 0.6 // 60%共识度
	}
	if config.ModelTimeoutSeconds <= 0 {
		config.ModelTimeoutSeconds = 30
	}
	if config.TotalTimeoutSeconds <= 0 {
		config.TotalTimeoutSeconds = 90
	}
	if config.VolatilityThreshold <= 0 {
		config.VolatilityThreshold = 0.05 // 5%波动率
	}
	if config.TrendThreshold <= 0 {
		config.TrendThreshold = 0.02 // 2%趋势
	}
	if config.MaxRiskScore <= 0 {
		config.MaxRiskScore = 0.8 // 最大80%风险
	}

	// 初始化模型性能
	performances := make(map[ModelType]ModelPerformance)
	for _, model := range config.EnabledModels {
		performances[model] = ModelPerformance{
			ModelType:  model,
			IsActive:   true,
			LastActive: time.Now().UTC(),
		}
	}

	return &AICommittee{
		config: config,
		state: AICommitteeState{
			CurrentStrategy:   "balanced",
			MarketCondition:   MarketSideways,
			ModelPerformances: performances,
			ActiveModels:      config.EnabledModels,
		},
		decisionHistory: make([]CommitteeDecision, 0),
		logger:          log.New(log.Writer(), "[AICommittee] ", log.LstdFlags),
	}
}

// MakeDecision 进行委员会决策
func (ac *AICommittee) MakeDecision(symbol string, marketData map[string]interface{}) (*CommitteeDecision, error) {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	startTime := time.Now()
	decisionID := fmt.Sprintf("committee_%s_%d", symbol, startTime.Unix())

	// 检测市场状态
	marketCondition := ac.detectMarketCondition(marketData)
	
	// 选择策略
	strategy := ac.selectStrategy(marketCondition)
	
	// 收集各模型决策
	modelDecisions := make(map[ModelType]ModelDecision)
	
	for _, modelType := range ac.state.ActiveModels {
		decision, err := ac.getModelDecision(modelType, symbol, marketData, strategy)
		if err != nil {
			ac.logger.Printf("模型 %s 决策失败: %v", modelType, err)
			continue
		}
		modelDecisions[modelType] = *decision
	}

	// 如果没有足够的模型响应，使用备用策略
	if len(modelDecisions) == 0 {
		return ac.fallbackDecision(symbol, marketCondition, decisionID)
	}

	// 进行投票和共识分析
	finalDecision, confidence, consensusLevel := ac.analyzeConsensus(modelDecisions)
	
	// 风险评估
	riskAssessment := ac.assessRisk(modelDecisions, finalDecision)
	
	// 应用保守策略（如果启用）
	if ac.config.ConservativeMode {
		finalDecision, confidence = ac.applyConservativeStrategy(finalDecision, confidence, consensusLevel)
	}

	// 创建委员会决策
	committeeDecision := CommitteeDecision{
		ID:              decisionID,
		Symbol:          symbol,
		FinalDecision:   finalDecision,
		Confidence:      confidence,
		ConsensusLevel:  consensusLevel,
		ModelDecisions:  modelDecisions,
		VotingResults:   ac.calculateVotingResults(modelDecisions),
		MarketCondition: marketCondition,
		Strategy:        strategy,
		Reasoning:       ac.generateReasoning(modelDecisions, finalDecision),
		RiskAssessment:  riskAssessment,
		Timestamp:       startTime,
		ProcessingTime:  time.Since(startTime),
	}

	// 记录决策
	ac.decisionHistory = append(ac.decisionHistory, committeeDecision)
	
	// 更新状态
	ac.updateState(committeeDecision)
	
	ac.logger.Printf("委员会决策完成: %s - %s (共识度: %.2f, 置信度: %.2f)", 
		symbol, finalDecision, consensusLevel, confidence)

	return &committeeDecision, nil
}

// detectMarketCondition 检测市场状态
func (ac *AICommittee) detectMarketCondition(marketData map[string]interface{}) MarketCondition {
	// 从市场数据中提取关键指标
	volatility, _ := marketData["volatility"].(float64)
	trend, _ := marketData["trend"].(float64)
	volume, _ := marketData["volume"].(float64)
	
	// 判断市场状态
	if volatility > ac.config.VolatilityThreshold*2 {
		return MarketVolatile
	} else if trend > ac.config.TrendThreshold {
		return MarketBullish
	} else if trend < -ac.config.TrendThreshold {
		return MarketBearish
	} else {
		// 使用volume来辅助判断横盘市场的活跃度
		_ = volume // 标记volume已使用
		return MarketSideways
	}
}

// selectStrategy 选择策略
func (ac *AICommittee) selectStrategy(condition MarketCondition) string {
	switch condition {
	case MarketBullish:
		return "aggressive_long"
	case MarketBearish:
		return "defensive_short"
	case MarketVolatile:
		return "scalping"
	default:
		return "balanced"
	}
}

// getModelDecision 获取单个模型的决策（模拟）
func (ac *AICommittee) getModelDecision(modelType ModelType, symbol string, marketData map[string]interface{}, strategy string) (*ModelDecision, error) {
	startTime := time.Now()
	
	// 这里应该调用实际的AI模型API
	// 现在使用模拟逻辑
	decision := ac.simulateModelDecision(modelType, symbol, marketData, strategy)
	
	decision.ResponseTime = time.Since(startTime)
	decision.Timestamp = startTime
	
	// 更新模型性能
	ac.updateModelPerformance(modelType, decision.ResponseTime)
	
	return &decision, nil
}

// simulateModelDecision 模拟模型决策
func (ac *AICommittee) simulateModelDecision(modelType ModelType, symbol string, marketData map[string]interface{}, strategy string) ModelDecision {
	// 基于模型类型和策略生成不同的决策倾向
	var decision DecisionType
	var confidence float64
	var riskScore float64
	var reasoning string

	trend, _ := marketData["trend"].(float64)
	volatility, _ := marketData["volatility"].(float64)

	switch modelType {
	case ModelQwen:
		// Qwen倾向于保守
		if trend > 0.01 {
			decision = DecisionLong
			confidence = 0.7 + trend*5
			reasoning = "技术指标显示上涨趋势，建议做多"
		} else if trend < -0.01 {
			decision = DecisionShort
			confidence = 0.7 - trend*5
			reasoning = "技术指标显示下跌趋势，建议做空"
		} else {
			decision = DecisionHold
			confidence = 0.6
			reasoning = "市场趋势不明确，建议观望"
		}
		riskScore = 0.3 + volatility*2

	case ModelDeepSeek:
		// DeepSeek倾向于激进
		if trend > 0.005 {
			decision = DecisionLong
			confidence = 0.8 + trend*8
			reasoning = "深度分析显示强烈上涨信号"
		} else if trend < -0.005 {
			decision = DecisionShort
			confidence = 0.8 - trend*8
			reasoning = "深度分析显示强烈下跌信号"
		} else {
			decision = DecisionHold
			confidence = 0.5
			reasoning = "信号不够强烈，暂时观望"
		}
		riskScore = 0.4 + volatility*3

	case ModelClaude:
		// Claude倾向于平衡
		if trend > 0.008 {
			decision = DecisionLong
			confidence = 0.75 + trend*6
			reasoning = "综合分析支持做多策略"
		} else if trend < -0.008 {
			decision = DecisionShort
			confidence = 0.75 - trend*6
			reasoning = "综合分析支持做空策略"
		} else {
			decision = DecisionHold
			confidence = 0.65
			reasoning = "市场信号混合，建议保持观望"
		}
		riskScore = 0.35 + volatility*2.5

	default:
		decision = DecisionHold
		confidence = 0.5
		reasoning = "默认保守策略"
		riskScore = 0.5
	}

	// 限制范围
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}
	if riskScore > 1.0 {
		riskScore = 1.0
	}
	if riskScore < 0.0 {
		riskScore = 0.0
	}

	return ModelDecision{
		ModelType:  modelType,
		Symbol:     symbol,
		Decision:   decision,
		Confidence: confidence,
		Reasoning:  reasoning,
		RiskScore:  riskScore,
	}
}

// analyzeConsensus 分析共识
func (ac *AICommittee) analyzeConsensus(decisions map[ModelType]ModelDecision) (DecisionType, float64, float64) {
	if len(decisions) == 0 {
		return DecisionHold, 0.0, 0.0
	}

	// 统计投票
	votes := make(map[DecisionType][]ModelDecision)
	totalConfidence := 0.0
	totalRiskScore := 0.0

	for _, decision := range decisions {
		votes[decision.Decision] = append(votes[decision.Decision], decision)
		totalConfidence += decision.Confidence
		totalRiskScore += decision.RiskScore
	}

	avgConfidence := totalConfidence / float64(len(decisions))
	avgRiskScore := totalRiskScore / float64(len(decisions))

	// 找到得票最多的决策
	var winningDecision DecisionType
	maxVotes := 0
	for decisionType, decisionList := range votes {
		if len(decisionList) > maxVotes {
			maxVotes = len(decisionList)
			winningDecision = decisionType
		}
	}

	// 计算共识度
	consensusLevel := float64(maxVotes) / float64(len(decisions))

	// 如果启用风险加权，调整置信度
	finalConfidence := avgConfidence
	if ac.config.RiskWeightEnabled {
		riskAdjustment := 1.0 - avgRiskScore*0.3 // 风险越高，置信度越低
		finalConfidence *= riskAdjustment
	}

	// 检查是否满足最小共识度要求
	if consensusLevel < ac.config.MinConsensusLevel {
		// 共识度不足，采用保守策略
		return DecisionHold, finalConfidence * 0.5, consensusLevel
	}

	// 检查是否需要一致同意
	if ac.config.RequireUnanimity && consensusLevel < 1.0 {
		return DecisionHold, finalConfidence * 0.7, consensusLevel
	}

	return winningDecision, finalConfidence, consensusLevel
}

// applyConservativeStrategy 应用保守策略
func (ac *AICommittee) applyConservativeStrategy(decision DecisionType, confidence, consensusLevel float64) (DecisionType, float64) {
	// 在保守模式下，只有高置信度和高共识度才执行开仓操作
	if (decision == DecisionLong || decision == DecisionShort) && 
	   (confidence < 0.8 || consensusLevel < 0.8) {
		return DecisionHold, confidence * 0.6
	}
	
	return decision, confidence
}

// assessRisk 评估风险
func (ac *AICommittee) assessRisk(decisions map[ModelType]ModelDecision, finalDecision DecisionType) string {
	if len(decisions) == 0 {
		return "无法评估风险：缺少模型决策"
	}

	totalRisk := 0.0
	riskFactors := make([]string, 0)

	for _, decision := range decisions {
		totalRisk += decision.RiskScore
		if decision.RiskScore > 0.7 {
			riskFactors = append(riskFactors, fmt.Sprintf("%s模型风险评分较高(%.2f)", decision.ModelType, decision.RiskScore))
		}
	}

	avgRisk := totalRisk / float64(len(decisions))
	
	riskLevel := "低"
	if avgRisk > 0.7 {
		riskLevel = "高"
	} else if avgRisk > 0.4 {
		riskLevel = "中"
	}

	assessment := fmt.Sprintf("风险等级: %s (平均风险评分: %.2f)", riskLevel, avgRisk)
	
	if len(riskFactors) > 0 {
		assessment += "; 风险因素: " + strings.Join(riskFactors, ", ")
	}

	if finalDecision != DecisionHold && avgRisk > ac.config.MaxRiskScore {
		assessment += "; 警告: 风险评分超过阈值，建议谨慎操作"
	}

	return assessment
}

// calculateVotingResults 计算投票结果
func (ac *AICommittee) calculateVotingResults(decisions map[ModelType]ModelDecision) map[DecisionType]int {
	results := make(map[DecisionType]int)
	
	for _, decision := range decisions {
		results[decision.Decision]++
	}
	
	return results
}

// generateReasoning 生成决策理由
func (ac *AICommittee) generateReasoning(decisions map[ModelType]ModelDecision, finalDecision DecisionType) string {
	if len(decisions) == 0 {
		return "无模型响应，采用默认策略"
	}

	reasons := make([]string, 0)
	supportingModels := make([]string, 0)

	for modelType, decision := range decisions {
		if decision.Decision == finalDecision {
			supportingModels = append(supportingModels, string(modelType))
			if decision.Reasoning != "" {
				reasons = append(reasons, fmt.Sprintf("%s: %s", modelType, decision.Reasoning))
			}
		}
	}

	reasoning := fmt.Sprintf("支持模型: %s", strings.Join(supportingModels, ", "))
	
	if len(reasons) > 0 {
		reasoning += "; 主要理由: " + strings.Join(reasons, "; ")
	}

	return reasoning
}

// fallbackDecision 备用决策
func (ac *AICommittee) fallbackDecision(symbol string, condition MarketCondition, decisionID string) (*CommitteeDecision, error) {
	decision := CommitteeDecision{
		ID:              decisionID,
		Symbol:          symbol,
		FinalDecision:   DecisionHold,
		Confidence:      0.3,
		ConsensusLevel:  0.0,
		ModelDecisions:  make(map[ModelType]ModelDecision),
		VotingResults:   map[DecisionType]int{DecisionHold: 1},
		MarketCondition: condition,
		Strategy:        "fallback",
		Reasoning:       "所有模型无响应，采用保守的观望策略",
		RiskAssessment:  "风险等级: 未知 (无模型评估)",
		Timestamp:       time.Now().UTC(),
		ProcessingTime:  0,
	}

	ac.logger.Printf("使用备用决策: %s - HOLD", symbol)
	return &decision, nil
}

// updateModelPerformance 更新模型性能
func (ac *AICommittee) updateModelPerformance(modelType ModelType, responseTime time.Duration) {
	if perf, exists := ac.state.ModelPerformances[modelType]; exists {
		perf.TotalDecisions++
		perf.LastActive = time.Now().UTC()
		
		// 更新平均响应时间
		if perf.AvgResponseTime == 0 {
			perf.AvgResponseTime = responseTime
		} else {
			perf.AvgResponseTime = (perf.AvgResponseTime + responseTime) / 2
		}
		
		ac.state.ModelPerformances[modelType] = perf
	}
}

// updateState 更新状态
func (ac *AICommittee) updateState(decision CommitteeDecision) {
	ac.state.LastDecisionTime = decision.Timestamp
	ac.state.TotalDecisions++
	ac.state.CurrentStrategy = decision.Strategy
	ac.state.MarketCondition = decision.MarketCondition

	if decision.ConsensusLevel >= ac.config.MinConsensusLevel {
		ac.state.ConsensusDecisions++
	} else {
		ac.state.ConflictDecisions++
	}

	// 更新平均共识度
	if ac.state.TotalDecisions == 1 {
		ac.state.AvgConsensusLevel = decision.ConsensusLevel
	} else {
		ac.state.AvgConsensusLevel = (ac.state.AvgConsensusLevel + decision.ConsensusLevel) / 2
	}
}

// GetCurrentState 获取当前状态
func (ac *AICommittee) GetCurrentState() AICommitteeState {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	return ac.state
}

// GetRecentDecisions 获取最近的决策
func (ac *AICommittee) GetRecentDecisions(limit int) []CommitteeDecision {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	if limit <= 0 || limit > len(ac.decisionHistory) {
		limit = len(ac.decisionHistory)
	}
	
	start := len(ac.decisionHistory) - limit
	result := make([]CommitteeDecision, limit)
	copy(result, ac.decisionHistory[start:])
	
	// 按时间倒序排列
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	
	return result
}

// GetConfig 获取配置
func (ac *AICommittee) GetConfig() AICommitteeConfig {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	return ac.config
}

// UpdateConfig 更新配置
func (ac *AICommittee) UpdateConfig(newConfig AICommitteeConfig) error {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	
	// 验证配置
	if newConfig.MinConsensusLevel < 0 || newConfig.MinConsensusLevel > 1 {
		return fmt.Errorf("最小共识度必须在0-1之间")
	}
	
	ac.config = newConfig
	ac.logger.Printf("配置已更新")
	return nil
}

// ToJSON 序列化为JSON
func (ac *AICommittee) ToJSON() ([]byte, error) {
	ac.mutex.RLock()
	defer ac.mutex.RUnlock()
	
	data := map[string]interface{}{
		"config":           ac.config,
		"state":            ac.state,
		"recent_decisions": ac.GetRecentDecisions(10),
	}
	
	return json.MarshalIndent(data, "", "  ")
}