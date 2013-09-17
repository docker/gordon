from github import Github as git

from web.config import properties
from web.app import sentry

def auth_git():
    return git(properties.get('GITHUB_USERNAME'), properties.get('GITHUB_PASSWORD'), timeout=3000)

def get_repo():
    g = auth_git()
    sentry.captureMessage('getting repo {0}'.format(properties.get('GITHUB_REPO')))
    docker_repo = g.get_repo(properties.get('GITHUB_REPO'))
    return docker_repo

def create_comment(number, body, *args, **kwargs):
    repo = get_repo()
    pull = repo.get_issue(number)
    pull.create_comment(body, *args, **kwargs)
    
def assign_issue(number, user):
    g = auth_git()
    r = g.get_repo(properties.get('GITHUB_REPO'))
    i = r.get_issue(number)
    sentry.captureMessage('assigning issue#{0} to {1} on repo {2}'.format(number, user, properties.get('GITHUB_REPO')))
    u = g.get_user(user)
    i.edit(assignee=u)

def update_status(commit_id, state, **kwargs):
    g = auth_git()
    repo = g.get_repo(properties.get('GITHUB_REPO'))
    commit = repo.get_commit(commit_id)
    commit.create_status(state, **kwargs)
    print "created status.."


def issues(*args, **kwargs):
    return [z for z in get_repo().get_issues(*args, **kwargs)]

def pulls(*args, **kwargs):
    return [z for z in get_repo().get_pulls(*args, **kwargs)]

def commits(*args, **kwargs):
    return [z for z in get_repo().get_commits(*args, **kwargs)]

