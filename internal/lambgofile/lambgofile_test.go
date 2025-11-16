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

	type mapFS map[string]any
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", nil),
				},
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", nil),
				},
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", nil),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-foo", "-bar", "baz qux"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-foo", "-bar", "baz qux"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", ""}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", ""}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", "value with spaces"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", "value with spaces"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", `value with "nested" quotes`}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", `value with "nested" quotes`}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", "value", "-other"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", "value", "-other"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags=-X main.version=1.0.0", "-tags=prod,dev"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags=-X main.version=1.0.0", "-tags=prod,dev"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag1", "double quotes", "-flag2", "single quotes"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag1", "double quotes", "-flag2", "single quotes"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.version=", "-tags"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.version=", "-tags"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.version=$VERSION", "-tags", "$BUILD_TAGS"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.version=$VERSION", "-tags", "$BUILD_TAGS"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", "value"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", "value"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag", "value with spaces"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag", "value with spaces"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-tags", "netgo,osusergo", "-ldflags=-s -w -X main.version=v1.2.3"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-tags", "netgo,osusergo", "-ldflags=-s -w -X main.version=v1.2.3"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.version=v1.2.3"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.version=v1.2.3"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.version=v2.0.0 -X main.commit=abc123", "-tags", "prod,netgo"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.version=v2.0.0 -X main.commit=abc123", "-tags", "prod,netgo"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.version=v3.0.0"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.version=v3.0.0"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-flag1=set", "-flag2=default2"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-flag1=set", "-flag2=default2"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-tags", "netgo", "-ldflags", "-X main.version=v1.0.0 -X main.literal=$LITERAL"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-tags", "netgo", "-ldflags", "-X main.version=v1.0.0 -X main.literal=$LITERAL"}),
				},
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
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", []string{"-ldflags", "-X main.set=set_value -X main.unset="}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags", "-X main.set=set_value -X main.unset="}),
				},
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", nil),
				},
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", nil),
				},
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
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/hello_world", nil),
					makeLambda("lambdas/goodbye_world", nil),
					makeLambda("functions/my_function", nil),
				},
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
			Name: "with valid config using lambdas field only",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", nil),
					makeLambda("lambdas/worker", nil),
				},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/api
  - path: lambdas/worker
`,
			}),
		},

		{
			Name: "with lambdas field having custom buildFlags",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", []string{"-tags", "prod"}),
					makeLambda("lambdas/worker", []string{"-ldflags=-s -w"}),
				},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildFlags: -ldflags="-s -w"
lambdas:
  - path: lambdas/api
    buildFlags: -tags prod
  - path: lambdas/worker
`,
			}),
		},

		{
			Name: "with lambdas field having empty buildFlags",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", nil)},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildFlags: -ldflags="-s -w"
lambdas:
  - path: lambdas/api
    buildFlags: ""
`,
			}),
		},

		{
			Name: "with both buildPaths and lambdas fields",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/legacy", nil),
					makeLambda("lambdas/api", nil),
				},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/legacy
lambdas:
  - path: lambdas/api
`,
			}),
		},

		{
			Name: "with lambdas field with environment variable expansion",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION": "v1.0.0",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", []string{"-ldflags", "-X main.version=v1.0.0"})},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/api
    buildFlags: -ldflags "-X main.version=$VERSION"
`,
			}),
		},

		{
			Name: "with lambdas field with complex buildFlags",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", []string{"-tags", "netgo,osusergo", "-ldflags=-s -w -X main.version=v1.2.3"})},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/api
    buildFlags: -tags netgo,osusergo -ldflags="-s -w -X main.version=v1.2.3"
`,
			}),
		},

		{
			Name: "with multiple lambdas with mixed buildFlags scenarios",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "tmp",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/api", []string{"-tags", "prod"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags=-s -w"}),
				},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildFlags: -ldflags="-s -w"
lambdas:
  - path: lambdas/api
    buildFlags: -tags prod
  - path: lambdas/worker
    buildFlags: ""
  - path: lambdas/simple
`,
			}),
		},

		{
			Name: "with complex config including all fields and comments",

			PWD: "/my/app",
			EnvVars: map[string]string{
				"VERSION":    "v2.0.0",
				"GIT_COMMIT": "abc123",
			},

			ExpectedConfig: &lambgofile.Config{
				RootPath:       "/my/app",
				ModulePath:     "github.com/my/app",
				OutDirectory:   "build",
				ZippedFileName: "bootstrap",
				Goos:           "linux",
				Goarch:         "arm64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/legacy", []string{"-ldflags=-s -w"}),
					makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w -X main.version=v2.0.0 -X main.commit=abc123"}),
					makeLambda("lambdas/worker", nil),
					makeLambda("lambdas/simple", []string{"-ldflags=-s -w"}),
				},
			},
			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `# Production configuration
# This is a comment
outDirectory: build

# Using bootstrap for provided.al2 runtime
zippedFileName: bootstrap

# Build flags for production
buildFlags: -ldflags="-s -w"

# Target ARM64 architecture
goos: linux
goarch: arm64

# Lambda functions to build
buildPaths:
  - lambdas/legacy
lambdas:
  - path: lambdas/api
    buildFlags: -tags prod -ldflags="-s -w -X main.version=$VERSION -X main.commit=$GIT_COMMIT"
  - path: lambdas/worker
    buildFlags: ""
  - path: lambdas/simple
`,
			}),
		}, {
			Name: "with various whitespace and formatting",

			PWD: "/my/app",

			ExpectedConfig: &lambgofile.Config{
				RootPath:     "/my/app",
				ModulePath:   "github.com/my/app",
				OutDirectory: "output",
				Goos:         "linux",
				Goarch:       "amd64",
				Lambdas: []*lambgofile.Lambda{
					makeLambda("lambdas/func1", nil),
					makeLambda("lambdas/func2", nil),
				},
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

		{
			Name: "when duplicate path between buildPaths and lambdas",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/duplicate
lambdas:
  - path: lambdas/duplicate
`,
			}),
		},

		{
			Name: "when duplicate path within lambdas array",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/duplicate
  - path: lambdas/duplicate
`,
			}),
		},

		{
			Name: "when duplicate path within buildPaths array",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/duplicate
  - lambdas/duplicate
`,
			}),
		},

		{
			Name: "when multiple duplicate paths",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/dup1
  - lambdas/dup1
  - lambdas/dup2
lambdas:
  - path: lambdas/dup2
`,
			}),
		},

		{
			Name: "when paths differ only by leading ./",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/hello_world
  - ./lambdas/hello_world
`,
			}),
		},

		{
			Name: "when paths differ only by trailing /",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - lambdas/hello_world
  - lambdas/hello_world/
`,
			}),
		},

		{
			Name: "when paths have multiple variations that normalize to same path",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - ./lambdas/api
  - lambdas/api/
lambdas:
  - path: lambdas/api
`,
			}),
		},

		{
			Name: "when buildPath is empty",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrEmptyLambdaPath,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
buildPaths:
  - ""
`,
			}),
		},

		{
			Name: "when lambda has empty path",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrEmptyLambdaPath,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: ""
    buildFlags: -tags prod
`,
			}),
		},

		{
			Name: "when lambdas section has paths with leading ./ that normalize to same path",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/api
    buildFlags: -tags prod
  - path: ./lambdas/api
    buildFlags: -tags dev
`,
			}),
		},

		{
			Name: "when lambdas section has paths with trailing / that normalize to same path",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrDuplicatePaths,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/worker
  - path: lambdas/worker/
    buildFlags: ""
`,
			}),
		},

		{
			Name: "when lambda has missing path field",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrEmptyLambdaPath,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - buildFlags: -tags prod
`,
			}),
		},

		{
			Name: "when per-lambda buildFlags has invalid syntax",

			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotParsePerLambdaFlags,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
				"my/app/.lambgo.yml": `
outDirectory: tmp
lambdas:
  - path: lambdas/api
    buildFlags: foo'
`,
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

func makeLambda(path string, buildFlags []string) *lambgofile.Lambda {
	return &lambgofile.Lambda{
		Path:       path,
		BuildFlags: buildFlags,
	}
}
