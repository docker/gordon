class BaseAdapter(object):
    def __init__(self):
        self.listeners = []
        pass

    def handle(self):
        raise Exception("must be overriden")

    def add_listener(self, listener):
        self.listeners.append(listener)


class PullRequestAdapter(BaseAdapter):
    def __init__(self):
        super(PullRequestAdapter, self).__init__()

    def handle(self):
        for listener in self.listeners:
            ret = listener.event_fired(content)

