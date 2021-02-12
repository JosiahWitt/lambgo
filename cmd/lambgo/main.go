package main

import (
	"fmt"
	"log"
	"os"

	"bursavich.dev/fs-shim/io/fs"
	"github.com/JosiahWitt/lambgo/internal/builder"
	"github.com/JosiahWitt/lambgo/internal/cmd"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

//nolint:gochecknoglobals // Allows injecting the version
// Version of the CLI.
// Should be tied to the release version.
var Version = "0.1.3"

func main() {
	app := cmd.App{
		Version: Version,

		Getwd:            os.Getwd,
		LambgoFileLoader: &lambgofile.Loader{FS: fs.DirFS("")},
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
