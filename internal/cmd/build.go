package cmd

import (
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/urfave/cli/v2"
)

const (
	allParallel       = "all"
	cpuParallelSuffix = "x"
)

type ErkCannotFilterBuildPaths struct{ erk.DefaultKind }

type ErkInvalidNumParallel struct{ erk.DefaultKind }

var (
	ErrCannotFilterBuildPaths = erk.New(ErkCannotFilterBuildPaths{},
		"Cannot filter build paths with --only, since '{{.filter}}' does not match any of:\n{{.buildPaths}}"+
			"\n\nRemember to end with a trailing `/` if you wish to match a directory.",
	)

	ErrInvalidNumParallel = erk.New(ErkInvalidNumParallel{},
		"Invalid value ({{.numParallel}}) provided for --num-parallel. "+
			"Only `all`, `<int>`, or `<float>x` are supported, where the resulting number is a non-zero value.",
	)
)

func (a *App) buildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "build Lambdas",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "disable-parallel",
				Usage: "Disables building in parallel. Overrides --num-parallel to 1.",
			},
			&cli.StringSliceFlag{
				Name: "only",
				Usage: "Only build the provided `path`, instead of all the paths in .lambgo.yml. " +
					"If you wish to build all Lambdas in a directory, you can provide a trailing `/`. " +
					"This flag can be used multiple times to build multiple Lambdas (or Lambda directories).",
			},
			&cli.StringFlag{
				Name: "num-parallel",
				Usage: "Number of Lambdas to build in parallel. Defaults to `all`, which builds all Lambdas in parallel at the same time. " +
					"An `x` suffix multiplies the prefix by the number of CPUs. For example `1.5x` builds `1.5 * <number of CPUs>` Lambdas in parallel. " +
					"The result is truncated.",
				Value: allParallel,
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

			if c.Bool("disable-parallel") {
				config.NumParallel = 1
			} else {
				numParallel, err := parseNumParallel(config, c.String("num-parallel"))
				if err != nil {
					return err
				}

				config.NumParallel = numParallel
			}

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

func parseNumParallel(config *lambgofile.Config, numParallel string) (int, error) {
	if numParallel == "" {
		return 0, erk.WithParams(ErrInvalidNumParallel, erk.Params{"numParallel": numParallel})
	}

	if numParallel == allParallel {
		return len(config.BuildPaths), nil
	}

	if strings.HasSuffix(numParallel, cpuParallelSuffix) {
		prefix := strings.TrimSuffix(numParallel, cpuParallelSuffix)
		cpuMultiplier, err := strconv.ParseFloat(prefix, 64) //nolint:gomnd
		if err != nil {
			return 0, erk.WrapWith(ErrInvalidNumParallel, err, erk.Params{"numParallel": numParallel})
		}

		numCPUs := float64(runtime.NumCPU())
		parallel := int(cpuMultiplier * numCPUs)
		if parallel < 1 {
			return 0, erk.WithParams(ErrInvalidNumParallel, erk.Params{"numParallel": numParallel})
		}

		return parallel, nil
	}

	parallel, err := strconv.Atoi(numParallel)
	if err != nil {
		return 0, erk.WrapWith(ErrInvalidNumParallel, err, erk.Params{"numParallel": numParallel})
	}

	if parallel < 1 {
		return 0, erk.WithParams(ErrInvalidNumParallel, erk.Params{"numParallel": numParallel})
	}

	return parallel, nil
}
