package analyzer

import (
	"testing"
	"tracker/types"
)

func TestFirstLastTransaction(t *testing.T) {
	transactions := []types.Transaction{
		{Id: "id1", Date: "2025-01-01", Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
		{Id: "id2", Date: "2025-02-01", Type: types.TransactionTypeBuy, Symbol: "AAPL", Pps: 1, Quantity: 1},
	}
	portfolio, err := AnalyzeTransactions(transactions, map[string]types.SymbolPrice{
		"AAPL": {AdjPrice: 12},
	})
	if err != nil {
		t.Fatalf("Error wasn't nil: %e\n", err)
	}

	if portfolio.FirstTransaction == (types.Transaction{}) {
		t.Fatalf("expected FirstTransaction to be NOT nil %e\n", err)
	}

	if portfolio.FirstTransaction.Id != "id1" {
		t.Fatalf("expected FirstTransaction.Id to be id1 but got %s\n", portfolio.FirstTransaction.Id)
	}

	if portfolio.LastTransaction.Id != "id2" {
		t.Fatalf("expected LastTransaction.Id to be id2 but got %s\n", portfolio.LastTransaction.Id)
	}
}
