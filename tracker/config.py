import json
import os

_SECRETS = None


def get_secret(key):
    global _SECRETS
    if _SECRETS is None:
        dir_path = os.path.dirname(os.path.realpath(__file__))
        if os.path.exists(f'{dir_path}/../.secrets.json'):
            _SECRETS = json.load(open(f'{dir_path}/../.secrets.json', 'r'))
        else:
            _SECRETS = json.load(open(f'{dir_path}/../secrets.json', 'r'))

    return _SECRETS[key]
