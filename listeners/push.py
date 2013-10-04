from adapters import git

class DCOPushListener(object):
    def __init__(self):
        pass

    def event_fired(self, content):
        # signed commit should look like
        # Signed-off-by: FirstName LastName <emailaddress>
        pusher = content.get('pusher').get('name')
        for commit in content.get('commits'):
            author = commit.get('author').get('username')
            email = commit.get('author').get('email')
            real_name = commit.get('author').get('name')
            message = commit.get('message')
            commit_id = commit.get('id')
            if self._check_commit_message(message, real_name, email):
                git.update_status(commit_id, "success", description="Commit has been properly signed.")
            else:
                git.update_status(commit_id, "error", description="This commit has not been properly signed. Please rebase with a proper commit message.")

    def _check_commit_message(self, message, name, email):
        ret = 'Signed-off-by: {0} <{1}>'.format(name, email) in message
        return ret
        


