package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"path"
)

var configPath = path.Join(os.Getenv("HOME"), ".maintainercfg")

type Config struct {
	Token string
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
func authCmd(c *cli.Context) {
	if token := c.String("add"); token != "" {
		if err := saveConfig(Config{token}); err != nil {
			writeError("%s", err)
		}
		return
	}

	// Display token and user information
	if config := loadConfig(); config.Token != "" {
		fmt.Fprintf(os.Stdout, "Token: %s\n", config.Token)
	} else {
		fmt.Fprintf(os.Stderr, "No token registered\n")
		os.Exit(1)
	}
}
