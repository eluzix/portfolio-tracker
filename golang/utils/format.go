package utils

import (
	"fmt"
	"strings"
)

func addCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	return addCommas(s[:n-3]) + "," + s[n-3:]
}

func ToCurrencyString(val int64, precision int, currency string, rate float64) string {
	amount := float64(val) * rate / 100.0
	formatStr := fmt.Sprintf("%%.%df", precision)
	formatted := fmt.Sprintf(formatStr, amount)

	if precision == 0 {
		intPart := formatted
		if strings.HasPrefix(intPart, "-") {
			intPart = "-" + addCommas(intPart[1:])
		} else {
			intPart = addCommas(intPart)
		}
		return fmt.Sprintf("%s%s", currency, intPart)
	}

	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	decPart := parts[1]

	if strings.HasPrefix(intPart, "-") {
		intPart = "-" + addCommas(intPart[1:])
	} else {
		intPart = addCommas(intPart)
	}

	return fmt.Sprintf("%s%s.%s", currency, intPart, decPart)
}

func ToCurrencyStringUSD(val int64, precision int) string {
	return ToCurrencyString(val, precision, "$", 1.0)
}

func ToYieldString(val float32) string {
	return fmt.Sprintf("%.2f%%", val*100)
}
