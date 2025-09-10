package types

import "time"

type SymbolPrice struct {
	Symbol    string
	AdjPrice  int32
	CreatedAt time.Time
}
