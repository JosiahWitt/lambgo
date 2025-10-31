package lambgofile_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensuring"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/mocks/io/mock_fs"
	"github.com/golang/mock/gomock"
)

func TestLoadConfig(t *testing.T) {
	ensure := ensure.New(t)

	const defaultGoModFile = "module github.com/my/app"

	type Mocks struct {
		FS *mock_fs.MockReadFileFS
	}

	type mapFS map[string]interface{}
	setupMapFS := func(mapFS mapFS) func(*Mocks) {
		return func(m *Mocks) {
			m.FS.EXPECT().ReadFile(gomock.Any()).AnyTimes().
				DoAndReturn(func(name string) ([]byte, error) {
					rawData, ok := mapFS[name]
					if !ok {
						return nil, fs.ErrNotExist
					}

					switch data := rawData.(type) {
					case string:
						return []byte(data), nil
					case error:
						return nil, data
					default:
						return nil, errors.New("unknown type")
					}
				})
		}
	}

	table := []struct {
		Name string

		PWD     string
		EnvVars map[string]string

		ExpectedConfig *lambgofile.Config
		ExpectedError  error

		Mocks      *Mocks
		SetupMocks func(*Mocks)
		Subject    *lambgofile.Loader
	}{
		{
			Name: "with valid config in current directory",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with valid config in parent directory",

			PWD: "/my/app/some/nested/pkg",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with valid config including zippedFileName",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:       "/my/app",
				ModulePath:     "github.com/my/app",
				ZippedFileName: "some-name",
				OutDirectory:   "tmp",
				Goos:           "linux",
				Goarch:         "amd64",
				BuildPaths:     []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
zippedFileName: some-name
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with valid config including buildFlags",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-foo -bar "baz qux"`,
				BuildFlags:    []string{"-foo", "-bar", "baz qux"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -foo -bar "baz qux"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing empty quoted strings",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag ""`,
				BuildFlags:    []string{"-flag", ""},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag ""
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing single quotes",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag 'value with spaces'`,
				BuildFlags:    []string{"-flag", "value with spaces"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag 'value with spaces'
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing escaped quotes",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag "value with \"nested\" quotes"`,
				BuildFlags:    []string{"-flag", `value with "nested" quotes`},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag "value with \"nested\" quotes"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing multiple spaces",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag    value    -other`,
				BuildFlags:    []string{"-flag", "value", "-other"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag    value    -other
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing special characters",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags="-X main.version=1.0.0" -tags=prod,dev`,
				BuildFlags:    []string{"-ldflags=-X main.version=1.0.0", "-tags=prod,dev"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags="-X main.version=1.0.0" -tags=prod,dev
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing mixed quote styles",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag1 "double quotes" -flag2 'single quotes'`,
				BuildFlags:    []string{"-flag1", "double quotes", "-flag2", "single quotes"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag1 "double quotes" -flag2 'single quotes'
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing undefined environment variables",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.version=$UNDEFINED_VAR" -tags $ALSO_UNDEFINED`,
				BuildFlags:    []string{"-ldflags", "-X main.version=", "-tags"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.version=$UNDEFINED_VAR" -tags $ALSO_UNDEFINED
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing escaped dollar signs",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.version=\$VERSION" -tags \$BUILD_TAGS`,
				BuildFlags:    []string{"-ldflags", "-X main.version=$VERSION", "-tags", "$BUILD_TAGS"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.version=\$VERSION" -tags \$BUILD_TAGS
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing trailing and leading whitespace",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag value`,
				BuildFlags:    []string{"-flag", "value"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags:   -flag value
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags containing backslash escapes",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag value\ with\ spaces`,
				BuildFlags:    []string{"-flag", "value with spaces"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag value\ with\ spaces
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with complex real-world buildFlags",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-tags netgo,osusergo -ldflags="-s -w -X main.version=v1.2.3"`,
				BuildFlags:    []string{"-tags", "netgo,osusergo", "-ldflags=-s -w -X main.version=v1.2.3"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -tags netgo,osusergo -ldflags="-s -w -X main.version=v1.2.3"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags using environment variable expansion",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION": "v1.2.3",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.version=$VERSION"`,
				BuildFlags:    []string{"-ldflags", "-X main.version=v1.2.3"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.version=$VERSION"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags using multiple environment variables",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION":    "v2.0.0",
				"GIT_COMMIT": "abc123",
				"BUILD_TAGS": "prod,netgo",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.version=$VERSION -X main.commit=$GIT_COMMIT" -tags $BUILD_TAGS`,
				BuildFlags:    []string{"-ldflags", "-X main.version=v2.0.0 -X main.commit=abc123", "-tags", "prod,netgo"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.version=$VERSION -X main.commit=$GIT_COMMIT" -tags $BUILD_TAGS
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags using braced variable syntax",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION": "v3.0.0",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.version=${VERSION}"`,
				BuildFlags:    []string{"-ldflags", "-X main.version=v3.0.0"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.version=${VERSION}"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags using braced variable syntax with fallbacks",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VAR1": "set",
				"VAR2": "",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-flag1=${VAR1:-default1} -flag2=${VAR2:-default2}`,
				BuildFlags:    []string{"-flag1=set", "-flag2=default2"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -flag1=${VAR1:-default1} -flag2=${VAR2:-default2}
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags mixing literal and expanded variables",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION": "v1.0.0",
				"LITERAL": "should not expand",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-tags netgo -ldflags "-X main.version=$VERSION -X main.literal=\$LITERAL"`,
				BuildFlags:    []string{"-tags", "netgo", "-ldflags", "-X main.version=v1.0.0 -X main.literal=$LITERAL"},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -tags netgo -ldflags "-X main.version=$VERSION -X main.literal=\$LITERAL"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with buildFlags using variables with default values when unset",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"SET_VAR": "set_value",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:      "/my/app",
				ModulePath:    "github.com/my/app",
				RawBuildFlags: `-ldflags "-X main.set=$SET_VAR -X main.unset=$UNSET_VAR"`,
				BuildFlags:    []string{"-ldflags", "-X main.set=set_value -X main.unset="},
				OutDirectory:  "tmp",
				Goos:          "linux",
				Goarch:        "amd64",
				BuildPaths:    []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
buildFlags: -ldflags "-X main.set=$SET_VAR -X main.unset=$UNSET_VAR"
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with valid config including goos",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "plan9",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
goos: plan9
` + lambgofile.ExampleFile,
			}),
		},
		{
			Name: "with valid config including goarch",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "arm64",
				BuildPaths:   []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
goarch: arm64
` + lambgofile.ExampleFile,
			}),
		},

		{
			Name: "with multiple buildPaths",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/hello_world", "lambdas/goodbye_world", "functions/my_function"},
			},
			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/hello_world
  - lambdas/goodbye_world
  - functions/my_function
`,
			}),
		},

		{
			Name: "with complex config including all fields and comments",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:       "/my/app",
				ModulePath:     "github.com/my/app",
				OutDirectory:   "build",
				ZippedFileName: "bootstrap",
				RawBuildFlags:  `-tags prod -ldflags="-s -w"`,
				BuildFlags:     []string{"-tags", "prod", "-ldflags=-s -w"},
				Goos:           "linux",
				Goarch:         "arm64",
				BuildPaths:     []string{"lambdas/api", "lambdas/worker"},
			},
			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `# Production configuration
# This is a comment
outDirectory: build

# Using bootstrap for provided.al2 runtime
zippedFileName: bootstrap

# Build flags for production
buildFlags: -tags prod -ldflags="-s -w"

# Target ARM64 architecture
goos: linux
goarch: arm64

# Lambda functions to build
buildPaths:
  - lambdas/api
  - lambdas/worker
`,
			}),
		},

		{
			Name: "with various whitespace and formatting",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "output",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/func1", "lambdas/func2"},
			},
			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory:    output


buildPaths:

  -    lambdas/func1
  -  lambdas/func2


`,
			}),
		},

		{
			Name: "when missing go.mod file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotFindGoModule,

			SetupMocks: setupMapFS(mapFS{
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name: "when cannot open go.mod file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      fs.ErrPermission,
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name: "when cannot parse go.mod file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotParseGoModule,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      "something is broken",
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name: "when cannot find .lambgo.yml file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
			}),
		},

		{
			Name: "when cannot open .lambgo.yml file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": fs.ErrPermission,
			}),
		},

		{
			Name: "when cannot parse .lambgo.yml file",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotUnmarshalFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": "{{{{{{ Not YAML",
			}),
		},

		{
			Name: "when .lambgo.yml file has bad flag syntax",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotParseFlags,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": "buildFlags: foo'",
			}),
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensuring.E, i int) {
		entry := table[i]

		for key, value := range entry.EnvVars {
			ensure.T().Setenv(key, value)
		}

		config, err := entry.Subject.LoadConfig(entry.PWD)
		ensure(err).IsError(entry.ExpectedError)
		ensure(config).Equals(entry.ExpectedConfig)
	})
}
