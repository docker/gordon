package gordon

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	MaintainerFileName = "MAINTAINERS"
	NumWorkers         = 10
)

// GetMaintainersFromRepo returns the maintainers for a repo with the username
// as the key and the file's that they own as a slice in the value
func GetMaintainersFromRepo(repoPath string) (map[string][]string, error) {
	current := make(map[string][]string)

	if err := getMaintainersForDirectory(repoPath, repoPath, current); err != nil {
		return nil, err
	}
	return current, nil
}

func getMaintainersForDirectory(root, dir string, current map[string][]string) error {
	maintainersPerFile, err := getMaintainersFromFile(dir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for m, files := range maintainersPerFile {
		for _, f := range files {
			p, err := filepath.Rel(root, filepath.Join(dir, f))
			if err != nil {
				return err
			}
			current[m] = append(current[m], p)
		}
	}

	contents, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range contents {
		if fi.IsDir() && fi.Name() != ".git" {
			if err := getMaintainersForDirectory(root, filepath.Join(dir, fi.Name()), current); err != nil {
				return err
			}
		}
	}
	return nil
}

func getMaintainersFromFile(dir string) (map[string][]string, error) {
	maintainerFile := filepath.Join(dir, MaintainerFileName)
	f, err := os.Open(maintainerFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		maintainer = make(map[string][]string)
		s          = bufio.NewScanner(f)
	)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}
		t := s.Text()
		if t == "" || t[0] == '#' {
			continue
		}
		m := parseMaintainer(t)
		if m.Email == "" || m.Username == "" {
			return nil, fmt.Errorf("invalid maintainer file format %s in %s", t, maintainerFile)
		}
		target := m.Target
		if target == "" {
			target = "*"
		}
		maintainer[m.Username] = append(maintainer[m.Username], target)
	}
	return maintainer, nil
}

// this function basically reverses the maintainers format so that file paths can be looked
// up by path and the maintainers are the value.  We have to parse the directories differently
// at first then lookup per path when we actually have the files so that it is much faster
// and cleaner than walking a fill dir tree looking at files and placing them into memeory.
//
// I swear I'm not crazy
func buildFileIndex(maintainers map[string][]string) map[string]map[string]bool {
	index := make(map[string]map[string]bool)

	for m, files := range maintainers {
		for _, f := range files {
			nm, exists := index[f]
			if !exists {
				nm = make(map[string]bool)
				index[f] = nm
			}
			nm[m] = true
		}
	}
	return index
}
