package trader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// FrequencyMode 频率模式类型
type FrequencyMode string

const (
	BasicMode   FrequencyMode = "basic"   // 基础模式
	ElasticMode FrequencyMode = "elastic" // 弹性模式
)

// FrequencyLimits 频率限制配置
type FrequencyLimits struct {
	BasicMode struct {
		HourlyLimit int `json:"hourly_limit"` // 基础模式每小时限制
		DailyLimit  int `json:"daily_limit"`  // 基础模式每日限制
	} `json:"basic_mode"`

	ElasticMode struct {
		HourlyLimit int `json:"hourly_limit"` // 弹性模式每小时限制
		DailyLimit  int `json:"daily_limit"`  // 弹性模式每日限制 (-1表示无限制)
	} `json:"elastic_mode"`

	AbsoluteLimit struct {
		HourlyMax int `json:"hourly_max"` // 绝对每小时上限
	} `json:"absolute_limit"`

	Thresholds struct {
		UpgradePnLPercent   float64 `json:"upgrade_pnl_percent"`   // 升级到弹性模式的盈利阈值
		DowngradePnLPercent float64 `json:"downgrade_pnl_percent"` // 降级到基础模式的盈利阈值
	} `json:"thresholds"`
}

// FrequencyManager 频率管理器
type FrequencyManager struct {
	CurrentMode        FrequencyMode `json:"current_mode"`         // 当前模式
	DailyRealizedPnL   float64       `json:"daily_realized_pnl"`   // 当日已实现盈亏
	DailyPnLPercent    float64       `json:"daily_pnl_percent"`    // 当日盈亏百分比
	HourlyTradeCount   int           `json:"hourly_trade_count"`   // 当前小时交易次数
	DailyTradeCount    int           `json:"daily_trade_count"`    // 当日交易次数
	LastModeSwitch     time.Time     `json:"last_mode_switch"`     // 最后模式切换时间
	AccountEquity      float64       `json:"account_equity"`       // 账户净值
	LastHourlyReset    time.Time     `json:"last_hourly_reset"`    // 最后小时重置时间
	LastDailyReset     time.Time     `json:"last_daily_reset"`     // 最后日重置时间
	ModeUpgradeCount   int           `json:"mode_upgrade_count"`   // 今日升级次数
	ModeDowngradeCount int           `json:"mode_downgrade_count"` // 今日降级次数
	RejectedTradeCount int           `json:"rejected_trade_count"` // 今日被拒绝交易次数

	// 配置
	Limits FrequencyLimits `json:"limits"`

	// 状态文件路径
	stateFilePath string
}

// NewFrequencyManager 创建新的频率管理器
func NewFrequencyManager(stateFilePath string) *FrequencyManager {
	fm := &FrequencyManager{
		CurrentMode:      BasicMode,
		LastModeSwitch:   time.Now(),
		LastHourlyReset:  time.Now(),
		LastDailyReset:   time.Now(),
		stateFilePath:    stateFilePath,
		Limits: FrequencyLimits{
			BasicMode: struct {
				HourlyLimit int `json:"hourly_limit"`
				DailyLimit  int `json:"daily_limit"`
			}{
				HourlyLimit: 4,  // 2 → 4 (提高100%)
				DailyLimit:  20, // 10 → 20 (提高100%)
			},
			ElasticMode: struct {
				HourlyLimit int `json:"hourly_limit"`
				DailyLimit  int `json:"daily_limit"`
			}{
				HourlyLimit: 8,  // 5 → 8 (提高60%)
				DailyLimit:  50, // -1 → 50 (设置上限防止过度交易)
			},
			AbsoluteLimit: struct {
				HourlyMax int `json:"hourly_max"`
			}{
				HourlyMax: 10, // 6 → 10 (提高硬上限)
			},
			Thresholds: struct {
				UpgradePnLPercent   float64 `json:"upgrade_pnl_percent"`
				DowngradePnLPercent float64 `json:"downgrade_pnl_percent"`
			}{
				UpgradePnLPercent:   0.5, // 2.0% → 0.5% (更容易触发弹性模式)
				DowngradePnLPercent: 0.2, // 1.0% → 0.2% (降级阈值也相应降低)
			},
		},
	}

	// 尝试加载已保存的状态
	if err := fm.LoadState(); err != nil {
		log.Printf("⚠️ 无法加载频率管理器状态，使用默认配置: %v", err)
	}

	return fm
}

// CalculateDailyPnLPercent 计算当日已实现盈亏百分比
func (fm *FrequencyManager) CalculateDailyPnLPercent() float64 {
	if fm.AccountEquity <= 0 {
		return 0
	}
	return (fm.DailyRealizedPnL / fm.AccountEquity) * 100
}

// UpdateAccountEquity 更新账户净值
func (fm *FrequencyManager) UpdateAccountEquity(equity float64) {
	fm.AccountEquity = equity
	fm.DailyPnLPercent = fm.CalculateDailyPnLPercent()
}

// UpdateDailyPnL 更新当日已实现盈亏
func (fm *FrequencyManager) UpdateDailyPnL(realizedPnL float64) {
	fm.DailyRealizedPnL += realizedPnL
	fm.DailyPnLPercent = fm.CalculateDailyPnLPercent()
	log.Printf("📊 [频率管理] 更新当日PnL: +%.2f USDT, 累计: %.2f USDT (%.2f%%)",
		realizedPnL, fm.DailyRealizedPnL, fm.DailyPnLPercent)
}

// UpdateFrequencyMode 检查并更新频率模式
func (fm *FrequencyManager) UpdateFrequencyMode() (bool, string) {
	oldMode := fm.CurrentMode
	pnlPercent := fm.DailyPnLPercent

	switch fm.CurrentMode {
	case BasicMode:
		if pnlPercent > fm.Limits.Thresholds.UpgradePnLPercent {
			fm.CurrentMode = ElasticMode
			fm.LastModeSwitch = time.Now()
			fm.ModeUpgradeCount++
			msg := fmt.Sprintf("✅ 当日利润超过%.1f%%，进入弹性频率模式", fm.Limits.Thresholds.UpgradePnLPercent)
			log.Printf("🔄 [模式切换] %s -> %s | %s", oldMode, fm.CurrentMode, msg)
			return true, msg
		}
	case ElasticMode:
		if pnlPercent < fm.Limits.Thresholds.DowngradePnLPercent {
			fm.CurrentMode = BasicMode
			fm.LastModeSwitch = time.Now()
			fm.ModeDowngradeCount++
			msg := fmt.Sprintf("⚠️ 当日利润回撤至%.1f%%以下，退回基础频率模式", fm.Limits.Thresholds.DowngradePnLPercent)
			log.Printf("🔄 [模式切换] %s -> %s | %s", oldMode, fm.CurrentMode, msg)
			return true, msg
		}
	}

	return false, ""
}

// CheckTradeAllowance 检查是否允许新开仓
func (fm *FrequencyManager) CheckTradeAllowance() (bool, string) {
	// 先重置计数器（如果需要）
	fm.resetCountersIfNeeded()

	// 1. 检查绝对硬限制
	if fm.HourlyTradeCount >= fm.Limits.AbsoluteLimit.HourlyMax {
		fm.RejectedTradeCount++
		reason := fmt.Sprintf("🚫 已达到每小时绝对上限(%d笔)，拒绝开仓", fm.Limits.AbsoluteLimit.HourlyMax)
		log.Printf("🚫 [交易拒绝] %s | 当前模式:%s | 小时计数:%d | 日计数:%d",
			reason, fm.CurrentMode, fm.HourlyTradeCount, fm.DailyTradeCount)
		return false, reason
	}

	// 2. 根据当前模式检查限制
	switch fm.CurrentMode {
	case BasicMode:
		if fm.HourlyTradeCount >= fm.Limits.BasicMode.HourlyLimit {
			fm.RejectedTradeCount++
			reason := fmt.Sprintf("⏸️ 基础模式：已达到每小时上限(%d笔)", fm.Limits.BasicMode.HourlyLimit)
			log.Printf("🚫 [交易拒绝] %s | 小时计数:%d | 日计数:%d",
				reason, fm.HourlyTradeCount, fm.DailyTradeCount)
			return false, reason
		}
		if fm.DailyTradeCount >= fm.Limits.BasicMode.DailyLimit {
			fm.RejectedTradeCount++
			reason := fmt.Sprintf("⏸️ 基础模式：已达到每日上限(%d笔)", fm.Limits.BasicMode.DailyLimit)
			log.Printf("🚫 [交易拒绝] %s | 小时计数:%d | 日计数:%d",
				reason, fm.HourlyTradeCount, fm.DailyTradeCount)
			return false, reason
		}
	case ElasticMode:
		if fm.HourlyTradeCount >= fm.Limits.ElasticMode.HourlyLimit {
			fm.RejectedTradeCount++
			reason := fmt.Sprintf("⚡ 弹性模式：已达到每小时上限(%d笔)", fm.Limits.ElasticMode.HourlyLimit)
			log.Printf("🚫 [交易拒绝] %s | 小时计数:%d | 日计数:%d",
				reason, fm.HourlyTradeCount, fm.DailyTradeCount)
			return false, reason
		}
		// 弹性模式无每日限制（DailyLimit = -1）
	}

	reason := "✅ 频率检查通过，允许开仓"
	log.Printf("📊 [频率检查] 模式:%s | 当日PnL:%.2f%% | 小时交易:%d | 日交易:%d | 结果:%s",
		fm.CurrentMode, fm.DailyPnLPercent, fm.HourlyTradeCount, fm.DailyTradeCount, reason)
	return true, reason
}

// IncrementTradeCount 增加交易计数
func (fm *FrequencyManager) IncrementTradeCount() {
	fm.HourlyTradeCount++
	fm.DailyTradeCount++
	log.Printf("📈 [交易计数] 小时:%d | 日:%d | 模式:%s",
		fm.HourlyTradeCount, fm.DailyTradeCount, fm.CurrentMode)
}

// resetCountersIfNeeded 重置计数器（如果需要）
func (fm *FrequencyManager) resetCountersIfNeeded() {
	now := time.Now()

	// 检查是否需要重置小时计数器
	if now.Hour() != fm.LastHourlyReset.Hour() || now.Day() != fm.LastHourlyReset.Day() {
		fm.HourlyTradeCount = 0
		fm.LastHourlyReset = now
		log.Printf("🔄 [计数重置] 小时交易计数已重置")
	}

	// 检查是否需要重置日计数器
	if now.Day() != fm.LastDailyReset.Day() || now.Month() != fm.LastDailyReset.Month() {
		fm.DailyTradeCount = 0
		fm.DailyRealizedPnL = 0
		fm.DailyPnLPercent = 0
		fm.ModeUpgradeCount = 0
		fm.ModeDowngradeCount = 0
		fm.RejectedTradeCount = 0
		fm.LastDailyReset = now
		// 重置后回到基础模式
		if fm.CurrentMode != BasicMode {
			fm.CurrentMode = BasicMode
			fm.LastModeSwitch = now
			log.Printf("🔄 [模式重置] 新的一天开始，回到基础模式")
		}
		log.Printf("🔄 [计数重置] 日交易计数和PnL已重置")
	}
}

// GetCurrentLimits 获取当前模式的限制
func (fm *FrequencyManager) GetCurrentLimits() (hourlyLimit, dailyLimit int) {
	switch fm.CurrentMode {
	case BasicMode:
		return fm.Limits.BasicMode.HourlyLimit, fm.Limits.BasicMode.DailyLimit
	case ElasticMode:
		return fm.Limits.ElasticMode.HourlyLimit, fm.Limits.ElasticMode.DailyLimit
	default:
		return fm.Limits.BasicMode.HourlyLimit, fm.Limits.BasicMode.DailyLimit
	}
}

// GetNextModeThreshold 获取下一个模式切换阈值
func (fm *FrequencyManager) GetNextModeThreshold() float64 {
	switch fm.CurrentMode {
	case BasicMode:
		return fm.Limits.Thresholds.UpgradePnLPercent
	case ElasticMode:
		return fm.Limits.Thresholds.DowngradePnLPercent
	default:
		return fm.Limits.Thresholds.UpgradePnLPercent
	}
}

// SaveState 保存频率管理器状态到文件
func (fm *FrequencyManager) SaveState() error {
	if fm.stateFilePath == "" {
		return nil // 如果没有指定文件路径，跳过保存
	}

	data, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化状态失败: %w", err)
	}

	if err := ioutil.WriteFile(fm.stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("写入状态文件失败: %w", err)
	}

	return nil
}

// LoadState 从文件加载频率管理器状态
func (fm *FrequencyManager) LoadState() error {
	if fm.stateFilePath == "" {
		return nil // 如果没有指定文件路径，跳过加载
	}

	data, err := ioutil.ReadFile(fm.stateFilePath)
	if err != nil {
		return fmt.Errorf("读取状态文件失败: %w", err)
	}

	var loadedState FrequencyManager
	if err := json.Unmarshal(data, &loadedState); err != nil {
		return fmt.Errorf("反序列化状态失败: %w", err)
	}

	// 保留当前的配置和文件路径
	limits := fm.Limits
	stateFilePath := fm.stateFilePath

	// 复制加载的状态
	*fm = loadedState

	// 恢复配置和文件路径
	fm.Limits = limits
	fm.stateFilePath = stateFilePath

	log.Printf("📂 [状态加载] 频率管理器状态已从文件加载: %s", fm.stateFilePath)
	return nil
}

// GetMetrics 获取频率管理器指标
func (fm *FrequencyManager) GetMetrics() map[string]interface{} {
	hourlyLimit, dailyLimit := fm.GetCurrentLimits()
	nextThreshold := fm.GetNextModeThreshold()

	return map[string]interface{}{
		"current_mode":         string(fm.CurrentMode),
		"daily_pnl_percent":    fm.DailyPnLPercent,
		"hourly_trade_count":   fm.HourlyTradeCount,
		"daily_trade_count":    fm.DailyTradeCount,
		"hourly_limit":         hourlyLimit,
		"daily_limit":          dailyLimit,
		"next_mode_threshold":  nextThreshold,
		"mode_upgrade_count":   fm.ModeUpgradeCount,
		"mode_downgrade_count": fm.ModeDowngradeCount,
		"rejected_trade_count": fm.RejectedTradeCount,
		"last_mode_switch":     fm.LastModeSwitch.Format("2006-01-02 15:04:05"),
		"account_equity":       fm.AccountEquity,
		"daily_realized_pnl":   fm.DailyRealizedPnL,
	}
}

// UpdateLimits 更新频率限制配置
func (fm *FrequencyManager) UpdateLimits(limits FrequencyLimits) {
	log.Printf("⚙️ [配置更新] 开始更新频率限制配置")
	log.Printf("⚙️ [配置更新] 新配置 - 基础模式: 小时%d/日%d", limits.BasicMode.HourlyLimit, limits.BasicMode.DailyLimit)
	log.Printf("⚙️ [配置更新] 新配置 - 弹性模式: 小时%d/日%d", limits.ElasticMode.HourlyLimit, limits.ElasticMode.DailyLimit)
	log.Printf("⚙️ [配置更新] 新配置 - 绝对限制: 小时%d", limits.AbsoluteLimit.HourlyMax)
	log.Printf("⚙️ [配置更新] 新配置 - 阈值: 升级%.1f%%/降级%.1f%%", limits.Thresholds.UpgradePnLPercent, limits.Thresholds.DowngradePnLPercent)
	
	fm.Limits = limits
	
	// 保存状态到文件
	if err := fm.SaveState(); err != nil {
		log.Printf("⚠️ [配置更新] 保存状态失败: %v", err)
	} else {
		log.Printf("✅ [配置更新] 频率限制配置已更新并保存")
	}
}