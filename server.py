from app import app
from api import index
from api import cache

app.add_url_rule('/', 'index', index)
app.add_url_rule('/cache', 'cache', cache)

if __name__=="__main__":
    app.run('0.0.0.0')
