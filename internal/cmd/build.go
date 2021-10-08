package cmd

import (
	"strings"

	"github.com/JosiahWitt/erk"
	"github.com/urfave/cli/v2"
)

type ErkCannotFilterBuildPaths struct{ erk.DefaultKind }

var ErrCannotFilterBuildPaths = erk.New(ErkCannotFilterBuildPaths{}, "Cannot filter build paths, since '{{.allowedPath}}' is not in: {{.buildPaths}}")

func (a *App) buildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "build Lambdas",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "disable-parallel",
				Usage: "Disables building in parallel",
			},
			&cli.StringFlag{
				Name:  "only",
				Usage: "Only build the provided comma-separated paths, instead of all the paths in .lambgo.yml",
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

			if rawOnlyFlag := c.String("only"); rawOnlyFlag != "" {
				filteredBuildPaths, err := filterBuildPaths(config.BuildPaths, rawOnlyFlag)
				if err != nil {
					return err
				}

				config.BuildPaths = filteredBuildPaths
			}

			config.DisableParallelBuild = c.Bool("disable-parallel")
			return a.Builder.BuildBinaries(config)
		},
	}
}

func filterBuildPaths(buildPaths []string, filter string) ([]string, error) {
	buildPathsMap := make(map[string]bool, len(buildPaths))
	for _, buildPath := range buildPaths {
		buildPathsMap[buildPath] = true
	}

	allowedPaths := strings.Split(filter, ",")
	for _, allowedPath := range allowedPaths {
		if !buildPathsMap[allowedPath] {
			return nil, erk.WithParams(ErrCannotFilterBuildPaths, erk.Params{
				"allowedPath": allowedPath,
				"buildPaths":  buildPaths,
			})
		}
	}

	return allowedPaths, nil
}
