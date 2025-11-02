package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/decision"
	"nofx/market"
	"nofx/mcp"
	"nofx/pool"
	"strings"
	"time"
)

func demoMain() {
	fmt.Println("=== AI交易决策演示模式 ===")
	fmt.Println("这是一个模拟模式，用于观察AI的决策过程")
	fmt.Println("不会执行真实交易，仅显示AI的分析和决策")
	fmt.Println()

	// 加载配置
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 找到启用的Binance trader
	var traderCfg *config.TraderConfig
	for _, trader := range cfg.Traders {
		if trader.Enabled && strings.Contains(trader.Name, "Binance") {
			traderCfg = &trader
			break
		}
	}

	if traderCfg == nil {
		log.Fatal("未找到启用的Binance trader")
	}

	fmt.Printf("使用交易者: %s\n", traderCfg.Name)
	fmt.Printf("AI模型: %s\n", traderCfg.AIModel)
	fmt.Printf("扫描间隔: %d分钟\n", traderCfg.ScanIntervalMinutes)
	fmt.Println()

	// 初始化MCP客户端
	mcpClient := mcp.New()
	
	// 根据AI模型设置客户端
	switch traderCfg.AIModel {
	case "deepseek":
		mcpClient.SetDeepSeekAPIKey(traderCfg.DeepSeekKey)
	case "qwen":
		mcpClient.SetQwenAPIKey(traderCfg.QwenKey, "")
	default:
		if traderCfg.CustomAPIURL != "" {
			mcpClient.SetCustomAPI(traderCfg.CustomAPIURL, traderCfg.CustomAPIKey, traderCfg.CustomModelName)
		} else {
			log.Fatal("未知的AI模型或未配置自定义API")
		}
	}

	// 设置币种池配置
	pool.SetCoinPoolAPI(cfg.CoinPoolAPIURL)
	pool.SetOITopAPI(cfg.OITopAPIURL)
	pool.SetUseDefaultCoins(cfg.UseDefaultCoins)
	pool.SetDefaultCoins(cfg.DefaultCoins)

	// 创建模拟的交易上下文
	ctx := createMockContext(traderCfg, cfg)

	// 运行AI决策循环
	for i := 1; i <= 3; i++ {
		fmt.Printf("=== AI决策循环 #%d ===\n", i)
		
		// 更新上下文
		ctx.CallCount = i
		ctx.RuntimeMinutes = i * 5 // 模拟运行时间

		// 获取AI决策
		fullDecision, err := decision.GetFullDecision(ctx, mcpClient)
		if err != nil {
			fmt.Printf("❌ AI决策失败: %v\n", err)
		} else {
			displayDecision(fullDecision)
		}

		if i < 3 {
			fmt.Println("\n等待下一次决策...")
			time.Sleep(10 * time.Second)
		}
	}

	fmt.Println("\n=== 演示结束 ===")
	fmt.Println("如需执行真实交易，请修复API权限问题后重启系统")
}

func createMockContext(traderCfg *config.TraderConfig, cfg *config.Config) *decision.Context {
	// 创建模拟账户信息
	account := decision.AccountInfo{
		TotalEquity:      1000.0,
		AvailableBalance: 800.0,
		TotalPnL:         50.0,
		TotalPnLPct:      5.0,
		MarginUsed:       200.0,
		MarginUsedPct:    20.0,
		PositionCount:    1,
	}

	// 创建模拟持仓
	positions := []decision.PositionInfo{
		{
			Symbol:           "BTCUSDT",
			Side:             "long",
			EntryPrice:       45000.0,
			MarkPrice:        46000.0,
			Quantity:         0.01,
			Leverage:         10,
			UnrealizedPnL:    10.0,
			UnrealizedPnLPct: 2.22,
			LiquidationPrice: 40500.0,
			MarginUsed:       46.0,
			UpdateTime:       time.Now().UnixMilli(),
		},
	}

	// 获取候选币种
	candidateCoins := []decision.CandidateCoin{}
	
	// 从默认币种获取候选币种
	for i, coin := range cfg.DefaultCoins {
		if i >= 10 { // 限制数量
			break
		}
		candidateCoins = append(candidateCoins, decision.CandidateCoin{
			Symbol:  coin,
			Sources: []string{"default"},
		})
	}

	return &decision.Context{
		CurrentTime:     time.Now().Format("2006-01-02 15:04:05"),
		RuntimeMinutes:  0,
		CallCount:       0,
		Account:         account,
		Positions:       positions,
		CandidateCoins:  candidateCoins,
		MarketDataMap:   make(map[string]*market.Data),
		OITopDataMap:    make(map[string]*decision.OITopData),
		BTCETHLeverage:  cfg.Leverage.BTCETHLeverage,
		AltcoinLeverage: cfg.Leverage.AltcoinLeverage,
	}
}

func displayDecision(fullDecision *decision.FullDecision) {
	fmt.Printf("⏰ 决策时间: %s\n", fullDecision.Timestamp.Format("15:04:05"))
	
	if fullDecision.CoTTrace != "" {
		fmt.Println("\n🧠 AI思维过程:")
		fmt.Println(strings.Repeat("-", 50))
		// 只显示前500个字符，避免输出过长
		trace := fullDecision.CoTTrace
		if len(trace) > 500 {
			trace = trace[:500] + "..."
		}
		fmt.Println(trace)
		fmt.Println(strings.Repeat("-", 50))
	}

	fmt.Printf("\n📊 决策数量: %d\n", len(fullDecision.Decisions))
	
	for i, d := range fullDecision.Decisions {
		fmt.Printf("\n决策 #%d:\n", i+1)
		fmt.Printf("  币种: %s\n", d.Symbol)
		fmt.Printf("  动作: %s\n", d.Action)
		
		if d.Action != "hold" && d.Action != "wait" {
			if d.Leverage > 0 {
				fmt.Printf("  杠杆: %dx\n", d.Leverage)
			}
			if d.PositionSizeUSD > 0 {
				fmt.Printf("  仓位大小: $%.2f\n", d.PositionSizeUSD)
			}
			if d.StopLoss > 0 {
				fmt.Printf("  止损: $%.2f\n", d.StopLoss)
			}
			if d.TakeProfit > 0 {
				fmt.Printf("  止盈: $%.2f\n", d.TakeProfit)
			}
			if d.RiskUSD > 0 {
				fmt.Printf("  风险: $%.2f\n", d.RiskUSD)
			}
		}
		
		if d.Confidence > 0 {
			fmt.Printf("  信心度: %d%%\n", d.Confidence)
		}
		
		if d.Reasoning != "" {
			fmt.Printf("  理由: %s\n", d.Reasoning)
		}
	}
}