package gordon

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"code.google.com/p/go.codereview/patch"
)

func GetReviewersForPR(patch io.Reader) (map[string][]string, error) {
	toplevel, err := GetTopLevelGitRepo()
	if err != nil {
		return nil, err
	}
	maintainers, err := GetMaintainersFromRepo(toplevel)
	if err != nil {
		return nil, err
	}

	return ReviewPatch(patch, maintainers)
}

// ReviewPatch reads a git-formatted patch from `src`, and for each file affected by the patch
// it assign its Maintainers based on the current repository tree directories
// The list of Maintainers are generated when the MaintainerManager object is instantiated.
//
// The result is a map where the keys are the paths of files affected by the patch,
// and the values are the maintainers assigned to review that partiular file.
//
// There is no duplicate checks: the same maintainer may be present in multiple entries
// of the map, or even multiple times in the same entry if the MAINTAINERS file has
// duplicate lines.
func ReviewPatch(src io.Reader, maintainers map[string][]string) (map[string][]string, error) {
	var (
		reviewers = make(map[string][]string)
		index     = buildFileIndex(maintainers)
	)

	input, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	set, err := patch.Parse(input)
	if err != nil {
		return nil, err
	}
	mapReviewers := func(rm map[string]bool) []string {
		var (
			i   int
			out = make([]string, len(rm))
		)
		for k := range rm {
			out[i] = k
			i++
		}
		return out
	}

	for _, f := range set.File {
		for _, originalTarget := range []string{f.Dst, f.Src} {
			if originalTarget == "" {
				continue
			}
			target := path.Clean(originalTarget)
			if _, exists := reviewers[target]; exists {
				continue
			}

			var fileMaintainers map[string]bool
			fileMaintainers = index[target]
			for len(fileMaintainers) == 0 {
				target = path.Dir(target)
				fileMaintainers = index[target]
			}
			reviewers[originalTarget] = mapReviewers(fileMaintainers)
		}
	}
	return reviewers, nil
}

type MaintainerFile map[string][]*Maintainer

type Maintainer struct {
	Username string
	FullName string
	Email    string
	Target   string
	Active   bool
	Lead     bool
	Raw      string
}

// Currently not being used
func LoadMaintainerFile(dir string) (MaintainerFile, error) {
	src, err := os.Open(path.Join(dir, "MAINTAINERS"))
	if err != nil {
		return nil, err
	}
	maintainers := make(MaintainerFile)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		m := parseMaintainer(scanner.Text())
		if m.Username == "" && m.Email == "" && m.FullName == "" {
			return nil, fmt.Errorf("Incorrect maintainer format: %s", m.Raw)
		}
		if _, exists := maintainers[m.Target]; !exists {
			maintainers[m.Target] = make([]*Maintainer, 0, 1)
		}
		maintainers[m.Target] = append(maintainers[m.Target], m)
	}
	return maintainers, nil
}

func parseMaintainer(line string) *Maintainer {
	const (
		commentIndex  = 1
		targetIndex   = 3
		fullnameIndex = 4
		emailIndex    = 5
		usernameIndex = 7
	)
	re := regexp.MustCompile("^[ \t]*(#|)((?P<target>[^: ]*) *:|) *(?P<fullname>[a-zA-Z][^<]*) *<(?P<email>[^>]*)> *(\\(@(?P<username>[^\\)]+)\\)|).*$")
	match := re.FindStringSubmatch(line)
	return &Maintainer{
		Active:   match[commentIndex] == "",
		Target:   path.Base(path.Clean(match[targetIndex])),
		Username: strings.Trim(match[usernameIndex], " \t"),
		Email:    strings.Trim(match[emailIndex], " \t"),
		FullName: strings.Trim(match[fullnameIndex], " \t"),
		Raw:      line,
	}
}

// Currently not being used
//
// TopMostMaintainerFile moves up the directory tree looking for a MAINTAINERS file,
// parses the top-most file it finds, and returns its contents.
// This is used to find the top-level maintainer of a project for certain
// privileged reviews, such as authorizing changes to a MAINTAINERS file.
func TopMostMaintainerFile(dir string) (MaintainerFile, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if dir == "/" {
		return make(MaintainerFile), nil
	}
	parent, err := TopMostMaintainerFile(path.Dir(dir))
	if err != nil {
		// Ignore recursive errors which might be caused by
		// permission errors on parts of the filesystem, etc.
		parent = make(MaintainerFile)
	}
	if len(parent) > 0 {
		return parent, nil
	}
	current, err := LoadMaintainerFile(dir)
	if os.IsNotExist(err) {
		return make(MaintainerFile), nil
	} else if err != nil {
		return nil, err
	}
	return current, nil
}
