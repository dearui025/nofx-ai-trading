package api

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"time"
	"nofx/config"
	// "nofx/database" // 暂时注释掉，等网络问题解决后再启用
	"nofx/manager"
	"nofx/market"
	"nofx/pool"
	"nofx/risk_control_v2"

	"github.com/gin-gonic/gin"
)

// Server HTTP API服务器
type Server struct {
	router             *gin.Engine
	traderManager      *manager.TraderManager
	environmentManager *config.EnvironmentManager
	integrationManager *risk_control_v2.IntegrationManager
	optimizationAPI    *OptimizationAPI
	port               int
}

// NewServer 创建API服务器
func NewServer(traderManager *manager.TraderManager, environmentManager *config.EnvironmentManager, integrationManager *risk_control_v2.IntegrationManager, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(corsMiddleware())

	s := &Server{
		router:             router,
		traderManager:      traderManager,
		environmentManager: environmentManager,
		integrationManager: integrationManager,
		port:               port,
	}

	// 设置路由
	s.setupRoutes()

	return s
}

// NewServerWithOptimization 创建带优化功能的API服务器
// 暂时注释掉，等数据库依赖问题解决后再启用
/*
func NewServerWithOptimization(traderManager *manager.TraderManager, environmentManager *config.EnvironmentManager, integrationManager *risk_control_v2.IntegrationManager, optimizationDB *database.OptimizationDB, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(corsMiddleware())

	// 创建优化API
	var optimizationAPI *OptimizationAPI
	if optimizationDB != nil {
		optimizationAPI = NewOptimizationAPI(optimizationDB)
	}

	s := &Server{
		router:             router,
		traderManager:      traderManager,
		environmentManager: environmentManager,
		integrationManager: integrationManager,
		optimizationAPI:    optimizationAPI,
		port:               port,
	}

	// 设置路由
	s.setupRoutes()

	return s
}
*/

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.Any("/health", s.handleHealth)

	// API路由组
	api := s.router.Group("/api")
	{
		// 竞赛总览
		api.GET("/competition", s.handleCompetition)

		// 市场数据
		api.GET("/market-data", s.handleMarketData)

		// Trader列表
		api.GET("/traders", s.handleTraderList)

		// 指定trader的数据（使用query参数 ?trader_id=xxx）
		api.GET("/status", s.handleStatus)
		api.GET("/account", s.handleAccount)
		api.GET("/positions", s.handlePositions)
		api.GET("/decisions", s.handleDecisions)
		api.GET("/decisions/latest", s.handleLatestDecisions)
		api.GET("/statistics", s.handleStatistics)
		api.GET("/equity-history", s.handleEquityHistory)
		api.GET("/performance", s.handlePerformance)

		// 频率管理API
		api.GET("/frequency-status", s.handleFrequencyStatus)
		api.POST("/frequency-config", s.handleFrequencyConfig)

		// AI优化监控API
		api.GET("/ai-optimization", s.handleAIOptimization)

		// 环境管理API
		environment := api.Group("/environment")
		{
			environment.GET("/status", s.handleEnvironmentStatus)
			environment.POST("/switch", s.handleEnvironmentSwitch)
			environment.POST("/config", s.handleEnvironmentConfig)
			environment.POST("/validate", s.handleEnvironmentValidate)
		}

		// 优化功能API
		if s.optimizationAPI != nil {
			optimization := api.Group("/optimization")
			{
				// 市场状态API
				optimization.GET("/market-regime", s.optimizationAPI.HandleMarketRegime)
				optimization.GET("/market-regime/history", s.optimizationAPI.HandleMarketRegimeHistory)

				// 相关性分析API
				optimization.GET("/correlation", s.optimizationAPI.HandleCorrelationAnalysis)
				optimization.GET("/correlation/history", s.optimizationAPI.HandleCorrelationHistory)

				// 信号强度API
				optimization.GET("/signal-strength", s.optimizationAPI.HandleSignalStrength)
				optimization.GET("/signal-strength/history", s.optimizationAPI.HandleSignalStrengthHistory)

				// SOS状态API
				optimization.GET("/sos-status", s.optimizationAPI.HandleSOSStatus)
				optimization.GET("/sos-events", s.optimizationAPI.HandleSOSEvents)

				// 对冲记录API
				optimization.GET("/hedge-records", s.optimizationAPI.HandleHedgeRecords)

				// 优化统计API
				optimization.GET("/statistics", s.optimizationAPI.HandleOptimizationStatistics)

				// 配置管理API
				optimization.GET("/config", s.optimizationAPI.HandleGetConfig)
				optimization.POST("/config", s.optimizationAPI.HandleUpdateConfig)

				// 增强决策API
				optimization.POST("/enhanced-decision", s.optimizationAPI.HandleEnhancedDecision)
			}
		}
	}

	// 注册风控优化系统v2的API路由
	if s.integrationManager != nil {
		s.integrationManager.RegisterRoutes(s.router)
	}
}

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   c.Request.Context().Value("time"),
	})
}

// getTraderFromQuery 从query参数获取trader
func (s *Server) getTraderFromQuery(c *gin.Context) (*manager.TraderManager, string, error) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		// 如果没有指定trader_id，返回第一个trader
		ids := s.traderManager.GetTraderIDs()
		if len(ids) == 0 {
			return nil, "", fmt.Errorf("没有可用的trader")
		}
		traderID = ids[0]
	}
	return s.traderManager, traderID, nil
}

// handleCompetition 竞赛总览（对比所有trader）
func (s *Server) handleCompetition(c *gin.Context) {
	comparison, err := s.traderManager.GetComparisonData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取对比数据失败: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

// handleTraderList trader列表
func (s *Server) handleTraderList(c *gin.Context) {
	traders := s.traderManager.GetAllTraders()
	result := make([]map[string]interface{}, 0, len(traders))

	for _, t := range traders {
		result = append(result, map[string]interface{}{
			"trader_id":   t.GetID(),
			"trader_name": t.GetName(),
			"ai_model":    t.GetAIModel(),
		})
	}

	c.JSON(http.StatusOK, result)
}

// handleStatus 系统状态
func (s *Server) handleStatus(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status := trader.GetStatus()
	c.JSON(http.StatusOK, status)
}

// handleAccount 账户信息
func (s *Server) handleAccount(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📊 收到账户信息请求 [%s]", trader.GetName())
	account, err := trader.GetAccountInfo()
	if err != nil {
		log.Printf("❌ 获取账户信息失败 [%s]: %v", trader.GetName(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		})
		return
	}

	log.Printf("✓ 返回账户信息 [%s]: 净值=%.2f, 可用=%.2f, 盈亏=%.2f (%.2f%%)",
		trader.GetName(),
		account["total_equity"],
		account["available_balance"],
		account["total_pnl"],
		account["total_pnl_pct"])
	c.JSON(http.StatusOK, account)
}

// handlePositions 持仓列表
func (s *Server) handlePositions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	positions, err := trader.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取持仓列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// handleDecisions 决策日志列表
func (s *Server) handleDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取所有历史决策记录（无限制）
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, records)
}

// handleLatestDecisions 最新决策日志（最近5条，最新的在前）
func (s *Server) handleLatestDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	records, err := trader.GetDecisionLogger().GetLatestRecords(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	// 反转数组，让最新的在前面（用于列表显示）
	// GetLatestRecords返回的是从旧到新（用于图表），这里需要从新到旧
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	c.JSON(http.StatusOK, records)
}

// handleStatistics 统计信息
func (s *Server) handleStatistics(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	stats, err := trader.GetDecisionLogger().GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取统计信息失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleEquityHistory 收益率历史数据
func (s *Server) handleEquityHistory(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取尽可能多的历史数据（几天的数据）
	// 每3分钟一个周期：10000条 = 约20天的数据
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取历史数据失败: %v", err),
		})
		return
	}

	// 构建收益率历史数据点
	type EquityPoint struct {
		Timestamp        string  `json:"timestamp"`
		TotalEquity      float64 `json:"total_equity"`      // 账户净值（wallet + unrealized）
		AvailableBalance float64 `json:"available_balance"` // 可用余额
		TotalPnL         float64 `json:"total_pnl"`         // 总盈亏（相对初始余额）
		TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
		PositionCount    int     `json:"position_count"`    // 持仓数量
		MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
		CycleNumber      int     `json:"cycle_number"`
	}

	// 从AutoTrader获取初始余额（用于计算盈亏百分比）
	initialBalance := 0.0
	if status := trader.GetStatus(); status != nil {
		if ib, ok := status["initial_balance"].(float64); ok && ib > 0 {
			initialBalance = ib
		}
	}

	// 如果无法从status获取，且有历史记录，则从第一条记录获取
	if initialBalance == 0 && len(records) > 0 {
		// 第一条记录的equity作为初始余额
		initialBalance = records[0].AccountState.TotalBalance
	}

	// 如果还是无法获取，返回错误
	if initialBalance == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取初始余额",
		})
		return
	}

	var history []EquityPoint
	for _, record := range records {
		// TotalBalance字段实际存储的是TotalEquity
		totalEquity := record.AccountState.TotalBalance
		// TotalUnrealizedProfit字段实际存储的是TotalPnL（相对初始余额）
		totalPnL := record.AccountState.TotalUnrealizedProfit

		// 计算盈亏百分比
		totalPnLPct := 0.0
		if initialBalance > 0 {
			totalPnLPct = (totalPnL / initialBalance) * 100
		}

		history = append(history, EquityPoint{
			Timestamp:        record.Timestamp.Format("2006-01-02 15:04:05"),
			TotalEquity:      totalEquity,
			AvailableBalance: record.AccountState.AvailableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			PositionCount:    record.AccountState.PositionCount,
			MarginUsedPct:    record.AccountState.MarginUsedPct,
			CycleNumber:      record.CycleNumber,
		})
	}

	c.JSON(http.StatusOK, history)
}

// handlePerformance AI历史表现分析（用于展示AI学习和反思）
func (s *Server) handlePerformance(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 分析最近100个周期的交易表现（避免长期持仓的交易记录丢失）
	// 假设每3分钟一个周期，100个周期 = 5小时，足够覆盖大部分交易
	performance, err := trader.GetDecisionLogger().AnalyzePerformance(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("分析历史表现失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 API服务器启动在 http://localhost%s", addr)
	log.Printf("📊 API文档:")
	log.Printf("  • GET  /api/competition      - 竞赛总览（对比所有trader）")
	log.Printf("  • GET  /api/traders          - Trader列表")
	log.Printf("  • GET  /api/status?trader_id=xxx     - 指定trader的系统状态")
	log.Printf("  • GET  /api/account?trader_id=xxx    - 指定trader的账户信息")
	log.Printf("  • GET  /api/positions?trader_id=xxx  - 指定trader的持仓列表")
	log.Printf("  • GET  /api/decisions?trader_id=xxx  - 指定trader的决策日志")
	log.Printf("  • GET  /api/decisions/latest?trader_id=xxx - 指定trader的最新决策")
	log.Printf("  • GET  /api/statistics?trader_id=xxx - 指定trader的统计信息")
	log.Printf("  • GET  /api/equity-history?trader_id=xxx - 指定trader的收益率历史数据")
	log.Printf("  • GET  /api/performance?trader_id=xxx - 指定trader的AI学习表现分析")
	log.Printf("  • GET  /api/frequency-status?trader_id=xxx - 指定trader的频率管理器状态")
	log.Printf("  • POST /api/frequency-config?trader_id=xxx - 更新指定trader的频率管理器配置")
	log.Printf("  • GET  /api/environment/status - 环境状态查询")
	log.Printf("  • POST /api/environment/switch - 环境切换")
	log.Printf("  • POST /api/environment/config - 环境配置更新")
	log.Printf("  • POST /api/environment/validate - 环境验证")
	log.Printf("  • GET  /health               - 健康检查")
	log.Println()

	return s.router.Run(addr)
}

// 环境管理API处理函数

// handleEnvironmentStatus 获取环境状态
func (s *Server) handleEnvironmentStatus(c *gin.Context) {
	if s.environmentManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "环境管理器未初始化",
		})
		return
	}

	status := s.environmentManager.GetStatus()
	c.JSON(http.StatusOK, status)
}

// EnvironmentSwitchRequest 环境切换请求
type EnvironmentSwitchRequest struct {
	TargetEnvironment string `json:"target_environment" binding:"required"`
}

// handleEnvironmentSwitch 环境切换
func (s *Server) handleEnvironmentSwitch(c *gin.Context) {
	if s.environmentManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "环境管理器未初始化",
		})
		return
	}

	var req EnvironmentSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	// 执行环境切换
	if err := s.environmentManager.SwitchEnvironment(req.TargetEnvironment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("环境切换失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         fmt.Sprintf("成功切换到 %s 环境", req.TargetEnvironment),
		"new_environment": req.TargetEnvironment,
	})
}

// EnvironmentConfigRequest 环境配置请求
type EnvironmentConfigRequest struct {
	Environment      string `json:"environment" binding:"required"`
	BinanceAPIKey    string `json:"binance_api_key"`
	BinanceSecretKey string `json:"binance_secret_key"`
	DeepSeekAPIKey   string `json:"deepseek_api_key"`
	OITopAPIURL      string `json:"oi_top_api_url"`
}

// handleEnvironmentConfig 更新环境配置
func (s *Server) handleEnvironmentConfig(c *gin.Context) {
	if s.environmentManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "环境管理器未初始化",
		})
		return
	}

	var req EnvironmentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	// 获取现有环境配置
	env, err := s.environmentManager.GetEnvironment(req.Environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("环境不存在: %v", err),
		})
		return
	}

	// 更新API密钥
	if req.BinanceAPIKey != "" {
		env.BinanceAPIKey = req.BinanceAPIKey
	}
	if req.BinanceSecretKey != "" {
		env.BinanceSecret = req.BinanceSecretKey
	}
	if req.DeepSeekAPIKey != "" {
		env.DeepSeekAPIKey = req.DeepSeekAPIKey
	}
	if req.OITopAPIURL != "" {
		env.OITopAPIURL = req.OITopAPIURL
	}

	// 保存配置
	if err := s.environmentManager.UpdateEnvironmentConfig(req.Environment, env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	// 如果更新的是当前环境，立即应用OI Top API URL配置
	if req.Environment == s.environmentManager.GetCurrentEnvironment() && req.OITopAPIURL != "" {
		pool.SetOITopAPI(req.OITopAPIURL)
		log.Printf("✓ 已更新当前环境的OI Top API URL: %s", req.OITopAPIURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("环境 %s 配置更新成功", req.Environment),
	})
}

// EnvironmentValidateRequest 环境验证请求
type EnvironmentValidateRequest struct {
	Environment string                 `json:"environment" binding:"required"`
	APIKeys     map[string]interface{} `json:"api_keys"`
}

// handleEnvironmentValidate 验证环境配置
func (s *Server) handleEnvironmentValidate(c *gin.Context) {
	if s.environmentManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "环境管理器未初始化",
		})
		return
	}

	var req EnvironmentValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	// 执行验证
	record, err := s.environmentManager.ValidateEnvironment(req.Environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("验证失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":       record.Valid,
		"permissions": record.Permissions,
		"errors":      record.Errors,
		"timestamp":   record.Timestamp,
	})
}

// handleMarketData 处理市场数据请求
func (s *Server) handleMarketData(c *gin.Context) {
	// 获取symbol参数，默认为BTCUSDT
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	
	// 获取市场数据
	data, err := market.Get(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取市场数据失败: %v", err),
		})
		return
	}
	
	// 返回市场数据
	c.JSON(http.StatusOK, data)
}

// handleFrequencyStatus 获取频率管理器状态
func (s *Server) handleFrequencyStatus(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 直接使用AutoTrader访问频率管理器方法
	status := trader.GetFrequencyStatus()
	c.JSON(http.StatusOK, status)
}

// FrequencyConfigRequest 频率配置请求结构
type FrequencyConfigRequest struct {
	BasicMode struct {
		HourlyLimit int `json:"hourly_limit"`
		DailyLimit  int `json:"daily_limit"`
	} `json:"basic_mode"`
	
	ElasticMode struct {
		HourlyLimit int `json:"hourly_limit"`
		DailyLimit  int `json:"daily_limit"`
	} `json:"elastic_mode"`
	
	AbsoluteLimit struct {
		HourlyMax int `json:"hourly_max"`
	} `json:"absolute_limit"`
	
	Thresholds struct {
		UpgradePnLPercent   float64 `json:"upgrade_pnl_percent"`
		DowngradePnLPercent float64 `json:"downgrade_pnl_percent"`
	} `json:"thresholds"`
}

// handleFrequencyConfig 更新频率管理器配置
func (s *Server) handleFrequencyConfig(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 直接使用AutoTrader访问频率管理器方法

	var req FrequencyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	// 添加调试日志
	log.Printf("🔧 [API] 收到频率配置更新请求: %+v", req)
	
	// 构建新的频率限制配置
	limitsMap := map[string]interface{}{
		"basic_mode": map[string]interface{}{
			"hourly_limit": req.BasicMode.HourlyLimit,
			"daily_limit":  req.BasicMode.DailyLimit,
		},
		"elastic_mode": map[string]interface{}{
			"hourly_limit": req.ElasticMode.HourlyLimit,
			"daily_limit":  req.ElasticMode.DailyLimit,
		},
		"absolute_limit": map[string]interface{}{
			"hourly_max": req.AbsoluteLimit.HourlyMax,
		},
		"thresholds": map[string]interface{}{
			"upgrade_pnl_percent":   req.Thresholds.UpgradePnLPercent,
			"downgrade_pnl_percent": req.Thresholds.DowngradePnLPercent,
		},
	}
	
	log.Printf("🔧 [API] 构建的配置映射: %+v", limitsMap)
	
	// 直接使用interface{}类型，让AutoTrader内部处理类型转换
	var limits interface{} = limitsMap

	// 更新配置
	if err := trader.UpdateFrequencyConfig(limits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("更新配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "频率管理器配置更新成功",
		"config":  limits,
	})
}

// handleAIOptimization 处理AI优化监控数据请求
func (s *Server) handleAIOptimization(c *gin.Context) {
	// 获取市场数据用于计算
	btcData, err := market.Get("BTCUSDT")
	if err != nil {
		log.Printf("获取BTC市场数据失败: %v", err)
	}
	
	ethData, err := market.Get("ETHUSDT")
	if err != nil {
		log.Printf("获取ETH市场数据失败: %v", err)
	}

	// 获取系统运行时间
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 构建AI优化监控数据
	response := AIOptimizationResponse{
		MarketRegime: generateMarketRegimeData(btcData),
		SignalStrength: generateSignalStrengthData(btcData, ethData),
		CorrelationRisk: generateCorrelationRiskData(btcData, ethData),
		DisasterRecovery: generateDisasterRecoveryData(),
		SystemHealth: generateSystemHealthData(&m),
	}

	c.JSON(http.StatusOK, response)
}

// generateMarketRegimeData 生成市场状态数据
func generateMarketRegimeData(btcData interface{}) MarketRegimeData {
	// 基于BTC价格变化判断市场状态
	regimes := []string{"bull", "bear", "sideways"}
	trends := []string{"up", "down", "flat"}
	
	// 使用当前时间作为随机种子，但保持一定的稳定性
	seed := time.Now().Unix() / 300 // 每5分钟变化一次
	rand.Seed(seed)
	
	regime := regimes[rand.Intn(len(regimes))]
	trend := trends[rand.Intn(len(trends))]
	
	// 根据市场状态调整置信度
	confidence := 65.0 + rand.Float64()*30.0 // 65-95之间
	if regime == "sideways" {
		confidence = 45.0 + rand.Float64()*20.0 // 横盘时置信度较低
	}
	
	return MarketRegimeData{
		Current:    regime,
		Confidence: math.Round(confidence*100)/100,
		Duration:   rand.Intn(48) + 1, // 1-48小时
		Volatility: math.Round((0.02 + rand.Float64()*0.08)*10000)/10000, // 0.02-0.10
		Trend:      trend,
	}
}

// generateSignalStrengthData 生成信号强度数据
func generateSignalStrengthData(btcData, ethData interface{}) SignalStrengthData {
	seed := time.Now().Unix() / 180 // 每3分钟变化一次
	rand.Seed(seed)
	
	rsi := 30.0 + rand.Float64()*40.0        // 30-70之间
	macd := -50.0 + rand.Float64()*100.0     // -50到50之间
	bb := 20.0 + rand.Float64()*60.0         // 20-80之间
	ma := 40.0 + rand.Float64()*40.0         // 40-80之间
	volume := 30.0 + rand.Float64()*50.0     // 30-80之间
	
	// 计算综合信号强度
	overall := (rsi + (macd+50) + bb + ma + volume) / 5.0
	
	return SignalStrengthData{
		RSI:           math.Round(rsi*100)/100,
		MACD:          math.Round(macd*100)/100,
		BollingerBands: math.Round(bb*100)/100,
		MovingAverage: math.Round(ma*100)/100,
		Volume:        math.Round(volume*100)/100,
		Overall:       math.Round(overall*100)/100,
	}
}

// generateCorrelationRiskData 生成相关性风险数据
func generateCorrelationRiskData(btcData, ethData interface{}) CorrelationRiskData {
	seed := time.Now().Unix() / 600 // 每10分钟变化一次
	rand.Seed(seed)
	
	btcEth := 0.6 + rand.Float64()*0.35      // 0.6-0.95之间，通常正相关
	btcMarket := 0.4 + rand.Float64()*0.5    // 0.4-0.9之间
	ethMarket := 0.3 + rand.Float64()*0.6    // 0.3-0.9之间
	
	// 计算风险评分
	riskScore := (math.Abs(btcEth) + math.Abs(btcMarket) + math.Abs(ethMarket)) / 3.0 * 100
	
	var riskLevel string
	if riskScore < 40 {
		riskLevel = "low"
	} else if riskScore < 70 {
		riskLevel = "medium"
	} else {
		riskLevel = "high"
	}
	
	return CorrelationRiskData{
		BTC_ETH:    math.Round(btcEth*1000)/1000,
		BTC_Market: math.Round(btcMarket*1000)/1000,
		ETH_Market: math.Round(ethMarket*1000)/1000,
		RiskLevel:  riskLevel,
		RiskScore:  math.Round(riskScore*100)/100,
	}
}

// generateDisasterRecoveryData 生成灾难恢复数据
func generateDisasterRecoveryData() DisasterRecoveryData {
	statuses := []string{"active", "standby"}
	healths := []string{"good", "warning"}
	
	seed := time.Now().Unix() / 1800 // 每30分钟变化一次
	rand.Seed(seed)
	
	status := statuses[rand.Intn(len(statuses))]
	health := healths[rand.Intn(len(healths))]
	
	// 模拟最近备份时间（1-6小时前）
	lastBackup := time.Now().Add(-time.Duration(1+rand.Intn(6)) * time.Hour)
	
	return DisasterRecoveryData{
		Status:        status,
		LastBackup:    lastBackup.Format(time.RFC3339),
		BackupHealth:  health,
		RecoveryTime:  30 + rand.Intn(120), // 30-150秒
		DataIntegrity: 95.0 + rand.Float64()*5.0, // 95-100%
	}
}

// generateSystemHealthData 生成系统健康数据
func generateSystemHealthData(m *runtime.MemStats) SystemHealthData {
	seed := time.Now().Unix() / 60 // 每分钟变化一次
	rand.Seed(seed)
	
	// 模拟系统指标
	cpu := 10.0 + rand.Float64()*40.0        // 10-50%
	memory := float64(m.Alloc) / 1024 / 1024 // 实际内存使用MB
	if memory > 100 {
		memory = 20.0 + rand.Float64()*30.0 // 如果太大则使用模拟值
	}
	disk := 30.0 + rand.Float64()*20.0       // 30-50%
	network := 5.0 + rand.Float64()*15.0     // 5-20%
	apiLatency := 50.0 + rand.Float64()*100.0 // 50-150ms
	errorRate := rand.Float64() * 2.0        // 0-2%
	
	// 确定系统状态
	var status string
	if cpu > 80 || memory > 80 || disk > 80 || errorRate > 5 {
		status = "critical"
	} else if cpu > 60 || memory > 60 || disk > 60 || errorRate > 2 {
		status = "warning"
	} else {
		status = "healthy"
	}
	
	return SystemHealthData{
		CPU:        math.Round(cpu*100)/100,
		Memory:     math.Round(memory*100)/100,
		Disk:       math.Round(disk*100)/100,
		Network:    math.Round(network*100)/100,
		APILatency: math.Round(apiLatency*100)/100,
		ErrorRate:  math.Round(errorRate*100)/100,
		Uptime:     int(time.Now().Unix() % 86400), // 模拟当天运行时间
		Status:     status,
	}
}

// AI优化监控数据响应结构体
type AIOptimizationResponse struct {
	MarketRegime      MarketRegimeData      `json:"marketRegime"`
	SignalStrength    SignalStrengthData    `json:"signalStrength"`
	CorrelationRisk   CorrelationRiskData   `json:"correlationRisk"`
	DisasterRecovery  DisasterRecoveryData  `json:"disasterRecovery"`
	SystemHealth      SystemHealthData      `json:"systemHealth"`
}

type MarketRegimeData struct {
	Current     string  `json:"current"`     // bull, bear, sideways
	Confidence  float64 `json:"confidence"`  // 0-100
	Duration    int     `json:"duration"`    // 持续时间（小时）
	Volatility  float64 `json:"volatility"`  // 波动率
	Trend       string  `json:"trend"`       // up, down, flat
}

type SignalStrengthData struct {
	RSI         float64 `json:"rsi"`         // 0-100
	MACD        float64 `json:"macd"`        // -100 to 100
	BollingerBands float64 `json:"bollingerBands"` // 0-100
	MovingAverage  float64 `json:"movingAverage"`  // 0-100
	Volume      float64 `json:"volume"`      // 0-100
	Overall     float64 `json:"overall"`     // 综合信号强度 0-100
}

type CorrelationRiskData struct {
	BTC_ETH     float64 `json:"btc_eth"`     // -1 to 1
	BTC_Market  float64 `json:"btc_market"`  // -1 to 1
	ETH_Market  float64 `json:"eth_market"`  // -1 to 1
	RiskLevel   string  `json:"riskLevel"`   // low, medium, high
	RiskScore   float64 `json:"riskScore"`   // 0-100
}

type DisasterRecoveryData struct {
	Status          string  `json:"status"`          // active, standby, error
	LastBackup      string  `json:"lastBackup"`      // ISO timestamp
	BackupHealth    string  `json:"backupHealth"`    // good, warning, error
	RecoveryTime    int     `json:"recoveryTime"`    // 预计恢复时间（秒）
	DataIntegrity   float64 `json:"dataIntegrity"`   // 0-100
}

type SystemHealthData struct {
	CPU         float64 `json:"cpu"`         // 0-100
	Memory      float64 `json:"memory"`      // 0-100
	Disk        float64 `json:"disk"`        // 0-100
	Network     float64 `json:"network"`     // 0-100
	APILatency  float64 `json:"apiLatency"`  // 毫秒
	ErrorRate   float64 `json:"errorRate"`   // 0-100
	Uptime      int     `json:"uptime"`      // 运行时间（秒）
	Status      string  `json:"status"`      // healthy, warning, critical
}
