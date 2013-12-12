package pulls

import (
	"bufio"
	"encoding/json"
	"fmt"
	gh "github.com/crosbymichael/octokat"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Top level type that manages a repository
type Maintainer struct {
	repo           gh.Repo
	client         *gh.Client
	email          string
	directoriesMap *MaintainerDirectoriesMap
}

type MaintainerDirectoriesMap struct {
	paths []string
}

type Config struct {
	Token string
}

const MaintainersFileName = "MAINTAINERS"

var configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")

var maintainerDirectoriesMap = MaintainerDirectoriesMap{}

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

func getEmailFromLine(str string) (string, bool, error) {
	exp, err := regexp.Compile(`([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})`)
	if err != nil {
		return "", false, err
	}
	return exp.FindString(str), exp.MatchString(str), err
}

func getRepoPath(pth, org string) string {
	flag := false
	i := 0
	repoPath := path.Dir("/")
	for _, dir := range strings.Split(pth, "/") {
		if strings.EqualFold(dir, org) {
			flag = true
		}
		if flag {
			if i >= 2 {
				repoPath = path.Join(repoPath, dir)
			}
			i++
		}
	}
	return repoPath
}

func getMaintainersEmails(pth string) (*[]string, error) {
	maintainersFileMap := []string{}
	file, _ := os.Open(pth)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		strEmail, isEmail, err := getEmailFromLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		if isEmail {
			email := []string{strEmail}
			maintainersFileMap = append(maintainersFileMap, email...)
		}
	}
	sort.Strings(maintainersFileMap)

	return &maintainersFileMap, nil
}

func createMaintainerDirectoriesMap(pth, cpth, maintainerEmail string, belongsToOthers bool) error {
	names, err := ioutil.ReadDir(pth)
	if err != nil {
		return err
	}
	// Look for the Maintainer File
	foundMaintainersFile := false
	iAmOneOfTheMaintainers := false
	belongsToOtherMaintainers := false
	for _, name := range names {
		if strings.EqualFold(name.Name(), MaintainersFileName) {
			foundMaintainersFile = true
			emails, err := getMaintainersEmails(path.Join(pth, name.Name()))
			if err != nil {
				return err
			}
			i := sort.SearchStrings(*emails, maintainerEmail)
			if i < len(*emails) && (*emails)[i] == maintainerEmail {
				iAmOneOfTheMaintainers = true
			}
		}
	}
	// Check if we need to add the directory to the maintainer's  directories mapping tree
	if (!foundMaintainersFile && !belongsToOthers) || iAmOneOfTheMaintainers {
		tmpcpth := cpth
		if cpth == "" {
			tmpcpth = "."
		}
		currentPath := []string{tmpcpth}
		maintainerDirectoriesMap.paths = append(maintainerDirectoriesMap.paths, currentPath...)
	} else if foundMaintainersFile || belongsToOthers {
		belongsToOtherMaintainers = true
	}
	for _, name := range names {
		if name.IsDir() && name.Name()[0] != '.' {
			tmpcpth := path.Join(cpth, name.Name())
			newPath := path.Join(pth, name.Name())
			createMaintainerDirectoriesMap(newPath, tmpcpth, maintainerEmail, belongsToOtherMaintainers)
		}
	}

	return err
}

func getOriginPath(org string) (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	originPath := path.Dir("/")
	for _, dir := range strings.Split(currentPath, "/") {
		originPath = path.Join(originPath, dir)
		if strings.EqualFold(dir, org) {
			break
		}
	}
	return originPath, err
}

func NewMaintainer(client *gh.Client, org, repo string) (*Maintainer, error) {

	config, err := LoadConfig()
	if err == nil {
		client.WithToken(config.Token)
	}

	originPath, err := getOriginPath(org)
	if err != nil {
		return nil, err
	}
	originPath = path.Join(originPath, repo)

	email, err := GetMaintainerEmail()
	if err != nil {
		return nil, err
	}

	err = createMaintainerDirectoriesMap(originPath, "", email, false)
	if err != nil {
		return nil, err
	}

	return &Maintainer{
		repo:           gh.Repo{Name: repo, UserName: org},
		client:         client,
		directoriesMap: &maintainerDirectoriesMap,
		email:          email,
	}, nil
}

func (m *Maintainer) Repository() (*gh.Repository, error) {
	return m.client.Repository(m.repo, nil)
}

// Return all the pull requests that I care about
func (m *Maintainer) GetPullRequestsThatICareAbout(showAll bool, state string) ([]*gh.PullRequest, error) {

	if showAll {
		return m.GetPullRequests(state)
	}

	filteredPrs := []*gh.PullRequest{}
	prs, err := m.GetPullRequests(state)
	if err != nil {
		return nil, err
	}
	for _, p := range prs {
		prfs, err := m.GetPullRequestFiles(strconv.Itoa(p.Number))
		if err != nil {
			return nil, err
		}
		for _, prf := range prfs {
			dirPath := filepath.Dir(prf.FileName)
			i := sort.SearchStrings((*m.directoriesMap).paths, dirPath)
			if i < len(m.directoriesMap.paths) && (*m.directoriesMap).paths[i] == dirPath {
				pr := []*gh.PullRequest{p}
				filteredPrs = append(filteredPrs, pr...)
				break
			}
		}
		fmt.Printf(".")

	}
	return filteredPrs, nil
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

// Return all pull request Files
func (m *Maintainer) GetPullRequestFiles(number string) ([]*gh.PullRequestFile, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{}
	allPrFiles := []*gh.PullRequestFile{}

	if prfs, err := m.client.PullRequestFiles(m.repo, number, o); err != nil {
		return nil, err
	} else {
		allPrFiles = append(allPrFiles, prfs...)

	}
	return allPrFiles, nil
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
