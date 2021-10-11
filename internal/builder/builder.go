package builder

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erg"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

type (
	ErkMultipleFailures struct{ erk.DefaultKind }
	ErkBuildError       struct{ erk.DefaultKind }
)

var (
	ErrMultipleBuildFailures = erk.New(ErkMultipleFailures{}, "Unable to build at least one Lambda")

	ErrGoBuildDependenciesFailed = erk.New(ErkBuildError{}, "Unable to build dependencies for all Lambdas with `go build`: {{.err}}")

	ErrGoBuildFailed = erk.New(ErkBuildError{}, "Unable to build '{{.buildPath}}' with `go build`: {{.err}}")
	ErrZipFailed     = erk.New(ErkBuildError{}, "Unable to zip '{{.buildPath}}' to '{{.buildPath}}.zip': {{.err}}")
)

type LambdaBuilderAPI interface {
	BuildBinaries(config *lambgofile.Config) error
}

type LambdaBuilder struct {
	Cmd    runcmd.RunnerAPI
	Zip    zipper.ZipAPI
	Logger *log.Logger
}

var _ LambdaBuilderAPI = &LambdaBuilder{}

// BuildBinaries defined in the config.
func (b *LambdaBuilder) BuildBinaries(config *lambgofile.Config) error {
	if config.OutDirectory == "" {
		config.OutDirectory = "tmp"
	}

	asyncParams := &buildBinaryAsyncParams{
		errors: erg.NewAs(ErrMultipleBuildFailures),
	}

	b.Logger.Println("Building Lambda Dependencies...")
	if err := b.buildDependencies(config); err != nil {
		return err
	}

	b.Logger.Println()
	b.Logger.Println("Building Lambdas:")
	for _, buildPath := range config.BuildPaths {
		if !config.DisableParallelBuild {
			asyncParams.wg.Add(1)
			go b.buildBinaryAsync(config, buildPath, asyncParams)
		} else {
			b.Logger.Printf(" - Building: '%s' -> '%s.zip'\n", buildPath, buildOutPath(config, buildPath))

			if err := b.buildBinary(config, buildPath); err != nil {
				asyncParams.errors = erg.Append(asyncParams.errors, err)
			}
		}
	}

	asyncParams.wg.Wait()
	if erg.Any(asyncParams.errors) {
		return asyncParams.errors
	}

	return nil
}

func (b *LambdaBuilder) buildDependencies(config *lambgofile.Config) error {
	// Skip building dependencies when there is only one Lambda, otherwise it will
	// build the executable instead of only populating the build cache
	if len(config.BuildPaths) < 2 { //nolint:gomnd
		return nil
	}

	buildPaths := make([]string, 0, len(config.BuildPaths))
	for _, buildPath := range config.BuildPaths {
		buildPaths = append(buildPaths, "./"+buildPath)
	}

	_, err := b.Cmd.Exec(&runcmd.ExecParams{
		PWD:  config.RootPath,
		CMD:  "go",
		Args: append([]string{"build", "-trimpath"}, buildPaths...),

		EnvVars: map[string]string{
			"GOOS":   "linux",
			"GOARCH": "amd64",
		},
	})
	if err != nil {
		return erk.WrapAs(ErrGoBuildDependenciesFailed, err)
	}

	return nil
}

type buildBinaryAsyncParams struct {
	wg       sync.WaitGroup
	errors   error
	errorsMu sync.Mutex
}

func (b *LambdaBuilder) buildBinaryAsync(config *lambgofile.Config, buildPath string, asyncParams *buildBinaryAsyncParams) {
	defer asyncParams.wg.Done()

	if err := b.buildBinary(config, buildPath); err != nil {
		asyncParams.errorsMu.Lock()
		defer asyncParams.errorsMu.Unlock()
		asyncParams.errors = erg.Append(asyncParams.errors, err)
		return
	}

	b.Logger.Printf(" - Built: '%s' -> '%s.zip'\n", buildPath, buildOutPath(config, buildPath))
}

func (b *LambdaBuilder) buildBinary(config *lambgofile.Config, buildPath string) error {
	outPath := buildOutPath(config, buildPath)

	fullArgs := []string{"build", "-trimpath", "-o", outPath}
	fullArgs = append(fullArgs, config.BuildFlags...)
	fullArgs = append(fullArgs, "./"+buildPath)

	_, err := b.Cmd.Exec(&runcmd.ExecParams{
		PWD:  config.RootPath,
		CMD:  "go",
		Args: fullArgs,

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

	zippedFileName := filepath.Base(outPath)
	if config.ZippedFileName != "" {
		zippedFileName = config.ZippedFileName
	}

	if err := b.Zip.ZipFile(outPath, zippedFileName); err != nil {
		return erk.WrapWith(ErrZipFailed, err, erk.Params{
			"buildPath": buildPath,
		})
	}

	return nil
}

func buildOutPath(config *lambgofile.Config, buildPath string) string {
	return filepath.Join(config.OutDirectory, buildPath)
}
