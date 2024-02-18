import typing

from rich.console import Console

from tracker import store
from tracker.cache_utils import get_cache
from scripts.google_sheets_utils import collect_all_transactions, get_accounts_meta_data
from tracker.models import Transaction
from tracker.providers import load_dividend_information, extract_symbols_prices, load_currency_information
from tracker.providers.exchange_rates import get_exchange_rates

console = Console(record=True)


class TerminalColors:
    HEADER = '\033[95m'
    OK_BLUE = '\033[94m'
    OK_CYAN = '\033[96m'
    OK_GREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    END = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

    @classmethod
    def color(cls, txt, start=None, end=None):
        if start is None:
            start = cls.OK_GREEN
        if end is None:
            end = cls.END
        return f'{start}{txt}{end}'


def load_transactions_from_sheets(filter_by_accounts: list = None):
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


def load_dividends(symbols: typing.Union[list, set]) -> dict[str, list[Transaction]]:
    with console.status("[bold green]Collecting dividends...") as status:
        dividends = store.load_dividends(symbols)
        status.update(f"[bold green]Collected {len(dividends)} dividends[/]")

    return dividends


def load_prices(symbols: list[str]):
    with console.status("[bold green]Collecting prices...") as status:
        prices = store.load_prices(symbols)

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



def load_currencies_metadata():
    cache = get_cache()
    currencies_metadata = cache.get('currencies_metadata')
    if not currencies_metadata:
        with console.status("[bold green]Collecting currencies metadata...") as status:
            currencies_metadata = load_currency_information()
            cache.set('currencies_metadata', currencies_metadata, 60 * 60 * 24 * 30)

    return currencies_metadata


def load_accounts_metadata_from_sheet():
    # cache = get_cache()
    # accounts_metadata = cache.get('accounts_metadata')
    # if not accounts_metadata:
    with console.status("[bold green]Collecting accounts metadata...") as status:
        accounts_metadata = get_accounts_meta_data()
        # cache.set('accounts_metadata', accounts_metadata, 60 * 60 * 24 * 7)

    return accounts_metadata
