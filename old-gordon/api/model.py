class Sender(object):
    def __init__(self):
        pass
    def from_json(self, json):
        pass

class User(object):
    def __init__(self):
        pass

    def from_json(self, json):
        self.login = json.get("login")

class Repository(object):
    def __init__(self):
        pass

    def from_json(self, json):
        pass
