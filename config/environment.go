package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Environment 环境配置结构
type Environment struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	BinanceBaseURL  string                 `json:"binance_base_url"`
	BinanceAPIKey   string                 `json:"binance_api_key"`
	BinanceSecret   string                 `json:"binance_secret_key"`
	DeepSeekAPIKey  string                 `json:"deepseek_api_key"`
	OITopAPIURL     string                 `json:"oi_top_api_url"`
	ProxyURL        string                 `json:"proxy_url"`
	RiskLimits      RiskLimits            `json:"risk_limits"`
	TradingSettings TradingSettings       `json:"trading_settings"`
	Status          string                 `json:"status"`
	LastValidated   *time.Time            `json:"last_validated"`
	APIPermissions  []string              `json:"api_permissions"`
}

// RiskLimits 风险限制配置
type RiskLimits struct {
	MaxPositionSize float64 `json:"max_position_size"`
	MaxDailyLoss    float64 `json:"max_daily_loss"`
	MaxDrawdown     float64 `json:"max_drawdown"`
	MaxLeverage     int     `json:"max_leverage"`
}

// TradingSettings 交易设置
type TradingSettings struct {
	ScanIntervalMinutes int     `json:"scan_interval_minutes"`
	InitialBalance      float64 `json:"initial_balance"`
	StopTradingMinutes  int     `json:"stop_trading_minutes"`
}

// ValidationRecord 验证记录
type ValidationRecord struct {
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
	Valid       bool      `json:"valid"`
	Permissions []string  `json:"permissions"`
	Errors      []string  `json:"errors"`
}

// EnvironmentSwitch 环境切换记录
type EnvironmentSwitch struct {
	FromEnvironment string    `json:"from_environment"`
	ToEnvironment   string    `json:"to_environment"`
	Timestamp       time.Time `json:"timestamp"`
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
}

// EnvironmentConfig 环境配置管理
type EnvironmentConfig struct {
	CurrentEnvironment  string                       `json:"current_environment"`
	Environments        map[string]*Environment      `json:"environments"`
	ValidationHistory   []ValidationRecord           `json:"validation_history"`
	EnvironmentSwitches []EnvironmentSwitch          `json:"environment_switches"`
	CreatedAt           time.Time                    `json:"created_at"`
	UpdatedAt           time.Time                    `json:"updated_at"`
	mu                  sync.RWMutex                 `json:"-"`
	filePath            string                       `json:"-"`
}

// EnvironmentManager 环境管理器
type EnvironmentManager struct {
	config *EnvironmentConfig
}

// NewEnvironmentManager 创建环境管理器
func NewEnvironmentManager(configPath string) (*EnvironmentManager, error) {
	manager := &EnvironmentManager{}
	
	// 确保配置文件路径存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	// 加载配置
	config, err := loadEnvironmentConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载环境配置失败: %v", err)
	}
	
	config.filePath = configPath
	manager.config = config
	
	return manager, nil
}

// loadEnvironmentConfig 加载环境配置
func loadEnvironmentConfig(filePath string) (*EnvironmentConfig, error) {
	// 如果文件不存在，创建默认配置
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("环境配置文件不存在，创建默认配置: %s", filePath)
		return createDefaultEnvironmentConfig(), nil
	}
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	var config EnvironmentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	return &config, nil
}

// createDefaultEnvironmentConfig 创建默认环境配置
func createDefaultEnvironmentConfig() *EnvironmentConfig {
	now := time.Now()
	return &EnvironmentConfig{
		CurrentEnvironment: "testnet",
		Environments: map[string]*Environment{
			"testnet": {
				Name:           "测试网环境",
				Description:    "安全的测试环境，用于验证交易策略",
				BinanceBaseURL: "https://testnet.binancefuture.com",
				BinanceAPIKey:  "",
				BinanceSecret:  "",
				DeepSeekAPIKey: "",
				RiskLimits: RiskLimits{
					MaxPositionSize: 100,
					MaxDailyLoss:    50,
					MaxDrawdown:     10.0,
					MaxLeverage:     10,
				},
				TradingSettings: TradingSettings{
					ScanIntervalMinutes: 3,
					InitialBalance:      1000,
					StopTradingMinutes:  60,
				},
				Status:         "inactive",
				LastValidated:  nil,
				APIPermissions: []string{},
			},
			"mainnet": {
				Name:           "真实环境",
				Description:    "真实交易环境，请谨慎操作",
				BinanceBaseURL: "https://fapi.binance.com",
				BinanceAPIKey:  "",
				BinanceSecret:  "",
				DeepSeekAPIKey: "",
				RiskLimits: RiskLimits{
					MaxPositionSize: 1000,
					MaxDailyLoss:    500,
					MaxDrawdown:     20.0,
					MaxLeverage:     20,
				},
				TradingSettings: TradingSettings{
					ScanIntervalMinutes: 5,
					InitialBalance:      10000,
					StopTradingMinutes:  30,
				},
				Status:         "inactive",
				LastValidated:  nil,
				APIPermissions: []string{},
			},
		},
		ValidationHistory:   []ValidationRecord{},
		EnvironmentSwitches: []EnvironmentSwitch{},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// GetCurrentEnvironment 获取当前环境
func (em *EnvironmentManager) GetCurrentEnvironment() string {
	em.config.mu.RLock()
	defer em.config.mu.RUnlock()
	return em.config.CurrentEnvironment
}

// GetEnvironment 获取指定环境配置
func (em *EnvironmentManager) GetEnvironment(envName string) (*Environment, error) {
	em.config.mu.RLock()
	defer em.config.mu.RUnlock()
	
	env, exists := em.config.Environments[envName]
	if !exists {
		return nil, fmt.Errorf("环境 '%s' 不存在", envName)
	}
	
	return env, nil
}

// GetAllEnvironments 获取所有环境配置
func (em *EnvironmentManager) GetAllEnvironments() map[string]*Environment {
	em.config.mu.RLock()
	defer em.config.mu.RUnlock()
	
	// 创建副本以避免并发问题
	result := make(map[string]*Environment)
	for k, v := range em.config.Environments {
		result[k] = v
	}
	
	return result
}

// SwitchEnvironment 切换环境
func (em *EnvironmentManager) SwitchEnvironment(targetEnv string) error {
	em.config.mu.Lock()
	defer em.config.mu.Unlock()
	
	// 检查目标环境是否存在
	if _, exists := em.config.Environments[targetEnv]; !exists {
		return fmt.Errorf("目标环境 '%s' 不存在", targetEnv)
	}
	
	oldEnv := em.config.CurrentEnvironment
	em.config.CurrentEnvironment = targetEnv
	em.config.UpdatedAt = time.Now()
	
	// 记录切换历史
	switchRecord := EnvironmentSwitch{
		FromEnvironment: oldEnv,
		ToEnvironment:   targetEnv,
		Timestamp:       time.Now(),
		Success:         true,
		Message:         fmt.Sprintf("成功从 %s 切换到 %s", oldEnv, targetEnv),
	}
	em.config.EnvironmentSwitches = append(em.config.EnvironmentSwitches, switchRecord)
	
	// 保存配置
	if err := em.saveConfig(); err != nil {
		// 回滚
		em.config.CurrentEnvironment = oldEnv
		switchRecord.Success = false
		switchRecord.Message = fmt.Sprintf("切换失败: %v", err)
		return fmt.Errorf("保存配置失败: %v", err)
	}
	
	log.Printf("✓ 环境切换成功: %s -> %s", oldEnv, targetEnv)
	return nil
}

// UpdateEnvironmentConfig 更新环境配置
func (em *EnvironmentManager) UpdateEnvironmentConfig(envName string, env *Environment) error {
	em.config.mu.Lock()
	defer em.config.mu.Unlock()
	
	if _, exists := em.config.Environments[envName]; !exists {
		return fmt.Errorf("环境 '%s' 不存在", envName)
	}
	
	em.config.Environments[envName] = env
	em.config.UpdatedAt = time.Now()
	
	return em.saveConfig()
}

// ValidateEnvironment 验证环境配置
func (em *EnvironmentManager) ValidateEnvironment(envName string) (*ValidationRecord, error) {
	env, err := em.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}
	
	record := &ValidationRecord{
		Environment: envName,
		Timestamp:   time.Now(),
		Valid:       true,
		Permissions: []string{},
		Errors:      []string{},
	}
	
	// 验证API密钥
	if env.BinanceAPIKey == "" {
		record.Valid = false
		record.Errors = append(record.Errors, "Binance API Key 未配置")
	}
	
	if env.BinanceSecret == "" {
		record.Valid = false
		record.Errors = append(record.Errors, "Binance Secret Key 未配置")
	}
	
	if env.DeepSeekAPIKey == "" {
		record.Valid = false
		record.Errors = append(record.Errors, "DeepSeek API Key 未配置")
	}
	
	// TODO: 实际验证API连接和权限
	// 这里可以添加实际的API调用来验证密钥有效性
	
	if record.Valid {
		record.Permissions = []string{"read", "trade", "futures"}
		env.Status = "active"
		env.LastValidated = &record.Timestamp
		env.APIPermissions = record.Permissions
	} else {
		env.Status = "error"
	}
	
	// 保存验证记录
	em.config.mu.Lock()
	em.config.ValidationHistory = append(em.config.ValidationHistory, *record)
	em.config.UpdatedAt = time.Now()
	em.saveConfig()
	em.config.mu.Unlock()
	
	return record, nil
}

// GetStatus 获取环境状态
func (em *EnvironmentManager) GetStatus() map[string]interface{} {
	em.config.mu.RLock()
	defer em.config.mu.RUnlock()
	
	currentEnv := em.config.Environments[em.config.CurrentEnvironment]
	
	return map[string]interface{}{
		"current_environment": em.config.CurrentEnvironment,
		"status":             currentEnv.Status,
		"api_status": map[string]interface{}{
			"binance_configured":   currentEnv.BinanceAPIKey != "",
			"deepseek_configured":  currentEnv.DeepSeekAPIKey != "",
			"last_validated":       currentEnv.LastValidated,
			"permissions":          currentEnv.APIPermissions,
		},
		"last_updated": em.config.UpdatedAt,
		"environments": func() map[string]interface{} {
			envs := make(map[string]interface{})
			for name, env := range em.config.Environments {
				// 返回完整的环境配置信息，但隐藏敏感信息
				envs[name] = map[string]interface{}{
					"name":               env.Name,
					"description":        env.Description,
					"status":            env.Status,
					"binance_api_key":   env.BinanceAPIKey != "",  // 只返回是否配置的布尔值
					"deepseek_api_key":  env.DeepSeekAPIKey != "", // 只返回是否配置的布尔值
					"risk_limits":       env.RiskLimits,
					"trading_settings":  env.TradingSettings,
					"last_validated":    env.LastValidated,
					"api_permissions":   env.APIPermissions,
				}
			}
			return envs
		}(),
	}
}

// saveConfig 保存配置到文件
func (em *EnvironmentManager) saveConfig() error {
	if em.config.filePath == "" {
		return fmt.Errorf("配置文件路径未设置")
	}
	
	data, err := json.MarshalIndent(em.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	return ioutil.WriteFile(em.config.filePath, data, 0644)
}