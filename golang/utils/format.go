package utils

import "fmt"

func ToCurrencyString(val int64) string {
	return fmt.Sprintf("$%.2f", float64(val)/100.0)
}

func ToYieldString(val float32) string {
	return fmt.Sprintf("%.2f%%", val*100)
}
