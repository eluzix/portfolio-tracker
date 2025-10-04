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

func ToCurrencyString(val int64, precision int) string {
	amount := float64(val) / 100.0
	formatStr := fmt.Sprintf("%%.%df", precision)
	formatted := fmt.Sprintf(formatStr, amount)

	if precision == 0 {
		intPart := formatted
		if strings.HasPrefix(intPart, "-") {
			intPart = "-" + addCommas(intPart[1:])
		} else {
			intPart = addCommas(intPart)
		}
		return fmt.Sprintf("$%s", intPart)
	}

	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	decPart := parts[1]

	if strings.HasPrefix(intPart, "-") {
		intPart = "-" + addCommas(intPart[1:])
	} else {
		intPart = addCommas(intPart)
	}

	return fmt.Sprintf("$%s.%s", intPart, decPart)
}

func ToYieldString(val float32) string {
	return fmt.Sprintf("%.2f%%", val*100)
}
