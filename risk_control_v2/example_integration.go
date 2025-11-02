package risk_control_v2

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ExampleIntegration 展示如何在主程序中集成新的风控系统
func ExampleIntegration() {
	// 1. 创建集成配置
	config := IntegrationConfig{
		DatabasePath:        "./data/risk_control_v2.db",
		ConfigDir:          "./config/risk_control_v2",
		LogLevel:           "info",
		HealthCheckInterval: time.Minute * 2,
		StatsUpdateInterval: time.Minute * 5,
		EnableMetrics:      true,
		EnableProfiling:    false,
	}

	// 2. 初始化集成管理器
	integrationManager, err := NewIntegrationManager(config)
	if err != nil {
		log.Fatalf("初始化集成管理器失败: %v", err)
	}

	// 3. 启动风控系统
	if err := integrationManager.Start(); err != nil {
		log.Fatalf("启动风控系统失败: %v", err)
	}

	// 4. 创建Gin路由器并注册API路由
	router := gin.Default()
	integrationManager.RegisterRoutes(router)

	// 5. 添加中间件（可选）
	router.Use(func(c *gin.Context) {
		// 在这里可以添加风控检查中间件
		// 例如：检查交易请求是否通过风控
		c.Next()
	})

	// 6. 示例：在交易决策中使用风控系统
	router.POST("/api/trade/decision", func(c *gin.Context) {
		var req struct {
			Symbol string  `json:"symbol"`
			Action string  `json:"action"` // buy, sell, hold
			Amount float64 `json:"amount"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 获取风控决策
		riskManager := integrationManager.GetRiskManager()
		riskDecision, err := riskManager.MakeRiskDecision(req.Symbol, "open_long", map[string]interface{}{})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 根据风控决策判断是否允许交易
		if riskDecision.Action == "block" {
			c.JSON(403, gin.H{
				"success": false,
				"message": "交易被风控系统阻止",
				"reason":  riskDecision.Reason,
				"risk_factors": riskDecision.RiskFactors,
			})
			return
		}

		// 如果风控通过，继续执行交易逻辑
		c.JSON(200, gin.H{
			"success": true,
			"message": "交易决策通过风控检查",
			"risk_decision": riskDecision,
		})
	})

	// 7. 示例：更新市场数据并触发风控检查
	router.POST("/api/market/update", func(c *gin.Context) {
		var req struct {
			Symbol        string  `json:"symbol"`
			Price         float64 `json:"price"`
			Volume        float64 `json:"volume"`
			OpenInterest  float64 `json:"open_interest"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		riskManager := integrationManager.GetRiskManager()

		// 更新流动性数据
		err := riskManager.UpdateLiquidity(req.Symbol, req.OpenInterest, req.Volume)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 更新权益数据（假设从账户余额计算）
		// 这里需要根据实际情况获取账户权益
		currentEquity := 1000000.0 // 示例值
		riskManager.UpdateEquity(currentEquity)

		c.JSON(200, gin.H{
			"success": true,
			"message": "市场数据更新成功",
		})
	})

	// 8. 启动HTTP服务器
	log.Printf("风控系统v2启动成功，监听端口 :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("启动HTTP服务器失败: %v", err)
	}

	// 9. 优雅关闭（在实际应用中，这应该在信号处理器中调用）
	defer func() {
		if err := integrationManager.Stop(); err != nil {
			log.Printf("停止风控系统失败: %v", err)
		}
	}()
}

// ExampleRiskControlMiddleware 风控中间件示例
func ExampleRiskControlMiddleware(integrationManager *IntegrationManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对交易相关的API进行风控检查
		if c.Request.URL.Path == "/api/trade/order" || 
		   c.Request.URL.Path == "/api/trade/position" {
			
			// 检查系统是否处于紧急停止状态
			riskManager := integrationManager.GetRiskManager()
			state := riskManager.GetCurrentState()
			
			if state.EmergencyStop {
				c.JSON(503, gin.H{
					"success": false,
					"error":   "系统处于紧急停止状态，暂停所有交易",
				})
				c.Abort()
				return
			}

			// 检查全局风险级别
			if state.GlobalRiskLevel == "critical" {
				c.JSON(429, gin.H{
					"success": false,
					"error":   "系统风险级别过高，限制交易操作",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ExampleConfigUpdate 配置更新示例
func ExampleConfigUpdate(integrationManager *IntegrationManager) {
	configManager := integrationManager.GetConfigManager()

	// 更新时间管理器配置
	timeConfig, _ := configManager.GetTimeManagerConfig()
	timeConfig.DailyResetHour = 8 // 改为北京时间8点重置
	timeConfig.EquityBufferPercent = 0.15 // 降低权益缓冲到15%
	configManager.SetTimeManagerConfig(timeConfig)

	// 更新流动性监控配置
	liquidityConfig, _ := configManager.GetLiquidityMonitorConfig()
	liquidityConfig.LowLiquidityThreshold = 2000000 // 提高低流动性阈值到2M
	liquidityConfig.ForceCloseEnabled = true // 启用强制平仓
	configManager.SetLiquidityMonitorConfig(liquidityConfig)

	// 更新AI委员会配置
	aiConfig, _ := configManager.GetAICommitteeConfig()
	aiConfig.MinConsensusLevel = 0.7 // 提高最小共识水平
	aiConfig.ConservativeMode = true // 启用保守模式
	configManager.SetAICommitteeConfig(aiConfig)

	log.Printf("配置更新完成")
}

// ExampleMonitoring 监控示例
func ExampleMonitoring(integrationManager *IntegrationManager) {
	// 定期检查系统健康状态
	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		defer ticker.Stop()

		for range ticker.C {
			if !integrationManager.IsRunning() {
				continue
			}

			// 获取系统统计
			stats := integrationManager.GetStats()
			log.Printf("系统统计 - 运行时间: %s, 总决策数: %d, 成功率: %.2f%%",
				stats.Uptime,
				stats.TotalDecisions,
				float64(stats.SuccessfulDecisions)/float64(stats.TotalDecisions)*100)

			// 检查风控状态
			riskManager := integrationManager.GetRiskManager()
			riskState := riskManager.GetCurrentState()
			
			if riskState.EmergencyStop {
				log.Printf("警告: 系统处于紧急停止状态")
			}

			if riskState.GlobalRiskLevel == "high" || riskState.GlobalRiskLevel == "critical" {
				log.Printf("警告: 全局风险级别为 %s", riskState.GlobalRiskLevel)
			}

			// 检查活跃警报
			activeAlerts := riskManager.GetActiveAlerts()
			if len(activeAlerts) > 0 {
				log.Printf("当前有 %d 个活跃警报", len(activeAlerts))
				for _, alert := range activeAlerts {
					if alert.Level == "critical" {
						log.Printf("严重警报: %s - %s", alert.Type, alert.Message)
					}
				}
			}
		}
	}()
}

// ExampleDataExport 数据导出示例
func ExampleDataExport(integrationManager *IntegrationManager) {
	dbManager := integrationManager.GetDatabaseManager()

	// 导出最近的风控决策
	decisions, err := dbManager.GetRiskDecisions(100)
	if err != nil {
		log.Printf("获取风控决策失败: %v", err)
		return
	}

	log.Printf("导出了 %d 条风控决策记录", len(decisions))

	// 导出夏普比率记录
	sharpeRecords, err := dbManager.GetSharpeRecords(50)
	if err != nil {
		log.Printf("获取夏普比率记录失败: %v", err)
		return
	}

	log.Printf("导出了 %d 条夏普比率记录", len(sharpeRecords))

	// 导出黑名单
	blacklist, err := dbManager.GetBlacklistEntries()
	if err != nil {
		log.Printf("获取黑名单失败: %v", err)
		return
	}

	log.Printf("当前黑名单包含 %d 个交易对", len(blacklist))
}

// ExampleBackupRestore 备份恢复示例
func ExampleBackupRestore(integrationManager *IntegrationManager) {
	configManager := integrationManager.GetConfigManager()

	// 备份当前配置
	if err := configManager.BackupConfigs(); err != nil {
		log.Printf("备份配置失败: %v", err)
		return
	}

	log.Printf("配置备份完成")

	// 在需要时恢复配置
	// backupPath := "./config/risk_control_v2/backups/config_backup_20240101_120000.json"
	// if err := configManager.RestoreConfigs(backupPath); err != nil {
	//     log.Printf("恢复配置失败: %v", err)
	//     return
	// }
	// log.Printf("配置恢复完成")
}

// ExampleCleanup 数据清理示例
func ExampleCleanup(integrationManager *IntegrationManager) {
	dbManager := integrationManager.GetDatabaseManager()

	// 清理30天前的旧数据
	if err := dbManager.CleanOldRecords(30); err != nil {
		log.Printf("清理旧数据失败: %v", err)
		return
	}

	log.Printf("数据清理完成")
}