from datetime import datetime

class Issue(object):
    def __init__(self, number=None, title=None, created_at=None, updated_at=None, closed_at=None, state=None, assignee=None):
        self.number = number
        self.title = title
        self.created_at = created_at
        self.updated_at = updated_at
        self.closed_at = closed_at
        self.state = state
        self.assignee = assignee
        if self.created_at:
            self.age = self.age()

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
        self.assignee = data.get('assignee')
        return self

class IssueCollection(object):
    def __init__(self, this_week_count=None, last_week_count=None):
        print "initialized"
        self.this_week_count = this_week_count
        self.last_week_count = last_week_count
        self.difference = self.calculate_difference()


    def calculate_difference(self):
        if self.last_week_count == 0 or self.this_week_count == 0:
            return 'N/A'
        # ;\
        a = float(self.last_week_count) - float(self.this_week_count)
        ret = (a / float(self.last_week_count)) * 100
        return "%.2f" % ret




class Committer(object):
    def __init__(self, name, commits):
        self.name = name
        self.commits = commits

