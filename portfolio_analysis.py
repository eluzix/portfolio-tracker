import argparse
import sys
from datetime import datetime

from diskcache import Cache

from providers import extract_symbols_prices_from_transactions
from providers.alpha_vantage_utils import get_dividends_as_transactions
from providers.exchange_rates import get_exchange_rates
from google_sheets_utils import collect_all_transactions, list_sheets
from utils import TerminalColors


def analyze_portfolio(transactions: list, prices: dict, dividends: dict = None,
                      filter_by_accounts=None,
                      dividend_tax_rate=None,
                      as_currency=None,
                      currency_symbol='$'):
    today = datetime.now().strftime('%Y-%m-%d')

    # Process transactions
    shares = {}
    avg_pps = {}
    total_invested = 0
    total_withdrawn = 0
    cash_flows = []

    if filter_by_accounts is not None and len(filter_by_accounts) > 0:
        filter_by_accounts = [a.lower() for a in filter_by_accounts]
        transactions = [t for t in transactions if t["account"] in filter_by_accounts]

    symbols = set(t["symbol"] for t in transactions)
    if not dividends:
        first_date = min(t["date"] for t in transactions)
        dividends = get_dividends_as_transactions(symbols, first_date, today)

    prices = {symbol: data['adj_close'] for symbol, data in prices.items()}
    for symbol in dividends:
        if symbol in prices:
            transactions.extend(dividends[symbol])
    transactions.sort(key=lambda t: t["date"])

    all_dividends = {symbol: 0 for symbol in symbols}
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
            cash_flows.append((date, quantity * pps))

        elif transaction["type"] == "sell":
            total_withdrawn += quantity * pps
            shares[symbol] -= quantity
            cash_flows.append((date, 0 - (quantity * pps)))

        elif transaction["type"] == "dividend" and symbol in shares:
            # multiple pps with current number of shares
            if dividend_tax_rate is not None:
                pps *= (1 - dividend_tax_rate)

            transaction_value = pps * shares[symbol]
            all_dividends[symbol] += transaction_value
            cash_flows.append((date, 0 - transaction_value))

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

    total_dividends = sum(all_dividends.values())

    exchange_rate = get_exchange_rates([as_currency])[as_currency] if as_currency is not None else 1
    total_dividends *= exchange_rate
    portfolio_gain *= exchange_rate
    current_portfolio_value *= exchange_rate
    total_invested *= exchange_rate
    total_withdrawn *= exchange_rate

    # Output results
    if filter_by_accounts is not None:
        print(f"Filtering by accounts: {TerminalColors.color(', '.join(filter_by_accounts), TerminalColors.WARNING)}\n")

    print("Total shares per symbol:")
    for symbol, value in shares.items():
        print(f"{symbol}: {value}")

    print("\nAverage purchase price per share (PPS) per symbol:")
    for symbol, value in avg_pps.items():
        print(f"{symbol}: {currency_symbol}{value * exchange_rate:.2f}")

    print("\nTotal dividends per symbol:")
    for symbol, value in all_dividends.items():
        print(f"{symbol}: {currency_symbol}{value * exchange_rate:.2f}")

    print(f"\nCurrent portfolio value: {currency_symbol}{current_portfolio_value:,.2f}")
    print(f"Total money invested: {currency_symbol}{total_invested:,.2f}")
    print(f"Total money withdrawn: {currency_symbol}{total_withdrawn:,.2f}")
    print(f"Total dividends: {TerminalColors.OK_CYAN}{currency_symbol}{total_dividends:,.2f}{TerminalColors.END}")
    print(f"Portfolio gain: {currency_symbol}{portfolio_gain:,.2f}")
    print(f"\nSimple Yield: {(portfolio_gain + total_dividends) / current_portfolio_value:.2%}")
    print(f"Annualized Yield: {annualized_yield:.2%}")
    print(f"Modified Dietz Yield: {TerminalColors.OK_GREEN}{modified_dietz_yield:.2%}{TerminalColors.END}")


def list_accounts(transactions):
    accounts = set()
    for transaction in transactions:
        accounts.add(transaction["account"])
    return accounts


if __name__ == '__main__':
    # create the parser object
    parser = argparse.ArgumentParser()

    # add the list sheets and clear cache arguments
    parser.add_argument('--list-accounts', action='store_true', help='List available accounts')
    parser.add_argument("--accounts", nargs='+', help="Provide a list of accounts")
    parser.add_argument('--reset-prices', action='store_true', help='re-fetch prices')
    parser.add_argument('--clear-cache', action='store_true', help='Clear the cache')

    # parse the command-line arguments
    args = parser.parse_args()
    if args.list_accounts:
        all_sheets = list_sheets()
        for sheet in all_sheets:
            print(sheet)
        sys.exit(0)

    cache = Cache('cache')
    if args.clear_cache:
        cache.clear()

    if args.reset_prices:
        cache.delete('prices')

    transactions = cache.get('transactions')
    if not transactions:
        transactions = collect_all_transactions()
        cache.set('transactions', transactions, 60 * 60 * 24 * 7)
    symbols = set(t["symbol"] for t in transactions)

    dividends = cache.get('dividends')
    if not dividends:
        dividends = get_dividends_as_transactions(symbols)
        cache.set('dividends', dividends, 60 * 60 * 24 * 7)

    prices = cache.get('prices')
    if prices is None:
        prices = extract_symbols_prices_from_transactions(transactions)
        cache.set('prices', prices, 60 * 60 * 12)

    kwargs = {
        'dividend_tax_rate': 0.25,
        # 'as_currency': 'ILS',
        # 'currency_symbol': '₪',
    }

    if args.accounts:
        kwargs['filter_by_accounts'] = args.accounts

    analyze_portfolio(transactions, prices, dividends=dividends, **kwargs)
