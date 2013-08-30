from adapters import PullRequestAdapter

def route_and_handle(headers, body):
    hooktype = headers.get('X-Github-Event')
    if hooktype == "pull_request":
        pr = PullRequestAdapter().handle(body)
