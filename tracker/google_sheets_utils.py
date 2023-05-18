import json
import operator
import os
import re

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient import discovery

from tracker.cache_utils import get_cache
from tracker.config import get_secret


def sanitize_currency(val):
    """Remove any non-numeric characters from the pps field."""
    return float(re.sub(r'[^\d.]', '', val))


# Set up the API client
SCOPES = ['https://www.googleapis.com/auth/spreadsheets.readonly']


def google_authenticate():
    cache = get_cache()

    creds = cache.get('google_token')
    if creds is not None:
        creds = Credentials.from_authorized_user_info(json.loads(creds), SCOPES)
        # creds = Credentials.from_authorized_user_file(token_path, SCOPES)

    # If there are no (valid) credentials available, let the user log in.
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            try:
                creds.refresh(Request())
            except Exception as e:
                print(f'Error refreshing creds: {e}')
                creds = None
        else:
            creds = None

        if creds is None:
            dir_path = os.path.dirname(os.path.realpath(__file__))
            cred_path = os.path.join(dir_path, '..', 'credentials.json')
            flow = InstalledAppFlow.from_client_secrets_file(cred_path, SCOPES)
            creds = flow.run_local_server(port=0)

        # Save the credentials for the next run
        cache.set('google_token', creds.to_json())

    return creds


def list_accounts(spreadsheet_id: str = None, creds: Credentials = None):
    if spreadsheet_id is None:
        spreadsheet_id = get_secret('spreadsheet_id')

    if creds is None:
        creds = google_authenticate()

    sheets_api = discovery.build('sheets', 'v4', credentials=creds)

    # Get the sheet names in the selected spreadsheet
    sheet_metadata = sheets_api.spreadsheets().get(spreadsheetId=spreadsheet_id).execute()
    sheets = sheet_metadata.get('sheets', '')

    sheet_names = []
    for sheet in sheets:
        title = sheet.get("properties", {}).get("title", "")
        if title.lower() != "summary":
            sheet_names.append(title)

    return sheet_names


def collect_all_transactions(spreadsheet_id: str = None, included_sheets: list = None, creds: Credentials = None):
    if spreadsheet_id is None:
        spreadsheet_id = get_secret('spreadsheet_id')

    if included_sheets is None:
        included_sheets = []

    # Convert included sheet names to lowercase for case-insensitive comparison
    sheet_names = [sheet.lower() for sheet in included_sheets]

    # Authenticate with Google Sheets
    if creds is None:
        creds = google_authenticate()

    sheets_api = discovery.build('sheets', 'v4', credentials=creds)
    if len(sheet_names) == 0:
        sheet_names = list_accounts(spreadsheet_id, creds)

    # Iterate over all sheets and gather transactions
    all_transactions = []
    for sheet_name in sheet_names:
        result = sheets_api.spreadsheets().values().get(
            spreadsheetId=spreadsheet_id, range=sheet_name).execute()
        values = result.get('values', [])

        for row in values:
            # Ensure the row has the correct number of columns
            if len(row) > 0 and row[0].lower() == 'date':
                continue

            if len(row) == 5:
                date, t_type, symbol, quantity, pps = row
                all_transactions.append({
                    'date': date,
                    'type': t_type,
                    'symbol': symbol,
                    'quantity': int(quantity),
                    'pps': float(sanitize_currency(pps)),
                    'account': sheet_name.lower()
                })

    # Sort transactions by date
    sorted_transactions = sorted(all_transactions, key=operator.itemgetter('date'))

    return sorted_transactions
    # Print the sorted transactions
