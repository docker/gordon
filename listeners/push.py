from adapters import git

class CLAPushListener(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        pusher = content.get('pusher').get('name')
        for commit in content.get('commits'):
            author = commit.get('author').get('username')
            commit_id = commit.get('id')
            if self._check_login(pusher):
                git.update_status(commit_id, "success", description="This person has signed the CLA")
            else:
                git.update_status(commit_id, "error", description="Prior to merging, this user must sign the CLA!")

    def _check_login(self, login):
        # check to see if this login has signed the cla...
        # for now use the mock..
        from tests.mock import CLAChecker
        c = CLAChecker()
        return c.check_signed_cla(login)
        


