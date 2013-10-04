from adapters import PullRequestAdapter
from adapters import PushAdapter

from listeners.pulls import AutomaticPR
from listeners.pulls import DCOPullListener

def route_and_handle(headers, body):
    hooktype = headers.get('X-Github-Event')
    if hooktype == "pull_request":
        pr = PullRequestAdapter()
        pr.add_listener(DCOPullListener())
        pr.handle(body)
    elif hooktype == "push":
        pu = PushAdapter()
        pu.handle(body)
