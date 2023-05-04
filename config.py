import json

_SECRETS = None


def get_secret(key):
    global _SECRETS
    if _SECRETS is None:
        _SECRETS = json.load(open('secrets.json', 'r'))

    return _SECRETS[key]
