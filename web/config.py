from os import environ

properties = {}

properties['GITHUB_USERNAME'] = environ['GITHUB_USERNAME']
properties['GITHUB_PASSWORD'] = environ['GITHUB_PASSWORD']
properties['GITHUB_REPO'] = environ['GITHUB_REPO']
properties['REDIS_PORT'] = environ['REDIS_PORT']
properties['REDIS_HOST'] = environ['REDIS_HOST']
