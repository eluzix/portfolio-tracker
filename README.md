# Using this repo

## Setup
```zsh
pip install -r requirements.txt
```
Create a file called `secrets.json` it should include the following keys:
1. `alpha_vantage_api_key` - your alpha vantage api key
2. `openai_secret` - your openapi secret key

Please your Google project `credentials.json` in the main folder. 
Make sure it has Google Sheets read and write permissions.

## Usage
Create a spreadsheet in Google Sheets and for each account you want to track create a sheet with the following columns:
1. `date` - date of the transaction (YYYY-MM-DD)
2. `type` - transaction type (buy/sell)
3. `symbol` - ticker symbol of the transaction
4. `quantity` - number of shares
5. `pps` - price per share
6. `account` - name of the account

In addition, create a sheet called `Summary` with the following columns:
1. `Account` - name of the account
2. `Description` - description of the account
3. `Owner` - name of the owner of the account
4. `Institution` - name of the institution
5. `Institution ID` - ID of the account in the institution
6. `Liquid` - yes/no if the account is liquid

Then run the following command:
```zsh
python main.py
```
You can also pass `dividend_tax_rate` to `analyze_portfolio` to calculate the impact of taxes on your dividends.


# #Radicle

To clone this repository on [Radicle](https://radicle.xyz), simply run:
```zsh
rad clone rad:z4TxWeegmqBVrJVwNAUHHRuFSGiTY
```
