package upgrade

import (
	"github.com/sado0823/go-kitx/cmd/kitx/internal"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
)

var (
	flagAll = "all"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:    "upgrade",
		Aliases: []string{"u"},
		Usage:   "upgrade kitx tools",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        flagAll,
				DefaultText: "true",
				Usage:       "update all tools including protoc-gen-go*",
				Aliases:     []string{"a"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			answer := false
			confirm := &survey.Confirm{
				Message: "do you want to upgrade kitx tools ❓",
				Help:    "upgrade kitx、 protoc-gen-go-http-kitx、 protoc-gen-go-errors-kitx",
			}
			if err := survey.AskOne(confirm, &answer); err != nil {
				return err
			}
			if !answer {
				return nil
			}
			var (
				toInstall = []string{
					"github.com/sado0823/go-kitx/cmd/kitx@latest",
					"github.com/sado0823/go-kitx/cmd/protoc-gen-go-http-kitx@latest",
					"github.com/sado0823/go-kitx/cmd//protoc-gen-go-errors-kitx@latest",
				}
				addAll = cCtx.Bool(flagAll)
			)
			if addAll {
				toInstall = append([]string{
					"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
					"google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
				}, toInstall...)
			}

			return internal.GoInstall(toInstall...)
		},
	}
}
