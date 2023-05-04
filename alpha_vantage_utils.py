from alpha_vantage.timeseries import TimeSeries

from config import get_secret


# ---- Alpha Vantage ----


def extract_symbols_prices_from_transactions(transactions):
    # Extract unique symbols from transactions
    unique_symbols = sorted(set(transaction['symbol'] for transaction in transactions))

    # Fetch current stock prices and store them in a dictionary
    symbol_prices = {}
    ts = TimeSeries(key=get_secret('alpha_vantage_api_key'), output_format='json')

    for symbol in unique_symbols:
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
