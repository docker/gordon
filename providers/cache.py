import redis 
from config import properties
from model import Issue

redis_connection = redis.StrictRedis(host=properties.get('REDIS_HOST'), port=int(properties.get('REDIS_PORT')))


def retrieve_view(key, start=0, stop=4):
    return map(lambda x: obj_from_key(x), redis_connection.lrange(key, start, stop))

def obj_from_key(key):
    if "issue:" in key or "pull:" in key:
        return Issue().from_dict(redis_connection.hgetall(key))

