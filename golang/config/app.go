package config

import (
	"math"
	"os"
	"strconv"
	"strings"
)

const DefaultDividendTaxRate = 0.25

type AppConfig struct {
	DividendTaxRate float64
}

func Load() AppConfig {
	rate := DefaultDividendTaxRate
	raw := strings.TrimSpace(os.Getenv("TRACKER_DIVIDEND_TAX_RATE"))
	if raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			if parsed > 1 {
				parsed = parsed / 100
			}
			if parsed >= 0 && parsed <= 1 {
				rate = parsed
			}
		}
	}

	return AppConfig{DividendTaxRate: rate}
}

func DividendsAfterTax(totalDividends int64, taxRate float64) int64 {
	netRatio := 1 - taxRate
	if netRatio < 0 {
		netRatio = 0
	}
	if netRatio > 1 {
		netRatio = 1
	}

	return int64(math.Round(float64(totalDividends) * netRatio))
}
