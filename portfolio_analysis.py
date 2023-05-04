import json
from datetime import datetime

from google_sheets_utils import collect_all_transactions


def analyze_portfolio(transactions: list, prices: dict):
    today = datetime.now().strftime('%Y-%m-%d')

    # Process transactions
    shares = {}
    avg_pps = {}
    total_invested = 0
    total_withdrawn = 0
    cash_flows = []

    for transaction in transactions:
        date = datetime.strptime(transaction["date"], "%Y-%m-%d")
        symbol = transaction["symbol"]
        quantity = float(transaction["quantity"])
        pps = float(transaction["pps"])

        if transaction["type"] == "buy":
            if symbol not in shares:
                shares[symbol] = 0
                avg_pps[symbol] = 0

            total_invested += quantity * pps
            avg_pps[symbol] = (avg_pps[symbol] * shares[symbol] + quantity * pps) / (shares[symbol] + quantity)
            shares[symbol] += quantity
            cash_flows.append((date, -quantity * pps))

        elif transaction["type"] == "sell":
            total_withdrawn += quantity * pps
            shares[symbol] -= quantity
            cash_flows.append((date, quantity * pps))

    # Calculate other metrics
    current_portfolio_value = sum(prices[symbol] * shares[symbol] for symbol in shares)
    portfolio_gain = current_portfolio_value - total_invested + total_withdrawn

    # Calculate Modified Dietz Yield
    today_date = datetime.strptime(today, "%Y-%m-%d")
    days_since_start = (today_date - cash_flows[0][0]).days
    weighted_cash_flows = sum((cf[1] * (today_date - cf[0]).days / days_since_start) for cf in cash_flows)
    modified_dietz_yield = (portfolio_gain) / (total_invested + weighted_cash_flows)

    # Calculate Annualized Yield
    years_since_start = days_since_start / 365
    annualized_yield = ((1 + modified_dietz_yield) ** (1 / years_since_start)) - 1

    # Output results
    print("Total shares per symbol:")
    for symbol, value in shares.items():
        print(f"{symbol}: {value}")

    print("\nAverage purchase price per share (PPS) per symbol:")
    for symbol, value in avg_pps.items():
        print(f"{symbol}: ${value:.2f}")

    print(f"\nCurrent portfolio value: ${current_portfolio_value:,.2f}")
    print(f"Total money invested: ${total_invested:,.2f}")
    print(f"Total money withdrawn: ${total_withdrawn:,.2f}")
    print(f"Portfolio gain: ${portfolio_gain:,.2f}")
    print(f"Modified Dietz Yield: {modified_dietz_yield:.2%}")
    print(f"Annualized Yield: {annualized_yield:.2%}")


if __name__ == '__main__':
    # all_sheets = list_sheets()
    # for sheet in all_sheets:
    #     print(sheet)
    transactions = collect_all_transactions()
    # with open('all_transactions.json', 'w') as f:
    #     json.dump(transactions, f)

    # with open('all_transactions.json', 'r') as f:
    #     transactions = json.load(f)

    # for transaction in transactions:
    #     print(transaction)

    # prices = extract_symbols_prices_from_transactions(transactions)
    # with open('prices.json', 'w') as f:
    #     json.dump(prices, f)

    with open('prices.json', 'r') as f:
        prices = json.load(f)

    analyze_portfolio(transactions, prices)
