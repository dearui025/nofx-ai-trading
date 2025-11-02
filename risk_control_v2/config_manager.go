package risk_control_v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configDir       string
	configs         map[string]interface{}
	watchers        map[string]*ConfigWatcher
	mutex           sync.RWMutex
	logger          *log.Logger
	hotReloadEnabled bool
}

// ConfigWatcher 配置文件监视器
type ConfigWatcher struct {
	filePath     string
	lastModified time.Time
	callback     func(interface{}) error
	configType   string
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	ConfigType string      `json:"config_type"`
	ConfigName string      `json:"config_name"`
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
	Timestamp  time.Time   `json:"timestamp"`
	Source     string      `json:"source"` // file, api, system
}

// DefaultConfigs 默认配置
type DefaultConfigs struct {
	TimeManager      TimeManagerConfig      `json:"time_manager"`
	LiquidityMonitor LiquidityMonitorConfig `json:"liquidity_monitor"`
	SharpeCalculator SharpeCalculatorConfig `json:"sharpe_calculator"`
	AICommittee      AICommitteeConfig      `json:"ai_committee"`
	RiskManager      RiskManagerConfig      `json:"risk_manager"`
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configDir string) (*ConfigManager, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}

	cm := &ConfigManager{
		configDir:        configDir,
		configs:          make(map[string]interface{}),
		watchers:         make(map[string]*ConfigWatcher),
		logger:           log.New(log.Writer(), "[ConfigManager] ", log.LstdFlags),
		hotReloadEnabled: true,
	}

	// 初始化默认配置
	if err := cm.initDefaultConfigs(); err != nil {
		return nil, fmt.Errorf("初始化默认配置失败: %v", err)
	}

	// 加载现有配置
	if err := cm.loadAllConfigs(); err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 启动热重载监控
	if cm.hotReloadEnabled {
		go cm.startHotReload()
	}

	return cm, nil
}

// initDefaultConfigs 初始化默认配置
func (cm *ConfigManager) initDefaultConfigs() error {
	defaults := DefaultConfigs{
		TimeManager: TimeManagerConfig{
			Timezone:                "UTC",
			DailyResetHour:          0,
			EquityBufferPercent:     0.05, // 5%权益缓冲
			DrawdownResetConditions: []DrawdownResetCondition{
				{
					Type:      "daily",
					Threshold: 0.0,
					Enabled:   true,
				},
				{
					Type:      "new_high",
					Threshold: 0.05, // 5%缓冲
					Enabled:   true,
				},
			},
		},
		LiquidityMonitor: LiquidityMonitorConfig{
			HighLiquidityThreshold:     50000000, // 50M USD
			MediumLiquidityThreshold:   15000000, // 15M USD
			LowLiquidityThreshold:      10000000, // 10M USD
			RapidDeclineThreshold:      0.2,      // 20%
			MonitoringIntervalMinutes:  5,        // 5分钟
			AlertCooldownMinutes:       30,       // 30分钟
			ForceCloseEnabled:          true,
			ForceCloseThreshold:        10000000, // 10M USD
			BlacklistEnabled:           true,
			BlacklistDurationHours:     24,       // 24小时
			AutoRemoveFromBlacklist:    true,
		},
		SharpeCalculator: SharpeCalculatorConfig{
			WindowSize:           50,   // 50个决策周期
			MinWindowSize:        10,   // 最小10个周期
			BufferCycles:         2,    // 2个周期缓冲
			ConfidenceThreshold:  0.8,  // 80%置信度
			ExcellentThreshold:   2.0,  // 优秀阈值
			GoodThreshold:        1.0,  // 良好阈值
			NeutralThreshold:     0.0,  // 中性阈值
			PoorThreshold:        -1.0, // 较差阈值
			RiskFreeRate:         0.02, // 2%年化无风险利率
			AnnualizationFactor:  math.Sqrt(365 * 24 * 12), // 年化因子
			OutlierThreshold:     3.0,  // 3倍标准差
		},
		AICommittee: AICommitteeConfig{
			EnabledModels: []ModelType{
				ModelQwen,
				ModelDeepSeek,
				ModelClaude,
			},
			PrimaryModel:         ModelQwen,
			FallbackModel:        ModelDeepSeek,
			MinConsensusLevel:    0.6,
			ConservativeMode:     false,
			RequireUnanimity:     false,
			ModelTimeoutSeconds:  30,
			TotalTimeoutSeconds:  90,
			VolatilityThreshold:  0.05, // 5%
			TrendThreshold:       0.02, // 2%
			MaxRiskScore:         0.8,  // 80%
			RiskWeightEnabled:    true,
		},
		RiskManager: RiskManagerConfig{
			GlobalRiskEnabled:         true,
			EmergencyStopEnabled:      true,
			MaxDrawdownPercent:        0.25, // 25%
			MaxDailyLossPercent:       0.10, // 10%
			MonitoringIntervalSeconds: 60,   // 60秒
		},
	}

	// 保存默认配置到文件
	return cm.saveDefaultConfigsToFile(defaults)
}

// saveDefaultConfigsToFile 保存默认配置到文件
func (cm *ConfigManager) saveDefaultConfigsToFile(defaults DefaultConfigs) error {
	configFiles := map[string]interface{}{
		"time_manager.json":      defaults.TimeManager,
		"liquidity_monitor.json": defaults.LiquidityMonitor,
		"sharpe_calculator.json": defaults.SharpeCalculator,
		"ai_committee.json":      defaults.AICommittee,
		"risk_manager.json":      defaults.RiskManager,
	}

	for filename, config := range configFiles {
		filePath := filepath.Join(cm.configDir, filename)
		
		// 如果文件已存在，跳过
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化配置失败 %s: %v", filename, err)
		}

		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("写入配置文件失败 %s: %v", filename, err)
		}

		cm.logger.Printf("创建默认配置文件: %s", filename)
	}

	return nil
}

// loadAllConfigs 加载所有配置
func (cm *ConfigManager) loadAllConfigs() error {
	configFiles := []string{
		"time_manager.json",
		"liquidity_monitor.json",
		"sharpe_calculator.json",
		"ai_committee.json",
		"risk_manager.json",
	}

	for _, filename := range configFiles {
		if err := cm.loadConfigFromFile(filename); err != nil {
			cm.logger.Printf("加载配置文件失败 %s: %v", filename, err)
			// 继续加载其他配置文件
		}
	}

	return nil
}

// loadConfigFromFile 从文件加载配置
func (cm *ConfigManager) loadConfigFromFile(filename string) error {
	filePath := filepath.Join(cm.configDir, filename)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	cm.mutex.Lock()
	cm.configs[filename] = config
	cm.mutex.Unlock()

	cm.logger.Printf("加载配置文件: %s", filename)
	return nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(configType string, result interface{}) error {
	filename := configType + ".json"
	
	cm.mutex.RLock()
	config, exists := cm.configs[filename]
	cm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("配置不存在: %s", configType)
	}

	// 将配置转换为目标类型
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("反序列化配置失败: %v", err)
	}

	return nil
}

// SetConfig 设置配置
func (cm *ConfigManager) SetConfig(configType string, config interface{}) error {
	filename := configType + ".json"
	filePath := filepath.Join(cm.configDir, filename)

	// 获取旧配置用于事件记录
	cm.mutex.RLock()
	oldConfig := cm.configs[filename]
	cm.mutex.RUnlock()

	// 保存到内存
	cm.mutex.Lock()
	cm.configs[filename] = config
	cm.mutex.Unlock()

	// 保存到文件
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 记录配置变更事件
	event := ConfigChangeEvent{
		ConfigType: configType,
		ConfigName: filename,
		OldValue:   oldConfig,
		NewValue:   config,
		Timestamp:  time.Now(),
		Source:     "api",
	}

	cm.logConfigChange(event)
	cm.logger.Printf("配置已更新: %s", configType)

	return nil
}

// GetTimeManagerConfig 获取时间管理器配置
func (cm *ConfigManager) GetTimeManagerConfig() (TimeManagerConfig, error) {
	var config TimeManagerConfig
	err := cm.GetConfig("time_manager", &config)
	return config, err
}

// GetLiquidityMonitorConfig 获取流动性监控配置
func (cm *ConfigManager) GetLiquidityMonitorConfig() (LiquidityMonitorConfig, error) {
	var config LiquidityMonitorConfig
	err := cm.GetConfig("liquidity_monitor", &config)
	return config, err
}

// GetSharpeCalculatorConfig 获取夏普比率计算器配置
func (cm *ConfigManager) GetSharpeCalculatorConfig() (SharpeCalculatorConfig, error) {
	var config SharpeCalculatorConfig
	err := cm.GetConfig("sharpe_calculator", &config)
	return config, err
}

// GetAICommitteeConfig 获取AI委员会配置
func (cm *ConfigManager) GetAICommitteeConfig() (AICommitteeConfig, error) {
	var config AICommitteeConfig
	err := cm.GetConfig("ai_committee", &config)
	return config, err
}

// GetRiskManagerConfig 获取风控管理器配置
func (cm *ConfigManager) GetRiskManagerConfig() (RiskManagerConfig, error) {
	var config RiskManagerConfig
	err := cm.GetConfig("risk_manager", &config)
	return config, err
}

// SetTimeManagerConfig 设置时间管理器配置
func (cm *ConfigManager) SetTimeManagerConfig(config TimeManagerConfig) error {
	return cm.SetConfig("time_manager", config)
}

// SetLiquidityMonitorConfig 设置流动性监控配置
func (cm *ConfigManager) SetLiquidityMonitorConfig(config LiquidityMonitorConfig) error {
	return cm.SetConfig("liquidity_monitor", config)
}

// SetSharpeCalculatorConfig 设置夏普比率计算器配置
func (cm *ConfigManager) SetSharpeCalculatorConfig(config SharpeCalculatorConfig) error {
	return cm.SetConfig("sharpe_calculator", config)
}

// SetAICommitteeConfig 设置AI委员会配置
func (cm *ConfigManager) SetAICommitteeConfig(config AICommitteeConfig) error {
	return cm.SetConfig("ai_committee", config)
}

// SetRiskManagerConfig 设置风控管理器配置
func (cm *ConfigManager) SetRiskManagerConfig(config RiskManagerConfig) error {
	return cm.SetConfig("risk_manager", config)
}

// RegisterWatcher 注册配置监视器
func (cm *ConfigManager) RegisterWatcher(configType string, callback func(interface{}) error) error {
	filename := configType + ".json"
	filePath := filepath.Join(cm.configDir, filename)

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("配置文件不存在: %v", err)
	}

	watcher := &ConfigWatcher{
		filePath:     filePath,
		lastModified: stat.ModTime(),
		callback:     callback,
		configType:   configType,
	}

	cm.mutex.Lock()
	cm.watchers[configType] = watcher
	cm.mutex.Unlock()

	cm.logger.Printf("注册配置监视器: %s", configType)
	return nil
}

// UnregisterWatcher 取消注册配置监视器
func (cm *ConfigManager) UnregisterWatcher(configType string) {
	cm.mutex.Lock()
	delete(cm.watchers, configType)
	cm.mutex.Unlock()

	cm.logger.Printf("取消注册配置监视器: %s", configType)
}

// startHotReload 启动热重载监控
func (cm *ConfigManager) startHotReload() {
	ticker := time.NewTicker(time.Second * 5) // 每5秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		cm.checkConfigChanges()
	}
}

// checkConfigChanges 检查配置变更
func (cm *ConfigManager) checkConfigChanges() {
	cm.mutex.RLock()
	watchers := make(map[string]*ConfigWatcher)
	for k, v := range cm.watchers {
		watchers[k] = v
	}
	cm.mutex.RUnlock()

	for configType, watcher := range watchers {
		stat, err := os.Stat(watcher.filePath)
		if err != nil {
			cm.logger.Printf("检查配置文件失败 %s: %v", configType, err)
			continue
		}

		if stat.ModTime().After(watcher.lastModified) {
			cm.logger.Printf("检测到配置文件变更: %s", configType)
			
			// 重新加载配置
			if err := cm.loadConfigFromFile(filepath.Base(watcher.filePath)); err != nil {
				cm.logger.Printf("重新加载配置失败 %s: %v", configType, err)
				continue
			}

			// 获取新配置
			cm.mutex.RLock()
			newConfig := cm.configs[filepath.Base(watcher.filePath)]
			cm.mutex.RUnlock()

			// 调用回调函数
			if err := watcher.callback(newConfig); err != nil {
				cm.logger.Printf("配置变更回调失败 %s: %v", configType, err)
			} else {
				cm.logger.Printf("配置热重载成功: %s", configType)
			}

			// 更新最后修改时间
			watcher.lastModified = stat.ModTime()
		}
	}
}

// logConfigChange 记录配置变更
func (cm *ConfigManager) logConfigChange(event ConfigChangeEvent) {
	logFile := filepath.Join(cm.configDir, "config_changes.log")
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		cm.logger.Printf("序列化配置变更事件失败: %v", err)
		return
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		cm.logger.Printf("打开配置变更日志失败: %v", err)
		return
	}
	defer file.Close()

	logEntry := fmt.Sprintf("[%s] %s\n", 
		event.Timestamp.Format("2006-01-02 15:04:05"), 
		string(eventJSON))
	
	if _, err := file.WriteString(logEntry); err != nil {
		cm.logger.Printf("写入配置变更日志失败: %v", err)
	}
}

// GetAllConfigs 获取所有配置
func (cm *ConfigManager) GetAllConfigs() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[string]interface{})
	for k, v := range cm.configs {
		result[k] = v
	}

	return result
}

// ValidateConfig 验证配置
func (cm *ConfigManager) ValidateConfig(configType string, config interface{}) error {
	switch configType {
	case "time_manager":
		return cm.validateTimeManagerConfig(config)
	case "liquidity_monitor":
		return cm.validateLiquidityMonitorConfig(config)
	case "sharpe_calculator":
		return cm.validateSharpeCalculatorConfig(config)
	case "ai_committee":
		return cm.validateAICommitteeConfig(config)
	case "risk_manager":
		return cm.validateRiskManagerConfig(config)
	default:
		return fmt.Errorf("未知的配置类型: %s", configType)
	}
}

// validateTimeManagerConfig 验证时间管理器配置
func (cm *ConfigManager) validateTimeManagerConfig(config interface{}) error {
	// 这里可以添加具体的验证逻辑
	// 例如检查时区是否有效、重置时间是否合理等
	return nil
}

// validateLiquidityMonitorConfig 验证流动性监控配置
func (cm *ConfigManager) validateLiquidityMonitorConfig(config interface{}) error {
	// 这里可以添加具体的验证逻辑
	// 例如检查阈值是否合理、监控间隔是否有效等
	return nil
}

// validateSharpeCalculatorConfig 验证夏普比率计算器配置
func (cm *ConfigManager) validateSharpeCalculatorConfig(config interface{}) error {
	// 这里可以添加具体的验证逻辑
	// 例如检查窗口大小是否合理、阈值是否有效等
	return nil
}

// validateAICommitteeConfig 验证AI委员会配置
func (cm *ConfigManager) validateAICommitteeConfig(config interface{}) error {
	// 这里可以添加具体的验证逻辑
	// 例如检查模型权重总和是否为1、共识阈值是否合理等
	return nil
}

// validateRiskManagerConfig 验证风控管理器配置
func (cm *ConfigManager) validateRiskManagerConfig(config interface{}) error {
	// 这里可以添加具体的验证逻辑
	// 例如检查风险限制是否合理、监控间隔是否有效等
	return nil
}

// EnableHotReload 启用热重载
func (cm *ConfigManager) EnableHotReload() {
	cm.mutex.Lock()
	cm.hotReloadEnabled = true
	cm.mutex.Unlock()

	if len(cm.watchers) > 0 {
		go cm.startHotReload()
	}
}

// DisableHotReload 禁用热重载
func (cm *ConfigManager) DisableHotReload() {
	cm.mutex.Lock()
	cm.hotReloadEnabled = false
	cm.mutex.Unlock()
}

// BackupConfigs 备份配置
func (cm *ConfigManager) BackupConfigs() error {
	backupDir := filepath.Join(cm.configDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("config_backup_%s.json", timestamp))

	allConfigs := cm.GetAllConfigs()
	data, err := json.MarshalIndent(allConfigs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("写入备份文件失败: %v", err)
	}

	cm.logger.Printf("配置备份完成: %s", backupPath)
	return nil
}

// RestoreConfigs 恢复配置
func (cm *ConfigManager) RestoreConfigs(backupPath string) error {
	data, err := ioutil.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	var configs map[string]interface{}
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("解析备份文件失败: %v", err)
	}

	cm.mutex.Lock()
	cm.configs = configs
	cm.mutex.Unlock()

	// 将配置写回文件
	for filename, config := range configs {
		filePath := filepath.Join(cm.configDir, filename)
		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			cm.logger.Printf("序列化配置失败 %s: %v", filename, err)
			continue
		}

		if err := ioutil.WriteFile(filePath, configData, 0644); err != nil {
			cm.logger.Printf("写入配置文件失败 %s: %v", filename, err)
			continue
		}
	}

	cm.logger.Printf("配置恢复完成: %s", backupPath)
	return nil
}