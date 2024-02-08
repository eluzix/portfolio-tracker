import typing
from dataclasses import asdict

from tracker.cache_utils import get_cache
from tracker.dynamodb import ddb
from tracker.google_sheets_utils import collect_all_transactions, get_accounts_meta_data
from tracker.models import Account, Transaction
from tracker.providers import extract_symbols_prices, load_dividend_information, load_currency_information
from tracker.providers.exchange_rates import get_exchange_rates
from tracker.utils import console


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


def get_all_symbols(transactions: list = None):
    if transactions is None:
        transactions = load_transactions_from_sheets()

    return set(t["symbol"] for t in transactions)


def load_dividends(symbols: typing.Union[list, set] = None) -> dict[str, list]:
    cache = get_cache()
    dividends = cache.get('dividends')
    if not dividends:
        with console.status("[bold green]Collecting dividends...") as status:
            symbols = get_all_symbols()
            dividends = load_dividend_information(symbols)
            cache.set('dividends', dividends, 60 * 60 * 24 * 7)
            status.update(f"[bold green]Collected {len(dividends)} dividends[/]")

    # if symbols is not None:
    #     dividends = [d for d in dividends if d["symbol"] in symbols]

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


# ---- new methods for ddb ----
def save_account_metadata(user_id: str, accounts: list[Account]):
    save_accounts = []
    for account in accounts:
        item = asdict(account)
        item.update({
            'PK': f'user#{user_id}',
            'SK': f'account#{account.id}',
        })
        save_accounts.append(item)

    ddb.batch_write_items('tracker-data', save_accounts)


def load_accounts_metadata(user_id: str):
    items = ddb.query('tracker-data', **{
        'KeyConditionExpression': 'PK = :pk and begins_with(SK, :sk)',
        'ExpressionAttributeValues': {
            ':pk': ddb.serialize_value(f'user#{user_id}'),
            ':sk': ddb.serialize_value('account#')
        }
    })

    return [Account.from_dict(item) for item in items]


def save_transactions(user_id: str, transactions: list[Transaction]):
    save_items = []
    for transaction in transactions:
        item = asdict(transaction)
        item.update({
            'PK': f'user#{user_id}',
            'SK': f'transaction#{transaction.account_id}#{transaction.id}',
        })
        save_items.append(item)
    ddb.batch_write_items('tracker-data', save_items)


def load_transactions(user_id: str):
    items = ddb.query('tracker-data', **{
        'KeyConditionExpression': 'PK = :pk and begins_with(SK, :sk)',
        'ExpressionAttributeValues': {
            ':pk': ddb.serialize_value(f'user#{user_id}'),
            ':sk': ddb.serialize_value('transaction#')
        }
    })

    return [Transaction.from_dict(item) for item in items]


def load_user_data(user_id: str) -> tuple[list[Account], dict[str, list[Transaction]]]:
    accounts: list[Account] = []
    transactions: dict[str, list[Transaction]] = {}
    items = ddb.query('tracker-data', **{
        'KeyConditionExpression': 'PK = :pk',
        'ExpressionAttributeValues': {
            ':pk': ddb.serialize_value(f'user#{user_id}'),
        }
    })

    for item in items:
        if item['SK'].startswith('account#'):
            accounts.append(Account.from_dict(item))
        elif item['SK'].startswith('transaction#'):
            account_id = item['account_id']
            if account_id not in transactions:
                transactions[account_id] = []
            transactions[account_id].append(Transaction.from_dict(item))

    # sort transactions by date
    for account_id in transactions:
        transactions[account_id] = sorted(transactions[account_id], key=lambda t: t.date)

    return accounts, transactions
