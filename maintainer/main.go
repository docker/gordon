package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func addProject(c *cli.Context) {

}

func loadCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "add",
			ShortName: "a",
			Usage:     "Add a new repository to the list",
			Action:    addProject,
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "maintainer"
	app.Usage = "Manage github issues and prs"
	loadCommands(app)

	app.Run(os.Args)
}
