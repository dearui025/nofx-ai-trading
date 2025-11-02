package risk

import (
	"log"
	"math"
	"nofx/market"
	"time"
)

// MarketEnvironmentFilter 市场环境过滤器
type MarketEnvironmentFilter struct {
	MinATRRatio          float64 // 最小ATR比率 (相对于平均ATR)
	MaxCorrelation       float64 // 最大相关性阈值
	LowLiquidityHours    []int   // 低流动性时段 (小时)
	NewsEventBuffer      int     // 新闻事件缓冲时间 (分钟)
	ChoppyMarketThreshold float64 // 震荡市场阈值
}

// NewMarketEnvironmentFilter 创建市场环境过滤器
func NewMarketEnvironmentFilter() *MarketEnvironmentFilter {
	return &MarketEnvironmentFilter{
		MinATRRatio:          0.6,  // ATR低于平均值60%时过滤
		MaxCorrelation:       0.8,  // 相关性超过80%时过滤
		LowLiquidityHours:    []int{0, 1, 2, 3, 4, 5, 22, 23}, // 低流动性时段
		NewsEventBuffer:      30,   // 新闻事件前后30分钟
		ChoppyMarketThreshold: 0.3, // 震荡市场阈值
	}
}

// IsFavorableMarket 检查是否为有利的市场环境
func (mef *MarketEnvironmentFilter) IsFavorableMarket(data *market.Data) bool {
	// 1. 过滤低波动期
	if !mef.isVolatilityAdequate(data) {
		log.Printf("⚠️ 市场环境过滤: 波动性不足")
		return false
	}
	
	// 2. 过滤流动性不足时段
	if !mef.isLiquidityAdequate() {
		log.Printf("⚠️ 市场环境过滤: 流动性不足时段")
		return false
	}
	
	// 3. 过滤震荡市场
	if !mef.isTrendClear(data) {
		log.Printf("⚠️ 市场环境过滤: 趋势不明朗")
		return false
	}
	
	return true
}

// isVolatilityAdequate 检查波动性是否充足
func (mef *MarketEnvironmentFilter) isVolatilityAdequate(data *market.Data) bool {
	if data.LongerTermContext == nil || data.LongerTermContext.ATR14 <= 0 {
		return true // 无数据时不过滤
	}
	
	// 如果有ATR3数据，可以比较短期和长期ATR
	if data.LongerTermContext.ATR3 > 0 {
		atrRatio := data.LongerTermContext.ATR3 / data.LongerTermContext.ATR14
		// ATR3相对ATR14的比率应该在合理范围内
		if atrRatio < mef.MinATRRatio {
			return false
		}
	}
	
	return true
}

// isLiquidityAdequate 检查流动性是否充足
func (mef *MarketEnvironmentFilter) isLiquidityAdequate() bool {
	now := time.Now().UTC()
	currentHour := now.Hour()
	
	// 检查是否在低流动性时段
	for _, hour := range mef.LowLiquidityHours {
		if currentHour == hour {
			return false
		}
	}
	
	return true
}

// isTrendClear 检查趋势是否明朗
func (mef *MarketEnvironmentFilter) isTrendClear(data *market.Data) bool {
	if data.IntradaySeries == nil || len(data.IntradaySeries.MidPrices) < 20 {
		return true // 数据不足时不过滤
	}
	
	prices := data.IntradaySeries.MidPrices
	n := len(prices)
	
	// 计算价格变化的标准差
	if n < 20 {
		return true
	}
	
	recent := prices[n-20:]
	mean := 0.0
	for _, price := range recent {
		mean += price
	}
	mean /= float64(len(recent))
	
	variance := 0.0
	for _, price := range recent {
		variance += math.Pow(price-mean, 2)
	}
	variance /= float64(len(recent))
	stdDev := math.Sqrt(variance)
	
	// 计算相对标准差 (变异系数)
	cv := stdDev / mean
	
	// 如果变异系数过小，说明市场过于平静
	// 如果变异系数过大，说明市场过于震荡
	if cv < 0.001 || cv > mef.ChoppyMarketThreshold {
		return false
	}
	
	return true
}

// CheckCorrelationRisk 检查相关性风险
func (mef *MarketEnvironmentFilter) CheckCorrelationRisk(newSymbol string, existingPositions []Position) bool {
	for _, pos := range existingPositions {
		correlation := mef.getCorrelation(newSymbol, pos.Symbol)
		if correlation > mef.MaxCorrelation {
			log.Printf("⚠️ 相关性风险: %s 与 %s 相关性 %.2f", newSymbol, pos.Symbol, correlation)
			return false
		}
	}
	return true
}

// Position 持仓信息 (简化版本)
type Position struct {
	Symbol string
	Side   string
	Size   float64
}

// getCorrelation 获取两个交易对的相关性 (简化实现)
func (mef *MarketEnvironmentFilter) getCorrelation(symbol1, symbol2 string) float64 {
	// 简化的相关性计算
	// 实际应该基于历史价格数据计算皮尔逊相关系数
	
	// 基于交易对名称的简单规则
	if symbol1 == symbol2 {
		return 1.0
	}
	
	// 同一基础货币的相关性较高
	if mef.hasSameBaseCurrency(symbol1, symbol2) {
		return 0.7
	}
	
	// 主流币种间的相关性
	if mef.isMajorPair(symbol1) && mef.isMajorPair(symbol2) {
		return 0.5
	}
	
	// 默认低相关性
	return 0.2
}

// hasSameBaseCurrency 检查是否有相同的基础货币
func (mef *MarketEnvironmentFilter) hasSameBaseCurrency(symbol1, symbol2 string) bool {
	// 简化实现：检查前3个字符
	if len(symbol1) >= 3 && len(symbol2) >= 3 {
		return symbol1[:3] == symbol2[:3]
	}
	return false
}

// isMajorPair 检查是否为主流交易对
func (mef *MarketEnvironmentFilter) isMajorPair(symbol string) bool {
	majorPairs := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "XRPUSDT", "SOLUSDT", "DOTUSDT", "LINKUSDT"}
	for _, pair := range majorPairs {
		if symbol == pair {
			return true
		}
	}
	return false
}

// IsEconomicDataTime 检查是否为重要经济数据发布时间
func (mef *MarketEnvironmentFilter) IsEconomicDataTime() bool {
	now := time.Now().UTC()
	
	// 简化实现：避免在特定时间交易
	// 实际应该集成经济日历API
	
	// 避免在美国市场开盘前后
	if now.Hour() == 13 && now.Minute() < 30 { // UTC 13:30 = 美东 8:30/9:30
		return true
	}
	
	// 避免在欧洲市场开盘前后
	if now.Hour() == 7 && now.Minute() < 30 { // UTC 7:30 = 欧洲时间 8:30/9:30
		return true
	}
	
	return false
}