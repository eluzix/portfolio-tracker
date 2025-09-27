package types

type AnalyzedPortfolio struct {
	Value              int64
	TotalInvested      int64
	TotalWithdrawn     int64
	TotalDividends     int64
	GainValue          int64
	Gain               float32
	AnnualizedYield    float32
	ModifiedDietzYield float32

	FirstTransaction Transaction
	LastTransaction  Transaction

	SymbolsCount map[string]int32
}

func NewAnalyzedPortfolio() AnalyzedPortfolio {
	return AnalyzedPortfolio{}
}
