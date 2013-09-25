from app import app
from views import index
from views import hook
from views import maintainers
from views import lead_maintainer

app.add_url_rule('/', 'index', index, methods=['GET'])
app.add_url_rule('/', 'hook', hook, methods=['POST'])
app.add_url_rule('/maintainers/<issue>', 'maintainer', maintainers, methods=['GET'])
app.add_url_rule('/lead_maintainer/<issue>', 'lead_maintainer', lead_maintainer, methods=['GET'])

if __name__=="__main__":
    app.run('0.0.0.0')
