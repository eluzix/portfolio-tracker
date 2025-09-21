package utils

import (
	"fmt"
	"time"
)

// const formatYYYYMMDD = "2006-01-02T00:00:00Z"
const formatYYYYMMDD = "2006-01-02"

func StringToDate(v string) time.Time {
	ret, err := time.Parse(formatYYYYMMDD, v)
	if err != nil {
		panic(fmt.Sprintf("[StringToDate] Error formating date %s\n", v))
	}
	return ret
}
