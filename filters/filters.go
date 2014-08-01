package filters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"github.com/docker/gordon"
)

func FilterPullRequests(c *cli.Context, prs []*gh.PullRequest) ([]*gh.PullRequest, error) {
	var (
		yesterday  = time.Now().Add(-24 * time.Hour)
		out        = []*gh.PullRequest{}
		email, err = gordon.GetMaintainerManagerEmail()
	)
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

		maintainer := c.String("maintainer")
		dir := c.String("dir")
		extension := c.String("extension")

		var diff []byte

		if maintainer != "" || dir != "" || extension != "" {
			diffResp, err := http.Get(pr.DiffURL)
			if err != nil {
				continue
			}

			diff, err = ioutil.ReadAll(diffResp.Body)
			if err != nil {
				continue
			}

			diffResp.Body.Close()
		}

		if dir != "" {
			dirs, err := gordon.GetDirsForPR(diff, dir)
			if err != nil {
				continue
			}

			if len(dirs) == 0 {
				continue
			}
		}

		if extension != "" {
			files, err := gordon.GetFileExtensionsForPR(diff, extension)
			if err != nil {
				continue
			}

			if len(files) == 0 {
				continue
			}
		}

		if maintainer != "" || c.Bool("mine") {
			if maintainer == "" {
				maintainer = email
			}

			var found bool
			reviewers, err := gordon.GetReviewersForPR(diff, true)
			if err != nil {
				continue
			}
			for file := range reviewers {
				for _, reviewer := range reviewers[file] {
					if reviewer == maintainer {
						found = true
					}
				}
			}
			if !found {
				continue
			}

		}

		if c.Bool("unassigned") && pr.Assignee != nil {
			continue
		} else if assigned := c.String("assigned"); assigned != "" && (pr.Assignee == nil || pr.Assignee.Login != assigned) {
			continue
		}

		if c.Bool("lgtm") {
			pr.ReviewComments = 0
			maintainersOccurrence := map[string]bool{}
			for _, comment := range pr.CommentsBody {
				// We should check it this LGTM is by a user in
				// the maintainers file
				userName := comment.User.Login
				if strings.Contains(comment.Body, "LGTM") && !maintainersOccurrence[userName] {
					maintainersOccurrence[userName] = true
					pr.ReviewComments += 1
				}
			}
		}

		if c.Bool("no-merge") && pr.Mergeable {
			continue
		}

		out = append(out, pr)
	}
	return out, nil

}

func FilterIssues(c *cli.Context, issues []*gh.Issue) ([]*gh.Issue, error) {
	var (
		yesterday      = time.Now().Add(-24 * time.Hour)
		out            = []*gh.Issue{}
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

	for _, issue := range issues {
		fmt.Printf(".")

		if c.Bool("new") && !issue.CreatedAt.After(yesterday) {
			continue
		}

		if milestone := c.String("milestone"); milestone != "" && issue.Milestone.Title != milestone {
			continue
		}

		if numVotes := c.Int("votes"); numVotes > 0 {
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
			if issue.Comments < numVotes {
				continue
			}
		}

		if c.Bool("proposals") && !strings.HasPrefix(issue.Title, "Proposal") {
			continue
		}

		out = append(out, issue)
	}
	return out, nil

}
