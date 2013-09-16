from flask import Flask
from raven.contrib.flask import Sentry

app = Flask(__name__)
app.debug = True

app.config['SENTRY_DSN'] = 'http://abda613aba1249e9803e6b589dbcff79:5154827e95374a68a01024cc89c0cebf@sentry.stinemat.es/2'
sentry = Sentry(app)
