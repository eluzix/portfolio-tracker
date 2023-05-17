import argparse
import sys

from diskcache import Cache
from rich.table import Table
from rich.traceback import install

from tracker import store
from tracker.google_sheets_utils import list_accounts
from tracker.portfolio_analysis import analyze_portfolio
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
    parser.add_argument("--currency", type=str, default="USD", help="Provide a currency")
    parser.add_argument("--dividend-rate", type=float, default=0.25, help="Provide a dividend tax rate")

    # parse the command-line arguments
    args = parser.parse_args()
    if args.list_accounts:
        all_sheets = list_accounts()
        for sheet in all_sheets:
            print(sheet)
        sys.exit(0)

    cache = Cache('cache')
    if args.clear_cache:
        cache.clear()

    if args.reset_prices:
        cache.delete('prices')

    filter_by_accounts = None
    if args.accounts:
        filter_by_accounts = args.accounts

    transactions = store.load_transactions(filter_by_accounts=filter_by_accounts)

    kwargs = {
        'dividend_tax_rate': args.dividend_rate,
    }

    currency_symbol = '$'
    if args.currency != 'USD':
        currency_symbol = '₪' if args.currency == 'ILS' else '$'
        kwargs['as_currency'] = args.currency

    all_data = analyze_portfolio(transactions, **kwargs)
    # console.print(all_data)

    totals = all_data.pop('total')
    exchange_rate = totals['exchange_rate']

    # Output results
    symbols_table = Table(show_header=True, header_style="bold pale_turquoise1", title="Symbols Summary")
    symbols_table.add_column("Symbol", style="dim", width=6)
    symbols_table.add_column("Shares", justify="right")
    symbols_table.add_column("PPS", justify="right")
    symbols_table.add_column("Share Value", justify="right")
    symbols_table.add_column("Dividends", justify="right")
    symbols_table.add_column("Value", justify="right")

    all_symbols = store.get_all_symbols(transactions)
    prices = store.load_prices()
    for symbol in all_symbols:
        quantity = totals['shares'].get(symbol, 0)
        avg_pps = totals['avg_pps'].get(symbol, 0)
        current_value = prices.get(symbol, {'adj_close': avg_pps})['adj_close']
        symbols_table.add_row(
            symbol,
            f"{quantity:,.0f}",
            f"{currency_symbol}{avg_pps * exchange_rate:,.2f}",
            f'{currency_symbol}{current_value * exchange_rate:,.2f}',
            f"{currency_symbol}{totals['all_dividends'].get(symbol, 0) * exchange_rate:,.0f}",
            f"{currency_symbol}{quantity * current_value * exchange_rate:,.0f}"
        )

    info_table = Table(show_header=True, header_style="bold pale_turquoise1", title="Portfolio Analysis")
    info_table.add_column('Account', style="dim", width=20)
    info_table.add_column('Total Invested', justify="right")
    info_table.add_column('Total Withdrawn', justify="right")
    info_table.add_column('Total Dividends', justify="right")
    info_table.add_column('Portfolio Gain', justify="right")
    info_table.add_column('Simple Yield', justify="right")
    info_table.add_column('Annualized Yield', justify="right")
    info_table.add_column('Modified Dietz Yield', justify="right", style="bold green_yellow")

    last_item = len(all_data) - 1
    for i, (account, data) in enumerate(all_data.items()):
        info_table.add_row(
            account,
            f"{currency_symbol}{data['total_invested']:,.0f}",
            f"{currency_symbol}{data['total_withdrawn']:,.0f}",
            f"{currency_symbol}{data['total_dividends']:,.0f}",
            f"{currency_symbol}{data['portfolio_gain']:,.0f}",
            f"{(data['portfolio_gain'] + data['total_dividends']) / data['current_portfolio_value']:.2%}",
            f"{data['annualized_yield']:.2%}",
            f"{data['modified_dietz_yield']:.2%}",
            end_section=i == last_item,
        )

    if last_item > 0:
        info_table.add_row(
            'Total',
            f"{currency_symbol}{totals['total_invested']:,.0f}",
            f"{currency_symbol}{totals['total_withdrawn']:,.0f}",
            f"{currency_symbol}{totals['total_dividends']:,.0f}",
            f"{currency_symbol}{totals['portfolio_gain']:,.0f}",
            f"{(totals['portfolio_gain'] + totals['total_dividends']) / totals['current_portfolio_value']:.2%}",
            f"{totals['annualized_yield']:.2%}",
            f"{totals['modified_dietz_yield']:.2%}",
            style="bold dark_orange",
        )

    console.print(symbols_table)
    console.print('\n')
    console.print(info_table)

    console.print(f'\n:raccoon: [bold purple]All Done![/] :raccoon:')
