import typing

import requests

from tracker.config import get_secret
from tracker.providers import alpha_vantage_utils


def extract_symbols_prices(symbols: typing.Union[list, set]):
    # Extract unique symbols from transactions
    unique_symbols = sorted(set(symbols))
    key = get_secret('marketstack_key')
    url = f'http://api.marketstack.com/v1/eod/latest?access_key={key}&symbols={",".join(unique_symbols)}'
    response = requests.get(url).json()
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


if __name__ == '__main__':
    extract_symbols_prices(['VT'])
