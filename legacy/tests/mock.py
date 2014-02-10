SIGNED_CLA_LOGIN = ['crosbymichael', 'shykes', 'creack', ]

class CLAChecker(object):
    def __init__(self):
        pass
    
    def check_signed_cla(self, login):
        if login in SIGNED_CLA_LOGIN:
            return True
        return False


