package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JosiahWitt/lambgo/internal/builder"
	"github.com/JosiahWitt/lambgo/internal/cmd"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

// Version of the CLI.
// Should be tied to the release version.
//
//nolint:gochecknoglobals // Allows injecting the version
var Version = "0.1.16"

func main() {
	app := cmd.App{
		Version: Version,

		Getwd:            os.Getwd,
		LambgoFileLoader: &lambgofile.Loader{FS: os.DirFS("/")},
		Builder: &builder.LambdaBuilder{
			Cmd:    &runcmd.Runner{},
			Zip:    &zipper.Zip{},
			Logger: log.New(os.Stdout, "", 0),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err) //nolint:forbidigo // Allow printing error messages
		os.Exit(1)
	}
}
