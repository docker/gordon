from flask import render_template
from app import app
from model import build_cache, get_oldest_issues
from model import get_oldest_pulls, get_least_pulls, get_least_issues, get_top_contributors
from model import Cache

def cache():
    build_cache()
    return render_template('cache.html')

def index():
    c = Cache()
    c.load()
    return render_template('index.html', 
            oldest_issues = get_oldest_issues(),
            oldest_pulls = get_oldest_pulls(),
            attention_issues = get_least_issues(),
            attention_pulls = get_least_pulls(),
            top_contributors = get_top_contributors(),
            cache = c,
            )

def robot():
    return render_template("robot.html")

