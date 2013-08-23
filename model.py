from datetime import datetime

class Issue(object):
    def __init__(self, number=None, title=None, created_at=None, updated_at=None, closed_at=None, state=None):
        self.number = number
        self.title = title
        self.created_at = created_at
        self.updated_at = updated_at
        self.closed_at = closed_at
        self.state = state

    def age(self):
        delta = (datetime.strptime(self.created_at, '%Y-%m-%dT%H:%M:%SZ') - datetime.now()).days
        return abs(delta)
    
    def __repr__(self):
        return "{0} - {1}".format(self.number, self.title)

    def __key__(self):
        return "issue:{0}".format(self.number)

    def from_dict(self, data):
        self.number = data.get('number')
        self.title = data.get('title')
        self.created_at = data.get('created_at')
        self.updated_at = data.get('updated_at')
        self.closed_at = data.get('closed_at')
        self.age = self.age()
        self.state = data.get('state')
        return self


class Cache(object):
    def __init__(self, last_updated):
        self.last_updated = None


class Committer(object):
    def __init__(self, name, commits):
        self.name = name
        self.commits = commits

