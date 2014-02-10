from adapters import PullRequestAdapter
from adapters import PushAdapter

from listeners.pulls import AutomaticPR
from listeners.pulls import DCOPullListener

import logging

log = logging.getLogger('router')

def route_and_handle(headers, body):
    hooktype = headers.get('X-Github-Event')
    log.debug('recieved hooktype {0}'.format(hooktype))

    if hooktype == "pull_request":
        pr = PullRequestAdapter()
        pr.add_listener(DCOPullListener())
        pr.handle(body)
    elif hooktype == "push":
        pu = PushAdapter()
        pu.handle(body)
