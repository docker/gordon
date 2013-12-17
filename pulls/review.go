package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"os"
	"strings"
	"code.google.com/p/go.codereview/patch"
)

// ReviewPatch reads a git-formatted patch from `src`, and for each file affected by the patch
// it looks up the hierarchy of MAINTAINERS files in the current directory to
// determine who should review the patch.
// The result is a map where the keys are the paths of files affected by the patch,
// and the values are the maintainers assigned to review that partiular file.
//
// There is no duplicate checks: the same maintainer may be present in multiple entries
// of the map, or even multiple times in the same entry if the MAINTAINERS file has
// duplicate lines.
func ReviewPatch(src io.Reader) (reviewers map[string][]*Maintainer, err error) {
	reviewers = make(map[string][]*Maintainer)
	input, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}
	set, err := patch.Parse(input)
	if err != nil {
		return nil, err
	}
	for _, f := range(set.File) {
		for _, target := range([]string{f.Dst, f.Src}) {
			if target == "" {
				continue
			}
			target = path.Clean(target)
			if _, exists := reviewers[target]; exists {
				continue
			}
			maintainers, err := getMaintainers(target)
			if err != nil {
				return nil, fmt.Errorf("%s: %s\n", target, err)
			}
			reviewers[target] = maintainers
		}
	}
	return reviewers, nil
}


func getMaintainers(target string) (maintainers []*Maintainer, err error) {
	if _, err := os.Stat(target); err != nil {
		return nil, err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return nil, err
	}
	if target == "/" {
		return []*Maintainer{}, nil
	}
	defer func() {
		if err == nil && (maintainers == nil || len(maintainers) == 0) {
			maintainers, err = getMaintainers(path.Dir(target))
		}
	}()
	if path.Base(target) == "MAINTAINERS" {
		tpmf, err := TopMostMaintainerFile(path.Dir(target))
		if err != nil {
			return nil, err
		}
		if maintainers, exists := tpmf["."]; !exists || len(maintainers) == 0 {
			return nil, fmt.Errorf("can't find the top-level maintainer to review MAINTAINERS change")
		} else if lead := maintainers[0]; !lead.Active {
			return nil, fmt.Errorf("can't review MAINTAINERS change: top-level maintainer %s is inactive", lead.FullName)
		} else {
			return []*Maintainer{lead}, nil
		}
	}
	maintainerFile, err := LoadMaintainerFile(path.Dir(target))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if fileMaintainers, exists := maintainerFile[path.Base(target)]; exists {
		return fileMaintainers, nil
	} else if dirMaintainers, exists := maintainerFile["."]; exists {
		return dirMaintainers, nil
	}
	return nil, nil
}

type MaintainerFile map[string][]*Maintainer

type Maintainer struct {
	Username	string
	FullName	string
	Email		string
	Target		string
	Active		bool
	Lead		bool
	Raw		string
}

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
		commentIndex	= 1
		targetIndex	= 3
		fullnameIndex	= 4
		emailIndex	= 5
		usernameIndex	= 7
	)
	re := regexp.MustCompile("^[ \t]*(#|)((?P<target>[^: ]*) *:|) *(?P<fullname>[a-zA-Z][^<]*) *<(?P<email>[^>]*)> *(\\(@(?P<username>[^\\)]+)\\)|).*$")
	match := re.FindStringSubmatch(line)
	return &Maintainer{
		Active:		match[commentIndex] == "",
		Target:		path.Base(path.Clean(match[targetIndex])),
		Username:	strings.Trim(match[usernameIndex], " \t"),
		Email:		strings.Trim(match[emailIndex], " \t"),
		FullName:	strings.Trim(match[fullnameIndex], " \t"),
		Raw:		line,
	}
}

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
