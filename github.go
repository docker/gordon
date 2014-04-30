package gordon

import (
	"bufio"
	"encoding/json"
	"fmt"
	gh "github.com/crosbymichael/octokat"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Top level type that manages a repository
type MaintainerManager struct {
	repo              gh.Repo
	client            *gh.Client
	email             string
	maintainerDirMap  *MaintainerManagerDirectoriesMap
	maintainersIds    *[]string
	maintainersDirMap *map[string][]*Maintainer
}

type MaintainerManagerDirectoriesMap struct {
	paths []string
}

type Config struct {
	Token    string
	UserName string
}

const (
	MaintainerManagersFileName = "MAINTAINERS"
	NumWorkers                 = 10
)

var (
	maintainerDirMap  = MaintainerManagerDirectoriesMap{}
	maintainersIds    = []string{}
	maintainersDirMap = map[string][]*Maintainer{}
	belongsToOthers   = false
	fileMaintainers   = []*Maintainer{}
	configPath        = path.Join(os.Getenv("HOME"), ".maintainercfg")
)

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

func getMaintainerManagersIds(pth string) (*[]string, []*Maintainer, error) {
	maintainersFileMap := []string{}
	file, _ := os.Open(pth)
	scanner := bufio.NewScanner(file)
	var maintainers = []*Maintainer{}
	for scanner.Scan() {

		if t := scanner.Text(); t != "" && t[0] != '#' {
			m := parseMaintainer(t)
			if m.Username == "" && m.Email == "" {
				return nil, nil, fmt.Errorf("Incorrect maintainer format: %s", m.Raw)
			}
			if m.Username != "" {
				maintainers = append(maintainers, []*Maintainer{m}...)
			}
			if m.Email != "" {
				email := []string{m.Email}
				maintainersFileMap = append(maintainersFileMap, email...)
			}
			if m.Username != "" {
				userName := []string{m.Username}
				maintainersFileMap = append(maintainersFileMap, userName...)
			}
		}
	}
	sort.Strings(maintainersFileMap)

	return &maintainersFileMap, maintainers, nil
}

func createMaintainerManagersDirectoriesMap(pth, cpth, maintainerEmail, userName string) error {
	names, err := ioutil.ReadDir(pth)
	if err != nil {
		return err
	}
	// Look for the MaintainerManager File
	var (
		foundMaintainerManagersFile      = false
		iAmOneOfTheMaintainerManagers    = false
		belongsToOtherMaintainerManagers = false
	)

	for _, name := range names {
		if strings.EqualFold(name.Name(), MaintainerManagersFileName) {
			foundMaintainerManagersFile = true
			var ids = &[]string{}
			ids, fileMaintainers, err = getMaintainerManagersIds(path.Join(pth, name.Name()))
			maintainersIds = append(maintainersIds, (*ids)...)
			sort.Strings(maintainersIds)
			if err != nil {
				return err
			}
			i := sort.SearchStrings(*ids, maintainerEmail)
			if i < len(*ids) && (*ids)[i] == maintainerEmail {
				iAmOneOfTheMaintainerManagers = true
			} else {
				i := sort.SearchStrings(*ids, userName)
				if i < len(*ids) && (*ids)[i] == userName {
					iAmOneOfTheMaintainerManagers = true
				}
			}
		}
	}

	// Save the maintainers list related to the current directory
	tmpcpth := cpth
	if cpth == "" {
		tmpcpth = "."
	}
	if foundMaintainerManagersFile {
		maintainersDirMap[tmpcpth] = fileMaintainers
	}
	// Check if we need to add the directory to the maintainer's  directories mapping tree
	if (!foundMaintainerManagersFile && !belongsToOthers) || iAmOneOfTheMaintainerManagers {
		currentPath := []string{tmpcpth}
		maintainerDirMap.paths = append(maintainerDirMap.paths, currentPath...)
	} else if foundMaintainerManagersFile || belongsToOthers {
		belongsToOtherMaintainerManagers = true
	}
	for _, name := range names {
		if name.IsDir() && name.Name()[0] != '.' {
			tmpcpth := path.Join(cpth, name.Name())
			newPath := path.Join(pth, name.Name())
			belongsToOthers = belongsToOtherMaintainerManagers
			createMaintainerManagersDirectoriesMap(newPath, tmpcpth, maintainerEmail, userName)
		}
	}

	return err
}

func getOriginPath(repo string) (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	originPath := path.Dir("/")
	for _, dir := range strings.Split(currentPath, "/") {
		originPath = path.Join(originPath, dir)
		if strings.EqualFold(dir, repo) {
			break
		}
	}
	return originPath, err
}

func NewMaintainerManager(client *gh.Client, org, repo string) (*MaintainerManager, error) {

	config, err := LoadConfig()
	if err == nil {
		client.WithToken(config.Token)
	}
	originPath, err := getOriginPath(repo)
	if err != nil {
		return nil, err
	}
	email, err := GetMaintainerManagerEmail()
	if err != nil {
		return nil, err
	}
	err = createMaintainerManagersDirectoriesMap(originPath, "", email, config.UserName)
	if err != nil {
		return nil, err
	}
	return &MaintainerManager{
		repo:              gh.Repo{Name: repo, UserName: org},
		client:            client,
		maintainerDirMap:  &maintainerDirMap,
		email:             email,
		maintainersIds:    &maintainersIds,
		maintainersDirMap: &maintainersDirMap,
	}, nil
}

func (m *MaintainerManager) Repository() (*gh.Repository, error) {
	return m.client.Repository(m.repo, nil)
}

func (m *MaintainerManager) IsMaintainer(userName string) bool {
	i := sort.SearchStrings(*(m.maintainersIds), userName)
	return (i < len(*(m.maintainersIds)) && (*(m.maintainersIds))[i] == userName)
}

func (m *MaintainerManager) GetMaintainersDirMap() *map[string][]*Maintainer {
	return m.maintainersDirMap
}

func (m *MaintainerManager) worker(prepr <-chan *gh.PullRequest, pospr chan<- *gh.PullRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range prepr {
		prfs, err := m.GetPullRequestFiles(strconv.Itoa(p.Number))
		if err != nil {
			return
		}
		for _, prf := range prfs {
			dirPath := filepath.Dir(prf.FileName)
			i := sort.SearchStrings((*m.maintainerDirMap).paths, dirPath)
			if i < len(m.maintainerDirMap.paths) && (*m.maintainerDirMap).paths[i] == dirPath {
				pospr <- p
				break
			}
		}
		fmt.Printf(".")
	}
}

func (m *MaintainerManager) filterPullRequests(prs []*gh.PullRequest) []*gh.PullRequest {
	var (
		producer      = make(chan *gh.PullRequest, NumWorkers)
		consumer      = make(chan *gh.PullRequest, NumWorkers)
		wg            = &sync.WaitGroup{}
		consumerGroup = &sync.WaitGroup{}
		filteredPrs   = []*gh.PullRequest{}
	)

	// take the finished results and put them into the list
	consumerGroup.Add(1)
	go func() {
		defer consumerGroup.Done()

		for p := range consumer {
			filteredPrs = append(filteredPrs, []*gh.PullRequest{p}...)
		}
	}()

	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go m.worker(producer, consumer, wg)
	}

	// add all jobs
	for _, p := range prs {
		producer <- p
	}
	// we are done sending jobs so close the channel
	close(producer)

	wg.Wait()

	close(consumer)
	// wait for the consumer to finish adding all the results to the list
	consumerGroup.Wait()

	return filteredPrs
}

// Return all the pull requests that I care about
func (m *MaintainerManager) GetPullRequestsThatICareAbout(showAll bool, state, sortQuery string) ([]*gh.PullRequest, error) {
	prs, err := m.GetPullRequests(state, sortQuery)
	if err != nil {
		return nil, err
	}

	if showAll {
		return prs, nil
	}

	return m.filterPullRequests(prs), nil
}

// Return all pull requests
func (m *MaintainerManager) GetPullRequests(state, sort string) ([]*gh.PullRequest, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"sort":      sort,
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
func (m *MaintainerManager) GetPullRequestFiles(number string) ([]*gh.PullRequestFile, error) {
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

func (m *MaintainerManager) GetFirstPullRequest(state, sortBy string) (*gh.PullRequest, error) {
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
func (m *MaintainerManager) GetPullRequest(number string, comments bool) (*gh.PullRequest, []gh.Comment, error) {
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

// Return a single issue
// Return issue's comments if requested
func (m *MaintainerManager) GetIssue(number string, comments bool) (*gh.Issue, []gh.Comment, error) {
	var c []gh.Comment
	num, err := strconv.Atoi(number)
	if err != nil {
		return nil, nil, err
	}
	issue, err := m.client.Issue(m.repo, num, nil)
	if err != nil {
		return nil, nil, err
	}
	if comments {
		c, err = m.GetComments(number)
		if err != nil {
			return nil, nil, err
		}
	}
	return issue, c, nil
}

// Return all issue found
func (m *MaintainerManager) GetIssuesFound(query string) ([]*gh.SearchItem, error) {
	o := &gh.Options{}
	o.QueryParams = map[string]string{
		"sort":     "updated",
		"order":    "asc",
		"per_page": "100",
	}
	prevSize := -1
	page := 1
	issuesFound := []*gh.SearchItem{}
	for len(issuesFound) != prevSize {
		o.QueryParams["page"] = strconv.Itoa(page)
		if issues, err := m.client.SearchIssues(query, o); err != nil {
			return nil, err
		} else {
			prevSize = len(issuesFound)
			issuesFound = append(issuesFound, issues...)
			page += 1
		}
		fmt.Printf(".")
	}
	return issuesFound, nil
}

// Return contributors list
func (m *MaintainerManager) GetContributors() ([]*gh.Contributor, error) {
	o := &gh.Options{}
	contributors, err := m.client.Contributors(m.repo, o)
	if err != nil {
		return nil, err
	}

	return contributors, nil
}

// Return all comments for an issue or pull request
func (m *MaintainerManager) GetComments(number string) ([]gh.Comment, error) {
	return m.client.Comments(m.repo, number, nil)
}

// Add a comment to an existing pull request
func (m *MaintainerManager) AddComment(number, comment string) (gh.Comment, error) {
	return m.client.AddComment(m.repo, number, comment)
}

// Merge a pull request
// If no LGTMs are in the comments require force to be true
func (m *MaintainerManager) MergePullRequest(number, comment string, force bool) (gh.Merge, error) {
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
func (m *MaintainerManager) Checkout(pr *gh.PullRequest) error {
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

// Get the user information from the authenticated user
func (m *MaintainerManager) GetGithubUser() (*gh.User, error) {
	user, err := m.client.User("", nil)
	if err != nil {
		return nil, nil
	}
	return user, err
}

// Patch an issue
func (m *MaintainerManager) PatchIssue(number string, issue *gh.Issue) (*gh.Issue, error) {
	o := &gh.Options{}
	o.Params = map[string]string{
		"title":    issue.Title,
		"body":     issue.Body,
		"assignee": issue.Assignee.Login,
	}
	patchedIssue, err := m.client.PatchIssue(m.repo, number, o)
	if err != nil {
		return nil, nil
	}
	return patchedIssue, err
}

// Patch a pull request
func (m *MaintainerManager) PatchPullRequest(number string, pr *gh.PullRequest) (*gh.PullRequest, error) {
	o := &gh.Options{}
	params := map[string]string{
		"title": pr.Title,
		"body":  pr.Body,
	}
	if pr.Assignee == nil {
		params["assignee"] = ""
	} else {
		params["assignee"] = pr.Assignee.Login
	}
	o.Params = params
	// octokat doesn't expose PatchPullRequest. Use PatchIssue instead.
	_, err := m.client.PatchIssue(m.repo, number, o)
	if err != nil {
		return nil, nil
	}
	// Simulate the result of the patching
	patchedPR := *pr
	return &patchedPR, nil
}

func (m *MaintainerManager) GetFirstIssue(state, sortBy string) (*gh.Issue, error) {
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
func (m *MaintainerManager) GetIssues(state, assignee string) ([]*gh.Issue, error) {
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
