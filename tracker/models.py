from dataclasses import dataclass, field
from typing import Optional


@dataclass
class Account:
    id: str
    name: str
    owner: str
    institution: str
    institution_id: Optional[str]
    description: Optional[str]
    tags: Optional[list[str]] = field(default_factory=list)
    created_at: Optional[int] = None
    updated_at: Optional[int] = None

    @classmethod
    def from_dict(cls, data: dict) -> 'Account':
        a = {
            'id': data['id'],
            'name': data['name'],
            'owner': data['owner'],
            'institution': data['institution'],
            'institution_id': data.get('institution_id'),
            'description': data.get('description'),
            'tags': data.get('tags', []),
            'created_at': data.get('created_at'),
            'updated_at': data.get('updated_at'),
        }

        return Account(**a)


@dataclass
class Transaction:
    id: str
    account_id: str
    symbol: str
    date: str
    type: str
    quantity: int
    pps: float

    @classmethod
    def from_dict(cls, data: dict) -> 'Transaction':
        t = {
            'id': data.get('id'),
            'account_id': data.get('account_id'),
            'symbol': data['symbol'],
            'date': data['date'],
            'type': data['type'],
            'quantity': data['quantity'],
            'pps': data['pps'],
        }

        return Transaction(**t)
