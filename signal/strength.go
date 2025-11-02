// signal/strength.go
package signal

import (
	"fmt"
	"math"
	"nofx/market"
)

// SignalStrengthCalculator 信号强度计算器
type SignalStrengthCalculator struct {
	Weights struct {
		TrendAnalysis       float64 // 趋势分析权重 (40分)
		MomentumAnalysis    float64 // 动量分析权重 (30分)
		MarketStructure     float64 // 市场结构权重 (30分)
		// 保留原有权重用于兼容性
		PriceAction         float64 // 价格行为权重
		VolumeConfirmation  float64 // 成交量确认权重
		IndicatorAlignment  float64 // 指标一致性权重
		TimeframeConfluence float64 // 多时间框架共振权重
	}
}

// SignalStrengthResult 信号强度结果
type SignalStrengthResult struct {
	OverallScore     float64 `json:"overall_score"`     // 总体评分 (0-100)
	// 新的三维度评分
	TrendScore       float64 `json:"trend_score"`       // 趋势维度评分 (0-40)
	MomentumScore    float64 `json:"momentum_score"`    // 动量维度评分 (0-30)
	StructureScore   float64 `json:"structure_score"`   // 市场结构评分 (0-30)
	// 保留原有评分用于兼容性
	PriceActionScore float64 `json:"price_action_score"` // 价格行为评分
	VolumeScore      float64 `json:"volume_score"`      // 成交量评分
	IndicatorScore   float64 `json:"indicator_score"`   // 指标评分
	TimeframeScore   float64 `json:"timeframe_score"`   // 时间框架评分
	Confidence       float64 `json:"confidence"`        // 置信度
	Direction        int     `json:"direction"`         // 方向 (1: 看涨, -1: 看跌, 0: 中性)
	Reasoning        string  `json:"reasoning"`         // 评分理由
}

// NewSignalStrengthCalculator 创建信号强度计算器
func NewSignalStrengthCalculator() *SignalStrengthCalculator {
	ssc := &SignalStrengthCalculator{}
	// 新的三维度权重配置
	ssc.Weights.TrendAnalysis = 0.4    // 趋势分析 40%
	ssc.Weights.MomentumAnalysis = 0.3 // 动量分析 30%
	ssc.Weights.MarketStructure = 0.3  // 市场结构 30%
	
	// 保留原有权重用于兼容性
	ssc.Weights.PriceAction = 0.3
	ssc.Weights.VolumeConfirmation = 0.25
	ssc.Weights.IndicatorAlignment = 0.25
	ssc.Weights.TimeframeConfluence = 0.2
	return ssc
}

// CalculateSignalStrength 计算信号强度
func (ssc *SignalStrengthCalculator) CalculateSignalStrength(data *market.Data) *SignalStrengthResult {
	// === 新的三维度评分系统 ===
	
	// 1. 趋势维度分析 (40分)
	trendScore, trendDirection := ssc.analyzeTrendDimension(data)
	
	// 2. 动量维度分析 (30分)
	momentumScore, momentumDirection := ssc.analyzeMomentumDimension(data)
	
	// 3. 市场结构维度分析 (30分)
	structureScore, structureDirection := ssc.analyzeMarketStructure(data)
	
	// 4. 计算新的总体评分
	newOverallScore := trendScore + momentumScore + structureScore
	
	// 5. 确定最终方向 (基于三维度投票)
	finalDirection := ssc.determineDirectionByVoting(trendDirection, momentumDirection, structureDirection)
	
	// 6. 计算置信度 (基于三维度一致性)
	confidence := ssc.calculateNewConfidence(trendScore, momentumScore, structureScore, trendDirection, momentumDirection, structureDirection)
	
	// 7. 生成新的评分理由
	reasoning := ssc.generateNewReasoning(trendScore, momentumScore, structureScore, finalDirection)
	
	// === 保持兼容性：计算原有评分 ===
	priceActionScore, priceDirection := ssc.analyzePriceAction(data)
	volumeScore := ssc.analyzeVolumeConfirmation(data, priceDirection)
	indicatorScore, indicatorDirection := ssc.analyzeIndicatorAlignment(data)
	timeframeScore := ssc.analyzeTimeframeConfluence(data)
	
	// 原有总体评分
	oldOverallScore := ssc.Weights.PriceAction*priceActionScore +
		ssc.Weights.VolumeConfirmation*volumeScore +
		ssc.Weights.IndicatorAlignment*indicatorScore +
		ssc.Weights.TimeframeConfluence*timeframeScore

	// 移除未使用的变量
	_ = indicatorDirection
	_ = oldOverallScore

	return &SignalStrengthResult{
		OverallScore:     newOverallScore, // 使用新的评分系统
		TrendScore:       trendScore,
		MomentumScore:    momentumScore,
		StructureScore:   structureScore,
		// 保留原有评分
		PriceActionScore: priceActionScore,
		VolumeScore:      volumeScore,
		IndicatorScore:   indicatorScore,
		TimeframeScore:   timeframeScore,
		Confidence:       confidence,
		Direction:        finalDirection,
		Reasoning:        reasoning,
	}
}

// analyzePriceAction 分析价格行为
func (ssc *SignalStrengthCalculator) analyzePriceAction(data *market.Data) (float64, int) {
	score := 0.0
	direction := 0

	if data.IntradaySeries == nil || len(data.IntradaySeries.MidPrices) < 10 {
		return 0.0, 0
	}

	prices := data.IntradaySeries.MidPrices

	// 1. 趋势强度分析
	trendScore := ssc.calculateTrendStrength(prices)

	// 2. 突破分析
	breakoutScore := ssc.analyzeBreakout(prices)

	// 3. 支撑阻力分析
	supportResistanceScore := ssc.analyzeSupportResistance(prices)

	// 4. 价格动量分析
	momentumScore, momentumDirection := ssc.analyzeMomentum(prices)

	// 综合价格行为评分
	score = (trendScore + breakoutScore + supportResistanceScore + momentumScore) / 4.0
	direction = momentumDirection

	return math.Min(score*100, 100), direction
}

// analyzeVolumeConfirmation 分析成交量确认
func (ssc *SignalStrengthCalculator) analyzeVolumeConfirmation(data *market.Data, priceDirection int) float64 {
	// 使用24小时成交量作为参考
	if data.Volume24h <= 0 {
		return 50.0 // 中性评分
	}

	// 简化处理，使用当前成交量与平均值的比较
	currentVolume := data.Volume24h
	avgVolume := currentVolume * 0.8 // 假设平均值

	// 当前成交量相对于平均值的比率
	volumeRatio := currentVolume / avgVolume

	// 成交量确认评分
	score := 50.0 // 基础分

	if priceDirection != 0 {
		// 价格有方向时，成交量应该放大确认
		if volumeRatio > 1.2 {
			score += 30.0 // 成交量放大，确认信号
		} else if volumeRatio < 0.8 {
			score -= 20.0 // 成交量萎缩，信号减弱
		}
	}

	// 简化成交量趋势分析
	volumeTrend := 0.0
	if volumeRatio > 1.0 {
		volumeTrend = 0.5
	} else if volumeRatio < 1.0 {
		volumeTrend = -0.5
	}
	score += volumeTrend * 20.0

	return math.Max(0, math.Min(score, 100))
}

// analyzeIndicatorAlignment 分析指标一致性
func (ssc *SignalStrengthCalculator) analyzeIndicatorAlignment(data *market.Data) (float64, int) {
	score := 0.0
	direction := 0
	indicatorCount := 0

	// RSI分析
	if data.CurrentRSI7 > 0 {
		rsiScore, rsiDir := ssc.analyzeRSI(data.CurrentRSI7)
		score += rsiScore
		direction += rsiDir
		indicatorCount++
	}

	// MACD分析
	if data.CurrentMACD != 0 {
		macdScore, macdDir := ssc.analyzeMACD(data.CurrentMACD)
		score += macdScore
		direction += macdDir
		indicatorCount++
	}

	// EMA分析
	if data.CurrentEMA20 > 0 && data.CurrentPrice > 0 {
		emaScore, emaDir := ssc.analyzeEMA(data.CurrentPrice, data.CurrentEMA20)
		score += emaScore
		direction += emaDir
		indicatorCount++
	}

	if indicatorCount == 0 {
		return 50.0, 0
	}

	// 计算平均分数和方向
	avgScore := score / float64(indicatorCount)
	finalDirection := 0
	if direction > 0 {
		finalDirection = 1
	} else if direction < 0 {
		finalDirection = -1
	}

	return avgScore, finalDirection
}

// analyzeTimeframeConfluence 分析多时间框架共振
func (ssc *SignalStrengthCalculator) analyzeTimeframeConfluence(data *market.Data) float64 {
	score := 50.0 // 基础分

	// 检查不同时间框架的趋势一致性
	shortTermTrend := ssc.getShortTermTrend(data)
	mediumTermTrend := ssc.getMediumTermTrend(data)
	longTermTrend := ssc.getLongTermTrend(data)

	// 计算趋势一致性
	consistency := 0
	if shortTermTrend == mediumTermTrend {
		consistency++
	}
	if mediumTermTrend == longTermTrend {
		consistency++
	}
	if shortTermTrend == longTermTrend {
		consistency++
	}

	// 根据一致性调整评分
	switch consistency {
	case 3:
		score += 40.0 // 完全一致
	case 2:
		score += 20.0 // 部分一致
	case 1:
		score += 10.0 // 轻微一致
	default:
		score -= 20.0 // 不一致
	}

	return math.Max(0, math.Min(score, 100))
}

// 辅助函数实现

func (ssc *SignalStrengthCalculator) calculateTrendStrength(prices []float64) float64 {
	if len(prices) < 5 {
		return 0.0
	}

	// 使用线性回归计算趋势强度
	n := float64(len(prices))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, price := range prices {
		x := float64(i)
		y := price
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0.0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	avgPrice := sumY / n
	normalizedSlope := math.Abs(slope / avgPrice)

	return math.Min(normalizedSlope*1000, 1.0) // 标准化到0-1
}

func (ssc *SignalStrengthCalculator) analyzeBreakout(prices []float64) float64 {
	if len(prices) < 20 {
		return 0.0
	}

	// 计算最近20期的最高价和最低价
	recent := prices[len(prices)-20:]
	high := recent[0]
	low := recent[0]

	for _, price := range recent {
		if price > high {
			high = price
		}
		if price < low {
			low = price
		}
	}

	currentPrice := prices[len(prices)-1]
	range_ := high - low

	// 检查是否突破
	if currentPrice > high-range_*0.05 {
		return 0.8 // 向上突破
	} else if currentPrice < low+range_*0.05 {
		return 0.8 // 向下突破
	}

	return 0.3 // 无明显突破
}

func (ssc *SignalStrengthCalculator) analyzeSupportResistance(prices []float64) float64 {
	if len(prices) < 10 {
		return 0.0
	}

	// 简化的支撑阻力分析
	currentPrice := prices[len(prices)-1]
	recentPrices := prices[len(prices)-10:]

	// 计算价格在区间中的位置
	minPrice := recentPrices[0]
	maxPrice := recentPrices[0]

	for _, price := range recentPrices {
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
	}

	if maxPrice == minPrice {
		return 0.5
	}

	position := (currentPrice - minPrice) / (maxPrice - minPrice)

	// 在支撑位附近给高分，在阻力位附近给低分
	if position < 0.2 || position > 0.8 {
		return 0.7 // 接近关键位置
	}

	return 0.4 // 中间位置
}

func (ssc *SignalStrengthCalculator) analyzeMomentum(prices []float64) (float64, int) {
	if len(prices) < 5 {
		return 0.0, 0
	}

	// 计算价格动量
	recent := prices[len(prices)-5:]
	momentum := (recent[4] - recent[0]) / recent[0]

	direction := 0
	if momentum > 0 {
		direction = 1
	} else if momentum < 0 {
		direction = -1
	}

	score := math.Min(math.Abs(momentum)*100, 1.0)
	return score, direction
}

func (ssc *SignalStrengthCalculator) calculateVolumeTrend(volumes []float64) float64 {
	if len(volumes) < 2 {
		return 0.0
	}

	// 简单的成交量趋势计算
	trend := (volumes[len(volumes)-1] - volumes[0]) / volumes[0]
	return math.Max(-1.0, math.Min(1.0, trend))
}

func (ssc *SignalStrengthCalculator) analyzeRSI(rsi float64) (float64, int) {
	score := 50.0
	direction := 0

	if rsi > 70 {
		score = 30.0 // 超买
		direction = -1
	} else if rsi < 30 {
		score = 70.0 // 超卖
		direction = 1
	} else if rsi > 50 {
		score = 60.0
		direction = 1
	} else {
		score = 40.0
		direction = -1
	}

	return score, direction
}

func (ssc *SignalStrengthCalculator) analyzeMACD(macd float64) (float64, int) {
	score := 50.0
	direction := 0

	if macd > 0 {
		score = 65.0
		direction = 1
	} else {
		score = 35.0
		direction = -1
	}

	return score, direction
}

func (ssc *SignalStrengthCalculator) analyzeEMA(price, ema float64) (float64, int) {
	score := 50.0
	direction := 0

	if price > ema {
		score = 60.0
		direction = 1
	} else {
		score = 40.0
		direction = -1
	}

	return score, direction
}

func (ssc *SignalStrengthCalculator) getShortTermTrend(data *market.Data) int {
	if data.IntradaySeries == nil || len(data.IntradaySeries.MidPrices) < 5 {
		return 0
	}

	prices := data.IntradaySeries.MidPrices
	n := len(prices)
	recent := prices[n-5:]

	if recent[4] > recent[0] {
		return 1
	} else if recent[4] < recent[0] {
		return -1
	}
	return 0
}

func (ssc *SignalStrengthCalculator) getMediumTermTrend(data *market.Data) int {
	if data.IntradaySeries == nil || len(data.IntradaySeries.MidPrices) < 20 {
		return 0
	}

	prices := data.IntradaySeries.MidPrices
	n := len(prices)
	if n < 20 {
		return 0
	}

	if prices[n-1] > prices[n-20] {
		return 1
	} else if prices[n-1] < prices[n-20] {
		return -1
	}
	return 0
}

func (ssc *SignalStrengthCalculator) getLongTermTrend(data *market.Data) int {
	if data.CurrentEMA20 > 0 && data.CurrentPrice > 0 {
		if data.CurrentPrice > data.CurrentEMA20 {
			return 1
		} else if data.CurrentPrice < data.CurrentEMA20 {
			return -1
		}
	}
	return 0
}

func (ssc *SignalStrengthCalculator) determineDirection(priceDirection, indicatorDirection int) int {
	// 综合价格方向和指标方向
	if priceDirection == indicatorDirection && priceDirection != 0 {
		return priceDirection // 方向一致
	}

	// 如果方向不一致，返回中性
	return 0
}

func (ssc *SignalStrengthCalculator) calculateConfidence(priceScore, volumeScore, indicatorScore, timeframeScore float64) float64 {
	// 基于各项评分的方差计算置信度
	scores := []float64{priceScore, volumeScore, indicatorScore, timeframeScore}
	mean := (priceScore + volumeScore + indicatorScore + timeframeScore) / 4.0

	variance := 0.0
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= 4.0

	// 方差越小，置信度越高
	confidence := 1.0 - math.Min(variance/1000.0, 1.0)
	return math.Max(0.1, confidence)
}

func (ssc *SignalStrengthCalculator) generateReasoning(priceScore, volumeScore, indicatorScore, timeframeScore float64, direction int) string {
	reasoning := ""

	// 价格行为分析
	if priceScore > 70 {
		reasoning += "价格行为强劲，"
	} else if priceScore < 30 {
		reasoning += "价格行为疲弱，"
	} else {
		reasoning += "价格行为中性，"
	}

	// 成交量分析
	if volumeScore > 70 {
		reasoning += "成交量放大确认，"
	} else if volumeScore < 30 {
		reasoning += "成交量萎缩，"
	} else {
		reasoning += "成交量正常，"
	}

	// 指标分析
	if indicatorScore > 70 {
		reasoning += "技术指标看涨，"
	} else if indicatorScore < 30 {
		reasoning += "技术指标看跌，"
	} else {
		reasoning += "技术指标中性，"
	}

	// 时间框架分析
	if timeframeScore > 70 {
		reasoning += "多时间框架共振"
	} else if timeframeScore < 30 {
		reasoning += "时间框架分歧"
	} else {
		reasoning += "时间框架部分一致"
	}

	// 方向总结
	switch direction {
	case 1:
		reasoning = fmt.Sprintf("%s，综合看涨", reasoning)
	case -1:
		reasoning = fmt.Sprintf("%s，综合看跌", reasoning)
	default:
		reasoning = fmt.Sprintf("%s，方向不明确", reasoning)
	}

	return reasoning
}

// === 新的三维度分析函数 ===

// analyzeTrendDimension 趋势维度分析 (40分)
func (ssc *SignalStrengthCalculator) analyzeTrendDimension(data *market.Data) (float64, int) {
	score := 0.0
	direction := 0
	
	// 1. 价格相对EMA20位置 (20分)
	if data.LongerTermContext != nil && data.LongerTermContext.EMA20 > 0 {
		if data.CurrentPrice > data.LongerTermContext.EMA20 {
			score += 20.0
			direction = 1
		} else {
			direction = -1
		}
	}
	
	// 2. MACD趋势确认 (20分)
	if data.CurrentMACD != 0 {
		if data.CurrentMACD > 0 {
			score += 20.0
			if direction == 0 {
				direction = 1
			}
		} else if direction == 0 {
			direction = -1
		}
	}
	
	return score, direction
}

// analyzeMomentumDimension 动量维度分析 (30分)
func (ssc *SignalStrengthCalculator) analyzeMomentumDimension(data *market.Data) (float64, int) {
	score := 0.0
	direction := 0
	
	// 1. RSI7合理区间检查 (15分)
	if data.CurrentRSI7 > 0 {
		rsi := data.CurrentRSI7
		if rsi > 30 && rsi < 70 {
			score += 15.0
			// RSI在30-50区间偏向多头
			if rsi < 50 {
				direction = 1
			} else {
				direction = -1
			}
		}
	}
	
	// 2. 成交量放大确认 (15分)
	if data.LongerTermContext != nil && data.LongerTermContext.CurrentVolume > 0 && data.LongerTermContext.AverageVolume > 0 {
		currentVolume := data.LongerTermContext.CurrentVolume
		avgVolume := data.LongerTermContext.AverageVolume
		
		// 成交量放大1.5倍以上
		if currentVolume > avgVolume*1.5 {
			score += 15.0
			if direction == 0 {
				direction = 1
			}
		}
	}
	
	return score, direction
}

// analyzeMarketStructure 市场结构维度分析 (30分)
func (ssc *SignalStrengthCalculator) analyzeMarketStructure(data *market.Data) (float64, int) {
	score := 0.0
	direction := 0
	
	// 1. 关键价位支撑阻力 (15分)
	if data.IntradaySeries != nil && len(data.IntradaySeries.MidPrices) > 0 {
		prices := data.IntradaySeries.MidPrices
		currentPrice := data.CurrentPrice
		
		// 简化的支撑阻力判断
		if len(prices) >= 50 {
			// 计算近期高低点
			recentHigh := 0.0
			recentLow := math.MaxFloat64
			for i := len(prices) - 50; i < len(prices); i++ {
				if prices[i] > recentHigh {
					recentHigh = prices[i]
				}
				if prices[i] < recentLow {
					recentLow = prices[i]
				}
			}
			
			// 判断是否在关键位置
			priceRange := recentHigh - recentLow
			if priceRange > 0 {
				distanceFromHigh := (recentHigh - currentPrice) / priceRange
				distanceFromLow := (currentPrice - recentLow) / priceRange
				
				// 在支撑位附近 (底部20%区域)
				if distanceFromLow < 0.2 {
					score += 15.0
					direction = 1
				}
				// 在阻力位附近 (顶部20%区域)
				if distanceFromHigh < 0.2 {
					score += 15.0
					direction = -1
				}
			}
		}
	}
	
	// 2. 非极端超买超卖状态 (15分)
	if data.CurrentRSI7 > 0 {
		rsi := data.CurrentRSI7
		// 避免极端超买超卖区域
		if rsi > 20 && rsi < 80 {
			score += 15.0
		}
	}
	
	return score, direction
}

// determineDirectionByVoting 基于三维度投票确定方向
func (ssc *SignalStrengthCalculator) determineDirectionByVoting(trendDir, momentumDir, structureDir int) int {
	votes := trendDir + momentumDir + structureDir
	
	if votes > 0 {
		return 1  // 看涨
	} else if votes < 0 {
		return -1 // 看跌
	}
	return 0 // 中性
}

// calculateNewConfidence 基于三维度一致性计算置信度
func (ssc *SignalStrengthCalculator) calculateNewConfidence(trendScore, momentumScore, structureScore float64, trendDir, momentumDir, structureDir int) float64 {
	// 基础置信度：基于总分
	totalScore := trendScore + momentumScore + structureScore
	baseConfidence := totalScore // 0-100分直接作为基础置信度
	
	// 方向一致性加成
	directions := []int{trendDir, momentumDir, structureDir}
	consistency := 0
	for _, dir := range directions {
		if dir != 0 {
			consistency++
		}
	}
	
	// 一致性越高，置信度越高
	consistencyBonus := float64(consistency) * 5.0 // 每个一致方向+5分
	
	confidence := baseConfidence + consistencyBonus
	if confidence > 100 {
		confidence = 100
	}
	
	return confidence
}

// generateNewReasoning 生成新的评分理由
func (ssc *SignalStrengthCalculator) generateNewReasoning(trendScore, momentumScore, structureScore float64, direction int) string {
	reasoning := fmt.Sprintf("三维度评分: 趋势%.0f分, 动量%.0f分, 结构%.0f分", 
		trendScore, momentumScore, structureScore)
	
	directionText := "中性"
	if direction == 1 {
		directionText = "看涨"
	} else if direction == -1 {
		directionText = "看跌"
	}
	
	reasoning += fmt.Sprintf(", 综合方向: %s", directionText)
	
	// 添加具体分析
	if trendScore >= 30 {
		reasoning += ", 趋势强劲"
	} else if trendScore >= 20 {
		reasoning += ", 趋势适中"
	} else {
		reasoning += ", 趋势偏弱"
	}
	
	if momentumScore >= 20 {
		reasoning += ", 动量充足"
	} else if momentumScore >= 15 {
		reasoning += ", 动量一般"
	} else {
		reasoning += ", 动量不足"
	}
	
	if structureScore >= 20 {
		reasoning += ", 结构良好"
	} else if structureScore >= 15 {
		reasoning += ", 结构一般"
	} else {
		reasoning += ", 结构偏弱"
	}
	
	return reasoning
}