from providers.cache import retrieve_view as rv
from api.router import route_and_handle
from model import IssueCollection


class ApiController(object):
    def __init__(self):
        pass

    def route(self, headers, request):
        ret = route_and_handle(headers, request)

        return ret


class IssueController(object):
    def __init__(self):
        pass

    def get_oldest_issues(self):
        return rv("oldest-issues")
    
    def get_oldest_pulls(self):
        return rv("oldest-pulls")

    def get_least_issues(self):
        return rv('least-updated-issues')

    def get_least_pulls(self):
        return rv('least-updated-pulls')

    def get_top_contributors(self):
        return rv('top-contributors')


class IssueCollectionController(object):
    def __init__(self):
        pass

    def get_issues_opened_count(self):
        # last week
        ls1 = len(rv('issues-open-since-0-7', stop=-1))
        # week before
        ls2 = len(rv('issues-open-since-7-14', stop=-1))
        return IssueCollection(ls1, ls2)
        
    def get_issues_closed_count(self):
        # last week
        ls1 = len(rv('issues-closed-since-0-7', stop=-1))
        print ls1
        # week before
        ls2 = len(rv('issues-closed-since-7-14', stop=-1))
        
        return IssueCollection(ls1, ls2)

        


class CommitController(object):
    def __init__(self):
        pass


