// Package lambgofile assists with loading and representing .lambgo.yml files.
package lambgofile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/JosiahWitt/erk"
	"github.com/goccy/go-yaml"
	"golang.org/x/mod/modfile"
	"mvdan.cc/sh/v3/shell"
)

// ExampleFile for use in error messages and CLI help menus.
const ExampleFile = `# Directory to use as the root for build artifacts.
# Optional, defaults to tmp.
outDirectory: tmp

# File name to use for all zipped binaries.
# Useful when using provided.al2 instead of go1.x for the Lambda runtime.
# Optional, defaults to the name of the Lambda's directory.
# zippedFileName: bootstrap

# Additional build flags passed to "go build".
# For example, if you want to provide extra compiler or linker options.
# Supports environment variable expansion: $VAR or ${VAR}.
# This serves as the default for all lambdas unless overridden per-lambda.
# buildFlags: -tags extra,tags -ldflags="-linker -flags"

# Allow overriding the GOOS and GOARCH environment variables to
# cross compile for a different operating system or architecture.
# Optional, defaults to GOOS=linux and GOARCH=amd64.
# goos: linux
# goarch: amd64

# Option 1: Simple paths
# Paths to build into Lambda zip files.
# Each path should contain a main package.
# The artifacts are built to: <outDirectory>/<buildPath>.zip
buildPaths:
  - lambdas/hello_world

# Option 2: Per-lambda configuration with custom build flags.
lambdas:
  - path: lambdas/api
    buildFlags: -tags prod -ldflags="-s -w"
  - path: lambdas/worker
    buildFlags: "" # Forces no flags, even if some are defined on the top-level option
  - path: lambdas/simple
    # Inherits top-level buildFlags if not specified

# Both buildPaths and lambdas can be used together, but duplicate paths will result in an error.
`

const (
	gomodFileName  = "go.mod"
	configFileName = ".lambgo.yml"
)

type ErkCannotLoadConfig struct{ erk.DefaultKind }

var (
	ErrCannotFindGoModule        = erk.New(ErkCannotLoadConfig{}, "Cannot find root go.mod file by searching parent working directories")
	ErrCannotParseGoModule       = erk.New(ErkCannotLoadConfig{}, "Cannot read module path from go.mod file: {{.path}}")
	ErrCannotOpenFile            = erk.New(ErkCannotLoadConfig{}, "Cannot open the file '{{.path}}': {{.err}}")
	ErrCannotUnmarshalFile       = erk.New(ErkCannotLoadConfig{}, "Cannot parse the file '{{.path}}': {{.err}}")
	ErrCannotParseFlags          = erk.New(ErkCannotLoadConfig{}, "Cannot parse build flags '{{.flags}}': {{.err}}")
	ErrCannotParsePerLambdaFlags = erk.New(ErkCannotLoadConfig{}, "Cannot parse build flags for lambda '{{.path}}' with flags '{{.flags}}': {{.err}}")
	ErrDuplicatePaths            = erk.New(ErkCannotLoadConfig{}, "Duplicate lambda paths found: {{.paths}}")
	ErrEmptyLambdaPath           = erk.New(ErkCannotLoadConfig{}, "Lambda has an empty path")
)

type LoaderAPI interface {
	LoadConfig(pwd string) (*Config, error)
}

// Loader allows loading the project's .lambgo.yml file.
type Loader struct {
	FS fs.FS
}

var _ LoaderAPI = &Loader{}

// rawConfig is the internal struct used for unmarshaling from .lambgo.yml.
// It contains all the YAML tags and raw fields that need processing.
type rawConfig struct {
	OutDirectory   string       `yaml:"outDirectory"`
	ZippedFileName string       `yaml:"zippedFileName"`
	RawBuildFlags  string       `yaml:"buildFlags"`
	Goos           string       `yaml:"goos"`
	Goarch         string       `yaml:"goarch"`
	BuildPaths     []string     `yaml:"buildPaths"`
	RawLambdas     []*rawLambda `yaml:"lambdas"`
}

// rawLambda is the internal struct used for unmarshaling lambda configurations.
type rawLambda struct {
	Path          string  `yaml:"path"`
	RawBuildFlags *string `yaml:"buildFlags,omitempty"`
}

// Config is the root configuration after processing .lambgo.yml.
type Config struct {
	NumParallel    int
	RootPath       string
	ModulePath     string
	OutDirectory   string
	ZippedFileName string
	Goos           string
	Goarch         string
	Lambdas        []*Lambda
}

// Lambda represents a single lambda function with its build configuration.
type Lambda struct {
	Path       string
	BuildFlags []string
}

// LoadConfig from the .lambgo.yml file that is located in pwd or a parent of pwd.
func (l *Loader) LoadConfig(pwd string) (*Config, error) {
	pwd = strings.TrimPrefix(pwd, "/")
	gomodFilePath := filepath.Join(pwd, gomodFileName)

	gomodFileData, err := fs.ReadFile(l.FS, gomodFilePath)
	if errors.Is(err, fs.ErrNotExist) {
		newPWD := filepath.Dir(pwd)
		if pwd == newPWD {
			return nil, ErrCannotFindGoModule
		}

		return l.LoadConfig(newPWD)
	}

	if err != nil {
		return nil, erk.WrapWith(ErrCannotOpenFile, err, erk.Params{
			"path": gomodFilePath,
		})
	}

	modulePath := modfile.ModulePath(gomodFileData)
	if modulePath == "" {
		return nil, erk.WrapWith(ErrCannotParseGoModule, err, erk.Params{
			"path": gomodFilePath,
		})
	}

	config, err := l.parseConfigFile(pwd, modulePath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (l *Loader) parseConfigFile(pwd, modulePath string) (*Config, error) {
	configFilePath := filepath.Join(pwd, configFileName)
	configFileData, err := fs.ReadFile(l.FS, configFilePath)
	if err != nil {
		return nil, erk.WrapWith(ErrCannotOpenFile, err, erk.Params{
			"path": configFilePath,
		})
	}

	rawCfg := rawConfig{}
	if err := yaml.Unmarshal(configFileData, &rawCfg); err != nil {
		return nil, erk.WrapWith(ErrCannotUnmarshalFile, err, erk.Params{
			"path": configFilePath,
		})
	}

	buildFlags, err := parseBuildFlags(rawCfg.RawBuildFlags)
	if err != nil {
		return nil, erk.WrapWith(ErrCannotParseFlags, err, erk.Params{
			"flags": rawCfg.RawBuildFlags,
		})
	}

	lambdas, err := rawCfg.mergeLambdas(buildFlags)
	if err != nil {
		return nil, err
	}

	config := &Config{
		RootPath:       "/" + pwd,
		ModulePath:     modulePath,
		OutDirectory:   rawCfg.OutDirectory,
		ZippedFileName: rawCfg.ZippedFileName,
		Goos:           rawCfg.Goos,
		Goarch:         rawCfg.Goarch,
		Lambdas:        lambdas,
	}

	config.setDefaults()
	return config, nil
}

func (raw *rawConfig) mergeLambdas(defaultBuildFlags []string) ([]*Lambda, error) {
	lambdas := make([]*Lambda, 0, len(raw.BuildPaths)+len(raw.RawLambdas))

	for _, buildPath := range raw.BuildPaths {
		lambda, err := transformBuildPathToLambda(buildPath, defaultBuildFlags)
		if err != nil {
			return nil, err
		}

		lambdas = append(lambdas, lambda)
	}

	for _, rawLambda := range raw.RawLambdas {
		lambda, err := rawLambda.transform(defaultBuildFlags)
		if err != nil {
			return nil, err
		}

		lambdas = append(lambdas, lambda)
	}

	seenPaths := make(map[string]struct{})
	var duplicates []string
	for _, lambda := range lambdas {
		if _, exists := seenPaths[lambda.Path]; exists {
			duplicates = append(duplicates, lambda.Path)
		}
		seenPaths[lambda.Path] = struct{}{}
	}

	if len(duplicates) > 0 {
		return nil, erk.WithParams(ErrDuplicatePaths, erk.Params{
			"paths": strings.Join(duplicates, ", "),
		})
	}

	return lambdas, nil
}

func transformBuildPathToLambda(buildPath string, defaultBuildFlags []string) (*Lambda, error) {
	if buildPath == "" {
		return nil, ErrEmptyLambdaPath
	}

	normalizedPath := filepath.Clean(buildPath)
	return &Lambda{
		Path:       normalizedPath,
		BuildFlags: defaultBuildFlags,
	}, nil
}

func (rawLambda *rawLambda) transform(defaultBuildFlags []string) (*Lambda, error) {
	if rawLambda.Path == "" {
		return nil, ErrEmptyLambdaPath
	}

	normalizedPath := filepath.Clean(rawLambda.Path)
	lambda := &Lambda{Path: normalizedPath}

	if rawLambda.RawBuildFlags == nil {
		lambda.BuildFlags = defaultBuildFlags
	} else {
		buildFlags, err := parseBuildFlags(*rawLambda.RawBuildFlags)
		if err != nil {
			return nil, erk.WrapWith(ErrCannotParsePerLambdaFlags, err, erk.Params{
				"path":  rawLambda.Path,
				"flags": *rawLambda.RawBuildFlags,
			})
		}

		lambda.BuildFlags = buildFlags
	}

	return lambda, nil
}

func parseBuildFlags(rawFlags string) ([]string, error) {
	if rawFlags == "" {
		return nil, nil
	}

	buildFlags, err := shell.Fields(rawFlags, os.Getenv)
	if err != nil {
		return nil, err
	}

	return buildFlags, nil
}

func (config *Config) setDefaults() {
	if config.Goos == "" {
		config.Goos = "linux"
	}

	if config.Goarch == "" {
		config.Goarch = "amd64"
	}
}
