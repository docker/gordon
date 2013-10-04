import operator

from urllib import urlopen
from adapters import git
from adapters.git import get_all_maintainers
from adapters.git import assign_issue
from adapters.git import create_comment

from web.config import properties

class AutomaticPR(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        if content.get('action') != "opened":
            # we really don't care about events that are related to pull requests being opened.
            return

        num = content.get('pull_request').get('number')
        maintainers = get_all_maintainers(num)
        assign_issue(num, 'gordn')
        create_comment(num, 'Hey {0}, can you please take a look at this issue?', ','.join(maintainers))

