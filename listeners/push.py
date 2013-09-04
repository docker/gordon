from adapters import git

class CLAPushListener(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        print "cla event fired"
        pusher = content.get('pusher').get('name')
        for commit in content.get('commits'):
            author = commit.get('author').get('username')
            commit_id = commit.get('id')
            print author
            if self._check_login(pusher):
                print "adding a comment to the issue saying it's ok to merge this issue."
                git.update_status(commit_id, "success", description="This person has signed the CLA")
            else:
                print "adding a comment to the issue saying it's *not* ok to merge this issue, setting unmergeable flag."
                git.update_status(commit_id, "error", description="Prior to merging, this user must sign the CLA!")


    def _check_login(self, login):
        # check to see if this login has signed the cla...
        # for now use the mock..
        from tests.mock import CLAChecker
        c = CLAChecker()
        return c.check_signed_cla(login)
        

