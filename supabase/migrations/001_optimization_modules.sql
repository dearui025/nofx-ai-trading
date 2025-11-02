-- NOFX AI决策系统优化模块数据库迁移
-- 创建时间: 2024-12-27
-- 描述: 添加市场状态检测、相关性风险控制、信号强度量化、灾难恢复管理等优化功能的数据表

-- 1. 市场状态检测表
CREATE TABLE IF NOT EXISTS market_regime_analysis (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    symbol VARCHAR(20) NOT NULL DEFAULT 'BTCUSDT',
    regime VARCHAR(20) NOT NULL, -- 'bull_market', 'bear_market', 'sideways'
    confidence DECIMAL(5,4) NOT NULL, -- 0.0000 to 1.0000
    volatility_level VARCHAR(10) NOT NULL, -- 'low', 'medium', 'high'
    trend_strength DECIMAL(5,4) NOT NULL, -- -1.0000 to 1.0000
    reasoning TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. 相关性分析表
CREATE TABLE IF NOT EXISTS correlation_analysis (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    symbol1 VARCHAR(20) NOT NULL,
    symbol2 VARCHAR(20) NOT NULL,
    correlation DECIMAL(6,4) NOT NULL, -- -1.0000 to 1.0000
    risk_level VARCHAR(10) NOT NULL, -- 'low', 'medium', 'high'
    lookback_period INTEGER NOT NULL DEFAULT 20,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. 信号强度分析表
CREATE TABLE IF NOT EXISTS signal_strength_analysis (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    symbol VARCHAR(20) NOT NULL,
    score DECIMAL(5,2) NOT NULL, -- 0.00 to 100.00
    direction VARCHAR(10) NOT NULL, -- 'bullish', 'bearish', 'neutral'
    confidence DECIMAL(5,4) NOT NULL, -- 0.0000 to 1.0000
    price_action_score DECIMAL(5,2) NOT NULL,
    volume_score DECIMAL(5,2) NOT NULL,
    indicator_score DECIMAL(5,2) NOT NULL,
    timeframe_score DECIMAL(5,2) NOT NULL,
    reasoning TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. SOS事件记录表
CREATE TABLE IF NOT EXISTS sos_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trigger_type VARCHAR(20) NOT NULL, -- 'drawdown', 'margin', 'equity'
    trigger_value DECIMAL(10,4) NOT NULL,
    threshold_value DECIMAL(10,4) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'active', 'resolved'
    recommended_actions TEXT[], -- 数组存储建议行动
    actual_actions TEXT[], -- 数组存储实际执行的行动
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. 对冲记录表
CREATE TABLE IF NOT EXISTS hedge_records (
    id BIGSERIAL PRIMARY KEY,
    sos_event_id BIGINT REFERENCES sos_events(id),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hedge_type VARCHAR(20) NOT NULL, -- 'emergency_close', 'partial_hedge', 'full_hedge'
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- 'long', 'short'
    quantity DECIMAL(20,8) NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    hedge_ratio DECIMAL(5,4), -- 对冲比例
    success BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. 优化配置表
CREATE TABLE IF NOT EXISTS optimization_config (
    id BIGSERIAL PRIMARY KEY,
    module_name VARCHAR(50) NOT NULL UNIQUE,
    config_data JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_market_regime_timestamp ON market_regime_analysis(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_market_regime_symbol ON market_regime_analysis(symbol);

CREATE INDEX IF NOT EXISTS idx_correlation_timestamp ON correlation_analysis(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_correlation_symbols ON correlation_analysis(symbol1, symbol2);

CREATE INDEX IF NOT EXISTS idx_signal_strength_timestamp ON signal_strength_analysis(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_signal_strength_symbol ON signal_strength_analysis(symbol);

CREATE INDEX IF NOT EXISTS idx_sos_events_timestamp ON sos_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sos_events_status ON sos_events(status);

CREATE INDEX IF NOT EXISTS idx_hedge_records_timestamp ON hedge_records(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_hedge_records_sos_event ON hedge_records(sos_event_id);

CREATE INDEX IF NOT EXISTS idx_optimization_config_module ON optimization_config(module_name);

-- 插入默认配置
INSERT INTO optimization_config (module_name, config_data) VALUES
('market_regime', '{
    "lookback_period": 20,
    "volatility_threshold": 0.02,
    "trend_threshold": 0.1,
    "confidence_threshold": 0.7
}'),
('correlation_risk', '{
    "max_correlation": 0.8,
    "lookback_period": 20,
    "update_interval": 300
}'),
('signal_strength', '{
    "min_score_threshold": 75.0,
    "confidence_threshold": 0.75,
    "timeframes": ["1m", "5m", "15m", "1h"]
}'),
('disaster_recovery', '{
    "max_drawdown_pct": 0.15,
    "max_margin_used_pct": 0.90,
    "min_equity_threshold": 1000.0
}')
ON CONFLICT (module_name) DO NOTHING;

-- 添加注释
COMMENT ON TABLE market_regime_analysis IS '市场状态检测分析结果';
COMMENT ON TABLE correlation_analysis IS '币种相关性分析结果';
COMMENT ON TABLE signal_strength_analysis IS '信号强度量化分析结果';
COMMENT ON TABLE sos_events IS 'SOS紧急事件记录';
COMMENT ON TABLE hedge_records IS '对冲操作记录';
COMMENT ON TABLE optimization_config IS '优化模块配置';