package cmd

import (
	"context"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/urfave/cli/v3"
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

		Action: a.runBuild,
	}
}

func (a *App) runBuild(ctx context.Context, cmd *cli.Command) error {
	pwd, err := a.Getwd()
	if err != nil {
		return err
	}

	config, err := a.LambgoFileLoader.LoadConfig(pwd)
	if err != nil {
		return err
	}

	if rawOnlyFlags := cmd.StringSlice("only"); len(rawOnlyFlags) > 0 {
		filteredLambdas, err := filterLambdas(config.Lambdas, rawOnlyFlags)
		if err != nil {
			return err
		}

		config.Lambdas = filteredLambdas
	}

	if cmd.Bool("disable-parallel") {
		config.NumParallel = 1
	} else {
		numParallel, err := parseNumParallel(config, cmd.String("num-parallel"))
		if err != nil {
			return err
		}

		config.NumParallel = numParallel
	}

	return a.Builder.BuildBinaries(config)
}

func filterLambdas(lambdas []*lambgofile.Lambda, filters []string) ([]*lambgofile.Lambda, error) {
	matched := make([]*lambgofile.Lambda, 0, len(lambdas))
	seenPaths := make(map[string]struct{}, len(lambdas))

	for _, filter := range filters {
		foundMatch := false
		filterIsDir := strings.HasSuffix(filter, "/")

		for _, lambda := range lambdas {
			isChild := filterIsDir && strings.HasPrefix(lambda.Path, filter)

			if lambda.Path == filter || isChild {
				foundMatch = true

				if _, seen := seenPaths[lambda.Path]; !seen {
					matched = append(matched, lambda)
					seenPaths[lambda.Path] = struct{}{}
				}
			}
		}

		if !foundMatch {
			return nil, erk.WithParams(ErrCannotFilterBuildPaths, erk.Params{
				"filter":     filter,
				"buildPaths": extractLambdaPaths(lambdas),
			})
		}
	}

	slices.SortFunc(matched, func(a, b *lambgofile.Lambda) int {
		return strings.Compare(a.Path, b.Path)
	})

	return matched, nil
}

func extractLambdaPaths(lambdas []*lambgofile.Lambda) []string {
	paths := make([]string, 0, len(lambdas))
	for _, lambda := range lambdas {
		paths = append(paths, lambda.Path)
	}

	return paths
}

func parseNumParallel(config *lambgofile.Config, numParallel string) (int, error) {
	if numParallel == allParallel {
		return len(config.Lambdas), nil
	}

	if prefix, matched := strings.CutSuffix(numParallel, cpuParallelSuffix); matched {
		cpuMultiplier, err := strconv.ParseFloat(prefix, 64)
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
