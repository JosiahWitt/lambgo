package cmd

import (
	"github.com/urfave/cli/v2"
)

func (a *App) buildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "build Lambdas",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "disable-parallel",
				Usage: "Disables building in parallel",
			},
		},

		Action: func(c *cli.Context) error {
			pwd, err := a.Getwd()
			if err != nil {
				return err
			}

			config, err := a.LambgoFileLoader.LoadConfig(pwd)
			if err != nil {
				return err
			}

			config.DisableParallelBuild = c.Bool("disable-parallel")
			return a.Builder.BuildBinaries(config)
		},
	}
}
