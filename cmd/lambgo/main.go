package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/JosiahWitt/lambgo/internal/builder"
	"github.com/JosiahWitt/lambgo/internal/cmd"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

func main() {
	app := cmd.App{
		Version: getVersion(),

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

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}

	return strings.TrimPrefix(info.Main.Version, "v")
}
