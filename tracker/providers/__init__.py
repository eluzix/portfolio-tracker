import typing

import requests

from tracker.config import get_secret
from tracker.providers import alpha_vantage_utils
from tracker.utils import console


def extract_symbols_prices(symbols: typing.Union[list, set]):
    # Extract unique symbols from transactions
    unique_symbols = sorted(set(symbols))
    key = get_secret('marketstack_key')
    url = f'https://api.marketstack.com/v1/eod/latest?access_key={key}&symbols={",".join(unique_symbols)}'
    response = requests.get(url).json()
    if 'error' in response:
        console.print(f'[bold red]Error fetching prices from marketstack: {response["error"]["message"]}[/]')
        return None

    data = response['data']
    ret = {item['symbol']: item for item in data if 'symbol' in item}

    base_symbols = set(unique_symbols)
    found_symbols = set(ret.keys())
    missing_symbols = base_symbols - found_symbols
    if len(missing_symbols) > 0:
        av_data = alpha_vantage_utils.get_symbols_data(missing_symbols)
        for symbol in av_data:
            ret[symbol] = {'symbol': symbol, 'adj_close': av_data[symbol]}

    return ret


def load_dividend_information(symbols: list | set, date_from: str = '2014-01-01', limit: int = 1000) -> dict[str, list]:
    unique_symbols = sorted(set(symbols))
    key = get_secret('marketstack_key')
    url = f'https://api.marketstack.com/v1/dividends?access_key={key}&symbols={",".join(unique_symbols)}&limit={limit}'

    response = requests.get(url).json()
    if 'error' in response:
        console.print(f'[bold red]Error fetching prices from marketstack: {response["error"]["message"]}[/]')
        return None

    data = response['data']
    dividends = {s: [] for s in unique_symbols}
    for item in data:
        dividends[item['symbol']].append({
            'date': item['date'],
            'type': 'dividend',
            'symbol': item['symbol'],
            'quantity': 0,
            'pps': float(item['dividend']),
            'account': ''
        })

    return dividends


def load_currency_information():
    key = get_secret('marketstack_key')
    url = f'https://api.marketstack.com/v1/currencies?access_key={key}'
    response = requests.get(url).json()
    if 'error' in response:
        console.print(f'[bold red]Error fetching currencies from marketstack: {response["error"]["message"]}[/]')
        return None

    data = response['data']
    ret = {
        item['code']: item
        for item in data
    }

    return ret


if __name__ == '__main__':
    # extract_symbols_prices(['VT'])
    load_dividend_information(['VT', 'HDV'])
