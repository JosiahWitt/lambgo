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

# Additional build flags passed to "go build"
# For example, if you want to provide extra compiler or linker options
# Supports environment variable expansion: $VAR or ${VAR}
# buildFlags: -tags extra,tags -ldflags="-linker -flags"

# Allow overriding the GOOS and GOARCH environment variables to
# cross compile for a different operating system or architecture.
# Optional, defaults to GOOS=linux and GOARCH=amd64.
# goos: linux
# goarch: amd64

# Paths to build into Lambda zip files.
# Each path should contain a main package.
# The artifacts are built to: <outDirectory>/<buildPath>.zip
buildPaths:
  - lambdas/hello_world
`

const (
	gomodFileName  = "go.mod"
	configFileName = ".lambgo.yml"
)

type ErkCannotLoadConfig struct{ erk.DefaultKind }

var (
	ErrCannotFindGoModule  = erk.New(ErkCannotLoadConfig{}, "Cannot find root go.mod file by searching parent working directories")
	ErrCannotParseGoModule = erk.New(ErkCannotLoadConfig{}, "Cannot read module path from go.mod file: {{.path}}")
	ErrCannotOpenFile      = erk.New(ErkCannotLoadConfig{}, "Cannot open the file '{{.path}}': {{.err}}")
	ErrCannotUnmarshalFile = erk.New(ErkCannotLoadConfig{}, "Cannot parse the file '{{.path}}': {{.err}}")
	ErrCannotParseFlags    = erk.New(ErkCannotLoadConfig{}, "Cannot parse build flags '{{.flags}}': {{.err}}")
)

type LoaderAPI interface {
	LoadConfig(pwd string) (*Config, error)
}

// Loader allows loading the project's .lambgo.yml file.
type Loader struct {
	FS fs.FS
}

var _ LoaderAPI = &Loader{}

// Config is the root of the .lambgo.yml file.
type Config struct {
	NumParallel int    `yaml:"-"`
	RootPath    string `yaml:"-"`
	ModulePath  string `yaml:"-"`

	OutDirectory   string   `yaml:"outDirectory"`
	ZippedFileName string   `yaml:"zippedFileName"`
	RawBuildFlags  string   `yaml:"buildFlags"`
	Goos           string   `yaml:"goos"`
	Goarch         string   `yaml:"goarch"`
	BuildFlags     []string `yaml:"-"`
	BuildPaths     []string `yaml:"buildPaths"`
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

	configFilePath := filepath.Join(pwd, configFileName)
	configFileData, err := fs.ReadFile(l.FS, configFilePath)
	if err != nil {
		return nil, erk.WrapWith(ErrCannotOpenFile, err, erk.Params{
			"path": configFilePath,
		})
	}

	config := Config{}
	if err := yaml.Unmarshal(configFileData, &config); err != nil {
		return nil, erk.WrapWith(ErrCannotUnmarshalFile, err, erk.Params{
			"path": configFilePath,
		})
	}

	if config.RawBuildFlags != "" {
		buildFlags, err := shell.Fields(config.RawBuildFlags, os.Getenv)
		if err != nil {
			return nil, erk.WrapWith(ErrCannotParseFlags, err, erk.Params{
				"flags": config.RawBuildFlags,
			})
		}

		config.BuildFlags = buildFlags
	}

	config.setDefaults()

	config.RootPath = "/" + pwd
	config.ModulePath = modulePath
	return &config, nil
}

func (config *Config) setDefaults() {
	if config.Goos == "" {
		config.Goos = "linux"
	}

	if config.Goarch == "" {
		config.Goarch = "amd64"
	}
}
