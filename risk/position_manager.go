package risk

import (
	"log"
	"math"
)

// PositionManager 渐进式仓位管理器
type PositionManager struct {
	BaseRiskPercent    float64 // 基础风险百分比 (2%)
	MaxRiskPercent     float64 // 最大风险百分比 (5%)
	MinConfidence      int     // 最小信心度阈值
	MaxPositions       int     // 最大持仓数量
	CorrelationLimit   float64 // 相关性限制
}

// NewPositionManager 创建仓位管理器
func NewPositionManager() *PositionManager {
	return &PositionManager{
		BaseRiskPercent:  0.02, // 2%基础风险
		MaxRiskPercent:   0.05, // 5%最大风险
		MinConfidence:    65,   // 最小65分信心度
		MaxPositions:     3,    // 最多3个持仓
		CorrelationLimit: 0.8,  // 80%相关性限制
	}
}

// CalculatePositionSize 根据信心度计算仓位大小
func (pm *PositionManager) CalculatePositionSize(confidence int, accountEquity float64, existingPositions []Position) float64 {
	// 1. 信心度检查
	if confidence < pm.MinConfidence {
		log.Printf("📊 仓位管理: 信心度%d低于最小阈值%d，不开仓", confidence, pm.MinConfidence)
		return 0
	}
	
	// 2. 持仓数量检查
	if len(existingPositions) >= pm.MaxPositions {
		log.Printf("📊 仓位管理: 已达最大持仓数量%d，不开仓", pm.MaxPositions)
		return 0
	}
	
	// 3. 基于信心度的仓位计算
	baseSize := accountEquity * pm.BaseRiskPercent
	
	var sizeMultiplier float64
	var description string
	
	switch {
	case confidence >= 85:
		sizeMultiplier = 1.0 // 100%基础仓位
		description = "高信心度"
	case confidence >= 80:
		sizeMultiplier = 0.8 // 80%基础仓位
		description = "较高信心度"
	case confidence >= 75:
		sizeMultiplier = 0.6 // 60%基础仓位
		description = "中等信心度"
	case confidence >= 70:
		sizeMultiplier = 0.4 // 40%基础仓位
		description = "一般信心度"
	case confidence >= 65:
		sizeMultiplier = 0.2 // 20%基础仓位 (试单)
		description = "低信心度试单"
	default:
		sizeMultiplier = 0.0
		description = "信心度不足"
	}
	
	finalSize := baseSize * sizeMultiplier
	
	// 4. 风险限制检查
	maxAllowedSize := accountEquity * pm.MaxRiskPercent
	if finalSize > maxAllowedSize {
		finalSize = maxAllowedSize
		log.Printf("📊 仓位管理: 仓位被限制到最大风险%.1f%%", pm.MaxRiskPercent*100)
	}
	
	log.Printf("📊 仓位管理: %s(信心度%d) -> 仓位大小$%.2f (%.2f%%)", 
		description, confidence, finalSize, (finalSize/accountEquity)*100)
	
	return finalSize
}

// CalculateDynamicRisk 动态风险计算
func (pm *PositionManager) CalculateDynamicRisk(confidence int, marketVolatility float64, portfolioHeat float64) float64 {
	// 基础风险
	baseRisk := pm.BaseRiskPercent
	
	// 信心度调整
	confidenceMultiplier := pm.getConfidenceMultiplier(confidence)
	
	// 市场波动性调整
	volatilityMultiplier := pm.getVolatilityMultiplier(marketVolatility)
	
	// 组合热度调整 (已有持仓的风险程度)
	heatMultiplier := pm.getHeatMultiplier(portfolioHeat)
	
	// 计算最终风险
	finalRisk := baseRisk * confidenceMultiplier * volatilityMultiplier * heatMultiplier
	
	// 确保在合理范围内
	if finalRisk > pm.MaxRiskPercent {
		finalRisk = pm.MaxRiskPercent
	}
	if finalRisk < pm.BaseRiskPercent * 0.1 { // 最小10%基础风险
		finalRisk = pm.BaseRiskPercent * 0.1
	}
	
	log.Printf("📊 动态风险: 基础%.1f%% × 信心%.2f × 波动%.2f × 热度%.2f = %.2f%%",
		baseRisk*100, confidenceMultiplier, volatilityMultiplier, heatMultiplier, finalRisk*100)
	
	return finalRisk
}

// getConfidenceMultiplier 获取信心度乘数
func (pm *PositionManager) getConfidenceMultiplier(confidence int) float64 {
	switch {
	case confidence >= 90:
		return 1.5 // 超高信心度，增加50%风险
	case confidence >= 85:
		return 1.2 // 高信心度，增加20%风险
	case confidence >= 80:
		return 1.0 // 标准风险
	case confidence >= 75:
		return 0.8 // 降低20%风险
	case confidence >= 70:
		return 0.6 // 降低40%风险
	case confidence >= 65:
		return 0.4 // 降低60%风险
	default:
		return 0.2 // 最小风险
	}
}

// getVolatilityMultiplier 获取波动性乘数
func (pm *PositionManager) getVolatilityMultiplier(volatility float64) float64 {
	// volatility 是相对于平均波动性的比率
	switch {
	case volatility > 2.0:
		return 0.5 // 极高波动，减少50%风险
	case volatility > 1.5:
		return 0.7 // 高波动，减少30%风险
	case volatility > 1.2:
		return 0.9 // 较高波动，减少10%风险
	case volatility > 0.8:
		return 1.0 // 正常波动
	case volatility > 0.5:
		return 1.1 // 低波动，可适当增加风险
	default:
		return 0.8 // 极低波动，可能是假突破
	}
}

// getHeatMultiplier 获取组合热度乘数
func (pm *PositionManager) getHeatMultiplier(heat float64) float64 {
	// heat 是当前组合的风险暴露程度 (0-1)
	switch {
	case heat > 0.8:
		return 0.3 // 组合过热，大幅减少新仓位
	case heat > 0.6:
		return 0.5 // 组合较热，减少新仓位
	case heat > 0.4:
		return 0.7 // 组合温热，适当减少
	case heat > 0.2:
		return 1.0 // 正常状态
	default:
		return 1.2 // 组合冷却，可适当增加
	}
}

// CheckRiskLimits 检查风险限制
func (pm *PositionManager) CheckRiskLimits(newPositionRisk float64, existingPositions []Position, accountEquity float64) bool {
	// 1. 计算现有风险
	totalExistingRisk := 0.0
	for _, pos := range existingPositions {
		positionRisk := math.Abs(pos.Size) / accountEquity
		totalExistingRisk += positionRisk
	}
	
	// 2. 计算新的总风险
	newTotalRisk := totalExistingRisk + newPositionRisk
	
	// 3. 检查是否超过最大风险
	maxTotalRisk := pm.MaxRiskPercent * float64(pm.MaxPositions) // 每个仓位最大风险 × 最大仓位数
	
	if newTotalRisk > maxTotalRisk {
		log.Printf("⚠️ 风险限制: 新总风险%.2f%%超过限制%.2f%%", 
			newTotalRisk*100, maxTotalRisk*100)
		return false
	}
	
	log.Printf("✅ 风险检查: 现有风险%.2f%% + 新仓位%.2f%% = 总风险%.2f%% (限制%.2f%%)",
		totalExistingRisk*100, newPositionRisk*100, newTotalRisk*100, maxTotalRisk*100)
	
	return true
}

// CalculatePortfolioHeat 计算组合热度
func (pm *PositionManager) CalculatePortfolioHeat(positions []Position, accountEquity float64) float64 {
	if len(positions) == 0 {
		return 0.0
	}
	
	totalRisk := 0.0
	for _, pos := range positions {
		positionRisk := math.Abs(pos.Size) / accountEquity
		totalRisk += positionRisk
	}
	
	// 热度 = 总风险 / 最大允许风险
	maxAllowedRisk := pm.MaxRiskPercent * float64(pm.MaxPositions)
	heat := totalRisk / maxAllowedRisk
	
	if heat > 1.0 {
		heat = 1.0 // 最大热度为1.0
	}
	
	return heat
}

// ShouldReducePosition 是否应该减仓
func (pm *PositionManager) ShouldReducePosition(portfolioHeat float64, recentPnL float64) bool {
	// 1. 组合过热
	if portfolioHeat > 0.8 {
		log.Printf("📊 仓位建议: 组合热度%.2f过高，建议减仓", portfolioHeat)
		return true
	}
	
	// 2. 近期亏损过多
	if recentPnL < -0.05 { // 近期亏损超过5%
		log.Printf("📊 仓位建议: 近期亏损%.2f%%过多，建议减仓", recentPnL*100)
		return true
	}
	
	return false
}

// GetPositionSizeRecommendation 获取仓位大小建议
func (pm *PositionManager) GetPositionSizeRecommendation(confidence int, accountEquity float64, existingPositions []Position, marketVolatility float64) (float64, string) {
	// 计算组合热度
	portfolioHeat := pm.CalculatePortfolioHeat(existingPositions, accountEquity)
	
	// 动态风险计算
	dynamicRisk := pm.CalculateDynamicRisk(confidence, marketVolatility, portfolioHeat)
	
	// 计算建议仓位大小
	recommendedSize := accountEquity * dynamicRisk
	
	// 生成建议说明
	var recommendation string
	if confidence >= 85 && portfolioHeat < 0.3 && marketVolatility < 1.5 {
		recommendation = "🟢 优质机会，建议标准仓位"
	} else if confidence >= 75 && portfolioHeat < 0.5 {
		recommendation = "🟡 一般机会，建议适中仓位"
	} else if confidence >= 65 {
		recommendation = "🟠 试探机会，建议小仓位"
	} else {
		recommendation = "🔴 信号不足，建议观望"
		recommendedSize = 0
	}
	
	return recommendedSize, recommendation
}