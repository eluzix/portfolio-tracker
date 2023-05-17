import argparse
import sys

from diskcache import Cache
from rich.traceback import install

from tracker.google_sheets_utils import list_sheets, collect_all_transactions
from tracker.portfolio_analysis import analyze_portfolio
from tracker.providers import extract_symbols_prices_from_transactions
from tracker.providers.alpha_vantage_utils import get_dividends_as_transactions
from tracker.utils import console

install()

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
        with console.status("[bold green]Collecting transactions..."):
            transactions = collect_all_transactions()
            cache.set('transactions', transactions, 60 * 60 * 24 * 7)
    symbols = set(t["symbol"] for t in transactions)

    dividends = cache.get('dividends')
    if not dividends:
        with console.status("[bold green]Collecting dividends..."):
            dividends = get_dividends_as_transactions(symbols)
            cache.set('dividends', dividends, 60 * 60 * 24 * 7)

    prices = cache.get('prices')
    if prices is None:
        with console.status("[bold green]Updating prices..."):
            prices = extract_symbols_prices_from_transactions(transactions)
            cache.set('prices', prices, 60 * 60 * 12)

    kwargs = {
        'dividend_tax_rate': 0.25,
        # 'as_currency': 'ILS',
        # 'currency_symbol': '₪',
    }

    filter_by_accounts = None
    if args.accounts:
        kwargs['filter_by_accounts'] = args.accounts
        filter_by_accounts = args.accounts

    data = analyze_portfolio(transactions, prices, dividends=dividends, **kwargs)
    currency_symbol = data['currency_symbol']
    exchange_rate = data['exchange_rate']

    # Output results
    if filter_by_accounts is not None:
        console.print(f"Filtering by accounts: {', '.join(filter_by_accounts)}\n")

    console.print("Total shares per symbol:")
    for symbol, value in data['shares'].items():
        console.print(f"{symbol}: {value}")

    console.print("\nAverage purchase price per share (PPS) per symbol:")
    for symbol, value in data['avg_pps'].items():
        console.print(f"{symbol}: {currency_symbol}{value * exchange_rate:.2f}")

    console.print("\nTotal dividends per symbol:")
    for symbol, value in data['all_dividends'].items():
        console.print(f"{symbol}: {currency_symbol}{value * exchange_rate:.2f}")

    console.print(f"\nCurrent portfolio value: {currency_symbol}{data['current_portfolio_value']:,.2f}")
    console.print(f"Total money invested: {currency_symbol}{data['total_invested']:,.2f}")
    console.print(f"Total money withdrawn: {currency_symbol}{data['total_withdrawn']:,.2f}")
    console.print(f"Total dividends: [bold green]{currency_symbol}{data['total_dividends']:,.2f}[/]")
    console.print(f"Portfolio gain: {currency_symbol}{data['portfolio_gain']:,.2f}")
    console.print(f"\nSimple Yield: {(data['portfolio_gain'] + data['total_dividends']) / data['current_portfolio_value']:.2%}")
    console.print(f"Annualized Yield: {data['annualized_yield']:.2%}")
    console.print(f"Modified Dietz Yield: [bold green]{data['modified_dietz_yield']:.2%}[/]")

    console.print(f'\n:raccoon: [bold purple]All Done![/] :raccoon:')
