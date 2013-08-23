from tasks.cache import cache_issues
from tasks.cache import cache_pulls
from tasks.cache import cache_commits
from tasks.cache import oldest_issues
from tasks.cache import oldest_pulls
from tasks.cache import least_issues
from tasks.cache import least_pulls

#base stuff

cache_issues()
cache_pulls()
cache_commits()

# filters / views
oldest_issues()
oldest_pulls()
least_issues()
least_pulls()

