import enum
from decimal import Decimal

import boto3
from boto3.dynamodb.types import TypeSerializer, TypeDeserializer

from tracker.utils import decimal_to_number


class Dynamodb(object):
    def __init__(self):
        self._connection = None
        self._type_deserializer = TypeDeserializer()
        self._type_serializer = TypeSerializer()
        self._tmp_credentials = None

    def get_connection(self):

        if self._connection is None:
            extra_args = {}

            if self._tmp_credentials is not None:
                extra_args['aws_access_key_id'] = self._tmp_credentials['AccessKeyId'],
                extra_args['aws_secret_access_key'] = self._tmp_credentials['SecretAccessKey'],
                extra_args['aws_session_token'] = self._tmp_credentials['SessionToken'],

            self._connection = boto3.client('dynamodb', **extra_args)
        return self._connection

    def set_tmp_credentials(self, credentials=None):
        """Set or reset (by passing None) temporary credentials from AssumeRole"""
        self._tmp_credentials = credentials
        self._connection = None

    def serialize_value(self, value):
        if type(value) == float:
            value = Decimal(str(value))
        elif isinstance(value, str):
            # looks redundant but help skip if cycles which we don't need and might throw exceptions
            value = value
        elif isinstance(value, enum.Enum):
            value = value.value
        elif isinstance(value, dict):
            value = self._to_floats(value)
        elif isinstance(value, list) or isinstance(value, set):
            value = self._to_floats(value)
        # elif self._is_np_val(value):
        #     value = 0.0

        return self._type_serializer.serialize(value)

    def deserialize_value(self, value):
        return self._type_deserializer.deserialize(value)

    def _to_floats(self, item):
        is_iter = isinstance(item, list) or isinstance(item, set)
        ret = [] if is_iter else {}
        for k in item:
            val = k if is_iter else item[k]

            if val is None or val == '':
                continue

            if type(val) == float:
                val = Decimal(str(val))
            elif isinstance(val, enum.Enum):
                val = val.value
            elif isinstance(val, dict):
                val = self._to_floats(val)

            if is_iter:
                ret.append(val)
            else:
                ret[k] = val

        return ret

    def type_serialize(self, item):
        """

        :type item: dict
        :param item: parmas to encode

        :rtype: dict
        :return: encoded params
        """
        ret = {}
        for key in list(item.keys()):
            val = item[key]
            if val is None or val == '':
                continue

            if type(val) == float:
                val = Decimal(str(val))
            elif isinstance(val, enum.Enum):
                val = val.value
            elif isinstance(val, dict):
                if len(val) == 0:
                    continue
                val = self._to_floats(val)
            elif isinstance(val, list) or isinstance(val, set):
                if len(val) == 0:
                    continue
                val = self._to_floats(val)

            ret[key] = self._type_serializer.serialize(val)

        return ret

    def type_deserializer(self, item):
        """

        :type item: dict
        :param item: parmas to encode

        :rtype: dict
        :return: decoded params
        """
        ret = {}
        for key in list(item.keys()):
            ret[key] = self._type_deserializer.deserialize((item[key]))

        return ret

    def put_item(self, table_name, item, **kwargs):
        return self.get_connection().put_item(
            TableName=table_name,
            Item=ddb.type_serialize(item),
            **kwargs
        )

    def update_item(self, table_name, key, return_values='NONE', **kwargs):

        return self.get_connection().update_item(TableName=table_name,
                                                 Key=self.type_serialize(key),
                                                 ReturnValues=return_values, **kwargs)

    def delete_item(self, table_name, key, **kwargs):
        return self.get_connection().delete_item(
            TableName=table_name,
            Key=self.type_serialize(key),
            **kwargs
        )

    def get_item(self, table_name, key, **kwargs):
        response = self.get_connection().get_item(TableName=table_name, Key=self.type_serialize(key), **kwargs)

        if response is None or len(response) == 0 or 'Item' not in response:
            return None

        return self.type_deserializer(response['Item'])

    def batch_get_items(self, table_name, keys, projection_expression=None, expression_attribute_names=None,
                        decimal_factory=decimal_to_number, **kwargs) -> list:
        conn = self.get_connection()
        ret = []
        while len(keys) > 0:
            batch_keys = keys[:99]
            keys = keys[99:]

            req = {
                table_name: {
                    'Keys': [self.type_serialize(key) for key in batch_keys],
                }
            }
            if projection_expression is not None:
                req[table_name]['ProjectionExpression'] = projection_expression

            if expression_attribute_names is not None:
                req[table_name]['ExpressionAttributeNames'] = expression_attribute_names

            res = conn.batch_get_item(RequestItems=req, **kwargs)
            for obj in res['Responses'][table_name]:
                obj = self.type_deserializer(obj)
                for field in obj:
                    if type(obj[field]) == Decimal:
                        obj[field] = decimal_factory(obj[field])
                ret.append(obj)

            while len(res['UnprocessedKeys']) > 0:
                res = conn.batch_get_item(res['UnprocessedKeys'])
                for obj in res['Responses'][table_name]:
                    obj = self.type_deserializer(obj)
                    for field in obj:
                        if type(obj[field]) == Decimal:
                            obj[field] = decimal_factory(obj[field])
                    ret.append(obj)

        return ret

    def batch_write_items(self, table_name, items, **kwargs):
        conn = self.get_connection()
        ret = []
        while len(items) > 0:
            batch_items = items[:25]
            items = items[25:]

            req = {
                table_name: [{
                    'PutRequest': {
                        'Item': self.type_serialize(item)
                    }
                } for item in batch_items]
            }

            res = conn.batch_write_item(RequestItems=req, **kwargs)
            ret.append(res)

        return ret

    def batch_delete_items(self, table_name, keys, **kwargs):
        conn = self.get_connection()
        ret = []
        while len(keys) > 0:
            batch_items = keys[:25]
            keys = keys[25:]

            req = {
                table_name: [{
                    'DeleteRequest': {
                        'Key': self.type_serialize(item)
                    }
                } for item in batch_items]
            }

            res = conn.batch_write_item(RequestItems=req, **kwargs)
            ret.append(res)

        return ret

    def deserializer_items(self, items, decimal_factory=decimal_to_number):
        if type(items) not in (list, set):
            items = [items]

        ret = []
        for item in items:
            obj = self.type_deserializer(item)
            for field in obj:
                val = obj[field]
                if type(val) == Decimal:
                    obj[field] = decimal_factory(val)

            ret.append(obj)

        return ret

    def query(self, table_name, follow_lek=True, **kwargs):
        resp = self.get_connection().query(TableName=table_name, **kwargs)
        ret = ddb.deserializer_items(resp['Items'])
        lek = resp.get('LastEvaluatedKey')
        while lek is not None and follow_lek:
            resp = self.get_connection().query(TableName=table_name,
                                               ExclusiveStartKey=lek,
                                               **kwargs)
            ret.extend(ddb.deserializer_items(resp['Items']))
            lek = resp.get('LastEvaluatedKey')

        return ret

    def update_with_fields(self, table_name: str, key: any, item: dict, fields: list, additional_expression=None,
                           **kwargs):
        update_expression = []
        expression_attribute_names = kwargs.get('ExpressionAttributeNames', {})
        expression_attribute_values = kwargs.get('ExpressionAttributeValues', {})

        for field in fields:
            if field in item:
                k = '#{}'.format(field)
                v = ':{}'.format(field)
                update_expression.append('{} = {}'.format(k, v))
                expression_attribute_names[k] = field
                expression_attribute_values[v] = ddb.serialize_value(item[field])

        expression = 'SET {}'.format(','.join(update_expression))
        if additional_expression is not None:
            expression = '{} {}'.format(expression, additional_expression)

        kwargs['ExpressionAttributeNames'] = expression_attribute_names
        kwargs['ExpressionAttributeValues'] = expression_attribute_values

        return self.update_item(table_name, key, UpdateExpression=expression, **kwargs)


ddb = Dynamodb()
