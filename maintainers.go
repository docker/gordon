package gordon

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	MaintainerManagersFileName = "MAINTAINERS"
	NumWorkers                 = 10
)

var (
	maintainerDirMap  = MaintainerManagerDirectoriesMap{}
	maintainersIds    = []string{}
	maintainersDirMap = map[string][]*Maintainer{}
)

type MaintainerManagerDirectoriesMap struct {
	paths []string
}

func getMaintainerManagersIds(pth string) ([]string, []*Maintainer, error) {
	var (
		maintainers        = []*Maintainer{}
		maintainersFileMap = []string{}
	)

	f, err := os.Open(pth)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if t := scanner.Text(); t != "" && t[0] != '#' {
			m := parseMaintainer(t)

			if m.Username == "" && m.Email == "" {
				return nil, nil, fmt.Errorf("Incorrect maintainer format: %s", m.Raw)
			}

			if m.Username != "" {
				maintainers = append(maintainers, m)
				maintainersFileMap = append(maintainersFileMap, m.Username)
			}
		}
	}
	sort.Strings(maintainersFileMap)

	return maintainersFileMap, maintainers, nil
}

func createMaintainerManagersDirectoriesMap(pth, cpth, maintainerEmail, userName string) error {
	names, err := ioutil.ReadDir(pth)
	if err != nil {
		return err
	}

	var (
		fileMaintainers                  []*Maintainer
		foundMaintainerManagersFile      bool
		iAmOneOfTheMaintainerManagers    bool
		belongsToOtherMaintainerManagers bool
	)

	for _, name := range names {
		if strings.EqualFold(name.Name(), MaintainerManagersFileName) {
			var ids []string
			foundMaintainerManagersFile = true

			ids, fileMaintainers, err = getMaintainerManagersIds(path.Join(pth, name.Name()))
			if err != nil {
				return err
			}
			maintainersIds = append(maintainersIds, ids...)
			sort.Strings(maintainersIds)

			i := sort.SearchStrings(ids, maintainerEmail)
			if i < len(ids) && ids[i] == maintainerEmail {
				iAmOneOfTheMaintainerManagers = true
			} else {
				i := sort.SearchStrings(ids, userName)
				if i < len(ids) && ids[i] == userName {
					iAmOneOfTheMaintainerManagers = true
				}
			}
			break
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
		maintainerDirMap.paths = append(maintainerDirMap.paths, tmpcpth)
	} else if foundMaintainerManagersFile || belongsToOthers {
		belongsToOtherMaintainerManagers = true
	}

	for _, name := range names {
		if name.IsDir() && name.Name()[0] != '.' {
			var (
				tmpcpth = path.Join(cpth, name.Name())
				newPath = path.Join(pth, name.Name())
			)

			belongsToOthers = belongsToOtherMaintainerManagers
			createMaintainerManagersDirectoriesMap(newPath, tmpcpth, maintainerEmail, userName)
		}
	}
	return nil
}
