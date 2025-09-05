package types

type AnalyzedPortfolio struct {
	Value              int64
	TotalInvested      int64
	TotalWithdrawn     int64
	TotalDividends     int64
	GainValue          int64
	Gain               int32
	AnnualizedYield    int32
	ModifiedDietzYield int32

	FirstTransaction Transaction
	LastTransaction  Transaction
}

func NewAnalyzedPortfolio() AnalyzedPortfolio {
	return AnalyzedPortfolio{}
}
