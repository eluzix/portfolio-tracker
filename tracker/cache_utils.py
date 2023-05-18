import os

from diskcache import Cache

_cache = None


def get_cache():
    global _cache
    if _cache is None:
        dir_path = os.path.dirname(os.path.realpath(__file__))
        _cache = Cache(f'{dir_path}/../cache')

    return _cache
