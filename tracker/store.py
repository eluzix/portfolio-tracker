import typing
from dataclasses import asdict

from tracker.cache_utils import get_cache
from tracker.dynamodb import ddb
from tracker.models import Account, Transaction
from tracker.providers import extract_symbols_prices, load_dividend_information


def load_dividends(symbols: typing.Union[list, set]) -> dict[str, list[Transaction]]:
    cache = get_cache()
    dividends = cache.get('dividends')
    if not dividends:
        dividends = load_dividend_information(symbols)
        cache.set('dividends', dividends, 60 * 60 * 24 * 7)
    else:
        _dividends = {}
        for symbol, transactions in dividends.items():
            _dividends[symbol] = [Transaction.from_dict(t) for t in transactions]
        dividends = _dividends

    return dividends


def load_prices(symbols: list[str]):
    cache = get_cache()
    prices = cache.get('prices')
    if not prices:
        prices = extract_symbols_prices(symbols)
        cache.set('prices', prices, 60 * 60 * 12)

    return prices


def get_all_symbols(transactions: list[Transaction] | dict[str, list[Transaction]]) -> list[str]:
    if isinstance(transactions, dict):
        _transactions = []
        for transactions_list in transactions.values():
            _transactions.extend(transactions_list)
        transactions = _transactions

    return list(set([t.symbol for t in transactions]))


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
