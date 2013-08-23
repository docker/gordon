from flask import render_template
from controller import IssueController as IC

def index():
    c = None
    ic = IC()
    print ic.get_oldest_issues()
    return render_template('index.html', 
            oldest_issues = ic.get_oldest_issues(),
            oldest_pulls = ic.get_oldest_pulls(),
            attention_issues = ic.get_least_issues(),
            attention_pulls = ic.get_least_pulls(),
            top_contributors = ic.get_top_contributors(),
            cache = c,
            )

def robot():
    return render_template("robot.html")

