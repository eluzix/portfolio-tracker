from datetime import datetime

from tracker import store
from tracker.models import Transaction
from tracker.store import get_all_symbols
from tracker.utils import today


def analyze_account(transactions: list[Transaction],
                    dividend_tax_rate=None,
                    load_dividends=True):
    now = today()

    # Process transactions
    shares = {}
    avg_pps = {}
    total_invested = 0
    total_withdrawn = 0
    cash_flows = []
    last_transactions = {}

    symbols = set()
    symbol_first_transaction = {}
    for transaction in transactions:
        symbols.add(transaction.symbol)
        symbol_date = symbol_first_transaction.get(transaction.symbol, transaction.date)
        if transaction.date <= symbol_date:
            symbol_first_transaction[transaction.symbol] = transaction.date

    all_symbols = get_all_symbols(transactions)
    if load_dividends:
        dividends = store.load_dividends(all_symbols)
    else:
        dividends = {}

    prices = store.load_prices(all_symbols)
    prices = {symbol: data['adj_close'] for symbol, data in prices.items()}
    for symbol in dividends:
        if symbol in symbols:
            symbol_dividends = [d for d in dividends[symbol] if d.date >= symbol_first_transaction[symbol]]
            transactions.extend(symbol_dividends)
    transactions.sort(key=lambda t: t.date)

    all_dividends = {symbol: 0 for symbol in symbols}
    for transaction in transactions:
        date = datetime.strptime(transaction.date, "%Y-%m-%d")
        symbol = transaction.symbol
        quantity = float(transaction.quantity)
        pps = float(transaction.pps)

        if transaction.type == "buy":
            if symbol not in shares:
                shares[symbol] = 0
                avg_pps[symbol] = 0

            total_invested += quantity * pps
            avg_pps[symbol] = (avg_pps[symbol] * shares[symbol] + quantity * pps) / (shares[symbol] + quantity)
            shares[symbol] += quantity
            cash_flows.append((date, quantity * pps))
            last_transactions[symbol] = {'date': date, 'type': 'buy'}

        elif transaction.type == "sell":
            total_withdrawn += quantity * pps
            shares[symbol] -= quantity
            cash_flows.append((date, 0 - (quantity * pps)))
            last_transactions[symbol] = {'date': date, 'type': 'sell'}

        elif transaction.type == "dividend" and symbol in shares:
            # multiple pps with current number of shares
            if dividend_tax_rate is not None:
                pps *= (1 - dividend_tax_rate)

            transaction_value = pps * shares[symbol]
            all_dividends[symbol] += transaction_value
            cash_flows.append((date, 0 - transaction_value))

    total_dividends = sum(all_dividends.values())

    # Calculate other metrics
    current_portfolio_value = sum(prices.get(symbol, 0) * shares[symbol] for symbol in shares)
    portfolio_gain = (current_portfolio_value + total_withdrawn + total_dividends) - total_invested

    # Calculate Modified Dietz Yield
    today_date = datetime.strptime(now, "%Y-%m-%d")
    days_since_start = (today_date - cash_flows[0][0]).days
    weighted_cash_flows = sum((cf[1] * (today_date - cf[0]).days / days_since_start) for cf in cash_flows)
    modified_dietz_yield = portfolio_gain / (total_invested + weighted_cash_flows)

    # Calculate Annualized Yield
    years_since_start = days_since_start / 365
    annualized_yield = ((1 + modified_dietz_yield) ** (1 / years_since_start)) - 1
    if isinstance(annualized_yield, complex):
        annualized_yield = annualized_yield.real

    # Simple yield
    if current_portfolio_value == 0:
        simple_yield = 0
    else:
        simple_yield = portfolio_gain / current_portfolio_value

    account_info = {
        "exchange_rate": 1,
        "shares": shares,
        "avg_pps": avg_pps,
        "total_invested": total_invested,
        "total_withdrawn": total_withdrawn,
        "all_dividends": all_dividends,
        "total_dividends": total_dividends,
        "portfolio_gain": portfolio_gain,
        "current_portfolio_value": current_portfolio_value,
        "annualized_yield": annualized_yield,
        "modified_dietz_yield": modified_dietz_yield,
        "simple_yield": simple_yield,
        'last_transactions': last_transactions
    }

    return account_info


def analyze_portfolio(transactions: dict[str, list[Transaction]],
                      dividend_tax_rate=None,
                      load_dividends=True):
    all_transactions = []
    # Group transactions by account
    # transactions_by_account = collections.defaultdict(list)
    # for transaction in transactions:
    #     account = transaction.get('account', 'default')
    #     transactions_by_account[account].append(transaction)
    #     transactions_by_account['total'].append(transaction)

    # Perform analysis for each account
    portfolio_summary = {}
    for account_id, account_transactions in transactions.items():
        all_transactions.extend(account_transactions)
        portfolio_summary[account_id] = analyze_account(account_transactions,
                                                        dividend_tax_rate,
                                                        load_dividends)

    portfolio_summary['total'] = analyze_account(all_transactions,
                                                 dividend_tax_rate,
                                                 load_dividends)

    return portfolio_summary


def list_accounts(transactions):
    accounts = set()
    for transaction in transactions:
        accounts.add(transaction["account"])
    return accounts
