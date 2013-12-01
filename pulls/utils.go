package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/aybabtme/color/brush"
	"os"
	"os/exec"
	"strings"
)

type remote struct {
	Name string
	Url  string
}

// Write the specific error message with the givin
// format and exit with a status code of 1
func writeError(format string, err error) {
	fmt.Fprintf(os.Stderr, format+"\n", brush.Red(err.Error()))
	os.Exit(1)
}

// Pulls looks at the local repos origin to
// know what github project to work with
func getOriginUrl() (string, string, error) {
	remotes, err := getRemotes()
	if err != nil {
		return "", "", nil
	}
	for _, r := range remotes {
		if r.Name == "origin" {
			parts := strings.Split(r.Url, "/")

			org := parts[len(parts)-2]
			if i := strings.LastIndex(org, ":"); i > 0 {
				org = org[i+1:]
			}

			name := parts[len(parts)-1]
			name = strings.TrimRight(name, ".git")

			return org, name, nil
		}
	}
	return "", "", nil
}

// Return the remotes for the current dir
func getRemotes() ([]remote, error) {
	output, err := exec.Command("git", "remote", "-v").Output()
	if err != nil {
		return nil, err
	}

	var (
		out = []remote{}
		s   = bufio.NewScanner(bytes.NewBuffer(output))
	)

	for s.Scan() {
		o := remote{}
		if _, err := fmt.Sscan(s.Text(), &o.Name, &o.Url); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}
