from flask import render_template
from controller import IssueController as IC
from controller import IssueCollectionController as ICC

def index():
    c = None
    ic = IC()
    icc = ICC()
    return render_template('index.html', 
            oldest_issues = ic.get_oldest_issues(),
            oldest_pulls = ic.get_oldest_pulls(),
            attention_issues = ic.get_least_issues(),
            attention_pulls = ic.get_least_pulls(),
            top_contributors = ic.get_top_contributors(),
            issue_open_collection = icc.get_issues_opened_count(),
            issue_closed_collection = icc.get_issues_closed_count(),
            cache = c,
            )

def robot():
    return render_template("robot.html")

