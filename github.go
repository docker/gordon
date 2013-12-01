package pulls

import (
	"fmt"
	"github.com/codegangsta/cli"
	gh "github.com/crosbymichael/octokat"
	"strconv"
	"strings"
)

// Top level type that manages a repository
type Maintainer struct {
	repo   gh.Repo
	client *gh.Client
}

func NewMaintainer(client *gh.Client, org, repo string) (*Maintainer, error) {
	return &Maintainer{
		repo:   gh.Repo{Name: repo, UserName: org},
		client: client,
	}, nil
}

func (m *Maintainer) Repository() (*gh.Repository, error) {
	return m.client.Repository(m.repo, nil)
}

// Return all pull requests
func (m *Maintainer) GetPullRequests(state string) ([]*gh.PullRequest, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"state":    state,
		"per_page": "100",
	}
	prevSize := -1
	page := 1
	allPRs := []*gh.PullRequest{}
	for len(allPRs) != prevSize {
		o.QueryParams["page"] = strconv.Itoa(page)
		if prs, err := m.client.PullRequests(m.repo, o); err != nil {
			return nil, err
		} else {
			prevSize = len(allPRs)
			allPRs = append(allPRs, prs...)
			page += 1
		}
		fmt.Printf(".")
	}
	return allPRs, nil
}

// Return a single pull request
// Return pr's comments if requested
func (m *Maintainer) GetPullRequest(number string, comments bool) (*gh.PullRequest, []gh.Comment, error) {
	var c []gh.Comment
	pr, err := m.client.PullRequest(m.repo, number, nil)
	if err != nil {
		return nil, nil, err
	}
	if comments {
		c, err = m.GetComments(number)
		if err != nil {
			return nil, nil, err
		}
	}
	return pr, c, nil
}

// Return all comments for an issue or pull request
func (m *Maintainer) GetComments(number string) ([]gh.Comment, error) {
	return m.client.Comments(m.repo, number, nil)
}

// Filter pull requests
func (m *Maintainer) FilterPullRequests(prs []*gh.PullRequest, c *cli.Context) ([]*gh.PullRequest, error) {
	out := []*gh.PullRequest{}
	noMerge := c.Bool("no-merge")
	fromUser := c.String("user")
	lgtm := c.Bool("lgtm")

	for _, pr := range prs {
		fmt.Printf(".")
		pr.ReviewComments = 0
		if fromUser == "" || fromUser == pr.User.Login {
			if noMerge {
				pr, _, err := m.GetPullRequest(strconv.Itoa(pr.Number), false)
				if err != nil {
					return nil, err
				}
				if pr.Mergeable {
					continue
				}
			}
			if lgtm {
				comments, err := m.GetComments(strconv.Itoa(pr.Number))
				if err != nil {
					return nil, err
				}
				for _, comment := range comments {
					if strings.Contains(comment.Body, "LGTM") {
						pr.ReviewComments += 1
					}
				}
			}
			out = append(out, pr)
		}
	}
	return out, nil
}

// Add a comment to an existing pull request
func (m *Maintainer) AddComment(number, comment string) (gh.Comment, error) {
	return m.client.AddComment(m.repo, number, comment)
}

// Merge a pull request
func (m *Maintainer) MergePullRequest(number, comment string) (gh.Merge, error) {
	o := &gh.Options{}
	o.Params = map[string]string{
		"commit_message": comment,
	}
	return m.client.MergePullRequest(m.repo, number, o)
}

// Checkout the pull request into the working tree of
// the users repository.
// This will mimic the operations on the manual merge view
func (m *Maintainer) Checkout(pr *gh.PullRequest) error {
	var (
		userBranch        = fmt.Sprintf("%s-%s", pr.User.Login, pr.Head.Ref)
		destinationBranch = pr.Base.Ref
	)

	// Checkout a new branch locally before pulling the changes
	if err := Git("checkout", "-b", userBranch, destinationBranch); err != nil {
		return err
	}

	if err := Git("pull", pr.Head.Repo.CloneURL, pr.Head.Ref); err != nil {
		return err
	}
	return nil
}
