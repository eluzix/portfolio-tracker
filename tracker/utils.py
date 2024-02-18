import uuid
from datetime import datetime


def today() -> str:
    return datetime.today().strftime('%Y-%m-%d')


def decimal_to_number(val):
    if int(val) == val:
        return int(val)
    else:
        return float(val)


def generate_id():
    return str(uuid.uuid4())
