import requests

from tracker.config import get_secret


def get_exchange_rates(target_currencies: list, base_currency="USD"):
    url = f"https://api.apilayer.com/exchangerates_data/latest?base={base_currency}&symbols={','.join(target_currencies)}"
    response = requests.get(url, headers={
        'apikey': get_secret('exchangerates_key')
    })

    data = response.json()
    rates = {}
    for currency in target_currencies:
        rate = data["rates"].get(currency)
        if rate:
            rates[currency] = rate
    return rates


if __name__ == '__main__':
    rates = get_exchange_rates(["ILS"])
    print(rates)
