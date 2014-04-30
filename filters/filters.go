package filters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/dotcloud/gordon"
)

func FilterPullRequests(c *cli.Context, prs []*gh.PullRequest) ([]*gh.PullRequest, error) {
	var (
		yesterday      = time.Now().Add(-24 * time.Hour)
		out            = []*gh.PullRequest{}
		client         = gh.NewClient()
		org, name, err = gordon.GetOriginUrl()
	)
	if err != nil {
		return nil, err
	}
	t, err := gordon.NewMaintainerManager(client, org, name)
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		fmt.Printf(".")

		if c.Bool("new") && !pr.CreatedAt.After(yesterday) {
			continue
		}

		if user := c.String("user"); user != "" {
			if pr.User.Login != user {
				continue
			}
		}

		if c.Bool("unassigned") && pr.Assignee != nil {
			continue
		} else if assigned := c.String("assigned"); assigned != "" && (pr.Assignee == nil || pr.Assignee.Login != assigned) {
			continue
		}

		if c.Bool("lgtm") {
			comments, err := t.GetComments(strconv.Itoa(pr.Number))
			if err != nil {
				return nil, err
			}
			pr.ReviewComments = 0
			maintainersOccurrence := map[string]bool{}
			for _, comment := range comments {
				// We should check it this LGTM is by a user in
				// the maintainers file
				userName := comment.User.Login
				if strings.Contains(comment.Body, "LGTM") && t.IsMaintainer(userName) && !maintainersOccurrence[userName] {
					maintainersOccurrence[userName] = true
					pr.ReviewComments += 1
				}
			}
		}

		if c.Bool("no-merge") {
			pr, _, err := t.GetPullRequest(strconv.Itoa(pr.Number), false)
			if err != nil {
				return nil, err
			}
			if pr.Mergeable {
				continue
			}

		}
		out = append(out, pr)
	}
	return out, nil

}

type IssuesFilter func(issues []*gh.Issue, err error) ([]*gh.Issue, error)

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

func combineIssues(filter, next IssuesFilter) IssuesFilter {
	return func(prs []*gh.Issue, err error) ([]*gh.Issue, error) {
		return next(filter(prs, err))
	}
}

func defaultIssuesFilter(prs []*gh.Issue, err error) ([]*gh.Issue, error) {
	return prs, err
}

func assignedPullRequestsFilter(prs []*gh.PullRequest, assignee string, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fmt.Printf(".")
		if (assignee == "" && pr.Assignee == nil) || (pr.Assignee != nil && pr.Assignee.Login == assignee) {
			out = append(out, pr)
		}
	}
	return out, nil
}

func unassignedPullRequestsFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	return assignedPullRequestsFilter(prs, "", err)
}

func voteIssuesFilter(issues []*gh.Issue, numVotes int, err error) ([]*gh.Issue, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.Issue{}
	for _, issue := range issues {
		fmt.Printf(".")
		client := gh.NewClient()
		org, name, err := gordon.GetOriginUrl()
		if err != nil {
			return nil, err
		}
		t, err := gordon.NewMaintainerManager(client, org, name)
		if err != nil {
			return nil, err
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
