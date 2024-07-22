import sys
sys.path.append('.')
import hashlib

import boto3

from scripts.script_utils import load_transactions_from_sheets
from tracker import store, utils
from tracker.dynamodb import ddb
from tracker.models import Account, Transaction
from tracker.portfolio_analysis import analyze_account

USER_ID = '1'


def dump_accounts_metadata_to_ddb():
    accounts: list[Account] = []
    metadata = store.load_accounts_metadata_from_sheet()
    for account_md in metadata.values():
        if account_md['account'].startswith('_'):
            continue

        accounts.append(Account(**{
            'id': account_md['id'],
            'name': account_md['account'],
            'owner': account_md['owner'],
            'institution': account_md['institution'],
            'institution_id': account_md.get('institution_id'),
            'description': account_md.get('description'),
            'tags': account_md.get('tags', []),
        }))

    store.save_account_metadata(USER_ID, accounts)


def dump_transactions_to_ddb():
    accounts = store.load_accounts_metadata(USER_ID)
    account_map = {account.name.lower(): account.id for account in accounts}
    transactions = load_transactions_from_sheets()
    saved_transactions = []
    tids = set()
    for transaction in transactions:
        account = account_map.get(transaction['account'].lower())
        if account is None:
            print(f'Account not found for {transaction["account"]}')
            continue

        transaction['symbol'] = transaction['symbol'].replace('.', '-')
        # transaction['id'] = utils.generate_id()
        transaction['account_id'] = account
        transaction_id = hashlib.sha224(
            f"{transaction['account_id']}#{transaction['symbol']}#{transaction['date']}#{transaction['type']}#{transaction['quantity']}#{transaction['pps']}".encode()).hexdigest()

        while transaction_id in tids:
            transaction_id = f'{transaction_id}0'

        tids.add(transaction_id)
        transaction['id'] = transaction_id
        t = Transaction.from_dict(transaction)
        saved_transactions.append(t)

    store.save_transactions(USER_ID, saved_transactions)
    print(f'Saved {len(saved_transactions)} transactions')


def clan_all_transactions():
    all_keys = []
    items = ddb.query('tracker-data', **{
        'ProjectionExpression': 'PK, SK',
        'KeyConditionExpression': 'PK = :pk and begins_with(SK, :sk)',
        'ExpressionAttributeValues': {
            ':pk': ddb.serialize_value(f'user#{USER_ID}'),
            ':sk': ddb.serialize_value('transaction#')
        }
    })
    for item in items:
        all_keys.append({
            'PK': item['PK'],
            'SK': item['SK']
        })

    ddb.batch_delete_items('tracker-data', all_keys)
    print(f'Deleted {len(all_keys)} transactions')


def test_load():
    accounts, transactions = store.load_user_data(USER_ID)

    # for account in accounts:
    #     print(account)
    #
    all_transactions = []
    for account_id in transactions:
        all_transactions.extend(transactions[account_id])

    print(f'Loaded {len(all_transactions)} transactions for {len(accounts)} accounts')

    res = analyze_account(transactions['7'])
    print(res)


if __name__ == '__main__':
    boto3.setup_default_session(profile_name='tracker')
    # test_load()
    # dump_accounts_metadata_to_ddb()
    dump_transactions_to_ddb()
    # clan_all_transactions()
