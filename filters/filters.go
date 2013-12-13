package filters

import (
	"fmt"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/crosbymichael/pulls"
	"strconv"
	"strings"
	"time"
)

type PullRequestsFilter func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error)
type IssuesFilter func(issues []*gh.Issue, err error) ([]*gh.Issue, error)

// Return the pr filter based on the context
func GetPullRequestFilter(c *cli.Context) PullRequestsFilter {
	filter := defaultPullRequestsFilter
	if c.Bool("new") {
		filter = combinePullRequests(filter, newPullRequestsFilter)
	}
	if user := c.String("user"); user != "" {
		filter = func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
			return userPullRequestsFilter(prs, user, err)
		}
	}
	if c.Bool("lgtm") {
		filter = combinePullRequests(filter, lgtmPullRequestsFilter)
	}
	if c.Bool("no-merge") {
		filter = combinePullRequests(filter, noMergePullRequestsFilter)
	}
	return filter
}

// Return the pr filter based on the context
func GetIssueFilter(c *cli.Context) IssuesFilter {
	filter := defaultIssuesFilter
	if c.Bool("new") {
		filter = combineIssues(filter, newIssuesFilter)
	}
	if numVotes := c.Int("votes"); numVotes > 0 {
		filter = func(issues []*gh.Issue, err error) ([]*gh.Issue, error) {
			return voteIssuesFilter(issues, numVotes, err)
		}
	}
	return filter
}

func combinePullRequests(filter, next PullRequestsFilter) PullRequestsFilter {
	return func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
		return next(filter(prs, err))
	}
}

func combineIssues(filter, next IssuesFilter) IssuesFilter {
	return func(prs []*gh.Issue, err error) ([]*gh.Issue, error) {
		return next(filter(prs, err))
	}
}

func defaultPullRequestsFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	return prs, err
}

func defaultIssuesFilter(prs []*gh.Issue, err error) ([]*gh.Issue, error) {
	return prs, err
}

func noMergePullRequestsFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fmt.Printf(".")
		// We have to fetch the single pr to get the merge state
		// it sucks but we have to do it
		client := gh.NewClient()
		org, name, err := pulls.GetOriginUrl()
		if err != nil {
			panic(err)
		}
		t, err := pulls.NewMaintainer(client, org, name)
		if err != nil {
			panic(err)
		}
		pr, _, err := t.GetPullRequest(strconv.Itoa(pr.Number), false)
		if err != nil {
			return nil, err
		}
		if !pr.Mergeable {
			out = append(out, pr)
		}
	}
	return out, nil
}

func userPullRequestsFilter(prs []*gh.PullRequest, user string, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fmt.Printf(".")
		if pr.User.Login == user {
			out = append(out, pr)
		}
	}
	return out, nil
}

func lgtmPullRequestsFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		fmt.Printf(".")
		client := gh.NewClient()
		org, name, err := pulls.GetOriginUrl()
		if err != nil {
			panic(err)
		}
		t, err := pulls.NewMaintainer(client, org, name)
		if err != nil {
			panic(err)
		}
		comments, err := t.GetComments(strconv.Itoa(pr.Number))
		if err != nil {
			return nil, err
		}
		pr.ReviewComments = 0
		for _, comment := range comments {
			// We should check it this LGTM is by a user in
			// the maintainers file
			if strings.Contains(comment.Body, "LGTM") {
				pr.ReviewComments += 1
			}
		}
	}
	return prs, nil
}

func newPullRequestsFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	yesterday := time.Now().Add(-24 * time.Hour)
	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fmt.Printf(".")
		if pr.CreatedAt.After(yesterday) {
			out = append(out, pr)
		}
	}
	return out, nil
}

func voteIssuesFilter(issues []*gh.Issue, numVotes int, err error) ([]*gh.Issue, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.Issue{}
	for _, issue := range issues {
		fmt.Printf(".")
		client := gh.NewClient()
		org, name, err := pulls.GetOriginUrl()
		if err != nil {
			panic(err)
		}
		t, err := pulls.NewMaintainer(client, org, name)
		if err != nil {
			panic(err)
		}
		comments, err := t.GetComments(strconv.Itoa(issue.Number))
		if err != nil {
			return nil, err
		}
		issue.Comments = 0
		for _, comment := range comments {
			if strings.Contains(comment.Body, "+1") {
				issue.Comments += 1
			}
		}
		if issue.Comments >= numVotes {
			out = append(out, issue)
		}
	}
	return out, nil
}

func newIssuesFilter(issues []*gh.Issue, err error) ([]*gh.Issue, error) {
	if err != nil {
		return nil, err
	}

	yesterday := time.Now().Add(-24 * time.Hour)
	out := []*gh.Issue{}
	for _, issue := range issues {
		fmt.Printf(".")
		if issue.CreatedAt.After(yesterday) {
			out = append(out, issue)
		}
	}
	return out, nil
}
