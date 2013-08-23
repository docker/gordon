from flask import render_template
from app import app
from adapters.fredis import RedisAdapter
from model import Cache

def cache():
    build_cache()
    return render_template('cache.html')

def index():
    c = Cache()
    c.load()
    a = RedisAdapter()
    return render_template('index.html', 
            oldest_issues = a.get_oldest_issues(),
            oldest_pulls = a.get_oldest_pulls(),
            attention_issues = a.get_least_issues(),
            attention_pulls = a.get_least_pulls(),
            top_contributors = a.get_top_contributors(),
            cache = c,
            )

def robot():
    return render_template("robot.html")

