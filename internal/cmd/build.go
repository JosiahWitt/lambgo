package cmd

import (
	"github.com/urfave/cli/v2"
)

func (a *App) buildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "build Lambdas",

		Action: func(c *cli.Context) error {
			pwd, err := a.Getwd()
			if err != nil {
				return err
			}

			config, err := a.LambgoFileLoader.LoadConfig(pwd)
			if err != nil {
				return err
			}

			return a.Builder.BuildBinaries(config)
		},
	}
}
