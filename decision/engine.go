package decision

import (
	"encoding/json"
	"fmt"
	"log"
	"nofx/disaster"
	"nofx/market"
	"nofx/mcp"
	"nofx/pool"
	"nofx/risk"
	"nofx/signal"
	"strings"
	"time"
)

// PositionInfo 持仓信息
type PositionInfo struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"` // "long" or "short"
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	Quantity         float64 `json:"quantity"`
	Leverage         int     `json:"leverage"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
	LiquidationPrice float64 `json:"liquidation_price"`
	MarginUsed       float64 `json:"margin_used"`
	UpdateTime       int64   `json:"update_time"` // 持仓更新时间戳（毫秒）
}

// AccountInfo 账户信息
type AccountInfo struct {
	TotalEquity      float64 `json:"total_equity"`      // 账户净值
	AvailableBalance float64 `json:"available_balance"` // 可用余额
	TotalPnL         float64 `json:"total_pnl"`         // 总盈亏
	TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
	MarginUsed       float64 `json:"margin_used"`       // 已用保证金
	MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
	PositionCount    int     `json:"position_count"`    // 持仓数量
}

// CandidateCoin 候选币种（来自币种池）
type CandidateCoin struct {
	Symbol  string   `json:"symbol"`
	Sources []string `json:"sources"` // 来源: "ai500" 和/或 "oi_top"
}

// OITopData 持仓量增长Top数据（用于AI决策参考）
type OITopData struct {
	Rank              int     // OI Top排名
	OIDeltaPercent    float64 // 持仓量变化百分比（1小时）
	OIDeltaValue      float64 // 持仓量变化价值
	PriceDeltaPercent float64 // 价格变化百分比
	NetLong           float64 // 净多仓
	NetShort          float64 // 净空仓
}

// Context 交易上下文（传递给AI的完整信息）
type Context struct {
	CurrentTime     string                  `json:"current_time"`
	RuntimeMinutes  int                     `json:"runtime_minutes"`
	CallCount       int                     `json:"call_count"`
	Account         AccountInfo             `json:"account"`
	Positions       []PositionInfo          `json:"positions"`
	CandidateCoins  []CandidateCoin         `json:"candidate_coins"`
	MarketDataMap   map[string]*market.Data `json:"-"` // 不序列化，但内部使用
	OITopDataMap    map[string]*OITopData   `json:"-"` // OI Top数据映射
	Performance     interface{}             `json:"-"` // 历史表现分析（logger.PerformanceAnalysis）
	BTCETHLeverage  int                     `json:"-"` // BTC/ETH杠杆倍数（从配置读取）
	AltcoinLeverage int                     `json:"-"` // 山寨币杠杆倍数（从配置读取）
	
	// === 新增优化模块 ===
	MarketRegimeDetector    *market.RegimeDetector           `json:"-"` // 市场状态检测器
	CorrelationRisk         *risk.CorrelationRiskManager     `json:"-"` // 相关性风险管理
	SignalStrength          *signal.SignalStrengthCalculator `json:"-"` // 信号强度计算器
	DisasterRecovery        *disaster.DisasterRecoveryManager `json:"-"` // 灾难恢复管理
	MarketFilter            *risk.MarketEnvironmentFilter    `json:"-"` // 市场环境过滤器
	PositionManager         *risk.PositionManager            `json:"-"` // 渐进式仓位管理器
	
	// 优化分析结果
	MarketRegimeResult      *market.RegimeAnalysis           `json:"-"` // 当前市场状态分析结果
	CorrelationReport       map[string]interface{}           `json:"-"` // 相关性风险报告
	SignalStrengthMap       map[string]*signal.SignalStrengthResult `json:"-"` // 各币种信号强度
	SOSStatus               map[string]interface{}           `json:"-"` // SOS状态
}

// Decision AI的交易决策
type Decision struct {
	Symbol          string  `json:"symbol"`
	Action          string  `json:"action"` // "open_long", "open_short", "close_long", "close_short", "hold", "wait"
	Leverage        int     `json:"leverage,omitempty"`
	PositionSizeUSD float64 `json:"position_size_usd,omitempty"`
	StopLoss        float64 `json:"stop_loss,omitempty"`
	TakeProfit      float64 `json:"take_profit,omitempty"`
	Confidence      int     `json:"confidence,omitempty"` // 信心度 (0-100)
	RiskUSD         float64 `json:"risk_usd,omitempty"`   // 最大美元风险
	Reasoning       string  `json:"reasoning"`
}

// FullDecision AI的完整决策（包含思维链）
type FullDecision struct {
	UserPrompt string     `json:"user_prompt"` // 发送给AI的输入prompt
	CoTTrace   string     `json:"cot_trace"`   // 思维链分析（AI输出）
	Decisions  []Decision `json:"decisions"`   // 具体决策列表
	Timestamp  time.Time  `json:"timestamp"`
}

// GetFullDecision 获取AI的完整交易决策（批量分析所有币种和持仓）
func GetFullDecision(ctx *Context, mcpClient *mcp.Client) (*FullDecision, error) {
	// 1. 为所有币种获取市场数据
	if err := fetchMarketDataForContext(ctx); err != nil {
		return nil, fmt.Errorf("获取市场数据失败: %w", err)
	}

	// 2. 执行优化分析模块
	if err := executeOptimizationAnalysis(ctx); err != nil {
		log.Printf("⚠️ 优化分析执行失败: %v", err)
		// 不中断主流程，继续执行
	}

	// 3. 检查SOS状态
	if ctx.DisasterRecovery != nil && ctx.DisasterRecovery.IsSOSActive() {
		log.Printf("🚨 SOS模式已激活，限制交易决策")
		// SOS模式下只允许平仓操作
		return generateSOSDecision(ctx), nil
	}

	// 4. 构建 System Prompt（固定规则）和 User Prompt（动态数据）
	systemPrompt := buildSystemPrompt(ctx.Account.TotalEquity, ctx.BTCETHLeverage, ctx.AltcoinLeverage)
	userPrompt := buildUserPrompt(ctx)

	// 5. 调用AI API（使用 system + user prompt）
	aiResponse, err := mcpClient.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("调用AI API失败: %w", err)
	}

	// 6. 解析AI响应
	decision, err := parseFullDecisionResponse(aiResponse, ctx.Account.TotalEquity, ctx.BTCETHLeverage, ctx.AltcoinLeverage)
	if err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %w", err)
	}

	// 7. 应用优化过滤器
	if err := applyOptimizationFilters(ctx, decision); err != nil {
		log.Printf("⚠️ 优化过滤器应用失败: %v", err)
	}

	decision.Timestamp = time.Now()
	decision.UserPrompt = userPrompt // 保存输入prompt
	return decision, nil
}

// fetchMarketDataForContext 为上下文中的所有币种获取市场数据和OI数据
func fetchMarketDataForContext(ctx *Context) error {
	ctx.MarketDataMap = make(map[string]*market.Data)
	ctx.OITopDataMap = make(map[string]*OITopData)

	// 收集所有需要获取数据的币种
	symbolSet := make(map[string]bool)

	// 1. 优先获取持仓币种的数据（这是必须的）
	for _, pos := range ctx.Positions {
		symbolSet[pos.Symbol] = true
	}

	// 2. 候选币种数量根据账户状态动态调整
	maxCandidates := calculateMaxCandidates(ctx)
	for i, coin := range ctx.CandidateCoins {
		if i >= maxCandidates {
			break
		}
		symbolSet[coin.Symbol] = true
	}

	// 并发获取市场数据
	// 持仓币种集合（用于判断是否跳过OI检查）
	positionSymbols := make(map[string]bool)
	for _, pos := range ctx.Positions {
		positionSymbols[pos.Symbol] = true
	}

	for symbol := range symbolSet {
		data, err := market.Get(symbol)
		if err != nil {
			// 单个币种失败不影响整体，只记录错误
			continue
		}

		// ⚠️ 流动性过滤：持仓价值低于15M USD的币种不做（多空都不做）
		// 持仓价值 = 持仓量 × 当前价格
		// 但现有持仓必须保留（需要决策是否平仓）
		isExistingPosition := positionSymbols[symbol]
		if !isExistingPosition && data.OpenInterest != nil && data.CurrentPrice > 0 {
			// 计算持仓价值（USD）= 持仓量 × 当前价格
			oiValue := data.OpenInterest.Latest * data.CurrentPrice
			oiValueInMillions := oiValue / 1_000_000 // 转换为百万美元单位
			if oiValueInMillions < 15 {
				log.Printf("⚠️  %s 持仓价值过低(%.2fM USD < 15M)，跳过此币种 [持仓量:%.0f × 价格:%.4f]",
					symbol, oiValueInMillions, data.OpenInterest.Latest, data.CurrentPrice)
				continue
			}
		}

		ctx.MarketDataMap[symbol] = data
	}

	// 加载OI Top数据（不影响主流程）
	oiPositions, err := pool.GetOITopPositions()
	if err == nil {
		for _, pos := range oiPositions {
			// 标准化符号匹配
			symbol := pos.Symbol
			ctx.OITopDataMap[symbol] = &OITopData{
				Rank:              pos.Rank,
				OIDeltaPercent:    pos.OIDeltaPercent,
				OIDeltaValue:      pos.OIDeltaValue,
				PriceDeltaPercent: pos.PriceDeltaPercent,
				NetLong:           pos.NetLong,
				NetShort:          pos.NetShort,
			}
		}
	}

	return nil
}

// === 新增优化模块集成函数 ===

// executeOptimizationAnalysis 执行优化分析模块
func executeOptimizationAnalysis(ctx *Context) error {
	// 初始化优化模块（如果尚未初始化）
	if err := initializeOptimizationModules(ctx); err != nil {
		return fmt.Errorf("初始化优化模块失败: %w", err)
	}

	// 1. 市场状态检测
	if err := analyzeMarketRegime(ctx); err != nil {
		log.Printf("⚠️ 市场状态分析失败: %v", err)
	}

	// 2. 相关性风险分析
	if err := analyzeCorrelationRisk(ctx); err != nil {
		log.Printf("⚠️ 相关性风险分析失败: %v", err)
	}

	// 3. 信号强度计算
	if err := calculateSignalStrength(ctx); err != nil {
		log.Printf("⚠️ 信号强度计算失败: %v", err)
	}

	// 4. SOS状态检查
	if err := checkSOSConditions(ctx); err != nil {
		log.Printf("⚠️ SOS状态检查失败: %v", err)
	}

	return nil
}

// initializeOptimizationModules 初始化优化模块
func initializeOptimizationModules(ctx *Context) error {
	// 初始化市场状态检测器
	if ctx.MarketRegimeDetector == nil {
		ctx.MarketRegimeDetector = market.NewRegimeDetector()
	}

	// 初始化相关性风险管理器
	if ctx.CorrelationRisk == nil {
		ctx.CorrelationRisk = risk.NewCorrelationRiskManager()
	}

	// 初始化信号强度计算器
	if ctx.SignalStrength == nil {
		ctx.SignalStrength = signal.NewSignalStrengthCalculator()
	}

	// 初始化灾难恢复管理器
	if ctx.DisasterRecovery == nil {
		ctx.DisasterRecovery = disaster.NewDisasterRecoveryManager()
		// 设置回调函数
		ctx.DisasterRecovery.OnSOSTriggered = func(event *disaster.SOSEvent) error {
			log.Printf("🚨 SOS事件触发: %s - %s", event.ID, event.Reason)
			return nil
		}
	}

	// 初始化市场环境过滤器
	if ctx.MarketFilter == nil {
		ctx.MarketFilter = risk.NewMarketEnvironmentFilter()
	}

	// 初始化渐进式仓位管理器
	if ctx.PositionManager == nil {
		ctx.PositionManager = risk.NewPositionManager()
	}

	return nil
}

// analyzeMarketRegime 分析市场状态
func analyzeMarketRegime(ctx *Context) error {
	if ctx.MarketRegimeDetector == nil {
		return fmt.Errorf("市场状态检测器未初始化")
	}

	// 使用BTC数据进行市场状态分析
	btcData, hasBTC := ctx.MarketDataMap["BTCUSDT"]
	if !hasBTC {
		return fmt.Errorf("缺少BTC数据进行市场状态分析")
	}

	// 提取价格序列
	priceSequence := market.ExtractPriceSequence(btcData)
	if len(priceSequence) < 20 {
		return fmt.Errorf("价格序列数据不足，需要至少20个数据点")
	}

	// 执行市场状态检测
	regimeAnalysis := ctx.MarketRegimeDetector.DetectMarketRegime(priceSequence)

	ctx.MarketRegimeResult = regimeAnalysis
	log.Printf("📊 市场状态: %s (置信度: %.2f)", regimeAnalysis.Regime, regimeAnalysis.Confidence)

	return nil
}

// analyzeCorrelationRisk 分析相关性风险
func analyzeCorrelationRisk(ctx *Context) error {
	if ctx.CorrelationRisk == nil {
		return fmt.Errorf("相关性风险管理器未初始化")
	}

	// 主动触发相关性矩阵更新
	// 构建一个虚拟的持仓列表来触发矩阵更新
	var existingPositions []string
	for _, pos := range ctx.Positions {
		existingPositions = append(existingPositions, pos.Symbol)
	}
	
	// 如果没有持仓，使用候选币种来触发更新
	if len(existingPositions) == 0 && len(ctx.CandidateCoins) > 0 {
		existingPositions = append(existingPositions, ctx.CandidateCoins[0].Symbol)
	}
	
	// 触发相关性矩阵更新（通过CheckCorrelationRisk方法）
	if len(existingPositions) > 0 && len(ctx.CandidateCoins) > 0 {
		// 使用第一个候选币种作为新币种来触发更新
		_ = ctx.CorrelationRisk.CheckCorrelationRisk(existingPositions, ctx.CandidateCoins[0].Symbol, ctx.MarketDataMap)
	}

	// 获取当前持仓的相关性报告
	correlationReport := ctx.CorrelationRisk.GetCorrelationReport()
	if correlationReport == nil {
		// 如果相关性报告为空，设置默认值
		ctx.CorrelationReport = map[string]interface{}{
			"symbols":      []string{},
			"matrix":       [][]float64{},
			"last_updated": nil,
		}
		log.Printf("⚠️ 相关性报告为空，可能需要更多市场数据")
	} else {
		ctx.CorrelationReport = map[string]interface{}{
			"symbols":      correlationReport.Symbols,
			"matrix":       correlationReport.Matrix,
			"last_updated": correlationReport.LastUpdated,
		}
		log.Printf("✓ 相关性矩阵已更新，包含 %d 个币种", len(correlationReport.Symbols))
	}

	// 检查高相关性对
	highCorrelationPairs := ctx.CorrelationRisk.GetHighCorrelationPairs(0.8)
	if len(highCorrelationPairs) > 0 {
		log.Printf("⚠️ 发现高相关性持仓对: %d个", len(highCorrelationPairs))
		for _, pair := range highCorrelationPairs {
			log.Printf("   %s - %s: %.3f", pair.Symbol1, pair.Symbol2, pair.Correlation)
		}
	}

	return nil
}

// calculateSignalStrength 计算信号强度
func calculateSignalStrength(ctx *Context) error {
	if ctx.SignalStrength == nil {
		return fmt.Errorf("信号强度计算器未初始化")
	}

	ctx.SignalStrengthMap = make(map[string]*signal.SignalStrengthResult)

	// 为每个候选币种计算信号强度
	for _, coin := range ctx.CandidateCoins {
		marketData, hasData := ctx.MarketDataMap[coin.Symbol]
		if !hasData {
			continue
		}

		// 计算信号强度
		signalResult := ctx.SignalStrength.CalculateSignalStrength(marketData)

		ctx.SignalStrengthMap[coin.Symbol] = signalResult
		log.Printf("📈 %s 信号强度: %.2f (方向: %d, 置信度: %.2f)", 
			coin.Symbol, signalResult.OverallScore, signalResult.Direction, signalResult.Confidence)
	}

	return nil
}

// checkSOSConditions 检查SOS触发条件
func checkSOSConditions(ctx *Context) error {
	if ctx.DisasterRecovery == nil {
		return fmt.Errorf("灾难恢复管理器未初始化")
	}

	// 检查SOS触发条件
	sosEvent, err := ctx.DisasterRecovery.CheckSOSConditions(
		ctx.Account.TotalEquity,
		ctx.Account.TotalPnL,
		ctx.Account.MarginUsedPct,
		"trader_001", // 这里应该从配置中获取trader ID
	)
	if err != nil {
		return fmt.Errorf("SOS条件检查失败: %w", err)
	}

	// 更新SOS状态
	ctx.SOSStatus = ctx.DisasterRecovery.GetSOSStatus()

	if sosEvent != nil {
		log.Printf("🚨 SOS事件触发: %s", sosEvent.Reason)
	}

	return nil
}

// generateSOSDecision 生成SOS模式下的决策
func generateSOSDecision(ctx *Context) *FullDecision {
	decisions := []Decision{}

	// SOS模式下只允许平仓操作
	for _, pos := range ctx.Positions {
		var action string
		if pos.Side == "long" {
			action = "close_long"
		} else {
			action = "close_short"
		}

		decision := Decision{
			Symbol:    pos.Symbol,
			Action:    action,
			Reasoning: "SOS紧急模式激活，执行风险控制平仓",
		}
		decisions = append(decisions, decision)
	}

	// 如果没有持仓，则等待
	if len(decisions) == 0 {
		decisions = append(decisions, Decision{
			Symbol:    "BTCUSDT",
			Action:    "wait",
			Reasoning: "SOS模式激活，暂停所有交易活动",
		})
	}

	return &FullDecision{
		CoTTrace:  "🚨 SOS紧急模式已激活，系统自动执行风险控制措施，停止新开仓并准备平仓现有持仓。",
		Decisions: decisions,
		Timestamp: time.Now(),
	}
}

// applyOptimizationFilters 应用优化过滤器
func applyOptimizationFilters(ctx *Context, decision *FullDecision) error {
	if decision == nil || len(decision.Decisions) == 0 {
		return nil
	}

	filteredDecisions := []Decision{}

	for _, d := range decision.Decisions {
		// 跳过非开仓决策
		if d.Action != "open_long" && d.Action != "open_short" {
			filteredDecisions = append(filteredDecisions, d)
			continue
		}

		// 0. 市场环境过滤 (新增)
		if ctx.MarketFilter != nil {
			if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData {
				if !ctx.MarketFilter.IsFavorableMarket(marketData) {
					log.Printf("🚫 %s 市场环境不利，暂停开仓", d.Symbol)
					d.Action = "wait"
					d.Reasoning = "市场环境不利（波动性不足/流动性不足/趋势不明朗），暂停开仓"
					// 清除开仓相关字段
					d.Leverage = 0
					d.PositionSizeUSD = 0
					d.StopLoss = 0
					d.TakeProfit = 0
					d.Confidence = 0
					d.RiskUSD = 0
				}
			}
			
			// 检查是否为重要经济数据发布时间
			if ctx.MarketFilter.IsEconomicDataTime() {
				log.Printf("⚠️ %s 重要经济数据发布时间，暂停开仓", d.Symbol)
				d.Action = "wait"
				d.Reasoning = "重要经济数据发布时间，暂停开仓"
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			}
		}

		// 0.5. 渐进式仓位管理过滤 (新增)
		if ctx.PositionManager != nil && (d.Action == "open_long" || d.Action == "open_short") {
			// 转换现有持仓格式
			existingPositions := make([]risk.Position, len(ctx.Positions))
			for i, pos := range ctx.Positions {
				existingPositions[i] = risk.Position{
					Symbol: pos.Symbol,
					Side:   pos.Side,
					Size:   pos.Quantity * pos.MarkPrice, // 转换为美元价值
				}
			}
			
			// 检查相关性风险
			if !ctx.MarketFilter.CheckCorrelationRisk(d.Symbol, existingPositions) {
				log.Printf("🚫 %s 相关性风险过高，暂停开仓", d.Symbol)
				d.Action = "wait"
				d.Reasoning = "相关性风险过高，暂停开仓"
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			} else {
				// 计算推荐仓位大小
				marketVolatility := 1.0 // 默认波动性
				if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData {
					if marketData.LongerTermContext != nil && marketData.LongerTermContext.ATR14 > 0 && marketData.LongerTermContext.ATR3 > 0 {
						marketVolatility = marketData.LongerTermContext.ATR3 / marketData.LongerTermContext.ATR14
					}
				}
				
				recommendedSize, recommendation := ctx.PositionManager.GetPositionSizeRecommendation(
					d.Confidence, ctx.Account.TotalEquity, existingPositions, marketVolatility)
				
				if recommendedSize == 0 {
					log.Printf("📊 %s %s", d.Symbol, recommendation)
					d.Action = "wait"
					d.Reasoning = recommendation
					// 清除开仓相关字段
					d.Leverage = 0
					d.PositionSizeUSD = 0
					d.StopLoss = 0
					d.TakeProfit = 0
					d.Confidence = 0
					d.RiskUSD = 0
				} else {
					// 调整仓位大小
					originalSize := d.PositionSizeUSD
					d.PositionSizeUSD = recommendedSize
					log.Printf("📊 %s %s，仓位从$%.0f调整至$%.0f", 
						d.Symbol, recommendation, originalSize, d.PositionSizeUSD)
				}
			}
		}

		// 1. 相关性风险过滤
		if ctx.CorrelationRisk != nil {
			// 提取现有持仓的币种列表
			existingSymbols := make([]string, len(ctx.Positions))
			for i, pos := range ctx.Positions {
				existingSymbols[i] = pos.Symbol
			}
			
			err := ctx.CorrelationRisk.CheckCorrelationRisk(existingSymbols, d.Symbol, ctx.MarketDataMap)
			if err != nil {
				log.Printf("⚠️ %s 相关性检查失败: %v", d.Symbol, err)
				// 转换为等待决策
				d.Action = "wait"
				d.Reasoning = "相关性风险检查失败，暂停开仓"
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			}
		}

		// 2. 信号强度过滤（动态阈值）
		if signalResult, hasSignal := ctx.SignalStrengthMap[d.Symbol]; hasSignal {
			// 根据夏普比率动态调整信号强度阈值
			minSignalScore := 65.0 // 优化：降低默认阈值从75到65
			minConfidence := 0.65  // 优化：降低默认置信度要求从70%到65%
			
			// 检查夏普比率状态
			if ctx.Performance != nil {
				type PerformanceData struct {
					SharpeRatio float64 `json:"sharpe_ratio"`
				}
				var perfData PerformanceData
				if jsonData, err := json.Marshal(ctx.Performance); err == nil {
					if err := json.Unmarshal(jsonData, &perfData); err == nil {
						// 基于夏普比率的动态调整策略
						if perfData.SharpeRatio < -0.3 {
							// 表现很差时稍微严格
							minSignalScore = 70.0
							minConfidence = 0.70
							log.Printf("📊 %s 夏普比率很差(%.3f)，稍微提高要求：≥%.0f分，置信度≥%.0f%%", 
								d.Symbol, perfData.SharpeRatio, minSignalScore, minConfidence*100)
						} else if perfData.SharpeRatio >= -0.3 && perfData.SharpeRatio <= 0 {
							// 轻微亏损时保持适中
							minSignalScore = 65.0
							minConfidence = 0.65
							log.Printf("📊 %s 夏普比率轻微亏损(%.3f)，保持适中要求：≥%.0f分，置信度≥%.0f%%", 
								d.Symbol, perfData.SharpeRatio, minSignalScore, minConfidence*100)
						} else if perfData.SharpeRatio > 0 {
							// 盈利时可以更积极
							minSignalScore = 60.0
							minConfidence = 0.60
							log.Printf("📊 %s 夏普比率盈利(%.3f)，降低要求更积极：≥%.0f分，置信度≥%.0f%%", 
								d.Symbol, perfData.SharpeRatio, minSignalScore, minConfidence*100)
						}
					}
				}
			}
			
			// 检查信号强度是否达标
			if signalResult.OverallScore < minSignalScore {
				log.Printf("🚫 %s 信号强度不足(%.2f < %.0f)，拒绝开仓", d.Symbol, signalResult.OverallScore, minSignalScore)
				d.Action = "wait"
				d.Reasoning = fmt.Sprintf("信号强度不足(%.1f分 < %.0f分)，暂停开仓", signalResult.OverallScore, minSignalScore)
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			} else if signalResult.Confidence < minConfidence {
				log.Printf("🚫 %s 信号置信度不足(%.1f%% < %.0f%%)，拒绝开仓", d.Symbol, signalResult.Confidence*100, minConfidence*100)
				d.Action = "wait"
				d.Reasoning = fmt.Sprintf("信号置信度不足(%.1f%% < %.0f%%)，暂停开仓", signalResult.Confidence*100, minConfidence*100)
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			}
			
			// 如果信号方向与决策不一致，拒绝开仓
			expectedDirection := 1 // 1 for bullish
			if d.Action == "open_short" {
				expectedDirection = -1 // -1 for bearish
			}
			if signalResult.Direction != expectedDirection && signalResult.Direction != 0 {
				directionStr := "看涨"
				if signalResult.Direction == -1 {
					directionStr = "看跌"
				} else if signalResult.Direction == 0 {
					directionStr = "中性"
				}
				expectedStr := "看涨"
				if expectedDirection == -1 {
					expectedStr = "看跌"
				}
				log.Printf("🚫 %s 信号方向不一致，预期%s但信号为%s", d.Symbol, expectedStr, directionStr)
				d.Action = "wait"
				d.Reasoning = fmt.Sprintf("信号方向不一致(预期%s，实际%s)，暂停开仓", expectedStr, directionStr)
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			}
		}

		// 2.5. 强化技术确认（新增）
		if d.Action == "open_long" || d.Action == "open_short" {
			if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData {
				// 检查技术确认条件
				techConfirmPassed := true
				var failReasons []string
				
				// 1. EMA20价格突破确认（至少1%）
				if marketData.CurrentEMA20 > 0 && marketData.CurrentPrice > 0 {
					priceEmaRatio := marketData.CurrentPrice / marketData.CurrentEMA20
					if d.Action == "open_long" {
						// 做多要求价格突破EMA20至少1%
						if priceEmaRatio < 1.01 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("价格未充分突破EMA20(%.3f < 1.01)", priceEmaRatio))
						}
					} else if d.Action == "open_short" {
						// 做空要求价格跌破EMA20至少1%
						if priceEmaRatio > 0.99 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("价格未充分跌破EMA20(%.3f > 0.99)", priceEmaRatio))
						}
					}
				}
				
				// 2. RSI明确信号确认
				if marketData.CurrentRSI7 > 0 {
					if d.Action == "open_long" {
						// 做多要求RSI < 30（超卖）
						if marketData.CurrentRSI7 >= 30 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("RSI7未达超卖区间(%.1f >= 30)", marketData.CurrentRSI7))
						}
					} else if d.Action == "open_short" {
						// 做空要求RSI > 70（超买）
						if marketData.CurrentRSI7 <= 70 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("RSI7未达超买区间(%.1f <= 70)", marketData.CurrentRSI7))
						}
					}
				}
				
				// 3. MACD方向确认
				if marketData.CurrentMACD != 0 {
					if d.Action == "open_long" {
						// 做多要求MACD > 0
						if marketData.CurrentMACD <= 0 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("MACD方向不支持做多(%.6f <= 0)", marketData.CurrentMACD))
						}
					} else if d.Action == "open_short" {
						// 做空要求MACD < 0
						if marketData.CurrentMACD >= 0 {
							techConfirmPassed = false
							failReasons = append(failReasons, fmt.Sprintf("MACD方向不支持做空(%.6f >= 0)", marketData.CurrentMACD))
						}
					}
				}
				
				// 如果技术确认未通过，拒绝开仓
				if !techConfirmPassed {
					log.Printf("🚫 %s 技术确认未通过：%s", d.Symbol, strings.Join(failReasons, "；"))
					d.Action = "wait"
					d.Reasoning = fmt.Sprintf("技术确认未通过：%s", strings.Join(failReasons, "；"))
					// 清除开仓相关字段
					d.Leverage = 0
					d.PositionSizeUSD = 0
					d.StopLoss = 0
					d.TakeProfit = 0
					d.Confidence = 0
					d.RiskUSD = 0
				} else {
					log.Printf("✅ %s 技术确认通过：EMA20突破、RSI信号、MACD方向均符合要求", d.Symbol)
				}
			}
		}

		// 3. 市场状态过滤
		if ctx.MarketRegimeResult != nil {
			// 在高波动率市场中降低仓位
			if ctx.MarketRegimeResult.Regime == market.HighVolatility {
				if d.Action == "open_long" || d.Action == "open_short" {
					d.PositionSizeUSD *= 0.7 // 降低30%仓位
					log.Printf("📉 %s 高波动率市场，降低仓位至%.0f", d.Symbol, d.PositionSizeUSD)
				}
			}
			
			// 在不确定市场中提高开仓门槛
			if ctx.MarketRegimeResult.Regime == market.Uncertain {
				if d.Confidence < 80 {
					log.Printf("🤔 %s 市场不确定且置信度不足(%d)，暂停开仓", d.Symbol, d.Confidence)
					d.Action = "wait"
					d.Reasoning = "市场状态不确定且信号置信度不足，暂停开仓"
					// 清除开仓相关字段
					d.Leverage = 0
					d.PositionSizeUSD = 0
					d.StopLoss = 0
					d.TakeProfit = 0
					d.Confidence = 0
					d.RiskUSD = 0
				}
			}
		}

		// 4. 优化仓位管理（负夏普比率时降低仓位 + 单笔风险控制）
		if d.Action == "open_long" || d.Action == "open_short" {
			accountEquity := ctx.Account.TotalEquity
			originalSize := d.PositionSizeUSD
			positionAdjusted := false
			var adjustmentReasons []string
			
			// 4.1 夏普比率仓位调整
			if ctx.Performance != nil {
				type PerformanceData struct {
					SharpeRatio float64 `json:"sharpe_ratio"`
				}
				var perfData PerformanceData
				if jsonData, err := json.Marshal(ctx.Performance); err == nil {
					if err := json.Unmarshal(jsonData, &perfData); err == nil {
						// 负夏普比率时仓位减半至8%
						if perfData.SharpeRatio < 0 {
							// 计算目标仓位（账户净值的8%）
							targetPositionSize := accountEquity * 0.08
							if d.PositionSizeUSD > targetPositionSize {
								d.PositionSizeUSD = targetPositionSize
								positionAdjusted = true
								adjustmentReasons = append(adjustmentReasons, 
									fmt.Sprintf("负夏普比率(%.3f)，仓位限制至8%%", perfData.SharpeRatio))
								log.Printf("📊 %s 负夏普比率(%.3f)，仓位从%.0f调整至%.0f(8%%)", 
									d.Symbol, perfData.SharpeRatio, originalSize, d.PositionSizeUSD)
							}
						}
					}
				}
			}
			
			// 4.2 单笔风险控制（≤2%）
			if d.StopLoss > 0 && d.PositionSizeUSD > 0 {
				// 计算当前风险
				var riskPercent float64
				if d.Action == "open_long" {
					// 做多风险 = (开仓价格 - 止损价格) / 开仓价格 * 仓位大小 / 账户净值
					if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData && marketData.CurrentPrice > 0 {
						priceRisk := (marketData.CurrentPrice - d.StopLoss) / marketData.CurrentPrice
						riskPercent = priceRisk * d.PositionSizeUSD / accountEquity
					}
				} else if d.Action == "open_short" {
					// 做空风险 = (止损价格 - 开仓价格) / 开仓价格 * 仓位大小 / 账户净值
					if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData && marketData.CurrentPrice > 0 {
						priceRisk := (d.StopLoss - marketData.CurrentPrice) / marketData.CurrentPrice
						riskPercent = priceRisk * d.PositionSizeUSD / accountEquity
					}
				}
				
				// 如果风险超过2%，调整仓位
				maxRiskPercent := 0.02 // 2%
				if riskPercent > maxRiskPercent {
					// 按风险比例调整仓位
					adjustmentFactor := maxRiskPercent / riskPercent
					d.PositionSizeUSD *= adjustmentFactor
					positionAdjusted = true
					adjustmentReasons = append(adjustmentReasons, 
						fmt.Sprintf("单笔风险控制(%.1f%% → 2.0%%)", riskPercent*100))
					log.Printf("⚠️ %s 单笔风险过高(%.1f%%)，仓位从%.0f调整至%.0f", 
						d.Symbol, riskPercent*100, originalSize, d.PositionSizeUSD)
				}
				
				// 更新RiskUSD字段
				d.RiskUSD = d.PositionSizeUSD * (riskPercent / (originalSize / d.PositionSizeUSD))
			}
			
			// 更新reasoning说明仓位调整原因
			if positionAdjusted {
				d.Reasoning = fmt.Sprintf("%s (%s)", d.Reasoning, strings.Join(adjustmentReasons, "；"))
			}
		}

		// 5. 动态止损策略（新增）
		if d.Action == "open_long" || d.Action == "open_short" {
			if marketData, hasMarketData := ctx.MarketDataMap[d.Symbol]; hasMarketData && marketData.CurrentPrice > 0 {
				// 计算更紧密的止损（1.5-2%）
				var tighterStopLoss float64
				var stopLossPercent float64 = 0.02 // 默认2%
				
				// 根据市场波动性调整止损幅度
				if ctx.MarketRegimeResult != nil {
					switch ctx.MarketRegimeResult.Regime {
					case market.HighVolatility:
						stopLossPercent = 0.02 // 高波动率时2%
					case market.LowVolatility:
						stopLossPercent = 0.015 // 低波动率时1.5%
					default:
						stopLossPercent = 0.0175 // 正常市场1.75%
					}
				}
				
				// 计算紧密止损价格
				if d.Action == "open_long" {
					tighterStopLoss = marketData.CurrentPrice * (1 - stopLossPercent)
				} else { // open_short
					tighterStopLoss = marketData.CurrentPrice * (1 + stopLossPercent)
				}
				
				// 如果原止损比紧密止损更宽松，则使用紧密止损
				var stopLossAdjusted bool
				if d.Action == "open_long" && d.StopLoss < tighterStopLoss {
					originalStopLoss := d.StopLoss
					d.StopLoss = tighterStopLoss
					stopLossAdjusted = true
					log.Printf("🎯 %s 做多止损收紧：%.2f → %.2f (%.1f%%)", 
						d.Symbol, originalStopLoss, d.StopLoss, stopLossPercent*100)
				} else if d.Action == "open_short" && d.StopLoss > tighterStopLoss {
					originalStopLoss := d.StopLoss
					d.StopLoss = tighterStopLoss
					stopLossAdjusted = true
					log.Printf("🎯 %s 做空止损收紧：%.2f → %.2f (%.1f%%)", 
						d.Symbol, originalStopLoss, d.StopLoss, stopLossPercent*100)
				}
				
				// 重新计算风险回报比，确保仍然满足要求
				if stopLossAdjusted {
					var riskPercent, rewardPercent, riskRewardRatio float64
					entryPrice := marketData.CurrentPrice
					
					if d.Action == "open_long" {
						riskPercent = (entryPrice - d.StopLoss) / entryPrice * 100
						rewardPercent = (d.TakeProfit - entryPrice) / entryPrice * 100
						if riskPercent > 0 {
							riskRewardRatio = rewardPercent / riskPercent
						}
					} else {
						riskPercent = (d.StopLoss - entryPrice) / entryPrice * 100
						rewardPercent = (entryPrice - d.TakeProfit) / entryPrice * 100
						if riskPercent > 0 {
							riskRewardRatio = rewardPercent / riskPercent
						}
					}
					
					// 如果风险回报比低于3:1，调整止盈目标
					if riskRewardRatio < 3.0 {
						if d.Action == "open_long" {
							// 调整止盈以维持3:1风险回报比
							riskAmount := entryPrice - d.StopLoss
							d.TakeProfit = entryPrice + (riskAmount * 3.0)
						} else {
							// 调整止盈以维持3:1风险回报比
							riskAmount := d.StopLoss - entryPrice
							d.TakeProfit = entryPrice - (riskAmount * 3.0)
						}
						log.Printf("📊 %s 调整止盈目标以维持3:1风险回报比：%.2f", d.Symbol, d.TakeProfit)
					}
					
					// 更新reasoning说明止损调整
					d.Reasoning = fmt.Sprintf("%s (紧密止损%.1f%%)", d.Reasoning, stopLossPercent*100)
				}
			}
		}

		// 6. 市场环境过滤（新增）
		if d.Action == "open_long" || d.Action == "open_short" {
			shouldBlockOpening := false
			var blockReasons []string
			
			// 6.1 低波动率过滤（置信度<50%时暂停开仓）
			if ctx.MarketRegimeResult != nil {
				if ctx.MarketRegimeResult.Confidence < 0.5 {
					shouldBlockOpening = true
					blockReasons = append(blockReasons, 
						fmt.Sprintf("市场置信度过低(%.1f%% < 50%%)", ctx.MarketRegimeResult.Confidence*100))
					log.Printf("🚫 %s 市场置信度过低(%.1f%%)，暂停开仓", d.Symbol, ctx.MarketRegimeResult.Confidence*100)
				}
			}
			
			// 6.2 夏普比率过滤（<-0.05时停止新开仓）
			if ctx.Performance != nil {
				type PerformanceData struct {
					SharpeRatio float64 `json:"sharpe_ratio"`
				}
				var perfData PerformanceData
				if jsonData, err := json.Marshal(ctx.Performance); err == nil {
					if err := json.Unmarshal(jsonData, &perfData); err == nil {
						if perfData.SharpeRatio < -0.05 {
							shouldBlockOpening = true
							blockReasons = append(blockReasons, 
								fmt.Sprintf("夏普比率过低(%.3f < -0.05)", perfData.SharpeRatio))
							log.Printf("🚫 %s 夏普比率过低(%.3f)，停止新开仓", d.Symbol, perfData.SharpeRatio)
						}
					}
				}
			}
			
			// 如果触发市场环境过滤，拒绝开仓
			if shouldBlockOpening {
				log.Printf("🚫 %s 市场环境过滤触发：%s", d.Symbol, strings.Join(blockReasons, "；"))
				d.Action = "wait"
				d.Reasoning = fmt.Sprintf("市场环境过滤：%s", strings.Join(blockReasons, "；"))
				// 清除开仓相关字段
				d.Leverage = 0
				d.PositionSizeUSD = 0
				d.StopLoss = 0
				d.TakeProfit = 0
				d.Confidence = 0
				d.RiskUSD = 0
			}
		}

		filteredDecisions = append(filteredDecisions, d)
	}

	decision.Decisions = filteredDecisions
	return nil
}

// calculateMaxCandidates 根据账户状态计算需要分析的候选币种数量
func calculateMaxCandidates(ctx *Context) int {
	// 直接返回候选池的全部币种数量
	// 因为候选池已经在 auto_trader.go 中筛选过了
	// 固定分析前20个评分最高的币种（来自AI500）
	return len(ctx.CandidateCoins)
}

// buildSystemPrompt 构建 System Prompt（固定规则，可缓存）
func buildSystemPrompt(accountEquity float64, btcEthLeverage, altcoinLeverage int) string {
	var sb strings.Builder

	// === 核心使命 ===
	sb.WriteString("你是专业的加密货币交易AI，在币安合约市场进行自主交易。\n\n")
	sb.WriteString("# 🎯 核心目标\n\n")
	sb.WriteString("**最大化夏普比率（Sharpe Ratio）**\n\n")
	sb.WriteString("夏普比率 = 平均收益 / 收益波动率\n\n")
	sb.WriteString("**这意味着**：\n")
	sb.WriteString("- ✅ 高质量交易（高胜率、大盈亏比）→ 提升夏普\n")
	sb.WriteString("- ✅ 稳定收益、控制回撤 → 提升夏普\n")
	sb.WriteString("- ✅ 耐心持仓、让利润奔跑 → 提升夏普\n")
	sb.WriteString("- ❌ 频繁交易、小盈小亏 → 增加波动，严重降低夏普\n")
	sb.WriteString("- ❌ 过度交易、手续费损耗 → 直接亏损\n")
	sb.WriteString("- ❌ 过早平仓、频繁进出 → 错失大行情\n\n")
	sb.WriteString("**关键认知**: 系统每3分钟扫描一次，但不意味着每次都要交易！\n")
	sb.WriteString("大多数时候应该是 `wait` 或 `hold`，只在极佳机会时才开仓。\n\n")

	// === 硬约束（风险控制）===
	sb.WriteString("# ⚖️ 硬约束（风险控制）\n\n")
	sb.WriteString("1. **风险回报比**: 必须 ≥ 1:3（冒1%风险，赚3%+收益）\n")
	sb.WriteString("2. **最多持仓**: 3个币种（质量>数量）\n")
	sb.WriteString(fmt.Sprintf("3. **单币仓位**: 山寨%.0f-%.0f U(%dx杠杆) | BTC/ETH %.0f-%.0f U(%dx杠杆)\n",
		accountEquity*0.8, accountEquity*1.5, altcoinLeverage, accountEquity*5, accountEquity*10, btcEthLeverage))
	sb.WriteString("4. **保证金**: 总使用率 ≤ 90%\n\n")

	// === 做空激励 ===
	sb.WriteString("# 📉 做多做空平衡\n\n")
	sb.WriteString("**重要**: 下跌趋势做空的利润 = 上涨趋势做多的利润\n\n")
	sb.WriteString("- 上涨趋势 → 做多\n")
	sb.WriteString("- 下跌趋势 → 做空\n")
	sb.WriteString("- 震荡市场 → 观望\n\n")
	sb.WriteString("**不要有做多偏见！做空是你的核心工具之一**\n\n")

	// === 交易频率认知 ===
	sb.WriteString("# ⏱️ 交易频率认知\n\n")
	sb.WriteString("**量化标准**:\n")
	sb.WriteString("- 优秀交易员：每天2-4笔 = 每小时0.1-0.2笔\n")
	sb.WriteString("- 过度交易：每小时>2笔 = 严重问题\n")
	sb.WriteString("- 最佳节奏：基于条件的动态退出规则（无强制观察期）\n\n")
	sb.WriteString("**自查**:\n")
	sb.WriteString("如果你发现自己每个周期都在交易 → 说明标准太低\n")
	sb.WriteString("如果你发现过度频繁交易 → 说明标准太低\n\n")

	// === 开仓信号强度 ===
	sb.WriteString("# 🎯 开仓标准（严格）\n\n")
	sb.WriteString("只在**强信号**时开仓，不确定就观望。\n\n")
	sb.WriteString("**你拥有的完整数据**：\n")
	sb.WriteString("- 📊 **原始序列**：3分钟价格序列(MidPrices数组) + 4小时K线序列\n")
	sb.WriteString("- 📈 **技术序列**：EMA20序列、MACD序列、RSI7序列、RSI14序列\n")
	sb.WriteString("- 💰 **资金序列**：成交量序列、持仓量(OI)序列、资金费率\n")
	sb.WriteString("- 🎯 **筛选标记**：AI500评分 / OI_Top排名（如果有标注）\n\n")
	sb.WriteString("**分析方法**（完全由你自主决定）：\n")
	sb.WriteString("- 自由运用序列数据，你可以做但不限于趋势分析、形态识别、支撑阻力、技术阻力位、斐波那契、波动带计算\n")
	sb.WriteString("- 多维度交叉验证（价格+量+OI+指标+序列形态）\n")
	sb.WriteString("- 用你认为最有效的方法发现高确定性机会\n")
	sb.WriteString("- 综合信心度 ≥ 75 才开仓\n\n")
	sb.WriteString("**🔍 优化分析工具**（系统自动提供）：\n")
	sb.WriteString("- **市场状态检测**：自动识别牛市/熊市/震荡市场，提供置信度和波动性分析\n")
	sb.WriteString("- **相关性风险控制**：检测币种间高相关性，避免重复风险敞口\n")
	sb.WriteString("- **信号强度量化**：多维度评分系统，提供客观的信号质量评估\n")
	sb.WriteString("- **灾难恢复管理**：监控回撤、保证金使用率，在极端情况下触发SOS保护\n\n")
	sb.WriteString("**如何使用优化分析**：\n")
	sb.WriteString("- 市场状态：在震荡市场中降低开仓频率，在趋势市场中积极跟随\n")
	sb.WriteString("- 相关性风险：避免开仓高相关性币种，分散投资组合风险\n")
	sb.WriteString("- 信号强度：优先选择高评分(>70分)的交易机会\n")
	sb.WriteString("- SOS状态：如果触发紧急状态，优先执行系统建议的保护性行动\n\n")
	sb.WriteString("**避免低质量信号**：\n")
	sb.WriteString("- 单一维度（只看一个指标）\n")
	sb.WriteString("- 相互矛盾（涨但量萎缩）\n")
	sb.WriteString("- 横盘震荡\n")
	sb.WriteString("- 刚平仓不久（<15分钟）\n\n")

	// === 夏普比率自我进化 ===
	sb.WriteString("# 🧬 夏普比率自我进化\n\n")
	sb.WriteString("每次你会收到**夏普比率**作为绩效反馈（周期级别）：\n\n")
	sb.WriteString("**夏普比率 < -0.5** (持续亏损):\n")
	sb.WriteString("  → 🛑 停止交易，连续观望至少6个周期（18分钟）\n")
	sb.WriteString("  → 🔍 深度反思：\n")
	sb.WriteString("     • 交易频率过高？（每小时>2次就是过度）\n")
	sb.WriteString("     • 交易过于频繁？（未遵循动态退出规则）\n")
	sb.WriteString("     • 信号强度不足？（信心度<65）\n")
	sb.WriteString("     • 是否在做空？（单边做多是错误的）\n\n")
	sb.WriteString("**夏普比率 -0.5 ~ 0** (轻微亏损):\n")
	sb.WriteString("  → ⚠️ 严格控制：只做信心度>75的交易\n")
	sb.WriteString("  → 减少交易频率：每小时最多1笔新开仓\n")
	sb.WriteString("  → 严格风控：只在明确止损信号时平仓\n")
	sb.WriteString("  → 降低仓位：单笔仓位从16%降低到8%（系统自动调整）\n\n")
	sb.WriteString("**夏普比率 0 ~ 0.7** (正收益):\n")
	sb.WriteString("  → ✅ 维持当前策略\n\n")
	sb.WriteString("**夏普比率 > 0.7** (优异表现):\n")
	sb.WriteString("  → 🚀 可适度扩大仓位\n\n")
	sb.WriteString("**关键**: 夏普比率是唯一指标，它会自然惩罚频繁交易和过度进出。\n\n")

	// === 持仓管理规则 ===
	sb.WriteString("# ⏱️ 持仓管理规则\n\n")
	sb.WriteString("**动态退出策略（无观察期限制）**：\n")
	sb.WriteString("- 基于市场条件和技术指标的实时退出决策\n")
	sb.WriteString("- 止损：价格触及初始止损线时立即执行\n")
	sb.WriteString("- 跟踪止盈：从浮盈峰值回撤20%时触发平仓\n")
	sb.WriteString("- 目标止盈：达到预设止盈目标时平仓\n")
	sb.WriteString("- RSI超买止盈：当RSI > 80且持仓盈利时，考虑部分止盈（50%仓位）\n")
	sb.WriteString("- 趋势反转：当技术指标显示趋势反转时及时退出\n\n")
	sb.WriteString("**平仓reasoning示例**：\n")
	sb.WriteString("- 止损退出：\"价格触及初始止损线，执行严格风控\"\n")
	sb.WriteString("- 跟踪止盈：\"价格从浮盈峰值回撤超过20%，触发跟踪止盈锁定利润\"\n")
	sb.WriteString("- 达到目标：\"达到初始止盈目标，锁定收益\"\n")
	sb.WriteString("- 趋势反转：\"技术指标显示趋势反转，及时退出避免回撤\"\n\n")

	// === 决策流程 ===
	sb.WriteString("# 📋 决策流程\n\n")
	sb.WriteString("1. **分析夏普比率**: 当前策略是否有效？需要调整吗？\n")
	sb.WriteString("2. **评估持仓**: 趋势是否改变？是否该止盈/止损？\n")
	sb.WriteString("3. **寻找新机会**: 有强信号吗？多空机会？\n")
	sb.WriteString("4. **输出决策**: 思维链分析 + JSON\n\n")

	// === 输出格式 ===
	sb.WriteString("# 📤 输出格式\n\n")
	sb.WriteString("**第一步: 思维链（纯文本）**\n")
	sb.WriteString("简洁分析你的思考过程\n\n")
	sb.WriteString("**第二步: JSON决策数组**\n\n")
	sb.WriteString("```json\n[\n")
	sb.WriteString(fmt.Sprintf("  {\"symbol\": \"BTCUSDT\", \"action\": \"open_short\", \"leverage\": %d, \"position_size_usd\": %.0f, \"stop_loss\": 97000, \"take_profit\": 91000, \"confidence\": 85, \"risk_usd\": 300, \"reasoning\": \"下跌趋势+MACD死叉\"},\n", btcEthLeverage, accountEquity*5))
	sb.WriteString("  {\"symbol\": \"ETHUSDT\", \"action\": \"close_long\", \"reasoning\": \"止盈离场\"}\n")
	sb.WriteString("]\n```\n\n")
	sb.WriteString("**字段说明**:\n")
	sb.WriteString("- `action`: 只能使用以下6种action: open_long | open_short | close_long | close_short | hold | wait\n")
	sb.WriteString("- `confidence`: 0-100（开仓建议≥75）\n")
	sb.WriteString("- 开仓时必填: leverage, position_size_usd, stop_loss, take_profit, confidence, risk_usd, reasoning\n")
	sb.WriteString("- **严禁使用其他action**（如update_stop_loss等），只能使用上述6种\n\n")

	// === 关键提醒 ===
	sb.WriteString("---\n\n")
	sb.WriteString("**记住**: \n")
	sb.WriteString("- 目标是夏普比率，不是交易频率\n")
	sb.WriteString("- 做空 = 做多，都是赚钱工具\n")
	sb.WriteString("- 宁可错过，不做低质量交易\n")
	sb.WriteString("- 风险回报比1:3是底线\n")
	sb.WriteString("- **严格遵守action限制**：只能使用6种有效action，禁止自创action\n")

	return sb.String()
}

// buildUserPrompt 构建 User Prompt（动态数据）
func buildUserPrompt(ctx *Context) string {
	var sb strings.Builder

	// 系统状态
	sb.WriteString(fmt.Sprintf("**时间**: %s | **周期**: #%d | **运行**: %d分钟\n\n",
		ctx.CurrentTime, ctx.CallCount, ctx.RuntimeMinutes))

	// BTC 市场
	if btcData, hasBTC := ctx.MarketDataMap["BTCUSDT"]; hasBTC {
		sb.WriteString(fmt.Sprintf("**BTC**: %.2f (1h: %+.2f%%, 4h: %+.2f%%) | MACD: %.4f | RSI: %.2f\n\n",
			btcData.CurrentPrice, btcData.PriceChange1h, btcData.PriceChange4h,
			btcData.CurrentMACD, btcData.CurrentRSI7))
	}

	// 账户
	sb.WriteString(fmt.Sprintf("**账户**: 净值%.2f | 余额%.2f (%.1f%%) | 盈亏%+.2f%% | 保证金%.1f%% | 持仓%d个\n\n",
		ctx.Account.TotalEquity,
		ctx.Account.AvailableBalance,
		(ctx.Account.AvailableBalance/ctx.Account.TotalEquity)*100,
		ctx.Account.TotalPnLPct,
		ctx.Account.MarginUsedPct,
		ctx.Account.PositionCount))

	// 持仓（完整市场数据）
	if len(ctx.Positions) > 0 {
		sb.WriteString("## 当前持仓\n")
		for i, pos := range ctx.Positions {
			// 计算持仓时长
			holdingDuration := ""
			if pos.UpdateTime > 0 {
				durationMs := time.Now().UnixMilli() - pos.UpdateTime
				durationMin := durationMs / (1000 * 60) // 转换为分钟
				if durationMin < 60 {
					holdingDuration = fmt.Sprintf(" | 持仓时长%d分钟", durationMin)
				} else {
					durationHour := durationMin / 60
					durationMinRemainder := durationMin % 60
					holdingDuration = fmt.Sprintf(" | 持仓时长%d小时%d分钟", durationHour, durationMinRemainder)
				}
			}

			sb.WriteString(fmt.Sprintf("%d. %s %s | 入场价%.4f 当前价%.4f | 盈亏%+.2f%% | 杠杆%dx | 保证金%.0f | 强平价%.4f%s\n\n",
				i+1, pos.Symbol, strings.ToUpper(pos.Side),
				pos.EntryPrice, pos.MarkPrice, pos.UnrealizedPnLPct,
				pos.Leverage, pos.MarginUsed, pos.LiquidationPrice, holdingDuration))

			// 使用FormatMarketData输出完整市场数据
			if marketData, ok := ctx.MarketDataMap[pos.Symbol]; ok {
				sb.WriteString(market.Format(marketData))
				sb.WriteString("\n")
			}
		}
	} else {
		sb.WriteString("**当前持仓**: 无\n\n")
	}

	// 候选币种（完整市场数据）
	sb.WriteString(fmt.Sprintf("## 候选币种 (%d个)\n\n", len(ctx.MarketDataMap)))
	displayedCount := 0
	for _, coin := range ctx.CandidateCoins {
		marketData, hasData := ctx.MarketDataMap[coin.Symbol]
		if !hasData {
			continue
		}
		displayedCount++

		sourceTags := ""
		if len(coin.Sources) > 1 {
			sourceTags = " (AI500+OI_Top双重信号)"
		} else if len(coin.Sources) == 1 && coin.Sources[0] == "oi_top" {
			sourceTags = " (OI_Top持仓增长)"
		}

		// 使用FormatMarketData输出完整市场数据
		sb.WriteString(fmt.Sprintf("### %d. %s%s\n\n", displayedCount, coin.Symbol, sourceTags))
		sb.WriteString(market.Format(marketData))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// 优化分析结果
	sb.WriteString("## 🔍 优化分析\n\n")
	
	// 市场状态
	if ctx.MarketRegimeResult != nil {
		sb.WriteString(fmt.Sprintf("**市场状态**: %s (置信度: %.1f%%, 波动率: %.3f, 趋势强度: %.3f)\n",
			ctx.MarketRegimeResult.Regime, ctx.MarketRegimeResult.Confidence*100,
			ctx.MarketRegimeResult.Volatility, ctx.MarketRegimeResult.TrendStrength))
		sb.WriteString("\n")
	}
	
	// 相关性风险
	if ctx.CorrelationReport != nil {
		if highCorrelationPairs, ok := ctx.CorrelationReport["HighCorrelationPairs"].([]risk.CorrelationPair); ok && len(highCorrelationPairs) > 0 {
			sb.WriteString("**相关性风险警告**:\n")
			for _, pair := range highCorrelationPairs {
				sb.WriteString(fmt.Sprintf("- %s ↔ %s: %.2f\n",
					pair.Symbol1, pair.Symbol2, pair.Correlation))
			}
			sb.WriteString("\n")
		}
	}
	
	// 信号强度
	if ctx.SignalStrengthMap != nil && len(ctx.SignalStrengthMap) > 0 {
		sb.WriteString("**信号强度分析**:\n")
		for symbol, strength := range ctx.SignalStrengthMap {
			directionStr := "中性"
			if strength.Direction == 1 {
				directionStr = "看涨"
			} else if strength.Direction == -1 {
				directionStr = "看跌"
			}
			sb.WriteString(fmt.Sprintf("- %s: %.1f分 (%s) | 置信度: %.1f%% | %s\n",
				symbol, strength.OverallScore, directionStr, strength.Confidence*100, strength.Reasoning))
		}
		sb.WriteString("\n")
	}
	
	// SOS状态
	if ctx.SOSStatus != nil {
		if isActive, ok := ctx.SOSStatus["IsActive"].(bool); ok && isActive {
			sb.WriteString("🚨 **紧急状态**: ")
			if status, ok := ctx.SOSStatus["Status"].(string); ok {
				sb.WriteString(status)
			}
			if triggerReason, ok := ctx.SOSStatus["TriggerReason"].(string); ok {
				sb.WriteString(" | 触发原因: " + triggerReason)
			}
			sb.WriteString("\n")
			if recommendedActions, ok := ctx.SOSStatus["RecommendedActions"].([]string); ok && len(recommendedActions) > 0 {
				sb.WriteString("建议行动: ")
				for i, action := range recommendedActions {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(action)
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	// 夏普比率（直接传值，不要复杂格式化）
	if ctx.Performance != nil {
		// 直接从interface{}中提取SharpeRatio
		type PerformanceData struct {
			SharpeRatio float64 `json:"sharpe_ratio"`
		}
		var perfData PerformanceData
		if jsonData, err := json.Marshal(ctx.Performance); err == nil {
			if err := json.Unmarshal(jsonData, &perfData); err == nil {
				sb.WriteString(fmt.Sprintf("## 📊 夏普比率: %.2f\n\n", perfData.SharpeRatio))
			}
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("现在请分析并输出决策（思维链 + JSON）\n")

	return sb.String()
}

// parseFullDecisionResponse 解析AI的完整决策响应
func parseFullDecisionResponse(aiResponse string, accountEquity float64, btcEthLeverage, altcoinLeverage int) (*FullDecision, error) {
	// 1. 提取思维链
	cotTrace := extractCoTTrace(aiResponse)

	// 2. 提取JSON决策列表
	decisions, err := extractDecisions(aiResponse)
	if err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: []Decision{},
		}, fmt.Errorf("提取决策失败: %w\n\n=== AI思维链分析 ===\n%s", err, cotTrace)
	}

	// 3. 验证决策
	if err := validateDecisions(decisions, accountEquity, btcEthLeverage, altcoinLeverage); err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: decisions,
		}, fmt.Errorf("决策验证失败: %w\n\n=== AI思维链分析 ===\n%s", err, cotTrace)
	}

	return &FullDecision{
		CoTTrace:  cotTrace,
		Decisions: decisions,
	}, nil
}

// extractCoTTrace 提取思维链分析
func extractCoTTrace(response string) string {
	// 查找JSON数组的开始位置
	jsonStart := strings.Index(response, "[")

	if jsonStart > 0 {
		// 思维链是JSON数组之前的内容
		return strings.TrimSpace(response[:jsonStart])
	}

	// 如果找不到JSON，整个响应都是思维链
	return strings.TrimSpace(response)
}

// extractDecisions 提取JSON决策列表
func extractDecisions(response string) ([]Decision, error) {
	// 直接查找JSON数组 - 找第一个完整的JSON数组
	arrayStart := strings.Index(response, "[")
	if arrayStart == -1 {
		return nil, fmt.Errorf("无法找到JSON数组起始")
	}

	// 从 [ 开始，匹配括号找到对应的 ]
	arrayEnd := findMatchingBracket(response, arrayStart)
	if arrayEnd == -1 {
		return nil, fmt.Errorf("无法找到JSON数组结束")
	}

	jsonContent := strings.TrimSpace(response[arrayStart : arrayEnd+1])

	// 🔧 修复常见的JSON格式错误：缺少引号的字段值
	// 匹配: "reasoning": 内容"}  或  "reasoning": 内容}  (没有引号)
	// 修复为: "reasoning": "内容"}
	// 使用简单的字符串扫描而不是正则表达式
	jsonContent = fixMissingQuotes(jsonContent)

	// 解析JSON
	var decisions []Decision
	if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w\nJSON内容: %s", err, jsonContent)
	}

	return decisions, nil
}

// fixMissingQuotes 替换中文引号为英文引号（避免输入法自动转换）
func fixMissingQuotes(jsonStr string) string {
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\"") // "
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\"") // "
	jsonStr = strings.ReplaceAll(jsonStr, "\u2018", "'")  // '
	jsonStr = strings.ReplaceAll(jsonStr, "\u2019", "'")  // '
	return jsonStr
}

// validateDecisions 验证所有决策（需要账户信息和杠杆配置）
func validateDecisions(decisions []Decision, accountEquity float64, btcEthLeverage, altcoinLeverage int) error {
	for i, decision := range decisions {
		if err := validateDecision(&decision, accountEquity, btcEthLeverage, altcoinLeverage); err != nil {
			return fmt.Errorf("决策 #%d 验证失败: %w", i+1, err)
		}
	}
	return nil
}

// findMatchingBracket 查找匹配的右括号
func findMatchingBracket(s string, start int) int {
	if start >= len(s) || s[start] != '[' {
		return -1
	}

	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// validateDecision 验证单个决策的有效性
func validateDecision(d *Decision, accountEquity float64, btcEthLeverage, altcoinLeverage int) error {
	// 验证action
	validActions := map[string]bool{
		"open_long":   true,
		"open_short":  true,
		"close_long":  true,
		"close_short": true,
		"hold":        true,
		"wait":        true,
	}

	if !validActions[d.Action] {
		return fmt.Errorf("无效的action: %s", d.Action)
	}

	// 开仓操作必须提供完整参数
	if d.Action == "open_long" || d.Action == "open_short" {
		// 根据币种使用配置的杠杆上限
		maxLeverage := altcoinLeverage          // 山寨币使用配置的杠杆
		maxPositionValue := accountEquity * 1.5 // 山寨币最多1.5倍账户净值
		if d.Symbol == "BTCUSDT" || d.Symbol == "ETHUSDT" {
			maxLeverage = btcEthLeverage          // BTC和ETH使用配置的杠杆
			maxPositionValue = accountEquity * 10 // BTC/ETH最多10倍账户净值
		}

		if d.Leverage <= 0 || d.Leverage > maxLeverage {
			return fmt.Errorf("杠杆必须在1-%d之间（%s，当前配置上限%d倍）: %d", maxLeverage, d.Symbol, maxLeverage, d.Leverage)
		}
		if d.PositionSizeUSD <= 0 {
			return fmt.Errorf("仓位大小必须大于0: %.2f", d.PositionSizeUSD)
		}
		// 验证仓位价值上限（加1%容差以避免浮点数精度问题）
		tolerance := maxPositionValue * 0.01 // 1%容差
		if d.PositionSizeUSD > maxPositionValue+tolerance {
			if d.Symbol == "BTCUSDT" || d.Symbol == "ETHUSDT" {
				return fmt.Errorf("BTC/ETH单币种仓位价值不能超过%.0f USDT（10倍账户净值），实际: %.0f", maxPositionValue, d.PositionSizeUSD)
			} else {
				return fmt.Errorf("山寨币单币种仓位价值不能超过%.0f USDT（1.5倍账户净值），实际: %.0f", maxPositionValue, d.PositionSizeUSD)
			}
		}
		if d.StopLoss <= 0 || d.TakeProfit <= 0 {
			return fmt.Errorf("止损和止盈必须大于0")
		}

		// 验证止损止盈的合理性
		if d.Action == "open_long" {
			if d.StopLoss >= d.TakeProfit {
				return fmt.Errorf("做多时止损价必须小于止盈价")
			}
		} else {
			if d.StopLoss <= d.TakeProfit {
				return fmt.Errorf("做空时止损价必须大于止盈价")
			}
		}

		// 验证风险回报比（必须≥1:3）
		// 计算入场价（假设当前市价）
		var entryPrice float64
		if d.Action == "open_long" {
			// 做多：入场价在止损和止盈之间
			entryPrice = d.StopLoss + (d.TakeProfit-d.StopLoss)*0.2 // 假设在20%位置入场
		} else {
			// 做空：入场价在止损和止盈之间
			entryPrice = d.StopLoss - (d.StopLoss-d.TakeProfit)*0.2 // 假设在20%位置入场
		}

		var riskPercent, rewardPercent, riskRewardRatio float64
		if d.Action == "open_long" {
			riskPercent = (entryPrice - d.StopLoss) / entryPrice * 100
			rewardPercent = (d.TakeProfit - entryPrice) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		} else {
			riskPercent = (d.StopLoss - entryPrice) / entryPrice * 100
			rewardPercent = (entryPrice - d.TakeProfit) / entryPrice * 100
			if riskPercent > 0 {
				riskRewardRatio = rewardPercent / riskPercent
			}
		}

		// 硬约束：风险回报比必须≥3.0
		if riskRewardRatio < 3.0 {
			return fmt.Errorf("风险回报比过低(%.2f:1)，必须≥3.0:1 [风险:%.2f%% 收益:%.2f%%] [止损:%.2f 止盈:%.2f]",
				riskRewardRatio, riskPercent, rewardPercent, d.StopLoss, d.TakeProfit)
		}
	}

	return nil
}
