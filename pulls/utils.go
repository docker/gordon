package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Token string
}

type remote struct {
	Name string
	Url  string
}

func writeError(format string, err error) {
	fmt.Fprintf(os.Stderr, format, err)
	os.Exit(1)
}

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
	out := []remote{}
	s := bufio.NewScanner(bytes.NewBuffer(output))
	for s.Scan() {
		o := remote{}
		if _, err := fmt.Sscan(s.Text(), &o.Name, &o.Url); err != nil {
			return nil, err
		}
		out = append(out, o)
	}

	return out, nil
}

func loadConfig() Config {
	var config Config
	f, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			writeError("%s", err)
		}
	} else {
		defer f.Close()

		dec := json.NewDecoder(f)
		if err := dec.Decode(&config); err != nil {
			writeError("%s", err)
		}
	}
	return config
}

func saveConfig(config Config) error {
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
