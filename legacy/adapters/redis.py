import redis 
from config import properties


class RedisAdapter(object):
    def __init__(self):
        self.connection = redis.StrictRedis(host=properties.get('REDIS_HOST'), port=int(properties.get('REDIS_PORT')))

    def add(prefix, num, di):
        if not isinstance(di, dict): return
        self.connection.hmset('{0}:{1}'.format(prefix, num), di)

    def zadd(key, weight, relationship):
        "# relationship is k:v"
        self.connection.zadd(key, weight, relationship)

    def push(name, prefix, num):
        key = "{0}:{1}".format(prefix, num)
        self.connection.rpush(name, key)

    def delete(key):
        self.connection.delete(key)

    def construct_issue_list_from_range(key, start=0, stop=4):
        print "getting key {0}".format(key)
        ret = []
        for issue in r.lrange(key, start, stop):
            j = r.hgetall(issue)
            i = Issue(j['number'], j['title'], j['created_at'])
            ret.append(i)
        return ret

    def construct_committer_list_from_range(key, start=0, stop=4):
        ret = []
        for committer in r.zrevrange(key, start, stop):
            c = r.hgetall(committer)
            i = Committer(c['name'], c['commits'])
            ret.append(i)
        return ret

    def get_top_contributors():
        return construct_committer_list_from_range("top-committers")

    def get_oldest_issues():
        return construct_issue_list_from_range("oldest-issues")

    def get_oldest_pulls():
        return construct_issue_list_from_range("oldest-pulls")

    def get_least_pulls():
        return construct_issue_list_from_range("least-updated-pulls")

    def get_least_issues():
        return construct_issue_list_from_range("least-updated-issues")

redis_connection = RedisAdapter()
