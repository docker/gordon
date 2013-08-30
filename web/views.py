from flask import render_template
from flask import request
from controller import IssueController as IC
from controller import IssueCollectionController as ICC
from controller import ApiController 
from api.router import route_and_handle

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

def hook():
    api = ApiController()
    data = api.route(request.headers, request.json)



def robot():
    return render_template("robot.html")

