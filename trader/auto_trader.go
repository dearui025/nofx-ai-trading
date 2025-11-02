package trader

import (
	"encoding/json"
	"fmt"
	"log"
	"nofx/decision"
	"nofx/logger"
	"nofx/market"
	"nofx/mcp"
	"nofx/pool"
	"strings"
	"time"
)

// AutoTraderConfig 自动交易配置（简化版 - AI全权决策）
type AutoTraderConfig struct {
	// Trader标识
	ID      string // Trader唯一标识（用于日志目录等）
	Name    string // Trader显示名称
	AIModel string // AI模型: "qwen" 或 "deepseek"

	// 交易平台选择
	Exchange string // "binance", "hyperliquid" 或 "aster"

	// 币安API配置
	BinanceAPIKey    string
	BinanceSecretKey string
	BinanceTestnet   bool

	// Hyperliquid配置
	HyperliquidPrivateKey string
	HyperliquidWalletAddr string
	HyperliquidTestnet    bool

	// Aster配置
	AsterUser       string // Aster主钱包地址
	AsterSigner     string // Aster API钱包地址
	AsterPrivateKey string // Aster API钱包私钥

	CoinPoolAPIURL string

	// AI配置
	UseQwen     bool
	DeepSeekKey string
	QwenKey     string

	// 自定义AI API配置
	CustomAPIURL    string
	CustomAPIKey    string
	CustomModelName string

	// 扫描配置
	ScanInterval time.Duration // 扫描间隔（建议3分钟）

	// 账户配置
	InitialBalance float64 // 初始金额（用于计算盈亏，需手动设置）

	// 杠杆配置
	BTCETHLeverage  int // BTC和ETH的杠杆倍数
	AltcoinLeverage int // 山寨币的杠杆倍数

	// 风险控制（仅作为提示，AI可自主决定）
	MaxDailyLoss    float64       // 最大日亏损百分比（提示）
	MaxDrawdown     float64       // 最大回撤百分比（提示）
	StopTradingTime time.Duration // 触发风控后暂停时长
}

// AutoTrader 自动交易器
type AutoTrader struct {
	id                    string // Trader唯一标识
	name                  string // Trader显示名称
	aiModel               string // AI模型名称
	exchange              string // 交易平台名称
	config                AutoTraderConfig
	trader                Trader // 使用Trader接口（支持多平台）
	mcpClient             *mcp.Client
	decisionLogger        *logger.DecisionLogger // 决策日志记录器
	frequencyManager      *FrequencyManager      // 频率管理器
	initialBalance        float64
	dailyPnL              float64
	lastResetTime         time.Time
	stopUntil             time.Time
	isRunning             bool
	startTime             time.Time        // 系统启动时间
	callCount             int              // AI调用次数
	positionFirstSeenTime map[string]int64 // 持仓首次出现时间 (symbol_side -> timestamp毫秒)
	positionPeakProfit    map[string]float64 // 持仓浮盈峰值 (symbol_side -> 峰值盈亏百分比)
}

// NewAutoTrader 创建自动交易器
func NewAutoTrader(config AutoTraderConfig) (*AutoTrader, error) {
	// 设置默认值
	if config.ID == "" {
		config.ID = "default_trader"
	}
	if config.Name == "" {
		config.Name = "Default Trader"
	}
	if config.AIModel == "" {
		if config.UseQwen {
			config.AIModel = "qwen"
		} else {
			config.AIModel = "deepseek"
		}
	}

	mcpClient := mcp.New()

	// 初始化AI
	if config.AIModel == "custom" {
		// 使用自定义API
		mcpClient.SetCustomAPI(config.CustomAPIURL, config.CustomAPIKey, config.CustomModelName)
		log.Printf("🤖 [%s] 使用自定义AI API: %s (模型: %s)", config.Name, config.CustomAPIURL, config.CustomModelName)
	} else if config.UseQwen || config.AIModel == "qwen" {
		// 使用Qwen
		mcpClient.SetQwenAPIKey(config.QwenKey, "")
		log.Printf("🤖 [%s] 使用阿里云Qwen AI", config.Name)
	} else {
		// 默认使用DeepSeek
		mcpClient.SetDeepSeekAPIKey(config.DeepSeekKey)
		log.Printf("🤖 [%s] 使用DeepSeek AI", config.Name)
	}

	// 初始化币种池API
	if config.CoinPoolAPIURL != "" {
		pool.SetCoinPoolAPI(config.CoinPoolAPIURL)
	}

	// 设置默认交易平台
	if config.Exchange == "" {
		config.Exchange = "binance"
	}

	// 根据配置创建对应的交易器
	var trader Trader
	var err error

	switch config.Exchange {
	case "binance":
		log.Printf("🏦 [%s] 使用币安合约交易", config.Name)
		trader = NewFuturesTrader(config.BinanceAPIKey, config.BinanceSecretKey, config.BinanceTestnet)
	case "hyperliquid":
		log.Printf("🏦 [%s] 使用Hyperliquid交易", config.Name)
		trader, err = NewHyperliquidTrader(config.HyperliquidPrivateKey, config.HyperliquidWalletAddr, config.HyperliquidTestnet)
		if err != nil {
			return nil, fmt.Errorf("初始化Hyperliquid交易器失败: %w", err)
		}
	case "aster":
		log.Printf("🏦 [%s] 使用Aster交易", config.Name)
		trader, err = NewAsterTrader(config.AsterUser, config.AsterSigner, config.AsterPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("初始化Aster交易器失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的交易平台: %s", config.Exchange)
	}

	// 验证初始金额配置
	if config.InitialBalance <= 0 {
		return nil, fmt.Errorf("初始金额必须大于0，请在配置中设置InitialBalance")
	}

	// 初始化决策日志记录器（使用trader ID创建独立目录）
	logDir := fmt.Sprintf("decision_logs/%s", config.ID)
	decisionLogger := logger.NewDecisionLogger(logDir)

	// 初始化频率管理器
	frequencyStateFile := fmt.Sprintf("data/frequency_state_%s.json", config.ID)
	frequencyManager := NewFrequencyManager(frequencyStateFile)
	frequencyManager.UpdateAccountEquity(config.InitialBalance)

	log.Printf("⚡ [%s] 频率管理器已初始化 - 模式: %s", config.Name, frequencyManager.CurrentMode)

	return &AutoTrader{
		id:                    config.ID,
		name:                  config.Name,
		aiModel:               config.AIModel,
		exchange:              config.Exchange,
		config:                config,
		trader:                trader,
		mcpClient:             mcpClient,
		decisionLogger:        decisionLogger,
		frequencyManager:      frequencyManager,
		initialBalance:        config.InitialBalance,
		lastResetTime:         time.Now(),
		startTime:             time.Now(),
		callCount:             0,
		isRunning:             false,
		positionFirstSeenTime: make(map[string]int64),
		positionPeakProfit:    make(map[string]float64),
	}, nil
}

// Run 运行自动交易主循环
func (at *AutoTrader) Run() error {
	at.isRunning = true
	log.Println("🚀 AI驱动自动交易系统启动")
	log.Printf("💰 初始余额: %.2f USDT", at.initialBalance)
	log.Printf("⚙️  扫描间隔: %v", at.config.ScanInterval)
	log.Println("🤖 AI将全权决定杠杆、仓位大小、止损止盈等参数")

	ticker := time.NewTicker(at.config.ScanInterval)
	defer ticker.Stop()

	// 首次立即执行
	if err := at.runCycle(); err != nil {
		log.Printf("❌ 执行失败: %v", err)
	}

	for at.isRunning {
		select {
		case <-ticker.C:
			if err := at.runCycle(); err != nil {
				log.Printf("❌ 执行失败: %v", err)
			}
		}
	}

	return nil
}

// Stop 停止自动交易
func (at *AutoTrader) Stop() {
	at.isRunning = false
	log.Println("⏹ 自动交易系统停止")
}

// runCycle 运行一个交易周期（使用AI全权决策）
func (at *AutoTrader) runCycle() error {
	at.callCount++

	log.Printf("\n" + strings.Repeat("=", 70))
	log.Printf("⏰ %s - AI决策周期 #%d", time.Now().Format("2006-01-02 15:04:05"), at.callCount)
	log.Printf(strings.Repeat("=", 70))

	// 创建决策记录
	record := &logger.DecisionRecord{
		ExecutionLog: []string{},
		Success:      true,
	}

	// 1. 检查是否需要停止交易
	if time.Now().Before(at.stopUntil) {
		remaining := at.stopUntil.Sub(time.Now())
		log.Printf("⏸ 风险控制：暂停交易中，剩余 %.0f 分钟", remaining.Minutes())
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("风险控制暂停中，剩余 %.0f 分钟", remaining.Minutes())
		at.decisionLogger.LogDecision(record)
		return nil
	}

	// 2. 重置日盈亏（每天重置）
	if time.Since(at.lastResetTime) > 24*time.Hour {
		at.dailyPnL = 0
		at.lastResetTime = time.Now()
		log.Println("📅 日盈亏已重置")
	}

	// 3. 收集交易上下文
	ctx, err := at.buildTradingContext()
	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("构建交易上下文失败: %v", err)
		at.decisionLogger.LogDecision(record)
		return fmt.Errorf("构建交易上下文失败: %w", err)
	}

	// 保存账户状态快照
	record.AccountState = logger.AccountSnapshot{
		TotalBalance:          ctx.Account.TotalEquity,
		AvailableBalance:      ctx.Account.AvailableBalance,
		TotalUnrealizedProfit: ctx.Account.TotalPnL,
		PositionCount:         ctx.Account.PositionCount,
		MarginUsedPct:         ctx.Account.MarginUsedPct,
	}

	// 保存持仓快照
	for _, pos := range ctx.Positions {
		record.Positions = append(record.Positions, logger.PositionSnapshot{
			Symbol:           pos.Symbol,
			Side:             pos.Side,
			PositionAmt:      pos.Quantity,
			EntryPrice:       pos.EntryPrice,
			MarkPrice:        pos.MarkPrice,
			UnrealizedProfit: pos.UnrealizedPnL,
			Leverage:         float64(pos.Leverage),
			LiquidationPrice: pos.LiquidationPrice,
		})
	}

	// 保存候选币种列表
	for _, coin := range ctx.CandidateCoins {
		record.CandidateCoins = append(record.CandidateCoins, coin.Symbol)
	}

	log.Printf("📊 账户净值: %.2f USDT | 可用: %.2f USDT | 持仓: %d",
		ctx.Account.TotalEquity, ctx.Account.AvailableBalance, ctx.Account.PositionCount)

	// 4. 更新频率管理器状态
	at.frequencyManager.UpdateAccountEquity(ctx.Account.TotalEquity)
	
	// 检查并更新频率模式
	if switched, msg := at.frequencyManager.UpdateFrequencyMode(); switched {
		log.Printf("🔄 [频率模式切换] %s", msg)
		record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("🔄 频率模式切换: %s", msg))
	}

	// 5. 调用AI获取完整决策
	log.Println("🤖 正在请求AI分析并决策...")
	decision, err := decision.GetFullDecision(ctx, at.mcpClient)

	// 即使有错误，也保存思维链、决策和输入prompt（用于debug）
	if decision != nil {
		record.InputPrompt = decision.UserPrompt
		record.CoTTrace = decision.CoTTrace
		if len(decision.Decisions) > 0 {
			decisionJSON, _ := json.MarshalIndent(decision.Decisions, "", "  ")
			record.DecisionJSON = string(decisionJSON)
		}
	}

	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("获取AI决策失败: %v", err)

		// 打印AI思维链（即使有错误）
		if decision != nil && decision.CoTTrace != "" {
			log.Printf("\n" + strings.Repeat("-", 70))
			log.Println("💭 AI思维链分析（错误情况）:")
			log.Println(strings.Repeat("-", 70))
			log.Println(decision.CoTTrace)
			log.Printf(strings.Repeat("-", 70) + "\n")
		}

		at.decisionLogger.LogDecision(record)
		return fmt.Errorf("获取AI决策失败: %w", err)
	}

	// 5. 打印AI思维链
	log.Printf("\n" + strings.Repeat("-", 70))
	log.Println("💭 AI思维链分析:")
	log.Println(strings.Repeat("-", 70))
	log.Println(decision.CoTTrace)
	log.Printf(strings.Repeat("-", 70) + "\n")

	// 6. 打印AI决策
	log.Printf("📋 AI决策列表 (%d 个):\n", len(decision.Decisions))
	for i, d := range decision.Decisions {
		log.Printf("  [%d] %s: %s - %s", i+1, d.Symbol, d.Action, d.Reasoning)
		if d.Action == "open_long" || d.Action == "open_short" {
			log.Printf("      杠杆: %dx | 仓位: %.2f USDT | 止损: %.4f | 止盈: %.4f",
				d.Leverage, d.PositionSizeUSD, d.StopLoss, d.TakeProfit)
		}
	}
	log.Println()

	// 7. 对决策排序：确保先平仓后开仓（防止仓位叠加超限）
	sortedDecisions := sortDecisionsByPriority(decision.Decisions)

	log.Println("🔄 执行顺序（已优化）: 先平仓→后开仓")
	for i, d := range sortedDecisions {
		log.Printf("  [%d] %s %s", i+1, d.Symbol, d.Action)
	}
	log.Println()

	// 执行决策并记录结果
	for _, d := range sortedDecisions {
		actionRecord := logger.DecisionAction{
			Action:    d.Action,
			Symbol:    d.Symbol,
			Quantity:  0,
			Leverage:  d.Leverage,
			Price:     0,
			Timestamp: time.Now(),
			Success:   false,
		}

		// 对于开仓操作，检查频率限制
		if d.Action == "open_long" || d.Action == "open_short" {
			if allowed, reason := at.frequencyManager.CheckTradeAllowance(); !allowed {
				log.Printf("🚫 频率限制阻止开仓 (%s %s): %s", d.Symbol, d.Action, reason)
				actionRecord.Error = reason
				record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("🚫 %s %s 被频率限制阻止: %s", d.Symbol, d.Action, reason))
				record.Decisions = append(record.Decisions, actionRecord)
				continue
			}
		}

		if err := at.executeDecisionWithRecord(&d, &actionRecord); err != nil {
			log.Printf("❌ 执行决策失败 (%s %s): %v", d.Symbol, d.Action, err)
			actionRecord.Error = err.Error()
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("❌ %s %s 失败: %v", d.Symbol, d.Action, err))
		} else {
			actionRecord.Success = true
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("✓ %s %s 成功", d.Symbol, d.Action))
			
			// 如果是成功的开仓操作，增加交易计数
			if d.Action == "open_long" || d.Action == "open_short" {
				at.frequencyManager.IncrementTradeCount()
			}
			
			// 成功执行后短暂延迟
			time.Sleep(1 * time.Second)
		}

		record.Decisions = append(record.Decisions, actionRecord)
	}

	// 8. 保存决策记录
	if err := at.decisionLogger.LogDecision(record); err != nil {
		log.Printf("⚠ 保存决策记录失败: %v", err)
	}

	// 9. 保存频率管理器状态
	if err := at.frequencyManager.SaveState(); err != nil {
		log.Printf("⚠ 保存频率管理器状态失败: %v", err)
	}

	return nil
}

// buildTradingContext 构建交易上下文
func (at *AutoTrader) buildTradingContext() (*decision.Context, error) {
	// 1. 获取账户信息
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("获取账户余额失败: %w", err)
	}

	// 获取账户字段
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Total Equity = 钱包余额 + 未实现盈亏
	totalEquity := totalWalletBalance + totalUnrealizedProfit

	// 2. 获取持仓信息
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var positionInfos []decision.PositionInfo
	totalMarginUsed := 0.0

	// 当前持仓的key集合（用于清理已平仓的记录）
	currentPositionKeys := make(map[string]bool)

	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity // 空仓数量为负，转为正数
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		// 计算占用保证金（估算）
		leverage := 10 // 默认值，实际应该从持仓信息获取
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed

		// 计算盈亏百分比
		pnlPct := 0.0
		if side == "long" {
			pnlPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
		} else {
			pnlPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
		}

		// 跟踪持仓首次出现时间
		posKey := symbol + "_" + side
		currentPositionKeys[posKey] = true
		if _, exists := at.positionFirstSeenTime[posKey]; !exists {
			// 新持仓，记录当前时间
			at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()
			// 初始化浮盈峰值
			at.positionPeakProfit[posKey] = pnlPct
		} else {
			// 更新浮盈峰值（只记录正向峰值）
			if pnlPct > 0 && pnlPct > at.positionPeakProfit[posKey] {
				at.positionPeakProfit[posKey] = pnlPct
			}
		}
		updateTime := at.positionFirstSeenTime[posKey]

		positionInfos = append(positionInfos, decision.PositionInfo{
			Symbol:           symbol,
			Side:             side,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			Quantity:         quantity,
			Leverage:         leverage,
			UnrealizedPnL:    unrealizedPnl,
			UnrealizedPnLPct: pnlPct,
			LiquidationPrice: liquidationPrice,
			MarginUsed:       marginUsed,
			UpdateTime:       updateTime,
		})
	}

	// 清理已平仓的持仓记录
	for key := range at.positionFirstSeenTime {
		if !currentPositionKeys[key] {
			delete(at.positionFirstSeenTime, key)
			delete(at.positionPeakProfit, key)
		}
	}

	// 3. 获取合并的候选币种池（AI500 + OI Top，去重）
	// 无论有没有持仓，都分析相同数量的币种（让AI看到所有好机会）
	// AI会根据保证金使用率和现有持仓情况，自己决定是否要换仓
	const ai500Limit = 20 // AI500取前20个评分最高的币种

	// 获取合并后的币种池（AI500 + OI Top）
	mergedPool, err := pool.GetMergedCoinPool(ai500Limit)
	if err != nil {
		return nil, fmt.Errorf("获取合并币种池失败: %w", err)
	}

	// 构建候选币种列表（包含来源信息）
	var candidateCoins []decision.CandidateCoin
	for _, symbol := range mergedPool.AllSymbols {
		sources := mergedPool.SymbolSources[symbol]
		candidateCoins = append(candidateCoins, decision.CandidateCoin{
			Symbol:  symbol,
			Sources: sources, // "ai500" 和/或 "oi_top"
		})
	}

	log.Printf("📋 合并币种池: AI500前%d + OI_Top20 = 总计%d个候选币种",
		ai500Limit, len(candidateCoins))

	// 4. 计算总盈亏
	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	// 5. 分析历史表现（最近100个周期，避免长期持仓的交易记录丢失）
	// 假设每3分钟一个周期，100个周期 = 5小时，足够覆盖大部分交易
	performance, err := at.decisionLogger.AnalyzePerformance(100)
	if err != nil {
		log.Printf("⚠️  分析历史表现失败: %v", err)
		// 不影响主流程，继续执行（但设置performance为nil以避免传递错误数据）
		performance = nil
	}

	// 6. 构建上下文
	ctx := &decision.Context{
		CurrentTime:     time.Now().Format("2006-01-02 15:04:05"),
		RuntimeMinutes:  int(time.Since(at.startTime).Minutes()),
		CallCount:       at.callCount,
		BTCETHLeverage:  at.config.BTCETHLeverage,  // 使用配置的杠杆倍数
		AltcoinLeverage: at.config.AltcoinLeverage, // 使用配置的杠杆倍数
		Account: decision.AccountInfo{
			TotalEquity:      totalEquity,
			AvailableBalance: availableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			MarginUsed:       totalMarginUsed,
			MarginUsedPct:    marginUsedPct,
			PositionCount:    len(positionInfos),
		},
		Positions:      positionInfos,
		CandidateCoins: candidateCoins,
		Performance:    performance, // 添加历史表现分析
	}

	return ctx, nil
}

// executeDecisionWithRecord 执行AI决策并记录详细信息
func (at *AutoTrader) executeDecisionWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	switch decision.Action {
	case "open_long":
		return at.executeOpenLongWithRecord(decision, actionRecord)
	case "open_short":
		return at.executeOpenShortWithRecord(decision, actionRecord)
	case "close_long":
		return at.executeCloseLongWithRecord(decision, actionRecord)
	case "close_short":
		return at.executeCloseShortWithRecord(decision, actionRecord)
	case "hold", "wait":
		// 无需执行，仅记录
		return nil
	default:
		return fmt.Errorf("未知的action: %s", decision.Action)
	}
}

// executeOpenLongWithRecord 执行开多仓并记录详细信息
func (at *AutoTrader) executeOpenLongWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  📈 开多仓: %s", decision.Symbol)

	// ⚠️ 关键：检查是否已有同币种同方向持仓，如果有则拒绝开仓（防止仓位叠加超限）
	positions, err := at.trader.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == decision.Symbol && pos["side"] == "long" {
				return fmt.Errorf("❌ %s 已有多仓，拒绝开仓以防止仓位叠加超限。如需换仓，请先给出 close_long 决策", decision.Symbol)
			}
		}
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// 计算数量
	quantity := decision.PositionSizeUSD / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// 开仓
	order, err := at.trader.OpenLong(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 开仓成功，订单ID: %v, 数量: %.4f", order["orderId"], quantity)

	// 记录开仓时间
	posKey := decision.Symbol + "_long"
	at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()

	// 设置止损止盈
	if err := at.trader.SetStopLoss(decision.Symbol, "LONG", quantity, decision.StopLoss); err != nil {
		log.Printf("  ⚠ 设置止损失败: %v", err)
	}
	if err := at.trader.SetTakeProfit(decision.Symbol, "LONG", quantity, decision.TakeProfit); err != nil {
		log.Printf("  ⚠ 设置止盈失败: %v", err)
	}

	return nil
}

// executeOpenShortWithRecord 执行开空仓并记录详细信息
func (at *AutoTrader) executeOpenShortWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  📉 开空仓: %s", decision.Symbol)

	// ⚠️ 关键：检查是否已有同币种同方向持仓，如果有则拒绝开仓（防止仓位叠加超限）
	positions, err := at.trader.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == decision.Symbol && pos["side"] == "short" {
				return fmt.Errorf("❌ %s 已有空仓，拒绝开仓以防止仓位叠加超限。如需换仓，请先给出 close_short 决策", decision.Symbol)
			}
		}
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// 计算数量
	quantity := decision.PositionSizeUSD / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// 开仓
	order, err := at.trader.OpenShort(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 开仓成功，订单ID: %v, 数量: %.4f", order["orderId"], quantity)

	// 记录开仓时间
	posKey := decision.Symbol + "_short"
	at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()

	// 设置止损止盈
	if err := at.trader.SetStopLoss(decision.Symbol, "SHORT", quantity, decision.StopLoss); err != nil {
		log.Printf("  ⚠ 设置止损失败: %v", err)
	}
	if err := at.trader.SetTakeProfit(decision.Symbol, "SHORT", quantity, decision.TakeProfit); err != nil {
		log.Printf("  ⚠ 设置止盈失败: %v", err)
	}

	return nil
}

// executeCloseLongWithRecord 执行平多仓并记录详细信息
func (at *AutoTrader) executeCloseLongWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  🔄 平多仓: %s", decision.Symbol)

	// 检查强制观察期和动态退出条件
	posKey := decision.Symbol + "_long"
	if !at.canClosePosition(posKey, decision) {
		return fmt.Errorf("❌ %s 处于强制观察期内，未触发止损条件，拒绝平仓", decision.Symbol)
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}
	actionRecord.Price = marketData.CurrentPrice

	// 平仓
	order, err := at.trader.CloseLong(decision.Symbol, 0) // 0 = 全部平仓
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 平仓成功")
	return nil
}

// executeCloseShortWithRecord 执行平空仓并记录详细信息
func (at *AutoTrader) executeCloseShortWithRecord(decision *decision.Decision, actionRecord *logger.DecisionAction) error {
	log.Printf("  🔄 平空仓: %s", decision.Symbol)

	// 检查强制观察期和动态退出条件
	posKey := decision.Symbol + "_short"
	if !at.canClosePosition(posKey, decision) {
		return fmt.Errorf("❌ %s 处于强制观察期内，未触发止损条件，拒绝平仓", decision.Symbol)
	}

	// 获取当前价格
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}
	actionRecord.Price = marketData.CurrentPrice

	// 平仓
	order, err := at.trader.CloseShort(decision.Symbol, 0) // 0 = 全部平仓
	if err != nil {
		return err
	}

	// 记录订单ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	log.Printf("  ✓ 平仓成功")
	return nil
}

// GetID 获取trader ID
func (at *AutoTrader) GetID() string {
	return at.id
}

// GetName 获取trader名称
func (at *AutoTrader) GetName() string {
	return at.name
}

// GetAIModel 获取AI模型
func (at *AutoTrader) GetAIModel() string {
	return at.aiModel
}

// GetDecisionLogger 获取决策日志记录器
func (at *AutoTrader) GetDecisionLogger() *logger.DecisionLogger {
	return at.decisionLogger
}

// canClosePosition 检查是否可以平仓（优化：移除观察期限制，实现纯条件驱动退出）
func (at *AutoTrader) canClosePosition(posKey string, decision *decision.Decision) bool {
	// 获取持仓开始时间（仅用于统计）
	firstSeenTime, exists := at.positionFirstSeenTime[posKey]
	if !exists {
		log.Printf("⚠ 未找到持仓记录: %s", posKey)
		return true // 如果没有记录，允许平仓
	}

	// 计算持仓时长（分钟）- 仅用于日志记录
	holdingTimeMs := time.Now().UnixMilli() - firstSeenTime
	holdingTimeMinutes := float64(holdingTimeMs) / (1000 * 60)

	log.Printf("📊 %s 持仓时长: %.1f分钟", posKey, holdingTimeMinutes)

	// 优化：移除强制观察期限制，实现纯条件驱动退出
	log.Printf("🎯 %s 纯条件驱动退出模式：检查止盈止损条件", posKey)

	// 获取当前持仓信息以计算浮盈
	positions, err := at.trader.GetPositions()
	if err != nil {
		log.Printf("⚠ 获取持仓信息失败: %v", err)
		return true // 如果无法获取持仓信息，允许平仓
	}

	// 查找对应的持仓
	var currentPnlPct float64
	var entryPrice float64
	found := false
	side := "long"
	if strings.HasSuffix(posKey, "_short") {
		side = "short"
	}

	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol && pos["side"] == side {
			if pnl, ok := pos["unrealizedPnl"].(float64); ok {
				if notional, ok := pos["notional"].(float64); ok && notional != 0 {
					currentPnlPct = (pnl / notional) * 100
					if entry, ok := pos["entryPrice"].(float64); ok {
						entryPrice = entry
					}
					found = true
					break
				}
			}
		}
	}

	if !found {
		log.Printf("⚠ 未找到对应持仓信息: %s", posKey)
		return true // 如果找不到持仓，允许平仓
	}

	// 获取市场数据用于ATR计算
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		log.Printf("⚠️ 无法获取%s市场数据，使用默认止损止盈", decision.Symbol)
		// 使用默认固定止损止盈
		if currentPnlPct >= 0.5 {
			log.Printf("✅ %s 达到默认止盈目标: %.2f%% >= 0.5%%", posKey, currentPnlPct)
			return true
		}
		if currentPnlPct <= -2.0 {
			log.Printf("🛑 %s 达到默认止损线: %.2f%% <= -2.0%%", posKey, currentPnlPct)
			return true
		}
		return true
	}

	// 基于ATR的动态止损止盈系统
	atr14 := marketData.LongerTermContext.ATR14
	currentPrice := marketData.CurrentPrice
	
	// 计算基于ATR的止损止盈距离
	atrMultiplier := 2.0 // 2倍ATR作为止损距离
	stopLossDistance := atr14 * atrMultiplier
	takeProfitDistance := stopLossDistance * 1.5 // 1.5倍风险回报比

	var stopLossPrice, takeProfitPrice float64
	var stopLossPct, takeProfitPct float64

	if side == "long" {
		stopLossPrice = entryPrice - stopLossDistance
		takeProfitPrice = entryPrice + takeProfitDistance
		stopLossPct = ((stopLossPrice - entryPrice) / entryPrice) * 100
		takeProfitPct = ((takeProfitPrice - entryPrice) / entryPrice) * 100
		
		// 检查止损条件
		if currentPrice <= stopLossPrice {
			log.Printf("🛑 %s 多头ATR止损: 当前价%.2f <= 止损价%.2f (ATR14=%.2f)", 
				posKey, currentPrice, stopLossPrice, atr14)
			return true
		}
		
		// 检查止盈条件
		if currentPrice >= takeProfitPrice {
			log.Printf("✅ %s 多头ATR止盈: 当前价%.2f >= 止盈价%.2f (ATR14=%.2f)", 
				posKey, currentPrice, takeProfitPrice, atr14)
			return true
		}
	} else {
		stopLossPrice = entryPrice + stopLossDistance
		takeProfitPrice = entryPrice - takeProfitDistance
		stopLossPct = ((entryPrice - stopLossPrice) / entryPrice) * 100
		takeProfitPct = ((entryPrice - takeProfitPrice) / entryPrice) * 100
		
		// 检查止损条件
		if currentPrice >= stopLossPrice {
			log.Printf("🛑 %s 空头ATR止损: 当前价%.2f >= 止损价%.2f (ATR14=%.2f)", 
				posKey, currentPrice, stopLossPrice, atr14)
			return true
		}
		
		// 检查止盈条件
		if currentPrice <= takeProfitPrice {
			log.Printf("✅ %s 空头ATR止盈: 当前价%.2f <= 止盈价%.2f (ATR14=%.2f)", 
				posKey, currentPrice, takeProfitPrice, atr14)
			return true
		}
	}

	log.Printf("📊 %s ATR动态风控: 止损%.2f%% 止盈%.2f%% (ATR14=%.2f)", 
		posKey, stopLossPct, takeProfitPct, atr14)

	// 分批止盈策略 - 在达到ATR止盈前进行分层止盈
	partialProfitTriggered := at.checkPartialProfitTargets(posKey, currentPnlPct, entryPrice, currentPrice, side)
	if partialProfitTriggered {
		return true
	}

	// 3. 趋势反转立即退出
	if at.isTrendReversed(decision.Symbol, side, entryPrice) {
		log.Printf("🔄 %s 趋势反转，立即退出", posKey)
		return true
	}

	// 4. RSI超买/超卖检查
	if at.checkRSIExtremeConditions(decision.Symbol, side, currentPnlPct) {
		log.Printf("📊 %s RSI极端条件触发，建议退出", posKey)
		return true
	}

	// 更新浮盈峰值（用于跟踪止盈）
	peakProfit, exists := at.positionPeakProfit[posKey]
	if !exists || currentPnlPct > peakProfit {
		at.positionPeakProfit[posKey] = currentPnlPct
		peakProfit = currentPnlPct
		log.Printf("📈 %s 更新浮盈峰值: %.2f%%", posKey, peakProfit)
	}

	// 5. 动态跟踪止盈：从峰值回撤超过30%
	var drawdownFromPeak float64
	if peakProfit > 0.2 { // 只有当峰值盈利超过0.2%时才启用跟踪止盈
		drawdownFromPeak = (peakProfit - currentPnlPct) / peakProfit
		if drawdownFromPeak >= 0.3 { // 从峰值回撤30%
			log.Printf("🎯 %s 触发跟踪止盈：从峰值%.2f%%回撤%.1f%% >= 30%%", 
				posKey, peakProfit, drawdownFromPeak*100)
			return true
		}
	}

	log.Printf("📊 %s 当前浮盈: %.2f%%, 峰值: %.2f%%, 回撤: %.1f%%, 继续持有", 
		posKey, currentPnlPct, peakProfit, drawdownFromPeak*100)

	// 默认：继续持有
	return true
}

// calculateDynamicObservationPeriod 计算动态观察期
func (at *AutoTrader) calculateDynamicObservationPeriod(posKey string, decision *decision.Decision) float64 {
	// 基础观察期：10分钟
	baseObservationPeriod := 10.0
	
	// 获取市场数据
	symbol := strings.Split(posKey, "_")[0] // 从posKey中提取symbol
	marketData, err := market.Get(symbol)
	if err != nil {
		log.Printf("⚠️ 无法获取%s市场数据，使用默认观察期%.1f分钟", symbol, baseObservationPeriod)
		return baseObservationPeriod
	}
	
	// 1. 根据市场不确定性调整
	uncertainMarketAdjustment := 1.0
	
	// 检查市场波动性（通过RSI和MACD判断）
	if marketData.CurrentRSI7 > 0 && marketData.CurrentMACD != 0 {
		// 市场不确定性指标
		rsiVolatility := false
		macdUncertainty := false
		
		// RSI在30-70之间表示不确定
		if marketData.CurrentRSI7 >= 30 && marketData.CurrentRSI7 <= 70 {
			rsiVolatility = true
		}
		
		// MACD接近0表示方向不明确
		if marketData.CurrentMACD > -0.001 && marketData.CurrentMACD < 0.001 {
			macdUncertainty = true
		}
		
		// 不确定市场环境下缩短观察期至5-8分钟
		if rsiVolatility && macdUncertainty {
			uncertainMarketAdjustment = 0.6 // 缩短至6分钟
			log.Printf("📊 %s 不确定市场环境(RSI:%.1f, MACD:%.6f)，观察期缩短", symbol, marketData.CurrentRSI7, marketData.CurrentMACD)
		} else if rsiVolatility || macdUncertainty {
			uncertainMarketAdjustment = 0.8 // 缩短至8分钟
			log.Printf("📊 %s 部分不确定市场环境，观察期适度缩短", symbol)
		}
	}
	
	// 2. 根据信号强度调整（需要从决策推理中推断）
	signalStrengthAdjustment := 1.0
	
	// 检查决策推理中是否提到低信号强度
	reasoning := strings.ToLower(decision.Reasoning)
	if strings.Contains(reasoning, "信号强度") && 
	   (strings.Contains(reasoning, "低") || strings.Contains(reasoning, "弱") || 
	    strings.Contains(reasoning, "不足") || strings.Contains(reasoning, "疲弱")) {
		// 低信号强度时观察期减半
		signalStrengthAdjustment = 0.5
		log.Printf("📊 %s 低信号强度，观察期减半", symbol)
	}
	
	// 计算最终观察期
	finalObservationPeriod := baseObservationPeriod * uncertainMarketAdjustment * signalStrengthAdjustment
	
	// 确保观察期在合理范围内（最少3分钟，最多15分钟）
	if finalObservationPeriod < 3.0 {
		finalObservationPeriod = 3.0
	} else if finalObservationPeriod > 15.0 {
		finalObservationPeriod = 15.0
	}
	
	log.Printf("🕐 %s 动态观察期：%.1f分钟 (基础%.1f × 市场%.1f × 信号%.1f)", 
		symbol, finalObservationPeriod, baseObservationPeriod, uncertainMarketAdjustment, signalStrengthAdjustment)
	
	return finalObservationPeriod
}

// isTrendReversed 检查趋势是否反转
func (at *AutoTrader) isTrendReversed(symbol string, side string, entryPrice float64) bool {
	marketData, err := market.Get(symbol)
	if err != nil {
		log.Printf("⚠️ 无法获取%s市场数据进行趋势检查", symbol)
		return false
	}

	currentPrice := marketData.CurrentPrice
	ema20 := marketData.CurrentEMA20
	macd := marketData.CurrentMACD

	if side == "long" {
		// 多头趋势反转：价格跌破EMA20且MACD转负
		if currentPrice < ema20 && macd < 0 {
			log.Printf("📉 %s 多头趋势反转：价格%.2f < EMA20(%.2f), MACD(%.6f) < 0", 
				symbol, currentPrice, ema20, macd)
			return true
		}
	} else {
		// 空头趋势反转：价格突破EMA20且MACD转正
		if currentPrice > ema20 && macd > 0 {
			log.Printf("📈 %s 空头趋势反转：价格%.2f > EMA20(%.2f), MACD(%.6f) > 0", 
				symbol, currentPrice, ema20, macd)
			return true
		}
	}

	return false
}

// checkRSIExtremeConditions 检查RSI极端条件
func (at *AutoTrader) checkRSIExtremeConditions(symbol string, side string, currentPnlPct float64) bool {
	marketData, err := market.Get(symbol)
	if err != nil {
		log.Printf("⚠️ 无法获取%s市场数据进行RSI检查", symbol)
		return false
	}

	rsi7 := marketData.CurrentRSI7

	if side == "long" {
		// 多头持仓：RSI超买(>80)且有盈利时建议退出
		if rsi7 > 80 && currentPnlPct > 0 {
			log.Printf("🔥 %s 多头RSI超买退出：RSI7(%.1f) > 80且盈利%.2f%%", 
				symbol, rsi7, currentPnlPct)
			return true
		}
	} else {
		// 空头持仓：RSI超卖(<20)且有盈利时建议退出
		if rsi7 < 20 && currentPnlPct > 0 {
			log.Printf("❄️ %s 空头RSI超卖退出：RSI7(%.1f) < 20且盈利%.2f%%", 
				symbol, rsi7, currentPnlPct)
			return true
		}
	}

	return false
}

// checkPartialProfitTargets 检查分批止盈目标
func (at *AutoTrader) checkPartialProfitTargets(posKey string, currentPnlPct, entryPrice, currentPrice float64, side string) bool {
	// 分层止盈策略
	// 第一层：盈利0.5%时，建议部分止盈30%
	if currentPnlPct >= 0.5 && currentPnlPct < 1.0 {
		log.Printf("🎯 %s 第一层止盈触发: %.2f%% >= 0.5%% (建议平仓30%，移动止损到保本)", posKey, currentPnlPct)
		// 在实际实现中，这里应该调用部分平仓API
		// 目前只是记录日志，实际平仓逻辑需要在调用方实现
		return false // 不完全退出，继续持有70%
	}

	// 第二层：盈利1.0%时，建议再平仓50%
	if currentPnlPct >= 1.0 && currentPnlPct < 2.0 {
		log.Printf("🎯 %s 第二层止盈触发: %.2f%% >= 1.0%% (建议再平仓50%)", posKey, currentPnlPct)
		return false // 不完全退出，继续持有20%
	}

	// 第三层：盈利2.0%时，平仓剩余20%
	if currentPnlPct >= 2.0 {
		log.Printf("🎯 %s 第三层止盈触发: %.2f%% >= 2.0%% (平仓剩余20%)", posKey, currentPnlPct)
		return true // 完全退出
	}

	// 动态追踪止盈：如果盈利超过0.3%，启用追踪止损
	if currentPnlPct >= 0.3 {
		// 计算从峰值回撤
		// 这里简化处理，实际应该跟踪历史最高盈利
		trailingStopPct := 0.2 // 20%回撤触发止盈
		if currentPnlPct > 0.5 { // 只有在盈利超过0.5%时才启用追踪止损
			// 简化的追踪止损逻辑
			// 实际实现需要跟踪峰值盈利
			log.Printf("📊 %s 追踪止盈监控: 当前盈利%.2f%% (追踪止损阈值%.1f%%)", 
				posKey, currentPnlPct, trailingStopPct*100)
		}
	}

	return false
}

// checkRSIForPartialProfit 检查RSI是否超过80且持仓盈利，建议部分止盈
func (at *AutoTrader) checkRSIForPartialProfit(symbol string, posKey string) bool {
	// 获取市场数据
	marketData, err := market.Get(symbol)
	if err != nil {
		log.Printf("⚠ 获取%s市场数据失败: %v", symbol, err)
		return false
	}

	// 检查RSI是否超过80
	if marketData.CurrentRSI7 > 80.0 {
		// 检查持仓是否盈利
		peakProfit, exists := at.positionPeakProfit[posKey]
		if exists && peakProfit > 0 {
			log.Printf("🚨 %s RSI超买警告: RSI7=%.1f > 80, 峰值盈利=%.2f%%", 
				symbol, marketData.CurrentRSI7, peakProfit)
			return true
		}
	}
	
	return false
}

// GetStatus 获取系统状态（用于API）
func (at *AutoTrader) GetStatus() map[string]interface{} {
	aiProvider := "DeepSeek"
	if at.config.UseQwen {
		aiProvider = "Qwen"
	}

	status := map[string]interface{}{
		"trader_id":       at.id,
		"trader_name":     at.name,
		"ai_model":        at.aiModel,
		"exchange":        at.exchange,
		"is_running":      at.isRunning,
		"start_time":      at.startTime.Format(time.RFC3339),
		"runtime_minutes": int(time.Since(at.startTime).Minutes()),
		"call_count":      at.callCount,
		"initial_balance": at.initialBalance,
		"scan_interval":   at.config.ScanInterval.String(),
		"stop_until":      at.stopUntil.Format(time.RFC3339),
		"last_reset_time": at.lastResetTime.Format(time.RFC3339),
		"ai_provider":     aiProvider,
	}

	// 添加频率管理器状态
	if at.frequencyManager != nil {
		status["frequency_manager"] = at.frequencyManager.GetMetrics()
	}

	return status
}

// GetFrequencyStatus 获取频率管理器状态（用于API）
func (at *AutoTrader) GetFrequencyStatus() map[string]interface{} {
	if at.frequencyManager == nil {
		return map[string]interface{}{
			"enabled": false,
			"error":   "频率管理器未初始化",
		}
	}

	metrics := at.frequencyManager.GetMetrics()
	metrics["enabled"] = true
	
	// 计算到下一个小时重置的时间
	now := time.Now()
	nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
	timeToReset := nextHour.Sub(now).String()
	metrics["time_to_hourly_reset"] = timeToReset

	return metrics
}

// UpdateFrequencyConfig 更新频率管理器配置（用于API）
func (at *AutoTrader) UpdateFrequencyConfig(limitsData interface{}) error {
	if at.frequencyManager == nil {
		return fmt.Errorf("频率管理器未初始化")
	}

	// 将interface{}转换为FrequencyLimits
	jsonData, err := json.Marshal(limitsData)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	var limits FrequencyLimits
	if err := json.Unmarshal(jsonData, &limits); err != nil {
		return fmt.Errorf("反序列化配置失败: %w", err)
	}

	at.frequencyManager.UpdateLimits(limits)
	
	// 保存更新后的状态
	if err := at.frequencyManager.SaveState(); err != nil {
		log.Printf("⚠ 保存频率管理器配置失败: %v", err)
		return fmt.Errorf("保存配置失败: %w", err)
	}

	log.Printf("⚙️ [%s] 频率管理器配置已更新", at.name)
	return nil
}

// GetAccountInfo 获取账户信息（用于API）
func (at *AutoTrader) GetAccountInfo() (map[string]interface{}, error) {
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	// 获取账户字段
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Total Equity = 钱包余额 + 未实现盈亏
	totalEquity := totalWalletBalance + totalUnrealizedProfit

	// 获取持仓计算总保证金
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	totalMarginUsed := 0.0
	totalUnrealizedPnL := 0.0
	for _, pos := range positions {
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		totalUnrealizedPnL += unrealizedPnl

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed
	}

	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	return map[string]interface{}{
		// 核心字段
		"total_equity":      totalEquity,           // 账户净值 = wallet + unrealized
		"wallet_balance":    totalWalletBalance,    // 钱包余额（不含未实现盈亏）
		"unrealized_profit": totalUnrealizedProfit, // 未实现盈亏（从API）
		"available_balance": availableBalance,      // 可用余额

		// 盈亏统计
		"total_pnl":            totalPnL,           // 总盈亏 = equity - initial
		"total_pnl_pct":        totalPnLPct,        // 总盈亏百分比
		"total_unrealized_pnl": totalUnrealizedPnL, // 未实现盈亏（从持仓计算）
		"initial_balance":      at.initialBalance,  // 初始余额
		"daily_pnl":            at.dailyPnL,        // 日盈亏

		// 持仓信息
		"position_count":  len(positions),  // 持仓数量
		"margin_used":     totalMarginUsed, // 保证金占用
		"margin_used_pct": marginUsedPct,   // 保证金使用率
	}, nil
}

// GetPositions 获取持仓列表（用于API）
func (at *AutoTrader) GetPositions() ([]map[string]interface{}, error) {
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity := pos["positionAmt"].(float64)
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}

		pnlPct := 0.0
		if side == "long" {
			pnlPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
		} else {
			pnlPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
		}

		marginUsed := (quantity * markPrice) / float64(leverage)

		result = append(result, map[string]interface{}{
			"symbol":             symbol,
			"side":               side,
			"entry_price":        entryPrice,
			"mark_price":         markPrice,
			"quantity":           quantity,
			"leverage":           leverage,
			"unrealized_pnl":     unrealizedPnl,
			"unrealized_pnl_pct": pnlPct,
			"liquidation_price":  liquidationPrice,
			"margin_used":        marginUsed,
		})
	}

	return result, nil
}

// sortDecisionsByPriority 对决策排序：先平仓，再开仓，最后hold/wait
// 这样可以避免换仓时仓位叠加超限
func sortDecisionsByPriority(decisions []decision.Decision) []decision.Decision {
	if len(decisions) <= 1 {
		return decisions
	}

	// 定义优先级
	getActionPriority := func(action string) int {
		switch action {
		case "close_long", "close_short":
			return 1 // 最高优先级：先平仓
		case "open_long", "open_short":
			return 2 // 次优先级：后开仓
		case "hold", "wait":
			return 3 // 最低优先级：观望
		default:
			return 999 // 未知动作放最后
		}
	}

	// 复制决策列表
	sorted := make([]decision.Decision, len(decisions))
	copy(sorted, decisions)

	// 按优先级排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if getActionPriority(sorted[i].Action) > getActionPriority(sorted[j].Action) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
