from app import app
from views import index
from views import hook
from views import maintainers

app.add_url_rule('/', 'index', index, methods=['GET'])
app.add_url_rule('/', 'hook', hook, methods=['POST'])
app.add_url_rule('/maintainers/<issue>', 'maintainer', maintainers, methods=['GET'])

if __name__=="__main__":
    app.run('0.0.0.0')
