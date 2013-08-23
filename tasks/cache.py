from adapters.git import issues, pulls, commits
from providers.cache import redis_connection as r
from model import Issue


def clear(whitelist=None, blacklist=None):
    for key in r.keys():
        if "issue:" in key or "pull:" in key or "commit" in key:
            r.delete(key)

    r.delete("oldest-issues")
    r.delete("oldest-pulls")
    r.delete("least-updated-pulls")
    r.delete("least-updated-issues")

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






def cache_issues():
    open_issue_list = issues(state="open")
    closed_issue_list = issues(state="closed")
    for issue in open_issue_list:
        prefix = "issue"
        suffix = issue.number
        json_dict = issue.__dict__.get('_rawData')
        r.hmset('{0}:{1}'.format(prefix, suffix), json_dict)

    for issue in closed_issue_list:
        prefix = "issue"
        suffix = issue.number
        json_dict = issue.__dict__.get('_rawData')
        r.hmset('{0}:{1}'.format(prefix, suffix), json_dict)
        
def cache_pulls():
    open_pull_list = pulls(state="open")
    closed_pull_list = pulls(state="closed")

    for issue in open_pull_list:
        prefix = "pull"
        suffix = issue.number
        json_dict = issue.__dict__.get('_rawData')
        r.hmset('{0}:{1}'.format(prefix, suffix), json_dict)

    for issue in closed_pull_list:
        prefix = "pull"
        suffix = issue.number
        json_dict = issue.__dict__.get('_rawData')
        r.hmset('{0}:{1}'.format(prefix, suffix), json_dict)


def cache_commits():
    commit_list = commits()
    for commit in commit_list:
        prefix = "commit"
        suffix = commit.sha[:5]
        json_dict = commit.__dict__
        r.hmset('{0}:{1}'.format(prefix, suffix), json_dict)


def filter_keys(key, func):
    lst = r.keys(key)
    return filter(func, lst)


def issue_from_key(key):
    i = Issue()
    data = r.hgetall(key)
    i.number = data.get('number')
    i.title = data.get('title')
    i.created_at = data.get('created_at')
    i.updated_at = data.get('updated_at')
    i.closed_at = data.get('closed_at')
    i.state = data.get('state')
    return i

def open_issues():
    open_issues = filter_keys("issue:*", lambda x: issue_from_key(x).state == "open")
    return open_issues

def open_pulls():
    open_issues = filter_keys("pull:*", lambda x: issue_from_key(x).state == "open")
    return open_issues

def closed_issues():
    closed_issues = filter_keys("issue:*", lambda x: issue_from_key(x).state == "closed")
    return closed_issues

def oldest(lst):
    return sorted(lst, key=lambda x: x.created_at)
    
def least_updated(lst):
    return sorted(lst, key=lambda x: x.updated_at)

def oldest_issues():
    op = open_issues()
    ls = [issue_from_key(z) for z in op]
    create_view('oldest-issues', oldest(ls))

def oldest_pulls():
    op = open_pulls()
    ls = [issue_from_key(z) for z in op]
    create_view('oldest-pulls', oldest(ls))

def least_issues():
    op = open_issues()
    ls = [issue_from_key(z) for z in op]
    create_view('least-updated-issues', least_updated(ls))

def least_pulls():
    op = open_pulls()
    ls = [issue_from_key(z) for z in op]
    create_view('least-updated-pulls', least_updated(ls))


def create_view(key, lst):
    r.delete(key)
    for i in lst:
        k = i.__key__()
        r.rpush(key, k)




def build_cache():
    pulls = [z for z in pulls()]
    commits = [z for z in commits()]
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
