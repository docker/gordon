from github import Github as git
from config import properties

def auth_git():
    return git(properties.get('GITHUB_USERNAME'), properties.get('GITHUB_PASSWORD'), timeout=3000)

g = auth_git()
user = g.get_organization("dotcloud")
docker_repo = user.get_repo("docker")
hooks = ['push', 'issues', 'issue_comment', 'commit_comment', 'pull_request',
        'pull_request_review_comment', 'watch', 'fork', 'fork_apply']
config = {'url': 'http://api.stinemat.es', 'content_type': 'json'}

docker_repo.create_hook('web', config, hooks, True)

