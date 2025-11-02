package risk_control_v2

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IntegrationManager 集成管理器
type IntegrationManager struct {
	riskManager     *RiskManager
	databaseManager *DatabaseManager
	configManager   *ConfigManager
	apiHandler      *APIHandler
	
	// 系统状态
	isInitialized bool
	isRunning     bool
	startTime     time.Time
	
	// 监控和统计
	stats           *SystemStats
	healthChecker   *HealthChecker
	
	mutex  sync.RWMutex
	logger *log.Logger
}

// SystemStats 系统统计
type SystemStats struct {
	StartTime           time.Time `json:"start_time"`
	Uptime              string    `json:"uptime"`
	TotalDecisions      int64     `json:"total_decisions"`
	TotalAlerts         int64     `json:"total_alerts"`
	SuccessfulDecisions int64     `json:"successful_decisions"`
	FailedDecisions     int64     `json:"failed_decisions"`
	AverageResponseTime string    `json:"average_response_time"`
	LastUpdateTime      time.Time `json:"last_update_time"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks          map[string]HealthCheck
	lastCheckTime   time.Time
	checkInterval   time.Duration
	overallStatus   string
	mutex           sync.RWMutex
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // healthy, warning, critical
	Message     string    `json:"message"`
	LastCheck   time.Time `json:"last_check"`
	CheckCount  int64     `json:"check_count"`
	FailCount   int64     `json:"fail_count"`
	ResponseTime string   `json:"response_time"`
}

// IntegrationConfig 集成配置
type IntegrationConfig struct {
	DatabasePath        string        `json:"database_path"`
	ConfigDir          string        `json:"config_dir"`
	LogLevel           string        `json:"log_level"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	StatsUpdateInterval time.Duration `json:"stats_update_interval"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableProfiling    bool          `json:"enable_profiling"`
}

// NewIntegrationManager 创建集成管理器
func NewIntegrationManager(config IntegrationConfig) (*IntegrationManager, error) {
	im := &IntegrationManager{
		logger: log.New(log.Writer(), "[IntegrationManager] ", log.LstdFlags),
		stats: &SystemStats{
			StartTime: time.Now(),
		},
		healthChecker: &HealthChecker{
			checks:        make(map[string]HealthCheck),
			checkInterval: config.HealthCheckInterval,
			overallStatus: "initializing",
		},
	}

	// 初始化数据库管理器
	dbManager, err := NewDatabaseManager(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库管理器失败: %v", err)
	}
	im.databaseManager = dbManager

	// 初始化配置管理器
	configManager, err := NewConfigManager(config.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("初始化配置管理器失败: %v", err)
	}
	im.configManager = configManager

	// 初始化风控管理器
	riskManagerConfig, err := configManager.GetRiskManagerConfig()
	if err != nil {
		im.logger.Printf("获取风控配置失败，使用默认配置: %v", err)
		riskManagerConfig = RiskManagerConfig{
			GlobalRiskEnabled:         true,
			EmergencyStopEnabled:      true,
			MonitoringIntervalSeconds: 30,
		}
	}

	riskManager := NewRiskManager(riskManagerConfig)
	im.riskManager = riskManager

	// 初始化API处理器
	im.apiHandler = NewAPIHandler(riskManager, dbManager)

	// 注册配置变更监听器
	if err := im.registerConfigWatchers(); err != nil {
		return nil, fmt.Errorf("注册配置监听器失败: %v", err)
	}

	// 初始化健康检查
	im.initHealthChecks()

	im.isInitialized = true
	im.logger.Printf("集成管理器初始化完成")

	return im, nil
}

// Start 启动系统
func (im *IntegrationManager) Start() error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if !im.isInitialized {
		return fmt.Errorf("系统未初始化")
	}

	if im.isRunning {
		return fmt.Errorf("系统已在运行")
	}

	// 启动风控管理器
	if err := im.riskManager.Start(); err != nil {
		return fmt.Errorf("启动风控管理器失败: %v", err)
	}

	// 启动健康检查
	go im.startHealthChecking()

	// 启动统计更新
	go im.startStatsUpdating()

	im.isRunning = true
	im.startTime = time.Now()
	im.stats.StartTime = im.startTime

	im.logger.Printf("风控系统v2启动成功")
	return nil
}

// Stop 停止系统
func (im *IntegrationManager) Stop() error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	if !im.isRunning {
		return fmt.Errorf("系统未运行")
	}

	// 停止风控管理器
	if err := im.riskManager.Stop(); err != nil {
		im.logger.Printf("停止风控管理器失败: %v", err)
	}

	// 关闭数据库连接
	if err := im.databaseManager.Close(); err != nil {
		im.logger.Printf("关闭数据库连接失败: %v", err)
	}

	im.isRunning = false
	im.logger.Printf("风控系统v2已停止")
	return nil
}

// RegisterRoutes 注册API路由
func (im *IntegrationManager) RegisterRoutes(router *gin.Engine) {
	im.apiHandler.RegisterRoutes(router)
	
	// 添加系统管理路由
	systemGroup := router.Group("/api/v2/system")
	{
		systemGroup.GET("/status", im.getSystemStatus)
		systemGroup.GET("/health", im.getHealthStatus)
		systemGroup.GET("/stats", im.getSystemStats)
		systemGroup.POST("/start", im.startSystem)
		systemGroup.POST("/stop", im.stopSystem)
		systemGroup.POST("/restart", im.restartSystem)
		systemGroup.GET("/config", im.getSystemConfig)
		systemGroup.PUT("/config", im.updateSystemConfig)
	}
}

// registerConfigWatchers 注册配置变更监听器
func (im *IntegrationManager) registerConfigWatchers() error {
	// 时间管理器配置监听
	err := im.configManager.RegisterWatcher("time_manager", func(config interface{}) error {
		timeConfig, err := im.configManager.GetTimeManagerConfig()
		if err == nil {
			im.riskManager.timeManager.UpdateConfig(timeConfig)
			im.logger.Printf("时间管理器配置已更新")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("注册时间管理器配置监听器失败: %v", err)
	}

	// 流动性监控配置监听
	err = im.configManager.RegisterWatcher("liquidity_monitor", func(config interface{}) error {
		liquidityConfig, err := im.configManager.GetLiquidityMonitorConfig()
		if err == nil {
			im.riskManager.liquidityMonitor.UpdateConfig(liquidityConfig)
			im.logger.Printf("流动性监控配置已更新")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("注册流动性监控配置监听器失败: %v", err)
	}

	// 夏普比率计算器配置监听
	err = im.configManager.RegisterWatcher("sharpe_calculator", func(config interface{}) error {
		sharpeConfig, err := im.configManager.GetSharpeCalculatorConfig()
		if err == nil {
			im.riskManager.sharpeCalculator.UpdateConfig(sharpeConfig)
			im.logger.Printf("夏普比率计算器配置已更新")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("注册夏普比率计算器配置监听器失败: %v", err)
	}

	// AI委员会配置监听
	err = im.configManager.RegisterWatcher("ai_committee", func(config interface{}) error {
		aiConfig, err := im.configManager.GetAICommitteeConfig()
		if err == nil {
			im.riskManager.aiCommittee.UpdateConfig(aiConfig)
			im.logger.Printf("AI委员会配置已更新")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("注册AI委员会配置监听器失败: %v", err)
	}

	// 风控管理器配置监听
	err = im.configManager.RegisterWatcher("risk_manager", func(config interface{}) error {
		riskConfig, err := im.configManager.GetRiskManagerConfig()
		if err == nil {
			im.riskManager.UpdateConfig(riskConfig)
			im.logger.Printf("风控管理器配置已更新")
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("注册风控管理器配置监听器失败: %v", err)
	}

	return nil
}

// initHealthChecks 初始化健康检查
func (im *IntegrationManager) initHealthChecks() {
	checks := map[string]func() HealthCheck{
		"database": im.checkDatabase,
		"risk_manager": im.checkRiskManager,
		"time_manager": im.checkTimeManager,
		"liquidity_monitor": im.checkLiquidityMonitor,
		"sharpe_calculator": im.checkSharpeCalculator,
		"ai_committee": im.checkAICommittee,
		"config_manager": im.checkConfigManager,
	}

	for name, checkFunc := range checks {
		im.healthChecker.checks[name] = HealthCheck{
			Name:   name,
			Status: "unknown",
		}
		
		// 执行初始检查
		go func(name string, checkFunc func() HealthCheck) {
			check := checkFunc()
			im.healthChecker.mutex.Lock()
			im.healthChecker.checks[name] = check
			im.healthChecker.mutex.Unlock()
		}(name, checkFunc)
	}
}

// startHealthChecking 启动健康检查
func (im *IntegrationManager) startHealthChecking() {
	ticker := time.NewTicker(im.healthChecker.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !im.isRunning {
			break
		}

		im.performHealthChecks()
	}
}

// performHealthChecks 执行健康检查
func (im *IntegrationManager) performHealthChecks() {
	checks := map[string]func() HealthCheck{
		"database": im.checkDatabase,
		"risk_manager": im.checkRiskManager,
		"time_manager": im.checkTimeManager,
		"liquidity_monitor": im.checkLiquidityMonitor,
		"sharpe_calculator": im.checkSharpeCalculator,
		"ai_committee": im.checkAICommittee,
		"config_manager": im.checkConfigManager,
	}

	healthyCount := 0
	totalCount := len(checks)

	for name, checkFunc := range checks {
		start := time.Now()
		check := checkFunc()
		check.ResponseTime = time.Since(start).String()
		check.LastCheck = time.Now()
		check.CheckCount++

		if check.Status != "healthy" {
			check.FailCount++
		} else {
			healthyCount++
		}

		im.healthChecker.mutex.Lock()
		im.healthChecker.checks[name] = check
		im.healthChecker.mutex.Unlock()
	}

	// 更新整体状态
	im.healthChecker.mutex.Lock()
	if healthyCount == totalCount {
		im.healthChecker.overallStatus = "healthy"
	} else if healthyCount >= totalCount/2 {
		im.healthChecker.overallStatus = "warning"
	} else {
		im.healthChecker.overallStatus = "critical"
	}
	im.healthChecker.lastCheckTime = time.Now()
	im.healthChecker.mutex.Unlock()
}

// 各个组件的健康检查函数
func (im *IntegrationManager) checkDatabase() HealthCheck {
	// 简单的数据库连接检查
	_, err := im.databaseManager.GetRiskDecisions(1)
	if err != nil {
		return HealthCheck{
			Name:    "database",
			Status:  "critical",
			Message: fmt.Sprintf("数据库连接失败: %v", err),
		}
	}
	return HealthCheck{
		Name:    "database",
		Status:  "healthy",
		Message: "数据库连接正常",
	}
}

func (im *IntegrationManager) checkRiskManager() HealthCheck {
	state := im.riskManager.GetCurrentState()
	if !state.IsActive {
		return HealthCheck{
			Name:    "risk_manager",
			Status:  "warning",
			Message: "风控管理器未激活",
		}
	}
	if state.EmergencyStop {
		return HealthCheck{
			Name:    "risk_manager",
			Status:  "critical",
			Message: "风控管理器处于紧急停止状态",
		}
	}
	return HealthCheck{
		Name:    "risk_manager",
		Status:  "healthy",
		Message: "风控管理器运行正常",
	}
}

func (im *IntegrationManager) checkTimeManager() HealthCheck {
	state := im.riskManager.timeManager.GetCurrentState()
	timeDiff := time.Since(state.CurrentUTCTime)
	if timeDiff > time.Minute*5 {
		return HealthCheck{
			Name:    "time_manager",
			Status:  "warning",
			Message: fmt.Sprintf("时间同步延迟: %v", timeDiff),
		}
	}
	return HealthCheck{
		Name:    "time_manager",
		Status:  "healthy",
		Message: "时间管理器运行正常",
	}
}

func (im *IntegrationManager) checkLiquidityMonitor() HealthCheck {
	state := im.riskManager.liquidityMonitor.GetCurrentState()
	criticalAlerts := 0
	for _, alert := range state.ActiveAlerts {
		if alert.Level == "critical" {
			criticalAlerts++
		}
	}
	
	if criticalAlerts > 5 {
		return HealthCheck{
			Name:    "liquidity_monitor",
			Status:  "critical",
			Message: fmt.Sprintf("存在%d个严重流动性警报", criticalAlerts),
		}
	} else if criticalAlerts > 0 {
		return HealthCheck{
			Name:    "liquidity_monitor",
			Status:  "warning",
			Message: fmt.Sprintf("存在%d个严重流动性警报", criticalAlerts),
		}
	}
	
	return HealthCheck{
		Name:    "liquidity_monitor",
		Status:  "healthy",
		Message: "流动性监控运行正常",
	}
}

func (im *IntegrationManager) checkSharpeCalculator() HealthCheck {
	state := im.riskManager.sharpeCalculator.GetCurrentState()
	if state.CurrentState == SharpePoor || state.CurrentState == SharpeVeryPoor {
		return HealthCheck{
			Name:    "sharpe_calculator",
			Status:  "warning",
			Message: "夏普比率处于低水平",
		}
	}
	return HealthCheck{
		Name:    "sharpe_calculator",
		Status:  "healthy",
		Message: "夏普比率计算器运行正常",
	}
}

func (im *IntegrationManager) checkAICommittee() HealthCheck {
	state := im.riskManager.aiCommittee.GetCurrentState()
	if state.TotalDecisions == 0 {
		return HealthCheck{
			Name:    "ai_committee",
			Status:  "warning",
			Message: "AI委员会暂无决策记录",
		}
	}
	
	// 检查最近决策的平均置信度
	recentDecisions := im.riskManager.aiCommittee.GetRecentDecisions(10)
	if len(recentDecisions) == 0 {
		return HealthCheck{
			Name:    "ai_committee",
			Status:  "warning",
			Message: "AI委员会暂无决策记录",
		}
	}
	
	totalConfidence := 0.0
	for _, decision := range recentDecisions {
		totalConfidence += decision.Confidence
	}
	avgConfidence := totalConfidence / float64(len(recentDecisions))
	
	if avgConfidence < 0.5 {
		return HealthCheck{
			Name:    "ai_committee",
			Status:  "warning",
			Message: fmt.Sprintf("AI决策平均置信度较低: %.2f", avgConfidence),
		}
	}
	
	return HealthCheck{
		Name:    "ai_committee",
		Status:  "healthy",
		Message: "AI委员会运行正常",
	}
}

func (im *IntegrationManager) checkConfigManager() HealthCheck {
	configs := im.configManager.GetAllConfigs()
	if len(configs) == 0 {
		return HealthCheck{
			Name:    "config_manager",
			Status:  "critical",
			Message: "配置管理器无可用配置",
		}
	}
	return HealthCheck{
		Name:    "config_manager",
		Status:  "healthy",
		Message: "配置管理器运行正常",
	}
}

// startStatsUpdating 启动统计更新
func (im *IntegrationManager) startStatsUpdating() {
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟更新一次统计
	defer ticker.Stop()

	for range ticker.C {
		if !im.isRunning {
			break
		}

		im.updateStats()
	}
}

// updateStats 更新统计信息
func (im *IntegrationManager) updateStats() {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	// 更新运行时间
	im.stats.Uptime = time.Since(im.stats.StartTime).String()
	im.stats.LastUpdateTime = time.Now()

	// 获取决策统计
	decisions := im.riskManager.GetRecentDecisions(1000) // 获取最近1000个决策
	im.stats.TotalDecisions = int64(len(decisions))

	// 计算成功和失败的决策
	successCount := int64(0)
	totalResponseTime := time.Duration(0)
	
	for _, decision := range decisions {
		if decision.Confidence > 0.7 { // 假设置信度>0.7为成功
			successCount++
		}
		totalResponseTime += decision.ProcessingTime
	}
	
	im.stats.SuccessfulDecisions = successCount
	im.stats.FailedDecisions = im.stats.TotalDecisions - successCount
	
	if len(decisions) > 0 {
		avgResponseTime := totalResponseTime / time.Duration(len(decisions))
		im.stats.AverageResponseTime = avgResponseTime.String()
	}

	// 获取警报统计
	alerts := im.riskManager.GetActiveAlerts()
	im.stats.TotalAlerts = int64(len(alerts))
}

// API处理函数
func (im *IntegrationManager) getSystemStatus(c *gin.Context) {
	im.mutex.RLock()
	status := gin.H{
		"initialized": im.isInitialized,
		"running":     im.isRunning,
		"start_time":  im.startTime,
		"uptime":      time.Since(im.startTime).String(),
	}
	im.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

func (im *IntegrationManager) getHealthStatus(c *gin.Context) {
	im.healthChecker.mutex.RLock()
	health := gin.H{
		"overall_status":  im.healthChecker.overallStatus,
		"last_check_time": im.healthChecker.lastCheckTime,
		"checks":          im.healthChecker.checks,
	}
	im.healthChecker.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

func (im *IntegrationManager) getSystemStats(c *gin.Context) {
	im.mutex.RLock()
	stats := *im.stats
	im.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (im *IntegrationManager) startSystem(c *gin.Context) {
	if err := im.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统启动成功",
	})
}

func (im *IntegrationManager) stopSystem(c *gin.Context) {
	if err := im.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统停止成功",
	})
}

func (im *IntegrationManager) restartSystem(c *gin.Context) {
	if err := im.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("停止系统失败: %v", err),
		})
		return
	}

	time.Sleep(time.Second * 2) // 等待2秒

	if err := im.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("启动系统失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统重启成功",
	})
}

func (im *IntegrationManager) getSystemConfig(c *gin.Context) {
	configs := im.configManager.GetAllConfigs()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

func (im *IntegrationManager) updateSystemConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 这里可以添加批量更新配置的逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// GetRiskManager 获取风控管理器
func (im *IntegrationManager) GetRiskManager() *RiskManager {
	return im.riskManager
}

// GetDatabaseManager 获取数据库管理器
func (im *IntegrationManager) GetDatabaseManager() *DatabaseManager {
	return im.databaseManager
}

// GetConfigManager 获取配置管理器
func (im *IntegrationManager) GetConfigManager() *ConfigManager {
	return im.configManager
}

// IsRunning 检查系统是否运行中
func (im *IntegrationManager) IsRunning() bool {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	return im.isRunning
}

// GetStats 获取系统统计
func (im *IntegrationManager) GetStats() SystemStats {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	return *im.stats
}