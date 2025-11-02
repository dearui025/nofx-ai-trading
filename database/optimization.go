package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"nofx/disaster"
	"nofx/market"
	"nofx/risk"
	"nofx/signal"

	_ "github.com/lib/pq"
)

// OptimizationDB 优化功能数据库访问层
type OptimizationDB struct {
	db *sql.DB
}

// NewOptimizationDB 创建新的优化数据库访问实例
func NewOptimizationDB(connectionString string) (*OptimizationDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OptimizationDB{db: db}, nil
}

// Close 关闭数据库连接
func (odb *OptimizationDB) Close() error {
	return odb.db.Close()
}

// === 市场状态检测相关 ===

// SaveMarketRegimeAnalysis 保存市场状态分析结果
func (odb *OptimizationDB) SaveMarketRegimeAnalysis(symbol string, result *market.RegimeAnalysis) error {
	query := `
		INSERT INTO market_regime_analysis (timestamp, symbol, regime, confidence, volatility_level, trend_strength, reasoning)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := odb.db.Exec(query, time.Now(), symbol, result.Regime, result.Confidence,
		result.Volatility, result.TrendStrength, "")
	return err
}

// GetLatestMarketRegime 获取最新的市场状态
func (odb *OptimizationDB) GetLatestMarketRegime(symbol string) (*market.RegimeAnalysis, error) {
	query := `
		SELECT regime, confidence, volatility_level, trend_strength, reasoning, timestamp
		FROM market_regime_analysis
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var result market.RegimeAnalysis
	var timestamp time.Time
	var reasoning string
	err := odb.db.QueryRow(query, symbol).Scan(
		&result.Regime, &result.Confidence, &result.Volatility,
		&result.TrendStrength, &reasoning, &timestamp,
	)
	if err != nil {
		return nil, err
	}
	result.LastUpdated = timestamp
	return &result, nil
}

// === 相关性分析相关 ===

// SaveCorrelationAnalysis 保存相关性分析结果
func (odb *OptimizationDB) SaveCorrelationAnalysis(pairs []risk.CorrelationPair) error {
	if len(pairs) == 0 {
		return nil
	}

	tx, err := odb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO correlation_analysis (timestamp, symbol1, symbol2, correlation, risk_level, lookback_period)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	timestamp := time.Now()
	for _, pair := range pairs {
		riskLevel := "low"
		if pair.Correlation > 0.8 {
			riskLevel = "high"
		} else if pair.Correlation > 0.6 {
			riskLevel = "medium"
		}
		_, err := stmt.Exec(timestamp, pair.Symbol1, pair.Symbol2, pair.Correlation, riskLevel, 20)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetHighCorrelationPairs 获取高相关性币种对
func (odb *OptimizationDB) GetHighCorrelationPairs(threshold float64, hours int) ([]risk.CorrelationPair, error) {
	query := `
		SELECT DISTINCT ON (symbol1, symbol2) symbol1, symbol2, correlation, risk_level
		FROM correlation_analysis
		WHERE timestamp > $1 AND ABS(correlation) > $2
		ORDER BY symbol1, symbol2, timestamp DESC
	`
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := odb.db.Query(query, since, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pairs []risk.CorrelationPair
	for rows.Next() {
		var pair risk.CorrelationPair
		var riskLevel string
		err := rows.Scan(&pair.Symbol1, &pair.Symbol2, &pair.Correlation, &riskLevel)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

// === 信号强度分析相关 ===

// SaveSignalStrengthAnalysis 保存信号强度分析结果
func (odb *OptimizationDB) SaveSignalStrengthAnalysis(symbol string, result *signal.SignalStrengthResult) error {
	query := `
		INSERT INTO signal_strength_analysis (timestamp, symbol, score, direction, confidence, 
			price_action_score, volume_score, indicator_score, timeframe_score, reasoning)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := odb.db.Exec(query, time.Now(), symbol, result.OverallScore, result.Direction,
		result.Confidence, result.PriceActionScore, result.VolumeScore,
		result.IndicatorScore, result.TimeframeScore, result.Reasoning)
	return err
}

// GetLatestSignalStrength 获取最新的信号强度
func (odb *OptimizationDB) GetLatestSignalStrength(symbol string) (*signal.SignalStrengthResult, error) {
	query := `
		SELECT score, direction, confidence, price_action_score, volume_score, 
			indicator_score, timeframe_score, reasoning, timestamp
		FROM signal_strength_analysis
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var result signal.SignalStrengthResult
	var timestamp time.Time
	err := odb.db.QueryRow(query, symbol).Scan(
		&result.OverallScore, &result.Direction, &result.Confidence,
		&result.PriceActionScore, &result.VolumeScore, &result.IndicatorScore,
		&result.TimeframeScore, &result.Reasoning, &timestamp,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTopSignalStrengths 获取信号强度最高的币种
func (odb *OptimizationDB) GetTopSignalStrengths(limit int, hours int) (map[string]*signal.SignalStrengthResult, error) {
	query := `
		SELECT DISTINCT ON (symbol) symbol, score, direction, confidence, 
			price_action_score, volume_score, indicator_score, timeframe_score, reasoning, timestamp
		FROM signal_strength_analysis
		WHERE timestamp > $1
		ORDER BY symbol, timestamp DESC, score DESC
		LIMIT $2
	`
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := odb.db.Query(query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*signal.SignalStrengthResult)
	for rows.Next() {
		var symbol string
		var result signal.SignalStrengthResult
		var timestamp time.Time
		err := rows.Scan(&symbol, &result.OverallScore, &result.Direction, &result.Confidence,
			&result.PriceActionScore, &result.VolumeScore, &result.IndicatorScore,
			&result.TimeframeScore, &result.Reasoning, &timestamp)
		if err != nil {
			return nil, err
		}
		results[symbol] = &result
	}
	return results, nil
}

// === SOS事件相关 ===

// SaveSOSEvent 保存SOS事件
func (odb *OptimizationDB) SaveSOSEvent(event *disaster.SOSEvent) (int64, error) {
	query := `
		INSERT INTO sos_events (timestamp, trigger_type, trigger_value, threshold_value, 
			status, recommended_actions)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int64
	actionsJSON, _ := json.Marshal(event.Actions)
	err := odb.db.QueryRow(query, event.ActivatedAt, event.TriggerCondition, 0.0,
		0.0, event.Status.String(), string(actionsJSON)).Scan(&id)
	return id, err
}

// UpdateSOSEvent 更新SOS事件状态
func (odb *OptimizationDB) UpdateSOSEvent(id int64, status string, actualActions []string) error {
	query := `
		UPDATE sos_events 
		SET status = $1, actual_actions = $2, resolved_at = $3
		WHERE id = $4
	`
	resolvedAt := time.Now()
	if status != "resolved" {
		resolvedAt = time.Time{}
	}
	_, err := odb.db.Exec(query, status, actualActions, resolvedAt, id)
	return err
}

// GetActiveSOSEvents 获取活跃的SOS事件
func (odb *OptimizationDB) GetActiveSOSEvents() ([]*disaster.SOSEvent, error) {
	query := `
		SELECT id, timestamp, trigger_type, trigger_value, threshold_value, 
			recommended_actions, actual_actions
		FROM sos_events
		WHERE status = 'active'
		ORDER BY timestamp DESC
	`
	rows, err := odb.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*disaster.SOSEvent
	for rows.Next() {
		var event disaster.SOSEvent
		var actualActions []string
		var timestamp time.Time
		var triggerType string
		var triggerValue, thresholdValue float64
		var recommendedActions string
		err := rows.Scan(&event.ID, &timestamp, &triggerType,
			&triggerValue, &thresholdValue, &recommendedActions, &actualActions)
		if err != nil {
			return nil, err
		}
		event.ActivatedAt = timestamp
		event.TriggerCondition = triggerType
		json.Unmarshal([]byte(recommendedActions), &event.Actions)
		events = append(events, &event)
	}
	return events, nil
}

// SaveHedgeRecord 保存对冲记录
func (odb *OptimizationDB) SaveHedgeRecord(record *disaster.HedgeRecord) error {
	query := `
		INSERT INTO hedge_records (sos_event_id, timestamp, hedge_type, symbol, side, 
			quantity, price, hedge_ratio, success, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := odb.db.Exec(query, "", record.ExecutedAt, "emergency",
		record.Symbol, record.HedgeSide, record.HedgeQuantity, 0.0, record.HedgeRatio,
		true, "")
	return err
}

// === 配置管理相关 ===

// GetOptimizationConfig 获取优化模块配置
func (odb *OptimizationDB) GetOptimizationConfig(moduleName string) (map[string]interface{}, error) {
	query := `
		SELECT config_data
		FROM optimization_config
		WHERE module_name = $1 AND is_active = true
	`
	var configData []byte
	err := odb.db.QueryRow(query, moduleName).Scan(&configData)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	err = json.Unmarshal(configData, &config)
	return config, err
}

// UpdateOptimizationConfig 更新优化模块配置
func (odb *OptimizationDB) UpdateOptimizationConfig(moduleName string, config map[string]interface{}) error {
	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	query := `
		UPDATE optimization_config 
		SET config_data = $1, updated_at = $2
		WHERE module_name = $3
	`
	_, err = odb.db.Exec(query, configData, time.Now(), moduleName)
	return err
}

// === 统计和分析相关 ===

// GetOptimizationStats 获取优化功能统计信息
func (odb *OptimizationDB) GetOptimizationStats(hours int) (map[string]interface{}, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	stats := make(map[string]interface{})
	
	// 市场状态分析统计
	var regimeCount int
	err := odb.db.QueryRow(`
		SELECT COUNT(*) FROM market_regime_analysis WHERE timestamp > $1
	`, since).Scan(&regimeCount)
	if err != nil {
		return nil, err
	}
	stats["market_regime_analyses"] = regimeCount
	
	// 相关性分析统计
	var correlationCount int
	err = odb.db.QueryRow(`
		SELECT COUNT(*) FROM correlation_analysis WHERE timestamp > $1
	`, since).Scan(&correlationCount)
	if err != nil {
		return nil, err
	}
	stats["correlation_analyses"] = correlationCount
	
	// 信号强度分析统计
	var signalCount int
	err = odb.db.QueryRow(`
		SELECT COUNT(*) FROM signal_strength_analysis WHERE timestamp > $1
	`, since).Scan(&signalCount)
	if err != nil {
		return nil, err
	}
	stats["signal_strength_analyses"] = signalCount
	
	// SOS事件统计
	var sosCount int
	err = odb.db.QueryRow(`
		SELECT COUNT(*) FROM sos_events WHERE timestamp > $1
	`, since).Scan(&sosCount)
	if err != nil {
		return nil, err
	}
	stats["sos_events"] = sosCount
	
	return stats, nil
}