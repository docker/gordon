package maintainer

import (
	gh "github.com/google/go-github/github"
	"time"
)

type Repository struct {
	Organization string
	Name         string
	c            *gh.Client
}

type Issue struct {
	gh.Issue
}

func (r *Repository) GetIssues(since time.Time) ([]*Issue, error) {
	opts := &gh.IssueListByRepoOptions{}
	opts.State = "open"
	opts.Since = since

	issues, resp, err := r.c.Issues.ListByRepo(r.Organization, r.Name, opts)
	if err != nil {
		return nil, err
	}
	out := make([]*Issue, len(issues))
	for i, issue := range issues {
		out[i] = &Issue{issue}
	}
	return out, nil
}
