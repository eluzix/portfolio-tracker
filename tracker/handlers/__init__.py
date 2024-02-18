import msgspec

from tracker import store


def sandbox(event, context):
    accounts, transactions = store.load_user_data('1')
    return msgspec.json.encode({
        'statusCode': 200,
        'body': {
            'accounts': accounts,
            'transactions': transactions
        }
    })
