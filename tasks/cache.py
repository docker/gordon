from tasks.cache import cache_issues
from tasks.cache import cache_pulls
from tasks.cache import cache_commits
from tasks.cache import oldest_issues
from tasks.cache import oldest_pulls
from tasks.cache import least_issues
from tasks.cache import least_pulls
from tasks.cache import issues_closed_since
from tasks.cache import issues_opened_since
from tasks.cache import unassigned_pulls

#base stuff

#cache_issues()
#cache_pulls()
#cache_commits()

# filters / views
#oldest_issues()
#oldest_pulls()
#least_issues()
#least_pulls()

#issues_closed_since(start=0, days=7)
#issues_closed_since(start=7, days=14)

issues_opened_since(start=0, days=7)
issues_opened_since(start=7, days=14)

#unassigned_pulls()

