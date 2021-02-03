package builder_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/lambgo/internal/builder"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/mocks/mock_runcmd"
	"github.com/JosiahWitt/lambgo/internal/mocks/mock_zipper"
	"github.com/JosiahWitt/lambgo/internal/runcmd"
	"github.com/golang/mock/gomock"
)

func TestBuildBinaries(t *testing.T) {
	ensure := ensure.New(t)

	type Mocks struct {
		Cmd *mock_runcmd.MockRunnerAPI
		Zip *mock_zipper.MockZipAPI
	}

	table := []struct {
		Name          string
		Config        *lambgofile.Config
		ExpectedError error

		Mocks      *Mocks
		SetupMocks func(*Mocks)
		Subject    *builder.LambdaBuilder
	}{
		{
			Name: "with valid config",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},

			SetupMocks: func(m *Mocks) {
				gomock.InOrder(
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path2").Return(nil),
				)
			},
		},

		{
			Name: "with valid config with default outDirectory",
			Config: &lambgofile.Config{
				RootPath:   "/my/root",
				BuildPaths: []string{"lambdas/path1", "lambdas/path2"},
			},

			SetupMocks: func(m *Mocks) {
				gomock.InOrder(
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path2").Return(nil),
				)
			},
		},

		{
			Name: "with error running go build",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},
			ExpectedError: builder.ErrGoBuildFailed,

			SetupMocks: func(m *Mocks) {
				gomock.InOrder(
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", errors.New("something is wrong")),
				)
			},
		},

		{
			Name: "with error zipping file",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},
			ExpectedError: builder.ErrZipFailed,

			SetupMocks: func(m *Mocks) {
				gomock.InOrder(
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1").Return(errors.New("something went wrong")),
				)
			},
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]

		err := entry.Subject.BuildBinaries(entry.Config)
		ensure(err).IsError(entry.ExpectedError)
	})
}
