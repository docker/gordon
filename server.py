from app import app
from views import index

app.add_url_rule('/', 'index', index)

if __name__=="__main__":
    app.run('0.0.0.0')
