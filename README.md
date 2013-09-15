# Gordon

Gordon is a multi-function robot providing a precence via the Web and IRC, while exposing a rich API. If you're contributing to Docker, his goal is to make your life AWESOME.

If you're interested in how non-trivial projects can be built, tested, distributed, or run via Docker, Gordon's goal is to provide a reference which adheres to the best practices developed by the community in these areas.

Gordon, like Docker, is only as strong as the community around it. Civil discourse is encouraged. Pull Requests and Issues are appreciated.

# Development Stack

* Python 
* RQ (Tasks)
* Flask (Web)
* Redis (Cache)
* MySQL (Persistence)

# Dependencies

A thorough list of python dependencies are available in the requirements.txt or README.md of each component and are buildable by their Dockerfile. See the `deps` directory.

* Bender - http://github.com/dotcloud/bender

# Components

## GitHub WebHooks

The GitHub API is feature rich and allows for immense workflow customization/integration. Gordon takes full advantage of this by implementing a router-based event dispatch (see: `api/router.py`). Adding custom functionality is as simple as creating a new python object which operates on `content`. 

Out of box, Gordon offers

* Automatic Pull Request Assignment to Maintainers
* Caching of data in realtime, as opposed polling

## IRC Precence

Command line automation and stewardship is the name of the game. 

### Standup (Bender)


## Web Precence

## Caching


