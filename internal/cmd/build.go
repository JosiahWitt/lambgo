package cmd

import (
	"sort"
	"strings"

	"github.com/JosiahWitt/erk"
	"github.com/urfave/cli/v2"
)

type ErkCannotFilterBuildPaths struct{ erk.DefaultKind }

var ErrCannotFilterBuildPaths = erk.New(ErkCannotFilterBuildPaths{},
	"Cannot filter build paths with --only, since '{{.filter}}' does not match any of:\n{{.buildPaths}}"+
		"\n\nRemember to end with a trailing `/` if you wish to match a directory.",
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
			&cli.StringSliceFlag{
				Name: "only",
				Usage: "Only build the provided `path`, instead of all the paths in .lambgo.yml. " +
					"If you wish to build all Lambdas in a directory, you can provide a trailing `/`. " +
					"This flag can be used multiple times to build multiple Lambdas (or Lambda directories).",
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

			if rawOnlyFlags := c.StringSlice("only"); len(rawOnlyFlags) > 0 {
				filteredBuildPaths, err := filterBuildPaths(config.BuildPaths, rawOnlyFlags)
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

func filterBuildPaths(buildPaths []string, filters []string) ([]string, error) {
	filteredBuildPathsMap := map[string]bool{}

	for _, filter := range filters {
		foundMatch := false
		filterIsDir := strings.HasSuffix(filter, "/")

		for _, buildPath := range buildPaths {
			isChild := filterIsDir && strings.HasPrefix(buildPath, filter)

			if buildPath == filter || isChild {
				filteredBuildPathsMap[buildPath] = true
				foundMatch = true
			}
		}

		if !foundMatch {
			return nil, erk.WithParams(ErrCannotFilterBuildPaths, erk.Params{
				"filter":     filter,
				"buildPaths": buildPaths,
			})
		}
	}

	filteredBuildPaths := make([]string, 0, len(filteredBuildPathsMap))
	for buildPath := range filteredBuildPathsMap {
		filteredBuildPaths = append(filteredBuildPaths, buildPath)
	}

	sort.Strings(filteredBuildPaths)
	return filteredBuildPaths, nil
}
