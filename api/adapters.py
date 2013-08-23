class BaseAdapter(object):
    def __init__(self):
        pass
    
    def handle(self):
        pass


class ActionAdapter(BaseAdapter):
    def __init__(self):
        super(ActionAdapter, self).__init__()

    def handle(self):
        print "i'm handling it!"
        pass
