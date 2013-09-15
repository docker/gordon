from github import Github as git
from config import properties

def auth_git():
    return git(properties.get('GITHUB_USERNAME'), properties.get('GITHUB_PASSWORD'), timeout=3000)

def get_repo():
    g = auth_git()
    org = g.get_organization("dotcloud")
    docker_repo = org.get_repo("docker")
    return docker_repo

def create_comment(number, body, *args, **kwargs):
    repo = get_repo()
    pull = repo.get_pul(number)
    pull.create_comment(body, *args, **kwargs)
    

def update_status(commit_id, state, **kwargs):
    g = auth_git()
    repo = g.get_repo("keeb/docker-build")
    commit = repo.get_commit(commit_id)
    commit.create_status(state, **kwargs)
    print "created status.."


def issues(*args, **kwargs):
    return [z for z in get_repo().get_issues(*args, **kwargs)]

def pulls(*args, **kwargs):
    return [z for z in get_repo().get_pulls(*args, **kwargs)]

def commits(*args, **kwargs):
    return [z for z in get_repo().get_commits(*args, **kwargs)]

