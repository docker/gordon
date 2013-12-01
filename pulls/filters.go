package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"strconv"
	"strings"
)

type Filter func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error)

// Return the pr filter based on the context
func getFilter(c *cli.Context) Filter {
	filter := defaultFilter
	if user := c.String("user"); user != "" {
		filter = func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
			return userFilter(prs, user, err)
		}
	}
	if c.Bool("lgtm") {
		filter = combine(filter, lgtmFilter)
	}
	if c.Bool("no-merge") {
		filter = combine(filter, noMergeFilter)
	}
	return filter
}

func combine(filter, next Filter) Filter {
	return func(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
		return next(filter(prs, err))
	}
}

func defaultFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	return prs, err
}
func noMergeFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fmt.Printf(".")
		// We have to fetch the single pr to get the merge state
		// it sucks but we have to do it
		pr, _, err := m.GetPullRequest(strconv.Itoa(pr.Number), false)
		if err != nil {
			return nil, err
		}
		if !pr.Mergeable {
			out = append(out, pr)
		}
	}
	return out, nil
}
func userFilter(prs []*gh.PullRequest, user string, err error) ([]*gh.PullRequest, error) {
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
func lgtmFilter(prs []*gh.PullRequest, err error) ([]*gh.PullRequest, error) {
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		fmt.Printf(".")
		comments, err := m.GetComments(strconv.Itoa(pr.Number))
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
