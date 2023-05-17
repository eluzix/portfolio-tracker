import json
import typing

import requests
from alpha_vantage.timeseries import TimeSeries

from tracker.config import get_secret


# ---- Alpha Vantage ----


def get_symbols_data(symbols):
    # Fetch current stock prices and store them in a dictionary
    symbol_prices = {}
    ts = TimeSeries(key=get_secret('alpha_vantage_api_key'), output_format='json')

    for symbol in symbols:
        try:
            data, _ = ts.get_quote_endpoint(symbol=symbol)
            current_price = float(data.get('05. price', None))
            if current_price:
                symbol_prices[symbol] = current_price
        except Exception as e:
            print(f'Error fetching price for {symbol}: {e}')

    # Print the symbol prices
    ret = {symbol: price for symbol, price in symbol_prices.items()}

    return ret


def get_dividends_as_transactions(stocks: typing.Union[list, set], start_date=None, end_date=None):
    api_key = get_secret("alpha_vantage_api_key")
    dividends_dict = {}
    for stock in stocks:
        url = f'https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&outputsize=full&symbol={stock}&apikey={api_key}'
        response = requests.get(url)
        data = json.loads(response.text)
        dividends = []
        for date, values in data["Time Series (Daily)"].items():
            if start_date is not None and date < start_date:
                continue

            if end_date is not None and date > end_date:
                continue

            dividend = values.get("7. dividend amount")
            if dividend and float(dividend) > 0:
                dividends.append({
                    'date': date,
                    'type': 'dividend',
                    'symbol': stock,
                    'quantity': 0,
                    'pps': float(dividend),
                    'account': ''
                })

        dividends_dict[stock] = dividends
    return dividends_dict


if __name__ == '__main__':
    stocks = ["VT"]
    start_date = "2014-01-01"
    end_date = "2023-12-31"
    dividends_dict = get_dividends_as_transactions(stocks, start_date, end_date)
    print(dividends_dict)
