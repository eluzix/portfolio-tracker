import datetime
import json

import openai

from tracker.config import get_secret


def get_completion(prompt, model="gpt-3.5-turbo"):
    messages = [{"role": "user", "content": prompt}]
    response = openai.ChatCompletion.create(
        model=model,
        messages=messages,
        temperature=0,  # this is the degree of randomness of the model's output
    )
    return response.choices[0].message["content"]


def transactions_to_csv(transactions):
    csv = 'date,type,symbol,quantity,pps\n'
    for transaction in transactions:
        csv += f"{transaction['date']},{transaction['type']},{transaction['symbol']},{transaction['quantity']},{transaction['pps']}\n"
    return csv


if __name__ == '__main__':
    # transactions = collect_all_transactions()
    # with open('all_transactions.json', 'w') as f:
    #     json.dump(transactions, f)

    with open('../all_transactions.json', 'r') as f:
        transactions = json.load(f)

    # for transaction in transactions:
    #     print(transaction)

    # prices = extract_symbols_prices_from_transactions(transactions)
    # with open('prices.json', 'w') as f:
    #     json.dump(prices, f)

    with open('../prices.json', 'r') as f:
        prices = json.load(f)

    # print(prices)

    # transactions = transactions_to_csv(transactions)

    today = datetime.datetime.now().strftime('%Y-%m-%d')
    _PROMPT = f"""\
you are an expert investment analyst and your job is to help me analyze my portfolio.
for the list of transactions surrounded by <tr> and </tr> i want you to provide me the following information:
1. total shares per symbol
2. avg. pps per symbol
3. my current portfolio value
4. Total money invested
5. Total money withdraw
6. anualized yield
7. modified dietz method yield
8. How much dividend payouts I had.
9. What’s the portfolio gain
10. any additional information you think is interesting about my portfolio.
11. Finally summarize everything in a short textual paragraph.

Additional notes:
1. All date are formatted as YYYY-MM-DD
2. These transaction are all the transactions made in this portfolio.
4. before the first transaction the portfolio had a value of 0.
5. The portfolio start date begins with the first transaction.
6. I don’t need advise about Diversification
7. Today is {today}

Additional information:

Current prices for the shares are:
{prices}

<tr>
{transactions}
</tr>
"""

    print(_PROMPT)

    openai.api_key = get_secret('openai_secret')

    # response = get_completion(_PROMPT)
    # print(response)
