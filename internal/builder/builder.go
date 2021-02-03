package builder

import (
	"fmt"
	"path/filepath"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

type ErkBuildError struct{ erk.DefaultKind }

var (
	ErrGoBuildFailed = erk.New(ErkBuildError{}, "Unable to build '{{.buildPath}}' with `go build`: {{.err}}")
	ErrZipFailed     = erk.New(ErkBuildError{}, "Unable to zip '{{.buildPath}}' to '{{.buildPath}}.zip': {{.err}}")
)

type LambdaBuilderAPI interface {
	BuildBinaries(config *lambgofile.Config) error
}

type LambdaBuilder struct {
	Cmd runcmd.RunnerAPI
	Zip zipper.ZipAPI
}

var _ LambdaBuilderAPI = &LambdaBuilder{}

// BuildBinaries defined in the config.
func (b *LambdaBuilder) BuildBinaries(config *lambgofile.Config) error {
	if config.OutDirectory == "" {
		config.OutDirectory = "tmp"
	}

	fmt.Println("Building:") //nolint:forbidigo
	for _, buildPath := range config.BuildPaths {
		if err := b.buildBinary(config, buildPath); err != nil {
			return err // TODO: Group errors
		}
	}

	return nil
}

func (b *LambdaBuilder) buildBinary(config *lambgofile.Config, buildPath string) error {
	outPath := filepath.Join(config.OutDirectory, buildPath)
	fmt.Printf(" - '%s' -> '%s.zip'\n", buildPath, outPath) //nolint:forbidigo

	_, err := b.Cmd.Exec(&runcmd.ExecParams{
		PWD:  config.RootPath,
		CMD:  "go",
		Args: []string{"build", "-trimpath", "-o", outPath, "./" + buildPath},

		EnvVars: map[string]string{
			"GOOS":   "linux",
			"GOARCH": "amd64",
		},
	})
	if err != nil {
		return erk.WrapWith(ErrGoBuildFailed, err, erk.Params{
			"buildPath": buildPath,
		})
	}

	if err := b.Zip.ZipFile(outPath); err != nil {
		return erk.WrapWith(ErrZipFailed, err, erk.Params{
			"buildPath": buildPath,
		})
	}

	return nil
}
