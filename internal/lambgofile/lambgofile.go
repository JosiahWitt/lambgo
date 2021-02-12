// Package lambgofile assists with loading and representing .lambgo.yml files.
package lambgofile

import (
	"errors"
	"path/filepath"
	"strings"

	"bursavich.dev/fs-shim/io/fs"
	"github.com/JosiahWitt/erk"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

// ExampleFile for use in error messages and CLI help menus.
const ExampleFile = `# Directory to use as the root for build artifacts.
# Optional, defaults to tmp.
outDirectory: tmp

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
	DisableParallelBuild bool   `yaml:"-"`
	RootPath             string `yaml:"-"`
	ModulePath           string `yaml:"-"`

	OutDirectory string   `yaml:"outDirectory"`
	BuildPaths   []string `yaml:"buildPaths"`
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

	config.RootPath = "/" + pwd
	config.ModulePath = modulePath
	return &config, nil
}
