package lambgofile_test

import (
	"errors"
	"testing"

	"bursavich.dev/fs-shim/io/fs"
	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/mocks/bursavich.dev/fs-shim/io/mock_fs"
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
				BuildPaths:   []string{"lambdas/hello_world"},
			},

			SetupMocks: setupMapFS(mapFS{
				"my/app/go.mod":      defaultGoModFile,
				"my/app/.lambgo.yml": lambgofile.ExampleFile,
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
	}

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]

		config, err := entry.Subject.LoadConfig(entry.PWD)
		ensure(err).IsError(entry.ExpectedError)
		ensure(config).Equals(entry.ExpectedConfig)
	})
}
