from dataclasses import asdict

import msgspec

from tracker import store
from tracker.portfolio_analysis import analyze_portfolio

USER_ID = '1'


def make_response(body: dict, status_code: int = 200):
    return {
        'statusCode': status_code,
        'body': msgspec.json.encode(body),
    }


def sandbox(event, context):
    return make_response({})


def portfolio_analysis(event, context):
    accounts, transactions = store.load_user_data(USER_ID)
    analysis = analyze_portfolio(transactions)

    accounts = [asdict(a) for a in accounts]
    # out_tr = {}
    # for k, v in transactions.items():
    #     out_tr[k] = [asdict(t) for t in v]

    body = {
        'accounts': accounts,
        # 'transactions': out_tr,
        'analysis': analysis
    }

    return make_response(body)


def all_transactions(event, context):
    transactions = store.load_transactions(USER_ID)

    body = {
        'transactions': [asdict(t) for t in transactions]
    }

    return make_response(body)
