import time

import msgspec.json

from tracker.dynamodb import ddb, hash_sort


class DdbCache:
    def __init__(self):
        self._local_cache = {}

    def get(self, key: str):
        if key in self._local_cache:
            return self._local_cache[key]

        item = ddb.get_item('tracker-data', hash_sort(f'CACHE', key))
        if item is not None:
            now = int(time.time())
            if item['ttl'] > now:
                val = msgspec.json.decode(item['value'].value)
                self._local_cache[key] = val
                return val

        return None

    def set(self, key: str, value, ttl: int = None):
        now = int(time.time())
        item = {
            'PK': 'CACHE',
            'SK': key,
            'value': msgspec.json.encode(value)
        }
        if ttl is not None:
            item['ttl'] = now + ttl

        ddb.put_item('tracker-data', item)

        if key in self._local_cache:
            self._local_cache[key] = value

    def delete(self, key: str):
        ddb.delete_item('tracker-data', hash_sort('CACHE', key))
        if key in self._local_cache:
            del self._local_cache[key]

    def clear(self):
        self._local_cache = {}

        all_keys = []
        items = ddb.query('tracker-data', **{
            'ProjectionExpression': 'PK, SK',
            'KeyConditionExpression': 'PK = :pk',
            'ExpressionAttributeValues': {
                ':pk': ddb.serialize_value('CACHE'),
            }
        })
        for item in items:
            all_keys.append({
                'PK': item['PK'],
                'SK': item['SK']
            })

        ddb.batch_delete_items('tracker-data', all_keys)


_cache = DdbCache()


def get_cache():
    # global _cache
    # if _cache is None:
    #     dir_path = os.path.dirname(os.path.realpath(__file__))
    #     _cache = Cache(f'{dir_path}/../cache')

    return _cache
