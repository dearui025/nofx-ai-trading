// market/regime.go
package market

import (
	"math"
	"time"
)

// MarketRegime 市场状态枚举
type MarketRegime int

const (
	TrendingBull   MarketRegime = iota // 牛市趋势
	TrendingBear                       // 熊市趋势
	Sideways                           // 横盘震荡
	HighVolatility                     // 高波动
	LowVolatility                      // 低波动
	Uncertain                          // 不确定状态
)

// String 返回市场状态的字符串表示
func (mr MarketRegime) String() string {
	switch mr {
	case TrendingBull:
		return "TrendingBull"
	case TrendingBear:
		return "TrendingBear"
	case Sideways:
		return "Sideways"
	case HighVolatility:
		return "HighVolatility"
	case LowVolatility:
		return "LowVolatility"
	case Uncertain:
		return "Uncertain"
	default:
		return "Unknown"
	}
}

// RegimeDetector 市场状态检测器
type RegimeDetector struct {
	VolatilityWindow int // 波动率计算窗口
	TrendWindow      int // 趋势强度计算窗口
	VolatilityThresholds struct {
		High float64 // 高波动阈值
		Low  float64 // 低波动阈值
	}
	TrendThresholds struct {
		Strong float64 // 强趋势阈值
		Weak   float64 // 弱趋势阈值
	}
}

// RegimeAnalysis 市场状态分析结果
type RegimeAnalysis struct {
	Regime         MarketRegime `json:"regime"`
	Volatility     float64      `json:"volatility"`
	TrendStrength  float64      `json:"trend_strength"`
	TrendDirection int          `json:"trend_direction"` // 1: 上涨, -1: 下跌, 0: 横盘
	Confidence     float64      `json:"confidence"`
	LastUpdated    time.Time    `json:"last_updated"`

	// 策略权重调整建议
	StrategyWeights struct {
		TrendFollowing float64 `json:"trend_following"`
		MeanReversion  float64 `json:"mean_reversion"`
		Breakout       float64 `json:"breakout"`
		Conservative   float64 `json:"conservative"`
	} `json:"strategy_weights"`
}

// NewRegimeDetector 创建新的市场状态检测器
func NewRegimeDetector() *RegimeDetector {
	rd := &RegimeDetector{
		VolatilityWindow: 20,
		TrendWindow:      20,
	}
	rd.VolatilityThresholds.High = 0.8
	rd.VolatilityThresholds.Low = 0.2
	rd.TrendThresholds.Strong = 0.05
	rd.TrendThresholds.Weak = 0.01
	return rd
}

// DetectMarketRegime 检测市场状态
func (rd *RegimeDetector) DetectMarketRegime(priceData []float64) *RegimeAnalysis {
	if len(priceData) < rd.TrendWindow {
		return &RegimeAnalysis{
			Regime:      Uncertain,
			Confidence:  0.0,
			LastUpdated: time.Now(),
		}
	}

	// 1. 计算波动率 (使用标准差)
	volatility := rd.calculateVolatility(priceData)

	// 2. 计算趋势强度 (使用线性回归斜率)
	trendStrength, trendDirection := rd.calculateTrendStrength(priceData)

	// 3. 确定市场状态
	regime := rd.classifyRegime(volatility, trendStrength, trendDirection)

	// 4. 计算置信度
	confidence := rd.calculateConfidence(volatility, trendStrength)

	// 5. 生成策略权重建议
	weights := rd.generateStrategyWeights(regime, volatility, trendStrength)

	return &RegimeAnalysis{
		Regime:          regime,
		Volatility:      volatility,
		TrendStrength:   trendStrength,
		TrendDirection:  trendDirection,
		Confidence:      confidence,
		LastUpdated:     time.Now(),
		StrategyWeights: weights,
	}
}

// calculateVolatility 计算价格波动率
func (rd *RegimeDetector) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}

	// 计算收益率序列
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	// 计算标准差
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance) * math.Sqrt(365*24*60/3) // 年化波动率
}

// calculateTrendStrength 计算趋势强度
func (rd *RegimeDetector) calculateTrendStrength(prices []float64) (float64, int) {
	n := float64(len(prices))
	if n < 2 {
		return 0.0, 0
	}

	// 线性回归计算斜率
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, price := range prices {
		x := float64(i)
		y := price
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 斜率 = (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0.0, 0
	}
	slope := (n*sumXY - sumX*sumY) / denominator

	// 标准化斜率 (相对于平均价格)
	avgPrice := sumY / n
	normalizedSlope := slope / avgPrice

	// 趋势强度 = |标准化斜率|
	trendStrength := math.Abs(normalizedSlope)

	// 趋势方向
	direction := 0
	if slope > 0 {
		direction = 1
	} else if slope < 0 {
		direction = -1
	}

	return trendStrength, direction
}

// classifyRegime 分类市场状态
func (rd *RegimeDetector) classifyRegime(volatility, trendStrength float64, trendDirection int) MarketRegime {
	// 高波动优先判断
	if volatility > rd.VolatilityThresholds.High {
		return HighVolatility
	}

	// 低波动判断
	if volatility < rd.VolatilityThresholds.Low {
		return LowVolatility
	}

	// 趋势强度判断
	if trendStrength > rd.TrendThresholds.Strong {
		if trendDirection > 0 {
			return TrendingBull
		} else if trendDirection < 0 {
			return TrendingBear
		}
	}

	if trendStrength < rd.TrendThresholds.Weak {
		return Sideways
	}

	return Uncertain
}

// calculateConfidence 计算置信度
func (rd *RegimeDetector) calculateConfidence(volatility, trendStrength float64) float64 {
	// 基于波动率和趋势强度的置信度计算
	volConfidence := 1.0
	if volatility > rd.VolatilityThresholds.High || volatility < rd.VolatilityThresholds.Low {
		volConfidence = 0.9
	} else {
		volConfidence = 0.6
	}

	trendConfidence := math.Min(trendStrength*10, 1.0) // 趋势强度越大置信度越高

	return (volConfidence + trendConfidence) / 2.0
}

// generateStrategyWeights 生成策略权重建议
func (rd *RegimeDetector) generateStrategyWeights(regime MarketRegime, volatility, trendStrength float64) struct {
	TrendFollowing float64 `json:"trend_following"`
	MeanReversion  float64 `json:"mean_reversion"`
	Breakout       float64 `json:"breakout"`
	Conservative   float64 `json:"conservative"`
} {
	weights := struct {
		TrendFollowing float64 `json:"trend_following"`
		MeanReversion  float64 `json:"mean_reversion"`
		Breakout       float64 `json:"breakout"`
		Conservative   float64 `json:"conservative"`
	}{}

	switch regime {
	case TrendingBull, TrendingBear:
		weights.TrendFollowing = 0.6
		weights.MeanReversion = 0.1
		weights.Breakout = 0.2
		weights.Conservative = 0.1
	case Sideways:
		weights.TrendFollowing = 0.1
		weights.MeanReversion = 0.6
		weights.Breakout = 0.2
		weights.Conservative = 0.1
	case HighVolatility:
		weights.TrendFollowing = 0.2
		weights.MeanReversion = 0.1
		weights.Breakout = 0.3
		weights.Conservative = 0.4
	case LowVolatility:
		weights.TrendFollowing = 0.3
		weights.MeanReversion = 0.4
		weights.Breakout = 0.2
		weights.Conservative = 0.1
	default:
		// 不确定状态，保守策略
		weights.TrendFollowing = 0.2
		weights.MeanReversion = 0.2
		weights.Breakout = 0.1
		weights.Conservative = 0.5
	}

	return weights
}

// ExtractPriceSequence 从市场数据中提取价格序列
func ExtractPriceSequence(data *Data) []float64 {
	if data == nil || data.IntradaySeries == nil {
		return nil
	}
	return data.IntradaySeries.MidPrices
}