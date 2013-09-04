class BaseAdapter(object):
    def __init__(self):
        self.listeners = []

    def add_listener(self, listener):
        self.listeners.append(listener)

    def handle(self, content):
        for listener in self.listeners:
            ret = listener.event_fired(content)


class PullRequestAdapter(BaseAdapter):
    def __init__(self):
        super(PullRequestAdapter, self).__init__()


class PushAdapter(BaseAdapter):
    def __init__(self):
        super(PushAdapter, self).__init__()


