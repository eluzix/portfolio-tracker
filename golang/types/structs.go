package types

type AnalysisData struct {
	Accounts     *[]Account
	AccountsData map[string]AnalyzedPortfolio
	ExchangeRate float64
	ExchaneSign  string
}

func NewAnalysisData(accounts *[]Account, accountsData map[string]AnalyzedPortfolio) AnalysisData {
	return AnalysisData{
		Accounts:     accounts,
		AccountsData: accountsData,
		ExchangeRate: 1.0,
		ExchaneSign:  "$",
	}
}
