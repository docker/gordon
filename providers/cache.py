import redis 
from config import properties

from model import Committer
from model import Issue
from model import Cache

redis_connection = redis.StrictRedis(host=properties.get('REDIS_HOST'), port=int(properties.get('REDIS_PORT')))

class CacheProvider(object):
    def construct_issue_list_from_range(key, start=0, stop=4):
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

