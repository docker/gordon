from github import Github as git
from config import properties

def auth_git():
    return git(properties.get('GITHUB_USERNAME'), properties.get('GITHUB_PASSWORD'), timeout=3000)

def get_repo():
    g = auth_git()
    org = g.get_organization("dotcloud")
    docker_repo = org.get_repo("docker")
    return docker_repo

def issues(*args, **kwargs):
    return [z for z in get_repo().get_issues(*args, **kwargs)]

def pulls(*args, **kwargs):
    return [z for z in get_repo().get_pulls(*args, **kwargs)]

def commits(*args, **kwargs):
    return [z for z in get_repo().get_commits(*args, **kwargs)]

