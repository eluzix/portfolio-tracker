import collections
from datetime import datetime

from tracker import store
from tracker.providers.exchange_rates import get_exchange_rates
from tracker.utils import console, today


def analyze_account(transactions: list,
                    dividend_tax_rate=None,
                    load_dividends=True,
                    as_currency=None):
    now = today()

    # Process transactions
    shares = {}
    avg_pps = {}
    total_invested = 0
    total_withdrawn = 0
    cash_flows = []

    symbols = set(t["symbol"] for t in transactions)
    if load_dividends:
        dividends = store.load_dividends()
    else:
        dividends = []

    prices = store.load_prices()
    prices = {symbol: data['adj_close'] for symbol, data in prices.items()}
    for symbol in dividends:
        if symbol in prices:
            transactions.extend(dividends[symbol])
    transactions.sort(key=lambda t: t["date"])

    all_dividends = {symbol: 0 for symbol in symbols}
    for transaction in transactions:
        date = datetime.strptime(transaction["date"], "%Y-%m-%d")
        symbol = transaction["symbol"]
        quantity = float(transaction["quantity"])
        pps = float(transaction["pps"])

        if transaction["type"] == "buy":
            if symbol not in shares:
                shares[symbol] = 0
                avg_pps[symbol] = 0

            total_invested += quantity * pps
            avg_pps[symbol] = (avg_pps[symbol] * shares[symbol] + quantity * pps) / (shares[symbol] + quantity)
            shares[symbol] += quantity
            cash_flows.append((date, quantity * pps))

        elif transaction["type"] == "sell":
            total_withdrawn += quantity * pps
            shares[symbol] -= quantity
            cash_flows.append((date, 0 - (quantity * pps)))

        elif transaction["type"] == "dividend" and symbol in shares:
            # multiple pps with current number of shares
            if dividend_tax_rate is not None:
                pps *= (1 - dividend_tax_rate)

            transaction_value = pps * shares[symbol]
            all_dividends[symbol] += transaction_value
            cash_flows.append((date, 0 - transaction_value))

    # Calculate other metrics
    current_portfolio_value = sum(prices[symbol] * shares[symbol] for symbol in shares)
    portfolio_gain = current_portfolio_value - total_invested + total_withdrawn

    # Calculate Modified Dietz Yield
    today_date = datetime.strptime(now, "%Y-%m-%d")
    days_since_start = (today_date - cash_flows[0][0]).days
    weighted_cash_flows = sum((cf[1] * (today_date - cf[0]).days / days_since_start) for cf in cash_flows)
    modified_dietz_yield = (portfolio_gain) / (total_invested + weighted_cash_flows)

    # Calculate Annualized Yield
    years_since_start = days_since_start / 365
    annualized_yield = ((1 + modified_dietz_yield) ** (1 / years_since_start)) - 1

    total_dividends = sum(all_dividends.values())

    exchange_rate = get_exchange_rates([as_currency])[as_currency] if as_currency is not None else 1
    total_dividends *= exchange_rate
    portfolio_gain *= exchange_rate
    current_portfolio_value *= exchange_rate
    total_invested *= exchange_rate
    total_withdrawn *= exchange_rate

    account_info = {
        "exchange_rate": exchange_rate,
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
    }

    return account_info


def analyze_portfolio(transactions: list,
                      per_account=False,
                      dividend_tax_rate=None,
                      load_dividends=True,
                      as_currency=None):

    with console.status("[bold green]Crunching numbers...") as status:

        # Group transactions by account
        transactions_by_account = collections.defaultdict(list)
        for transaction in transactions:
            if per_account:
                account = transaction.get('account', 'default')
            else:
                account = 'default'

            transactions_by_account[account].append(transaction)

        # Perform analysis for each account
        portfolio_summary = {}
        for account, account_transactions in transactions_by_account.items():
            portfolio_summary[account] = analyze_account(account_transactions,
                                                         dividend_tax_rate,
                                                         load_dividends,
                                                         as_currency)

    return portfolio_summary


def analyze_portfolio_(transactions: list,
                       dividend_tax_rate=None,
                       load_dividends=True,
                       as_currency=None,
                       currency_symbol='$'):
    with console.status("[bold green]Crunching numbers...") as status:
        today = datetime.now().strftime('%Y-%m-%d')

        # Process transactions
        shares = {}
        avg_pps = {}
        total_invested = 0
        total_withdrawn = 0
        cash_flows = []

        symbols = set(t["symbol"] for t in transactions)
        if load_dividends:
            dividends = store.load_dividends()
        else:
            dividends = []

        prices = store.load_prices()
        prices = {symbol: data['adj_close'] for symbol, data in prices.items()}
        for symbol in dividends:
            if symbol in prices:
                transactions.extend(dividends[symbol])
        transactions.sort(key=lambda t: t["date"])

        all_dividends = {symbol: 0 for symbol in symbols}
        for transaction in transactions:
            date = datetime.strptime(transaction["date"], "%Y-%m-%d")
            symbol = transaction["symbol"]
            quantity = float(transaction["quantity"])
            pps = float(transaction["pps"])

            if transaction["type"] == "buy":
                if symbol not in shares:
                    shares[symbol] = 0
                    avg_pps[symbol] = 0

                total_invested += quantity * pps
                avg_pps[symbol] = (avg_pps[symbol] * shares[symbol] + quantity * pps) / (shares[symbol] + quantity)
                shares[symbol] += quantity
                cash_flows.append((date, quantity * pps))

            elif transaction["type"] == "sell":
                total_withdrawn += quantity * pps
                shares[symbol] -= quantity
                cash_flows.append((date, 0 - (quantity * pps)))

            elif transaction["type"] == "dividend" and symbol in shares:
                # multiple pps with current number of shares
                if dividend_tax_rate is not None:
                    pps *= (1 - dividend_tax_rate)

                transaction_value = pps * shares[symbol]
                all_dividends[symbol] += transaction_value
                cash_flows.append((date, 0 - transaction_value))

        # Calculate other metrics
        current_portfolio_value = sum(prices[symbol] * shares[symbol] for symbol in shares)
        portfolio_gain = current_portfolio_value - total_invested + total_withdrawn

        # Calculate Modified Dietz Yield
        today_date = datetime.strptime(today, "%Y-%m-%d")
        days_since_start = (today_date - cash_flows[0][0]).days
        weighted_cash_flows = sum((cf[1] * (today_date - cf[0]).days / days_since_start) for cf in cash_flows)
        modified_dietz_yield = (portfolio_gain) / (total_invested + weighted_cash_flows)

        # Calculate Annualized Yield
        years_since_start = days_since_start / 365
        annualized_yield = ((1 + modified_dietz_yield) ** (1 / years_since_start)) - 1

        total_dividends = sum(all_dividends.values())

        exchange_rate = get_exchange_rates([as_currency])[as_currency] if as_currency is not None else 1
        total_dividends *= exchange_rate
        portfolio_gain *= exchange_rate
        current_portfolio_value *= exchange_rate
        total_invested *= exchange_rate
        total_withdrawn *= exchange_rate

    ret = {
        "currency_symbol": currency_symbol,
        "exchange_rate": exchange_rate,
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
    }

    return ret


def list_accounts(transactions):
    accounts = set()
    for transaction in transactions:
        accounts.add(transaction["account"])
    return accounts
