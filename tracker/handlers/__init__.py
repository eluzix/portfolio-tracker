import msgspec


def sandbox(event, context):
    return msgspec.json.encode({
        'statusCode': 200,
        'body': 'Hello, World!'
    })
