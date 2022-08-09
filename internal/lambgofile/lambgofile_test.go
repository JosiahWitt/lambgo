package lambgofile_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
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
		PWD  string

		ExpectedConfig *lambgofile.Config
		ExpectedError  error

		Mocks      *Mocks
		SetupMocks func(*Mocks)
		Subject    *lambgofile.Loader
	}{
		{
			Name: "with valid config in current directory",
			PWD:  "/my/app",
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
			PWD:  "/my/app/some/nested/pkg",
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
			PWD:  "/my/app",
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
			PWD:  "/my/app",
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
			Name: "with valid config including goos",
			PWD:  "/my/app",
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
			PWD:  "/my/app",
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
			Name:          "when missing go.mod file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotFindGoModule,

			SetupMocks: setupMapFS(mapFS{
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name:          "when cannot open go.mod file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      fs.ErrPermission,
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name:          "when cannot parse go.mod file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotParseGoModule,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      "something is broken",
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
			}),
		},

		{
			Name:          "when cannot find .lambgo.yml file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod": defaultGoModFile,
			}),
		},

		{
			Name:          "when cannot open .lambgo.yml file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotOpenFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": fs.ErrPermission,
			}),
		},

		{
			Name:          "when cannot parse .lambgo.yml file",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotUnmarshalFile,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": "{{{{{{ Not YAML",
			}),
		},

		{
			Name:          "when .lambgo.yml file has bad flag syntax",
			PWD:           "/my/app",
			ExpectedError: lambgofile.ErrCannotParseFlags,

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": "buildFlags: foo'",
			}),
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]

		config, err := entry.Subject.LoadConfig(entry.PWD)
		ensure(err).IsError(entry.ExpectedError)
		ensure(config).Equals(entry.ExpectedConfig)
	})
}
