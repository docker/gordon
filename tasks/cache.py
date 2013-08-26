from adapters.git import issues, pulls, commits
from providers.cache import redis_connection as r
from providers.cache import obj_from_key

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

def open_issues():
    open_issues = filter_keys("issue:*", lambda x: obj_from_key(x).state == "open")
    return open_issues

def open_pulls():
    open_issues = filter_keys("pull:*", lambda x: obj_from_key(x).state == "open")
    return open_issues

def closed_issues():
    closed_issues = filter_keys("issue:*", lambda x: obj_from_key(x).state == "closed")
    return closed_issues

def commits():
    commits = filter_keys("commit:*", lambda x: obj_from_key(x))

def oldest(lst):
    return sorted(lst, key=lambda x: x.created_at)
    
def least_updated(lst):
    return sorted(lst, key=lambda x: x.updated_at)

def oldest_issues():
    op = open_issues()
    ls = [obj_from_key(z) for z in op]
    create_view('oldest-issues', oldest(ls))

def oldest_pulls():
    op = open_pulls()
    ls = [obj_from_key(z) for z in op]
    create_view('oldest-pulls', oldest(ls))

def least_issues():
    op = open_issues()
    ls = [obj_from_key(z) for z in op]
    create_view('least-updated-issues', least_updated(ls))

def least_pulls():
    op = open_pulls()
    ls = [obj_from_key(z) for z in op]
    create_view('least-updated-pulls', least_updated(ls))

def issues_closed_since(start=0, days=7):
    cl = closed_issues()
    ls = [obj_from_key(z) for z in cl]
    create_view('issues-closed-since-{0}-{1}'.format(start, days), filter_since(ls, start, days))

def issues_opened_since(start=0, days=7):
    cl = open_issues()
    ls = [obj_from_key(z) for z in cl]
    create_view('issues-open-since-{0}-{1}'.format(start, days), filter_since(ls, start, days))

def unassigned_pulls():
    up = open_pulls()
    ls = [obj_from_key(z) for z in up]
    create_view('unassigned-prs', filter(lambda x: x.assignee is not None, ls))

def filter_since(ls, start, total_days=7):
    from datetime import timedelta
    from datetime import date
    if start == 0:
        base = date.today()
    else:
        base = date.today() - timedelta(start)
    date_range = [(base - timedelta(days=x)).isoformat() for x in range(0,total_days)]
    return filter(lambda x: x.closed_at.split("T",1)[0] in date_range, ls)


def create_view(key, lst):
    #lst should have objects which have a __key__ method.
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
