package utils

import "fmt"

func ToCurrencyString(val int64) string {
	return fmt.Sprintf("$%.2f", float64(val)/100.0)
}
