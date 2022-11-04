package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sado0823/go-kitx/cmd/kitx/internal/project"
	"github.com/sado0823/go-kitx/cmd/kitx/internal/upgrade"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "kitx",
		Usage:   "⭐️ cli command for go-kitx",
		Version: "v0.0.1",
		Suggest: true,
		Commands: []*cli.Command{
			// cmd upgrade
			upgrade.Cmd(),
			// new project
			project.Cmd(),
			{
				Name:    "complete",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("completed task: ", cCtx.Args().First())
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
