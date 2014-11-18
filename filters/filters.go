package filters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
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
		out        = filteredPullRequests{} //[]*gh.PullRequest{}
		email, err = gordon.GetMaintainerManagerEmail()
		chPrs      = make(chan *gh.PullRequest)
	)
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		go func(pr *gh.PullRequest) {
			if c.Bool("new") && !pr.CreatedAt.After(yesterday) {
				chPrs <- nil
				return
			}

			if user := c.String("user"); user != "" {
				if pr.User.Login != user {
					chPrs <- nil
					return
				}
			}

			if c.Bool("cleanup") {
				if !strings.HasPrefix(strings.ToLower(pr.Title), "cleanup") {
					chPrs <- nil
					return
				}
			}

			maintainer := c.String("maintainer")
			if maintainer == "" && c.Bool("mine") {
				maintainer = email
			}
			dir := c.String("dir")
			extension := c.String("extension")

			var diff []byte

			if maintainer != "" || dir != "" || extension != "" {
				diffResp, err := http.Get(pr.DiffURL)
				if err != nil {
					chPrs <- nil
					return
				}

				diff, err = ioutil.ReadAll(diffResp.Body)
				if err != nil {
					chPrs <- nil
					return
				}

				diffResp.Body.Close()
			}

			if dir != "" {
				dirs, err := gordon.GetDirsForPR(diff, dir)
				if err != nil {
					chPrs <- nil
					return
				}

				if len(dirs) == 0 {
					chPrs <- nil
					return
				}
			}

			if extension != "" {
				files, err := gordon.GetFileExtensionsForPR(diff, extension)
				if err != nil {
					chPrs <- nil
					return
				}

				if len(files) == 0 {
					chPrs <- nil
					return
				}
			}

			if maintainer != "" {
				var found bool
				reviewers, err := gordon.GetReviewersForPR(diff, true)
				if err != nil {
					chPrs <- nil
					return
				}
				for file := range reviewers {
					for _, reviewer := range reviewers[file] {
						if reviewer == maintainer {
							found = true
						}
					}
				}
				if !found {
					chPrs <- nil
					return
				}

			}

			if c.Bool("unassigned") && pr.Assignee != nil {
				chPrs <- nil
				return
			} else if assigned := c.String("assigned"); assigned != "" && (pr.Assignee == nil || pr.Assignee.Login != assigned) {
				chPrs <- nil
				return
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
				chPrs <- nil
				return
			}
			chPrs <- pr
		}(pr)
	}
	for i := 0; i < len(prs); i++ {
		if pr := <-chPrs; pr != nil {
			out = append(out, pr)
		}
	}
	sort.Sort(out)
	return out, nil
}

type filteredPullRequests []*gh.PullRequest

func (r filteredPullRequests) Len() int      { return len(r) }
func (r filteredPullRequests) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r filteredPullRequests) Less(i, j int) bool {
	return r[j].UpdatedAt.After(r[i].UpdatedAt)
}

func FilterIssues(c *cli.Context, issues []*gh.Issue) ([]*gh.Issue, error) {
	var (
		yesterday      = time.Now().Add(-24 * time.Hour)
		out            = []*gh.Issue{}
		client         = gh.NewClient()
		org, name, err = gordon.GetRemoteUrl(c.String("remote"))
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
