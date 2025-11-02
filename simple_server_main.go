package main

import (
	"fmt"
	"log"
	"net/http"
	"nofx/market"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置为Release模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 提供静态文件服务
	router.StaticFile("/", "./web/dist/index.html")
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/vite.svg", "./web/dist/vite.svg")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 模拟trader列表接口
	router.GET("/api/traders", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{
				"trader_id":   "test_trader",
				"trader_name": "Test Trader",
				"ai_model":    "deepseek",
			},
		})
	})

	// 市场数据接口
	router.GET("/api/market-data", func(c *gin.Context) {
		// 获取symbol参数，默认为BTCUSDT
		symbol := c.DefaultQuery("symbol", "BTCUSDT")

		fmt.Printf("📊 获取市场数据请求: %s\n", symbol)

		// 获取市场数据
		data, err := market.Get(symbol)
		if err != nil {
			fmt.Printf("❌ 获取市场数据失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("获取市场数据失败: %v", err),
			})
			return
		}

		fmt.Printf("✅ 成功获取 %s 市场数据: 价格=%.2f, 1h变化=%.2f%%, 4h变化=%.2f%%, 24h变化=%.2f%%\n",
			data.Symbol, data.CurrentPrice, data.PriceChange1h, data.PriceChange4h, data.PriceChange24h)

		// 返回市场数据
		c.JSON(http.StatusOK, data)
	})

	// 模拟竞赛数据接口
	router.GET("/api/competition", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"total_traders":  1,
			"active_traders": 1,
			"total_volume":   1000000,
		})
	})

	fmt.Println("🚀 简化测试服务器启动在 http://localhost:8080")
	fmt.Println("📊 市场数据接口: http://localhost:8080/api/market-data")
	fmt.Println("👥 Trader列表接口: http://localhost:8080/api/traders")
	fmt.Println("❤️  健康检查: http://localhost:8080/health")
	fmt.Println()

	log.Fatal(router.Run(":8080"))
}
