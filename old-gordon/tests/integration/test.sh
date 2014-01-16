curl -i \
    -H "Content-Type: application/json" \
    -H "X-Github-Event: pull_request" \
    -X POST -d "$(cat data.json)" \
    http://api.stinemat.es
