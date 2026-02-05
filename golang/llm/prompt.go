package llm

const SystemPrompt = `You are an expert financial analyst specializing in personal investment portfolios. 
Your role is to provide actionable insights and analysis of a user's investment portfolio.

When analyzing a portfolio, focus on:
1. **Diversification**: Evaluate how well the portfolio is diversified across different asset classes, sectors, and geographies.
2. **Sector Allocation**: Analyze the distribution of holdings across different sectors and identify any concentration risks.
3. **Risk Exposure**: Assess the overall risk profile of the portfolio, including volatility, correlation, and concentration risks.
4. **Performance vs Benchmarks**: Compare the portfolio's performance against relevant benchmarks and market indices.

Provide actionable insights that are:
- Clear and concise
- Based on the provided data
- Specific to the user's holdings
- Focused on potential concerns and opportunities

Format your response as markdown with clear headers and sections.

**IMPORTANT DISCLAIMER**: This analysis is for informational purposes only and should not be considered as financial advice. 
Always consult with a qualified financial advisor before making investment decisions.`

// EstimateTokens estimates the token count in a string (rough approximation: ~4 chars per token)
func EstimateTokens(text string) int {
	return len(text) / 4
}

// ContextBuilder builds a structured prompt context for portfolio analysis
func ContextBuilder(accountName string, holdings []PortfolioData, transactions []TransactionData, metrics MetricsData) string {
	context := ""

	context += "PORTFOLIO CONTEXT:\n"
	context += "==================\n\n"

	if accountName != "" {
		context += "Account: " + accountName + "\n"
	}
	context += "Analysis Date: " + metrics.Date + "\n\n"

	context += "HOLDINGS:\n"
	context += "---------\n"
	for _, h := range holdings {
		context += "- " + h.Symbol + " (" + h.Name + ")\n"
		context += "  Shares: " + h.Quantity + "\n"
		context += "  Current Price: $" + h.CurrentPrice + "\n"
		context += "  Current Value: $" + h.CurrentValue + "\n"
		context += "  Sector: " + h.Sector + "\n"
		context += "  Allocation: " + h.AllocationPercent + "%\n\n"
	}

	context += "PERFORMANCE METRICS:\n"
	context += "-------------------\n"
	context += "Total Portfolio Value: $" + metrics.TotalValue + "\n"
	context += "Total Cost Basis: $" + metrics.CostBasis + "\n"
	context += "Unrealized Gain/Loss: $" + metrics.UnrealizedGain + " (" + metrics.UnrealizedGainPercent + "%)\n"
	context += "Total Dividends Received: $" + metrics.DividendsReceived + "\n"
	context += "Yield on Cost: " + metrics.YieldOnCost + "%\n\n"

	if len(transactions) > 0 {
		context += "RECENT TRANSACTIONS:\n"
		context += "-------------------\n"
		context += buildTransactionContext(transactions, 8000)
	}

	return context
}

// buildTransactionContext builds transaction section and truncates oldest transactions if needed to stay under token limit
func buildTransactionContext(transactions []TransactionData, maxTokens int) string {
	const maxTotalTokens = 8000
	const systemPromptTokens = 500
	
	tokenBudget := maxTotalTokens - systemPromptTokens - maxTokens
	if tokenBudget < 200 {
		tokenBudget = 200
	}

	txnText := ""
	currentTokens := 0

	// Start from most recent (assume transactions are sorted newest first)
	for i := 0; i < len(transactions); i++ {
		t := transactions[i]
		line := "- " + t.Date + ": " + t.Action + " " + t.Quantity + " shares of " + t.Symbol + " @ $" + t.Price + "\n"
		lineTokens := EstimateTokens(line)

		if currentTokens+lineTokens > tokenBudget {
			txnText += "\n(Additional older transactions truncated to stay within token limit)\n"
			break
		}

		txnText += line
		currentTokens += lineTokens
	}

	return txnText + "\n"
}

// PortfolioData represents a single holding in the portfolio
type PortfolioData struct {
	Symbol             string
	Name               string
	Quantity           string
	CurrentPrice       string
	CurrentValue       string
	Sector             string
	AllocationPercent  string
}

// TransactionData represents a transaction (buy, sell, dividend)
type TransactionData struct {
	Date     string
	Action   string
	Symbol   string
	Quantity string
	Price    string
}

// MetricsData represents portfolio performance metrics
type MetricsData struct {
	Date                   string
	TotalValue            string
	CostBasis             string
	UnrealizedGain        string
	UnrealizedGainPercent string
	DividendsReceived     string
	YieldOnCost           string
}
