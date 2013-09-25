import operator

from github import Github as git
from urllib import urlopen

from web.config import properties
from web.app import sentry



def rank_file_changes(flist):
    ire = {}

    for f in flist:
        sentry.captureMessage('working on file: {0}'.format(f.filename))
        if "/" in f.filename:
            dire = '/'.join(f.filename.split("/")[:-1])
        else:
            dire = '/'

        if ire.get(dire):
            score = ire.get(dire) + f.changes
        else:
            score = f.changes
        ire[dire] = score

    sorted_ire = sorted(ire.iteritems(), key=operator.itemgetter(1))
    return sorted_ire.reverse()



def get_lead_maintainer(issue):
    repo = get_repo()
    p = repo.get_pull(issue)
    files = p.get_files()
    lead_maintainer_file = rank_file_changes(files)[0][0]
    maintainer_handle = _maintainer_from_path(lead_maintainer_file)
    return maintainer_handle

def get_all_maintainers(issue):
    repo = get_repo()
    p = repo.get_pull(issue)
    files = p.get_files()
    maintainers = []
    for f in files:
        maintainer = _maintainer_from_path(f.filename)
        print maintainer
        if maintainer not in maintainers:
            maintainers.append(maintainer)
    return maintainers



def _maintainer_from_path(path):
    repo_name = properties.get('GITHUB_REPO')

    base_url = "http://raw.github.com/{0}/master".format(repo_name)
    sentry.captureMessage('base_url is {0}'.format(base_url))
    # based on a path, traverse it backward until you find the maintainer.
    url = '{0}/{1}/MAINTAINERS'.format(base_url, path)
    maintainer = urlopen(url).readline()
    if not maintainer:
        sentry.captureMessage('maintainer not found for url {0}'.format(url))
        _maintainer_from_path('/'.join(f.split('/')[:-1]))
    try:
        maintainer_handle = maintainer.split('@')[2].strip()[:-1]
        sentry.captureMessage('read MAINTAINER from {0} and maintainer handle is {1}'.format(url, maintainer_handle))
        return maintainer_handle
    except:
        sentry.captureMessage('unable to parse maintainer file. invalid format.')


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

