package utils

import "testing"

func TestStringToDatePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_ = StringToDate("2006-01-02T00:00:00Z")
}
