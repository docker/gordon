import operator

from urllib import urlopen
from adapters import git
from adapters.git import assign_issue

from web.config import properties
from web.app import sentry

class AutomaticPR(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        if content.get('action') != "opened":
            return
        repo_name = properties.get('GITHUB_REPO')
        branch = content.get('pull_request').get('head').get('ref')
        base_url = "http://raw.github.com/{0}/{1}".format(repo_name, branch)
        sentry.captureMessage('base_url is {0}'.format(base_url))

        repo = git.get_repo()
        num = content.get('pull_request').get('number')
        sentry.captureMessage('pull_request number is {0}'.format(num))
        p = repo.get_pull(num)
        files = p.get_files()

        fd = {}
        ire = {}

        for f in files:
            sentry.captureMessage('working on file: {0}'.format(f.filename))
            if "/" in f.filename:
                dire = ''.join(f.filename.split("/")[:-1])
            else:
                dire = '/'

            fd[f.filename] = {'changes': f.changes, 
                    'additions': f.additions, 
                    'deletions': f.deletions,
                    }

            if ire.get(dire):
                score = ire.get(dire) + f.changes
            else:
                score = f.changes
            ire[dire] = score

        sorted_ire = sorted(ire.iteritems(), key=operator.itemgetter(1))
        sorted_ire.reverse()
        p = sorted_ire[0][0]
        url = '{0}/{1}/MAINTAINERS'.format(base_url, p)
        maintainer = urlopen(url).readline()
        maintainer_handle = maintainer.split('@')[2].strip()[:-1]
        sentry.captureMessage('read MAINTAINER from {0} and maintainer handle is {1}'.format(url, maintainer_handle))
        assign_issue(num, maintainer_handle)


