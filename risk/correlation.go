// risk/correlation.go
package risk

import (
	"fmt"
	"math"
	"time"
	"nofx/market"
)

// CorrelationMatrix 相关性矩阵
type CorrelationMatrix struct {
	Symbols     []string    `json:"symbols"`
	Matrix      [][]float64 `json:"matrix"`
	LastUpdated time.Time   `json:"last_updated"`
}

// CorrelationRiskManager 相关性风险管理器
type CorrelationRiskManager struct {
	MaxCorrelation    float64                // 最大允许相关性
	LookbackPeriod    int                    // 相关性计算回看期
	UpdateInterval    time.Duration          // 更新间隔
	correlationCache  *CorrelationMatrix     // 相关性矩阵缓存
	lastUpdate        time.Time              // 最后更新时间
}

// NewCorrelationRiskManager 创建相关性风险管理器
func NewCorrelationRiskManager() *CorrelationRiskManager {
	return &CorrelationRiskManager{
		MaxCorrelation: 0.8,                  // 默认最大相关性80%
		LookbackPeriod: 25,                   // 调整为25个数据点，匹配市场数据量
		UpdateInterval: 30 * time.Minute,     // 30分钟更新一次
	}
}

// CheckCorrelationRisk 检查新开仓的相关性风险
func (crm *CorrelationRiskManager) CheckCorrelationRisk(
	existingPositions []string,
	newSymbol string,
	marketDataMap map[string]*market.Data,
) error {
	if len(existingPositions) == 0 {
		return nil // 没有现有持仓，无相关性风险
	}

	// 更新相关性矩阵（如果需要）
	if err := crm.updateCorrelationMatrix(marketDataMap); err != nil {
		return fmt.Errorf("更新相关性矩阵失败: %w", err)
	}

	// 检查与每个现有持仓的相关性
	for _, existingSymbol := range existingPositions {
		correlation := crm.getCorrelation(existingSymbol, newSymbol)
		if math.Abs(correlation) > crm.MaxCorrelation {
			return fmt.Errorf("相关性风险过高: %s 与 %s 的相关性为 %.3f (阈值: %.3f)",
				existingSymbol, newSymbol, correlation, crm.MaxCorrelation)
		}
	}

	return nil
}

// updateCorrelationMatrix 更新相关性矩阵
func (crm *CorrelationRiskManager) updateCorrelationMatrix(marketDataMap map[string]*market.Data) error {
	now := time.Now()
	if crm.correlationCache != nil && now.Sub(crm.lastUpdate) < crm.UpdateInterval {
		return nil // 缓存仍然有效
	}

	// 提取所有币种的价格序列
	symbols := make([]string, 0, len(marketDataMap))
	priceMatrix := make([][]float64, 0, len(marketDataMap))

	for symbol, data := range marketDataMap {
		if data.IntradaySeries != nil && len(data.IntradaySeries.MidPrices) >= crm.LookbackPeriod {
			symbols = append(symbols, symbol)
			// 取最近的价格数据
			recentPrices := data.IntradaySeries.MidPrices[len(data.IntradaySeries.MidPrices)-crm.LookbackPeriod:]
			priceMatrix = append(priceMatrix, recentPrices)
		}
	}

	if len(symbols) < 2 {
		return fmt.Errorf("数据不足，无法计算相关性矩阵")
	}

	// 计算相关性矩阵
	n := len(symbols)
	correlationMatrix := make([][]float64, n)
	for i := range correlationMatrix {
		correlationMatrix[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				correlationMatrix[i][j] = 1.0
			} else {
				correlationMatrix[i][j] = crm.calculateCorrelation(priceMatrix[i], priceMatrix[j])
			}
		}
	}

	// 更新缓存
	crm.correlationCache = &CorrelationMatrix{
		Symbols:     symbols,
		Matrix:      correlationMatrix,
		LastUpdated: now,
	}
	crm.lastUpdate = now

	return nil
}

// calculateCorrelation 计算两个价格序列的皮尔逊相关系数
func (crm *CorrelationRiskManager) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0.0
	}

	n := float64(len(x))

	// 计算均值
	meanX, meanY := 0.0, 0.0
	for i := 0; i < len(x); i++ {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= n
	meanY /= n

	// 计算协方差和方差
	covariance := 0.0
	varianceX := 0.0
	varianceY := 0.0

	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		covariance += dx * dy
		varianceX += dx * dx
		varianceY += dy * dy
	}

	// 计算相关系数
	if varianceX == 0 || varianceY == 0 {
		return 0.0
	}

	return covariance / math.Sqrt(varianceX*varianceY)
}

// getCorrelation 获取两个币种的相关性
func (crm *CorrelationRiskManager) getCorrelation(symbol1, symbol2 string) float64 {
	if crm.correlationCache == nil {
		return 0.0
	}

	// 查找币种索引
	index1, index2 := -1, -1
	for i, symbol := range crm.correlationCache.Symbols {
		if symbol == symbol1 {
			index1 = i
		}
		if symbol == symbol2 {
			index2 = i
		}
	}

	if index1 == -1 || index2 == -1 {
		return 0.0
	}

	return crm.correlationCache.Matrix[index1][index2]
}

// GetCorrelationReport 获取相关性报告
func (crm *CorrelationRiskManager) GetCorrelationReport() *CorrelationMatrix {
	return crm.correlationCache
}

// GetHighCorrelationPairs 获取高相关性币种对
func (crm *CorrelationRiskManager) GetHighCorrelationPairs(threshold float64) []CorrelationPair {
	if crm.correlationCache == nil {
		return nil
	}

	var pairs []CorrelationPair
	symbols := crm.correlationCache.Symbols
	matrix := crm.correlationCache.Matrix

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			correlation := matrix[i][j]
			if math.Abs(correlation) > threshold {
				pairs = append(pairs, CorrelationPair{
					Symbol1:     symbols[i],
					Symbol2:     symbols[j],
					Correlation: correlation,
				})
			}
		}
	}

	return pairs
}

// CorrelationPair 相关性币种对
type CorrelationPair struct {
	Symbol1     string  `json:"symbol1"`
	Symbol2     string  `json:"symbol2"`
	Correlation float64 `json:"correlation"`
}

// ValidatePortfolioCorrelation 验证投资组合相关性
func (crm *CorrelationRiskManager) ValidatePortfolioCorrelation(positions []string) (bool, []string) {
	if len(positions) <= 1 {
		return true, nil // 单个或无持仓，无相关性风险
	}

	var warnings []string
	highCorrelationPairs := crm.GetHighCorrelationPairs(crm.MaxCorrelation)

	// 检查当前持仓中是否存在高相关性对
	for _, pair := range highCorrelationPairs {
		hasSymbol1 := false
		hasSymbol2 := false

		for _, pos := range positions {
			if pos == pair.Symbol1 {
				hasSymbol1 = true
			}
			if pos == pair.Symbol2 {
				hasSymbol2 = true
			}
		}

		if hasSymbol1 && hasSymbol2 {
			warning := fmt.Sprintf("持仓中存在高相关性风险: %s 与 %s 相关性为 %.3f",
				pair.Symbol1, pair.Symbol2, pair.Correlation)
			warnings = append(warnings, warning)
		}
	}

	return len(warnings) == 0, warnings
}

// SetMaxCorrelation 设置最大允许相关性
func (crm *CorrelationRiskManager) SetMaxCorrelation(maxCorr float64) {
	if maxCorr > 0 && maxCorr <= 1.0 {
		crm.MaxCorrelation = maxCorr
	}
}

// SetLookbackPeriod 设置回看期
func (crm *CorrelationRiskManager) SetLookbackPeriod(period int) {
	if period > 0 {
		crm.LookbackPeriod = period
	}
}

// SetUpdateInterval 设置更新间隔
func (crm *CorrelationRiskManager) SetUpdateInterval(interval time.Duration) {
	if interval > 0 {
		crm.UpdateInterval = interval
	}
}