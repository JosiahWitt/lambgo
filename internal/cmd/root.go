package cmd

import (
	"context"

	"github.com/JosiahWitt/lambgo/internal/builder"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/urfave/cli/v3"
)

// App is the CLI application for lambgo.
type App struct {
	Version string

	Getwd            func() (string, error)
	LambgoFileLoader lambgofile.LoaderAPI
	Builder          builder.LambdaBuilderAPI
}

// Run the application given the os.Args array.
func (a *App) Run(args []string) error {
	cliApp := &cli.Command{
		Name:    "lambgo",
		Usage:   "A simple framework for building AWS Lambdas in Go.",
		Version: a.Version,

		Commands: []*cli.Command{
			a.buildCmd(),
		},
	}

	return cliApp.Run(context.Background(), args)
}
