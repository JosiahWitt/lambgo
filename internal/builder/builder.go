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

	sharedParams := &sharedBuilderParams{
		config: config,
		errors: erg.NewAs(ErrMultipleBuildFailures),
	}

	b.Logger.Println("Building Lambda Dependencies...")
	if err := b.buildDependencies(config); err != nil {
		return err
	}

	ch := make(chan *builderParams)
	for range config.NumParallel {
		go b.launchBuilder(ch)
	}

	b.Logger.Println()
	if len(config.Lambdas) == 1 {
		b.Logger.Println("Building 1 Lambda")
	} else if config.NumParallel == 1 {
		b.Logger.Printf("Building %d Lambdas one at a time:\n", len(config.Lambdas))
	} else if config.NumParallel == len(config.Lambdas) {
		b.Logger.Printf("Building %d Lambdas all at once:\n", len(config.Lambdas))
	} else {
		b.Logger.Printf("Building %d Lambdas in parallel groups of %d:\n", len(config.Lambdas), config.NumParallel)
	}

	for _, lambda := range config.Lambdas {
		sharedParams.wg.Add(1)
		ch <- &builderParams{lambda: lambda, sharedBuilderParams: sharedParams}
	}

	sharedParams.wg.Wait()
	close(ch)

	if erg.Any(sharedParams.errors) {
		return sharedParams.errors
	}

	return nil
}

func (b *LambdaBuilder) buildDependencies(config *lambgofile.Config) error {
	// Skip building dependencies when there is only one Lambda, otherwise it will
	// build the executable instead of only populating the build cache
	if len(config.Lambdas) < 2 { //nolint:mnd
		return nil
	}

	buildPaths := make([]string, 0, len(config.Lambdas))
	for _, lambda := range config.Lambdas {
		buildPaths = append(buildPaths, "./"+lambda.Path)
	}

	_, err := b.Cmd.Exec(&runcmd.ExecParams{
		PWD:  config.RootPath,
		CMD:  "go",
		Args: append([]string{"build", "-trimpath"}, buildPaths...),

		EnvVars: buildEnvVars(config),
	})
	if err != nil {
		return erk.WrapAs(ErrGoBuildDependenciesFailed, err)
	}

	return nil
}

type builderParams struct {
	*sharedBuilderParams

	lambda *lambgofile.Lambda
}

func (b *LambdaBuilder) launchBuilder(ch chan *builderParams) {
	for params := range ch {
		b.buildBinaryAsync(params)
	}
}

type sharedBuilderParams struct {
	config *lambgofile.Config

	wg       sync.WaitGroup
	errors   error
	errorsMu sync.Mutex
}

func (b *LambdaBuilder) buildBinaryAsync(params *builderParams) {
	defer params.wg.Done()

	if err := b.buildBinary(params.config, params.lambda); err != nil {
		params.errorsMu.Lock()
		defer params.errorsMu.Unlock()
		params.errors = erg.Append(params.errors, err)
		return
	}

	b.Logger.Printf(" - Built: '%s' -> '%s.zip'\n", params.lambda.Path, buildOutPath(params.config, params.lambda.Path))
}

func (b *LambdaBuilder) buildBinary(config *lambgofile.Config, lambda *lambgofile.Lambda) error {
	outPath := buildOutPath(config, lambda.Path)

	fullArgs := []string{"build", "-trimpath", "-o", outPath}
	fullArgs = append(fullArgs, lambda.BuildFlags...)
	fullArgs = append(fullArgs, "./"+lambda.Path)

	_, err := b.Cmd.Exec(&runcmd.ExecParams{
		PWD:  config.RootPath,
		CMD:  "go",
		Args: fullArgs,

		EnvVars: buildEnvVars(config),
	})
	if err != nil {
		return erk.WrapWith(ErrGoBuildFailed, err, erk.Params{
			"buildPath": lambda.Path,
		})
	}

	zippedFileName := filepath.Base(outPath)
	if config.ZippedFileName != "" {
		zippedFileName = config.ZippedFileName
	}

	if err := b.Zip.ZipFile(outPath, zippedFileName); err != nil {
		return erk.WrapWith(ErrZipFailed, err, erk.Params{
			"buildPath": lambda.Path,
		})
	}

	return nil
}

func buildEnvVars(config *lambgofile.Config) map[string]string {
	return map[string]string{
		"GOOS":   config.Goos,
		"GOARCH": config.Goarch,
	}
}

func buildOutPath(config *lambgofile.Config, buildPath string) string {
	return filepath.Join(config.OutDirectory, buildPath)
}
