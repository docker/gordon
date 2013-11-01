package pulls

import (
	"fmt"
	gh "github.com/crosbymichael/octokat"
	"strconv"
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
func (m *Maintainer) GetPullRequest(number string) (*gh.PullRequest, error) {
	return m.client.PullRequest(m.repo, number, nil)
}

// Return all comments for an issue or pull request
func (m *Maintainer) GetComments(pr *gh.PullRequest) ([]gh.Comment, error) {
	number := strconv.Itoa(pr.Number)
	return m.client.Comments(m.repo, number, nil)
}

// Return all pull requests that cannot be merged cleanly
func (m *Maintainer) GetNoMergePullRequests() ([]*gh.PullRequest, error) {
	prs, err := m.GetPullRequests("open")
	if err != nil {
		return nil, err
	}
	out := []*gh.PullRequest{}
	for _, pr := range prs {
		fullPr, err := m.GetPullRequest(strconv.Itoa(pr.Number))
		fmt.Printf(".")
		if err != nil {
			return nil, err
		}
		if !fullPr.Mergeable {
			out = append(out, fullPr)
		}
	}
	fmt.Printf("%c[2K\r", 27)
	return out, nil
}

// Add a comment to an existing pull request
func (m *Maintainer) AddComment(pr *gh.PullRequest, comment string) (gh.Comment, error) {
	return m.client.AddComment(m.repo, strconv.Itoa(pr.Number), comment)
}
