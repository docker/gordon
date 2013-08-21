from github import Github as git
from config import properties
from datetime import datetime
import redis
import json

r = redis.StrictRedis(host=properties.get('REDIS_HOST'), port=int(properties.get('REDIS_PORT')))

def auth_git():
    return git(properties.get('GITHUB_USERNAME'), properties.get('GITHUB_PASSWORD'), timeout=3000)

def build_cache():
    clear_cache()
    g = auth_git()
    org = g.get_organization("dotcloud")
    docker_repo = org.get_repo("docker")
    issues = [z for z in docker_repo.get_issues()]
    pulls = [ z for z in docker_repo.get_pulls()]
    commits = [z for z in docker_repo.get_commits()]
    map(lambda x: add("issue", x.number, x.__dict__['_rawData']), issues)
    map(lambda x: add("pull", x.number, x.__dict__['_rawData']), pulls)
    map(lambda x: push("oldest-issues", "issue", x.number), oldest(issues))
    map(lambda x: push("oldest-pulls", "pull", x.number), oldest(pulls))
    map(lambda x: push("least-updated-pulls", "pull", x.number), least_updated(pulls))
    map(lambda x: push("least-updated-issues", "issue", x.number), least_updated(issues))
    map(lambda x: add("commit", x.sha[:5], x.__dict__['_rawData']), commits)
    

    x = 1
    for committer, commit_num in ranked_committers(commits).iteritems():
        add('committer', str(x), {'name': committer, 'commits': commit_num})
        zadd('top-committers', commit_num, "{0}:{1}".format('committer', str(x)))
        x += 1
        

    add("unique-committers-count", "1", {'total': str(len(ranked_committers(commits)))})
    add("total-commits", "1", {'total': str(len(commits))})
    add('github-cache', '1', {'last_updated': datetime.now()})



def ranked_committers(lst):
    ret = {}
    for commit in lst:
        try:
            login = commit.committer.login
        except Exception, e:
            print commit.sha
        
        if ret.get(login):
            i = ret[login]
            i+= 1
            ret[login] = i
        else:
            ret[login] = 1
    return ret

    return sorted(ret, key=lambda x: ret[x], reverse=True)


def oldest(lst):
    return sorted(lst, key=lambda x: x.created_at)

def least_updated(lst):
    return sorted(lst, key=lambda x: x.updated_at)

def add(prefix, num, di):
    #print "adding {0}:{1}".format(prefix, num)
    if not isinstance(di, dict): return
    r.hmset('{0}:{1}'.format(prefix, num), di)

def zadd(key, weight, relationship):
    "# relationship is k:v"
    r.zadd(key, weight, relationship)

def push(name, prefix, num):
    key = "{0}:{1}".format(prefix, num)
    r.rpush(name, key)

def clear_cache():
    for key in r.keys():
        if "issue:" in key or "pull:" in key or "commit" in key:
            r.delete(key)

    r.delete("oldest-issues")
    r.delete("oldest-pulls")
    r.delete("least-updated-pulls")
    r.delete("least-updated-issues")


class Issue(object):
    def __init__(self, number, title, created_at):
        self.number = number
        self.title = title
        self.created_at = created_at
        self.age = self.age()

    def age(self):
        delta = (datetime.strptime(self.created_at, '%Y-%m-%dT%H:%M:%SZ') - datetime.now()).days
        return abs(delta)

class Cache(object):
    def __init__(self):
        self.last_updated = None

    def get_last_updated(self):
        last_updated = r.hget("github-cache:1", 'last_updated')
        return last_updated

    def load(self):
        self.last_updated = self.get_last_updated()
        if not self.last_updated: self.last_updated = "not run"

class Committer(object):
    def __init__(self, name, commits):
        self.name = name
        self.commits = commits

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

