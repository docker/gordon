from adapters import PullRequestAdapter
from adapters import PushAdapter

from listeners.push import CLAPushListener

def route_and_handle(headers, body):
    hooktype = headers.get('X-Github-Event')
    if hooktype == "pull_request":
        pr = PullRequestAdapter()
        pr.handle(body)
    elif hooktype == "push":
        pu = PushAdapter()
        pu.add_listener(CLAPushListener())
        pu.handle(body)
