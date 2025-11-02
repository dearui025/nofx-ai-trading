package config

import (
	"os"
	"strconv"
)

// LoadFromEnv 从环境变量加载配置，覆盖JSON配置
func (c *Config) LoadFromEnv() {
	// API服务器端口
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.APIServerPort = p
		}
	}
	if port := os.Getenv("API_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.APIServerPort = p
		}
	}

	// 风控参数
	if maxLoss := os.Getenv("MAX_DAILY_LOSS"); maxLoss != "" {
		if ml, err := strconv.ParseFloat(maxLoss, 64); err == nil {
			c.MaxDailyLoss = ml
		}
	}
	if maxDrawdown := os.Getenv("MAX_DRAWDOWN"); maxDrawdown != "" {
		if md, err := strconv.ParseFloat(maxDrawdown, 64); err == nil {
			c.MaxDrawdown = md
		}
	}
	if stopMinutes := os.Getenv("STOP_TRADING_MINUTES"); stopMinutes != "" {
		if sm, err := strconv.Atoi(stopMinutes); err == nil {
			c.StopTradingMinutes = sm
		}
	}

	// 杠杆配置
	if btcEthLev := os.Getenv("BTC_ETH_LEVERAGE"); btcEthLev != "" {
		if lev, err := strconv.Atoi(btcEthLev); err == nil {
			c.Leverage.BTCETHLeverage = lev
		}
	}
	if altLev := os.Getenv("ALTCOIN_LEVERAGE"); altLev != "" {
		if lev, err := strconv.Atoi(altLev); err == nil {
			c.Leverage.AltcoinLeverage = lev
		}
	}

	// 币种池配置
	if coinPoolAPI := os.Getenv("COIN_POOL_API_URL"); coinPoolAPI != "" {
		c.CoinPoolAPIURL = coinPoolAPI
	}
	if oiTopAPI := os.Getenv("OI_TOP_API_URL"); oiTopAPI != "" {
		c.OITopAPIURL = oiTopAPI
	}
	if useDefault := os.Getenv("USE_DEFAULT_COINS"); useDefault != "" {
		if ud, err := strconv.ParseBool(useDefault); err == nil {
			c.UseDefaultCoins = ud
		}
	}

	// 为每个trader加载环境变量
	for i := range c.Traders {
		c.Traders[i].LoadFromEnv()
	}
}

// LoadFromEnv 为单个trader从环境变量加载配置
func (tc *TraderConfig) LoadFromEnv() {
	// 币安配置
	if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
		tc.BinanceAPIKey = apiKey
	}
	if secretKey := os.Getenv("BINANCE_SECRET_KEY"); secretKey != "" {
		tc.BinanceSecretKey = secretKey
	}
	if testnet := os.Getenv("BINANCE_TESTNET"); testnet != "" {
		if tn, err := strconv.ParseBool(testnet); err == nil {
			tc.BinanceTestnet = tn
		}
	}

	// Hyperliquid配置
	if privateKey := os.Getenv("HYPERLIQUID_PRIVATE_KEY"); privateKey != "" {
		tc.HyperliquidPrivateKey = privateKey
	}
	if walletAddr := os.Getenv("HYPERLIQUID_WALLET_ADDR"); walletAddr != "" {
		tc.HyperliquidWalletAddr = walletAddr
	}
	if testnet := os.Getenv("HYPERLIQUID_TESTNET"); testnet != "" {
		if tn, err := strconv.ParseBool(testnet); err == nil {
			tc.HyperliquidTestnet = tn
		}
	}

	// Aster配置
	if user := os.Getenv("ASTER_USER"); user != "" {
		tc.AsterUser = user
	}
	if signer := os.Getenv("ASTER_SIGNER"); signer != "" {
		tc.AsterSigner = signer
	}
	if privateKey := os.Getenv("ASTER_PRIVATE_KEY"); privateKey != "" {
		tc.AsterPrivateKey = privateKey
	}

	// AI模型配置
	if qwenKey := os.Getenv("QWEN_API_KEY"); qwenKey != "" {
		tc.QwenKey = qwenKey
	}
	if deepSeekKey := os.Getenv("DEEPSEEK_API_KEY"); deepSeekKey != "" {
		tc.DeepSeekKey = deepSeekKey
	}

	// 自定义API配置
	if customURL := os.Getenv("CUSTOM_API_URL"); customURL != "" {
		tc.CustomAPIURL = customURL
	}
	if customKey := os.Getenv("CUSTOM_API_KEY"); customKey != "" {
		tc.CustomAPIKey = customKey
	}
	if customModel := os.Getenv("CUSTOM_MODEL_NAME"); customModel != "" {
		tc.CustomModelName = customModel
	}
}