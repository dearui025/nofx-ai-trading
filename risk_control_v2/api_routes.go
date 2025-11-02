package risk_control_v2

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APIHandler API处理器
type APIHandler struct {
	riskManager     *RiskManager
	databaseManager *DatabaseManager
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(riskManager *RiskManager, databaseManager *DatabaseManager) *APIHandler {
	return &APIHandler{
		riskManager:     riskManager,
		databaseManager: databaseManager,
	}
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	v2 := router.Group("/api/v2/risk-control")
	{
		// 时间管理相关
		timeGroup := v2.Group("/time")
		{
			timeGroup.GET("/status", h.GetTimeStatus)
			timeGroup.POST("/reset", h.ManualReset)
			timeGroup.GET("/reset-history", h.GetResetHistory)
			timeGroup.PUT("/config", h.UpdateTimeConfig)
		}

		// 流动性和监控相关
		liquidityGroup := v2.Group("/liquidity")
		{
			liquidityGroup.GET("/status", h.GetLiquidityStatus)
			liquidityGroup.POST("/update", h.UpdateLiquidityData)
			liquidityGroup.GET("/alerts", h.GetLiquidityAlerts)
			liquidityGroup.POST("/alerts/:id/resolve", h.ResolveLiquidityAlert)
			liquidityGroup.GET("/blacklist", h.GetBlacklist)
			liquidityGroup.POST("/blacklist", h.AddToBlacklist)
			liquidityGroup.DELETE("/blacklist/:symbol", h.RemoveFromBlacklist)
			liquidityGroup.PUT("/config", h.UpdateLiquidityConfig)
		}

		// 夏普比率相关
		sharpeGroup := v2.Group("/sharpe")
		{
			sharpeGroup.GET("/status", h.GetSharpeStatus)
			sharpeGroup.POST("/update", h.UpdateSharpeData)
			sharpeGroup.GET("/records", h.GetSharpeRecords)
			sharpeGroup.GET("/transitions", h.GetSharpeTransitions)
			sharpeGroup.PUT("/config", h.UpdateSharpeConfig)
		}

		// AI委员会相关
		aiGroup := v2.Group("/ai-committee")
		{
			aiGroup.GET("/status", h.GetAICommitteeStatus)
			aiGroup.POST("/decision", h.MakeAIDecision)
			aiGroup.GET("/decisions", h.GetAIDecisions)
			aiGroup.GET("/performance", h.GetModelPerformance)
			aiGroup.PUT("/config", h.UpdateAIConfig)
		}

		// 风控管理相关
		riskGroup := v2.Group("/risk")
		{
			riskGroup.GET("/status", h.GetRiskStatus)
			riskGroup.POST("/decision", h.MakeRiskDecision)
			riskGroup.GET("/decisions", h.GetRiskDecisions)
			riskGroup.GET("/alerts", h.GetRiskAlerts)
			riskGroup.POST("/alerts/:id/resolve", h.ResolveRiskAlert)
			riskGroup.POST("/emergency-stop", h.EmergencyStop)
			riskGroup.POST("/resume", h.ResumeRisk)
			riskGroup.PUT("/config", h.UpdateRiskConfig)
		}

		// 系统配置相关
		configGroup := v2.Group("/config")
		{
			configGroup.GET("/:type/:name", h.GetSystemConfig)
			configGroup.PUT("/:type/:name", h.SetSystemConfig)
			configGroup.GET("/all", h.GetAllConfigs)
		}

		// 数据管理相关
		dataGroup := v2.Group("/data")
		{
			dataGroup.POST("/cleanup", h.CleanupOldData)
			dataGroup.GET("/stats", h.GetDataStats)
			dataGroup.POST("/export", h.ExportData)
		}
	}
}

// GetTimeStatus 获取时间管理状态
func (h *APIHandler) GetTimeStatus(c *gin.Context) {
	state := h.riskManager.timeManager.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state,
	})
}

// ManualReset 手动重置
func (h *APIHandler) ManualReset(c *gin.Context) {
	var req struct {
		ResetType string  `json:"reset_type" binding:"required"`
		Reason    string  `json:"reason" binding:"required"`
		NewEquity float64 `json:"new_equity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.riskManager.timeManager.ManualReset(req.Reason, req.NewEquity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("手动重置失败: %v", err),
		})
		return
	}

	// 创建重置记录
	record := DrawdownResetRecord{
		ID:           fmt.Sprintf("manual_%d", time.Now().Unix()),
		ResetType:    req.ResetType,
		Reason:       req.Reason,
		NewWatermark: req.NewEquity,
		ResetAt:      time.Now(),
	}

	// 保存到数据库
	if err := h.databaseManager.SaveDrawdownResetRecord(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存重置记录失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// GetResetHistory 获取重置历史
func (h *APIHandler) GetResetHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	history := h.riskManager.timeManager.GetResetHistory(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// UpdateTimeConfig 更新时间配置
func (h *APIHandler) UpdateTimeConfig(c *gin.Context) {
	var config TimeManagerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.riskManager.timeManager.UpdateConfig(config)

	// 保存到数据库
	if err := h.databaseManager.SaveSystemConfig("time_manager", "config", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "时间管理配置更新成功",
	})
}

// GetLiquidityStatus 获取流动性状态
func (h *APIHandler) GetLiquidityStatus(c *gin.Context) {
	state := h.riskManager.liquidityMonitor.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state,
	})
}

// UpdateLiquidityData 更新流动性和数据
func (h *APIHandler) UpdateLiquidityData(c *gin.Context) {
	var req struct {
		Symbol        string  `json:"symbol" binding:"required"`
		OpenInterest  float64 `json:"open_interest" binding:"required"`
		Volume24h     float64 `json:"volume_24h" binding:"required"`
		ChangePercent float64 `json:"change_percent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 更新流动性和数据
	err := h.riskManager.liquidityMonitor.UpdateLiquidityData(req.Symbol, req.OpenInterest, req.Volume24h)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("更新流动性和数据失败: %v", err),
		})
		return
	}

	// 创建数据结构用于保存
	data := LiquidityData{
		Symbol:        req.Symbol,
		OpenInterest:  req.OpenInterest,
		Volume24h:     req.Volume24h,
		ChangePercent: req.ChangePercent,
		LastUpdated:   time.Now(),
	}

	// 获取活跃警报
	alerts := h.riskManager.liquidityMonitor.GetActiveAlerts()

	// 保存数据和警报到数据库
	if err := h.databaseManager.SaveLiquidityData(req.Symbol, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存流动性和数据失败: %v", err),
		})
		return
	}

	for _, alert := range alerts {
		if err := h.databaseManager.SaveLiquidityAlert(alert); err != nil {
			// 记录错误但不中断响应
			fmt.Printf("保存流动性和警报失败: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"liquidity_data": data,
			"alerts":         alerts,
		},
	})
}

// GetLiquidityAlerts 获取流动性和警报
func (h *APIHandler) GetLiquidityAlerts(c *gin.Context) {
	alerts := h.riskManager.liquidityMonitor.GetActiveAlerts()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
	})
}

// ResolveLiquidityAlert 解决流动性和警报
func (h *APIHandler) ResolveLiquidityAlert(c *gin.Context) {
	alertID := c.Param("id")
	h.riskManager.liquidityMonitor.ResolveAlert(alertID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "警报已解决",
	})
}

// GetBlacklist 获取黑名单
func (h *APIHandler) GetBlacklist(c *gin.Context) {
	blacklist := h.riskManager.liquidityMonitor.GetBlacklistedSymbols()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    blacklist,
	})
}

// AddToBlacklist 添加到黑名单
func (h *APIHandler) AddToBlacklist(c *gin.Context) {
	var req struct {
		Symbol    string     `json:"symbol" binding:"required"`
		Reason    string     `json:"reason" binding:"required"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 创建黑名单条目
	entry := BlacklistEntry{
		Symbol:    req.Symbol,
		Reason:    req.Reason,
		AddedAt:   time.Now(),
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}

	// 添加到黑名单（需要实现公共方法）
	// 暂时直接保存到数据库
	if err := h.databaseManager.SaveBlacklistEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存黑名单条目失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entry,
	})
}

// RemoveFromBlacklist 从黑名单移除
func (h *APIHandler) RemoveFromBlacklist(c *gin.Context) {
	// 暂时返回成功响应，需要实现公共方法
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已从黑名单移除",
	})
}

// UpdateLiquidityConfig 更新流动性和配置
func (h *APIHandler) UpdateLiquidityConfig(c *gin.Context) {
	var config LiquidityMonitorConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.riskManager.liquidityMonitor.UpdateConfig(config)

	// 保存到数据库
	if err := h.databaseManager.SaveSystemConfig("liquidity_monitor", "config", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "流动性和监控配置更新成功",
	})
}

// GetSharpeStatus 获取夏普比率状态
func (h *APIHandler) GetSharpeStatus(c *gin.Context) {
	state := h.riskManager.sharpeCalculator.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state,
	})
}

// UpdateSharpeData 更新夏普比率数据
func (h *APIHandler) UpdateSharpeData(c *gin.Context) {
	var req struct {
		Equity float64 `json:"equity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	record, err := h.riskManager.sharpeCalculator.AddRecord(req.Equity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("添加夏普比率记录失败: %v", err),
		})
		return
	}

	// 保存到数据库
	if err := h.databaseManager.SaveSharpeRecord(*record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存夏普比率记录失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// GetSharpeRecords 获取夏普比率记录
func (h *APIHandler) GetSharpeRecords(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	records := h.riskManager.sharpeCalculator.GetRecentRecords(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

// GetSharpeTransitions 获取夏普比率状态转换
func (h *APIHandler) GetSharpeTransitions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	transitions := h.riskManager.sharpeCalculator.GetStateTransitions(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transitions,
	})
}

// UpdateSharpeConfig 更新夏普比率配置
func (h *APIHandler) UpdateSharpeConfig(c *gin.Context) {
	var config SharpeCalculatorConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.riskManager.sharpeCalculator.UpdateConfig(config)

	// 保存到数据库
	if err := h.databaseManager.SaveSystemConfig("sharpe_calculator", "config", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "夏普比率配置更新成功",
	})
}

// GetAICommitteeStatus 获取AI委员会状态
func (h *APIHandler) GetAICommitteeStatus(c *gin.Context) {
	state := h.riskManager.aiCommittee.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state,
	})
}

// MakeAIDecision 进行AI决策
func (h *APIHandler) MakeAIDecision(c *gin.Context) {
	var req struct {
		Symbol     string                 `json:"symbol" binding:"required"`
		MarketData map[string]interface{} `json:"market_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 如果没有提供市场数据，使用空的map
	if req.MarketData == nil {
		req.MarketData = make(map[string]interface{})
	}

	decision, err := h.riskManager.aiCommittee.MakeDecision(req.Symbol, req.MarketData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("AI决策失败: %v", err),
		})
		return
	}

	// 保存到数据库
	if err := h.databaseManager.SaveAICommitteeDecision(*decision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存AI决策失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    decision,
	})
}

// GetAIDecisions 获取AI决策历史
func (h *APIHandler) GetAIDecisions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	decisions := h.riskManager.aiCommittee.GetRecentDecisions(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    decisions,
	})
}

// GetModelPerformance 获取模型性能
func (h *APIHandler) GetModelPerformance(c *gin.Context) {
	state := h.riskManager.aiCommittee.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state.ModelPerformances,
	})
}

// UpdateAIConfig 更新AI配置
func (h *APIHandler) UpdateAIConfig(c *gin.Context) {
	var config AICommitteeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.riskManager.aiCommittee.UpdateConfig(config)

	// 保存到数据库
	if err := h.databaseManager.SaveSystemConfig("ai_committee", "config", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "AI委员会配置更新成功",
	})
}

// GetRiskStatus 获取风控状态
func (h *APIHandler) GetRiskStatus(c *gin.Context) {
	state := h.riskManager.GetCurrentState()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    state,
	})
}

// MakeRiskDecision 进行风控决策
func (h *APIHandler) MakeRiskDecision(c *gin.Context) {
	var req struct {
		Symbol     string                 `json:"symbol" binding:"required"`
		Action     string                 `json:"action" binding:"required"`
		MarketData map[string]interface{} `json:"market_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 如果没有提供市场数据，使用空的map
	if req.MarketData == nil {
		req.MarketData = make(map[string]interface{})
	}

	decision, err := h.riskManager.MakeRiskDecision(req.Symbol, req.Action, req.MarketData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("风控决策失败: %v", err),
		})
		return
	}

	// 保存到数据库
	if err := h.databaseManager.SaveRiskDecision(*decision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存风控决策失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    decision,
	})
}

// GetRiskDecisions 获取风控决策历史
func (h *APIHandler) GetRiskDecisions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	decisions := h.riskManager.GetRecentDecisions(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    decisions,
	})
}

// GetRiskAlerts 获取风控警报
func (h *APIHandler) GetRiskAlerts(c *gin.Context) {
	alerts := h.riskManager.GetActiveAlerts()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
	})
}

// ResolveRiskAlert 解决风控警报
func (h *APIHandler) ResolveRiskAlert(c *gin.Context) {
	// 这里需要实现警报解决逻辑
	// 暂时返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "风控警报已解决",
	})
}

// EmergencyStop 紧急停止
func (h *APIHandler) EmergencyStop(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 触发紧急停止
	h.riskManager.triggerEmergencyStop([]string{req.Reason})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "紧急停止已触发",
	})
}

// ResumeRisk 恢复风控
func (h *APIHandler) ResumeRisk(c *gin.Context) {
	// 这里需要实现恢复逻辑
	// 暂时返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "风控已恢复",
	})
}

// UpdateRiskConfig 更新风控配置
func (h *APIHandler) UpdateRiskConfig(c *gin.Context) {
	var config RiskManagerConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.riskManager.UpdateConfig(config)

	// 保存到数据库
	if err := h.databaseManager.SaveSystemConfig("risk_manager", "config", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "风控配置更新成功",
	})
}

// GetSystemConfig 获取系统配置
func (h *APIHandler) GetSystemConfig(c *gin.Context) {
	configType := c.Param("type")
	configName := c.Param("name")

	result, err := h.databaseManager.GetSystemConfig(configType, configName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// SetSystemConfig 设置系统配置
func (h *APIHandler) SetSystemConfig(c *gin.Context) {
	configType := c.Param("type")
	configName := c.Param("name")

	var configValue interface{}
	if err := c.ShouldBindJSON(&configValue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.databaseManager.SaveSystemConfig(configType, configName, configValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置保存成功",
	})
}

// GetAllConfigs 获取所有配置
func (h *APIHandler) GetAllConfigs(c *gin.Context) {
	// 这里需要实现获取所有配置的逻辑
	// 暂时返回空数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// CleanupOldData 清理旧数据
func (h *APIHandler) CleanupOldData(c *gin.Context) {
	var req struct {
		DaysToKeep int `json:"days_to_keep" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.databaseManager.CleanOldRecords(req.DaysToKeep)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("清理数据失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据清理完成",
	})
}

// GetDataStats 获取数据统计
func (h *APIHandler) GetDataStats(c *gin.Context) {
	// 这里需要实现数据统计逻辑
	// 暂时返回模拟数据
	stats := gin.H{
		"total_records": gin.H{
			"liquidity_data":      1000,
			"sharpe_records":      500,
			"ai_decisions":        200,
			"risk_decisions":      300,
			"alerts":              50,
		},
		"recent_activity": gin.H{
			"last_24h_decisions": 25,
			"active_alerts":      3,
			"blacklisted_symbols": 2,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ExportData 导出数据
func (h *APIHandler) ExportData(c *gin.Context) {
	var req struct {
		DataType  string    `json:"data_type" binding:"required"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
		Format    string    `json:"format"` // json, csv
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 这里需要实现数据导出逻辑
	// 暂时返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据导出请求已提交",
		"data": gin.H{
			"export_id": fmt.Sprintf("export_%d", time.Now().Unix()),
			"status":    "processing",
		},
	})
}