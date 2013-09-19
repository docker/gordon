import operator

from urllib import urlopen
from adapters import git
from adapters.git import assign_issue
from adapters.git import create_comment

from web.config import properties
from web.app import sentry

class AutomaticPR(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        if content.get('action') != "opened":
            return

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
                dire = '/'.join(f.filename.split("/")[:-1])
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
        maintainer_handle = self._get_maintainer(p)
        assign_issue(num, maintainer_handle)
        create_comment(num, 'cc @{0}, this issue was automatically assigned to you by Gordon'.format(maintainer_handle))

    def _get_maintainer(self, f):
        repo_name = properties.get('GITHUB_REPO')

        base_url = "http://raw.github.com/{0}/master".format(repo_name)
        sentry.captureMessage('base_url is {0}'.format(base_url))
        # based on a path, traverse it backward until you find the maintainer.
        url = '{0}/{1}/MAINTAINERS'.format(base_url, f)
        maintainer = urlopen(url).readline()
        if not maintainer:
            sentry.captureMessage('maintainer not found for url {0}'.format(url))
            self._get_maintainer('/'.join(f.split('/')[:-1]))
        try:
            maintainer_handle = maintainer.split('@')[2].strip()[:-1]
            sentry.captureMessage('read MAINTAINER from {0} and maintainer handle is {1}'.format(url, maintainer_handle))
            return maintainer_handle
        except:
            sentry.captureMessage('unable to parse maintainer file. invalid format.')


