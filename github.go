package pulls

import (
	"encoding/json"
	"fmt"
	gh "github.com/crosbymichael/octokat"
	"os"
	"path"
	"strconv"
	"strings"
)

// Top level type that manages a repository
type Maintainer struct {
	repo   gh.Repo
	client *gh.Client
}

type Config struct {
	Token string
}

var configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")

func LoadConfig() (*Config, error) {
	var config Config
	f, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return &config, err
		}
	} else {
		defer f.Close()

		dec := json.NewDecoder(f)
		if err := dec.Decode(&config); err != nil {
			return &config, err
		}
	}
	return &config, err
}

func SaveConfig(config Config) error {
	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(config); err != nil {
		return err
	}
	return nil
}

func NewMaintainer(client *gh.Client, org, repo string) (*Maintainer, error) {

	config, err := LoadConfig()
	if err == nil {
		client.WithToken(config.Token)
	}

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
		"sort":      "updated",
		"direction": "asc",
		"state":     state,
		"per_page":  "100",
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

func (m *Maintainer) GetFirstPullRequest(state, sortBy string) (*gh.PullRequest, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"state":     state,
		"per_page":  "1",
		"page":      "1",
		"sort":      sortBy,
		"direction": "asc",
	}
	prs, err := m.client.PullRequests(m.repo, o)
	if err != nil {
		return nil, err
	}
	if len(prs) == 0 {
		return nil, fmt.Errorf("No matching pull request")
	}
	return prs[0], nil
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

// Add a comment to an existing pull request
func (m *Maintainer) AddComment(number, comment string) (gh.Comment, error) {
	return m.client.AddComment(m.repo, number, comment)
}

// Merge a pull request
// If no LGTMs are in the comments require force to be true
func (m *Maintainer) MergePullRequest(number, comment string, force bool) (gh.Merge, error) {
	comments, err := m.GetComments(number)
	if err != nil {
		return gh.Merge{}, err
	}
	isApproved := false
	for _, c := range comments {
		// FIXME: Again should check for LGTM from a maintainer
		if strings.Contains(c.Body, "LGTM") {
			isApproved = true
			break
		}
	}
	if !isApproved && !force {
		return gh.Merge{}, fmt.Errorf("Pull request %s has not been approved", number)
	}
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

func (m *Maintainer) GetFirstIssue(state, sortBy string) (*gh.Issue, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"state":     state,
		"per_page":  "1",
		"page":      "1",
		"sort":      sortBy,
		"direction": "asc",
	}
	issues, err := m.client.Issues(m.repo, o)
	if err != nil {
		return &gh.Issue{}, err
	}
	if len(issues) == 0 {
		return &gh.Issue{}, fmt.Errorf("No matching issues")
	}
	return issues[0], nil
}

// GetIssues queries the GithubAPI for all issues matching the state `state` and the
// assignee `assignee`.
// See http://developer.github.com/v3/issues/#list-issues-for-a-repository
func (m *Maintainer) GetIssues(state, assignee string) ([]*gh.Issue, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"sort":      "updated",
		"direction": "asc",
		"state":     state,
		"per_page":  "100",
	}
	// If assignee == "", don't add it to the params.
	// This will show all issues, assigned or not.
	if assignee != "" {
		o.QueryParams["assignee"] = assignee
	}
	prevSize := -1
	page := 1
	all := []*gh.Issue{}
	for len(all) != prevSize {
		o.QueryParams["page"] = strconv.Itoa(page)
		if issues, err := m.client.Issues(m.repo, o); err != nil {
			return nil, err
		} else {
			prevSize = len(all)
			all = append(all, issues...)
			page += 1
		}
		fmt.Printf(".")
	}
	return all, nil
}
