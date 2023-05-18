import os
import typing

from diskcache import Cache

from tracker.google_sheets_utils import collect_all_transactions
from tracker.providers import extract_symbols_prices
from tracker.providers.alpha_vantage_utils import get_dividends_as_transactions
from tracker.providers.exchange_rates import get_exchange_rates
from tracker.utils import console

_cache = None


def get_cache():
    global _cache
    if _cache is None:
        dir_path = os.path.dirname(os.path.realpath(__file__))
        _cache = Cache(f'{dir_path}/../cache')

    return _cache


def load_transactions(filter_by_accounts: list = None):
    cache = get_cache()
    transactions = cache.get('transactions')
    if not transactions:
        with console.status("[bold green]Collecting transactions..."):
            transactions = collect_all_transactions()
            cache.set('transactions', transactions, 60 * 60 * 24 * 7)

    if transactions is None:
        console.print("[bold yellow]No transactions found![/]")
        transactions = []

    if filter_by_accounts is not None and len(filter_by_accounts) > 0:
        console.print(f"[bold]Filtering by accounts: {', '.join(filter_by_accounts)}[/]")
        filter_by_accounts = [a.lower() for a in filter_by_accounts]
        transactions = [t for t in transactions if t["account"] in filter_by_accounts]

    return transactions


def get_all_symbols(transactions: list = None):
    if transactions is None:
        transactions = load_transactions()

    return set(t["symbol"] for t in transactions)


def load_dividends(symbols: typing.Union[list, set] = None):
    cache = get_cache()
    dividends = cache.get('dividends')
    if not dividends:
        with console.status("[bold green]Collecting dividends...") as status:
            symbols = get_all_symbols()
            dividends = get_dividends_as_transactions(symbols)
            cache.set('dividends', dividends, 60 * 60 * 24 * 7)
            status.update(f"[bold green]Collected {len(dividends)} dividends[/]")

    if symbols is not None:
        dividends = [d for d in dividends if d["symbol"] in symbols]

    return dividends


def load_prices():
    cache = get_cache()
    prices = cache.get('prices')
    if not prices:
        with console.status("[bold green]Collecting prices...") as status:
            symbols = get_all_symbols()
            prices = extract_symbols_prices(symbols)
            cache.set('prices', prices, 60 * 60 * 12)
            status.update(f"[bold green]Collected {len(prices)} prices[/]")

    return prices


def load_exchange_rates(currency: str):
    cache = get_cache()
    key = f'exchange_rates:{currency}'
    exchange_rates = cache.get(key)
    if not exchange_rates:
        with console.status("[bold green]Collecting exchange rates...") as status:
            exchange_rates = get_exchange_rates([currency])[currency]
            cache.set(key, exchange_rates, 60 * 60 * 12)

    return exchange_rates
