package builder_test

import (
	"errors"
	"io/ioutil"
	"log"
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

	mockBuildDependencies := func(m *Mocks, lambdaPaths ...string) *gomock.Call {
		return m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
			PWD:  "/my/root",
			CMD:  "go",
			Args: append([]string{"build", "-trimpath"}, lambdaPaths...),

			EnvVars: map[string]string{
				"GOOS":   "linux",
				"GOARCH": "amd64",
			},
		}).Return("", nil)
	}

	table := []struct {
		Name          string
		Config        *lambgofile.Config
		ExpectedError error

		Mocks         *Mocks
		AssembleMocks func(*Mocks) []*gomock.Call
		Subject       *builder.LambdaBuilder
	}{
		{
			Name: "with valid config",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1", "path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path2", "path2").Return(nil),
				}
			},
		},

		{
			Name: "with valid config with default outDirectory",
			Config: &lambgofile.Config{
				RootPath:   "/my/root",
				Goos:       "linux",
				Goarch:     "amd64",
				BuildPaths: []string{"lambdas/path1", "lambdas/path2"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path1", "path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path2", "path2").Return(nil),
				}
			},
		},

		{
			Name: "with valid config with build flags",
			Config: &lambgofile.Config{
				RootPath:   "/my/root",
				BuildPaths: []string{"lambdas/path1", "lambdas/path2"},
				Goos:       "linux",
				Goarch:     "amd64",
				BuildFlags: []string{"-extra", "-stuff"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path1", "-extra", "-stuff", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path1", "path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "tmp/lambdas/path2", "-extra", "-stuff", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("tmp/lambdas/path2", "path2").Return(nil),
				}
			},
		},

		{
			Name: "with valid config with zippedFileName",
			Config: &lambgofile.Config{
				RootPath:       "/my/root",
				OutDirectory:   "out/dir",
				ZippedFileName: "bootstrap",
				Goos:           "linux",
				Goarch:         "amd64",
				BuildPaths:     []string{"lambdas/path1", "lambdas/path2"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1", "bootstrap").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path2", "bootstrap").Return(nil),
				}
			},
		},

		{
			Name: "with valid config with only one build path it does not build dependencies",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/path1"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1", "path1").Return(nil),
				}
			},
		},

		{
			Name: "when goos and goarch are configured",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "plan9",
				Goarch:       "arm64",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: append([]string{"build", "-trimpath"}, "./lambdas/path1", "./lambdas/path2"),

						EnvVars: map[string]string{
							"GOOS":   "plan9",
							"GOARCH": "arm64",
						},
					}).Return("", nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "plan9",
							"GOARCH": "arm64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1", "path1").Return(nil),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "plan9",
							"GOARCH": "arm64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path2", "path2").Return(nil),
				}
			},
		},

		{
			Name: "with error running go build for the dependencies",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},
			ExpectedError: builder.ErrGoBuildDependenciesFailed,

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "./lambdas/path1", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", errors.New("unable to build dependencies")),
				}
			},
		},

		{
			Name: "with error running go build for a Lambda",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},
			ExpectedError: builder.ErrGoBuildFailed,

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", errors.New("something is wrong 1")),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", errors.New("something is wrong 2")),
				}
			},
		},

		{
			Name: "with error zipping file",
			Config: &lambgofile.Config{
				RootPath:     "/my/root",
				OutDirectory: "out/dir",
				Goos:         "linux",
				Goarch:       "amd64",
				BuildPaths:   []string{"lambdas/path1", "lambdas/path2"},
			},
			ExpectedError: builder.ErrZipFailed,

			AssembleMocks: func(m *Mocks) []*gomock.Call {
				return []*gomock.Call{
					mockBuildDependencies(m, "./lambdas/path1", "./lambdas/path2"),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path1", "./lambdas/path1"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path1", "path1").Return(errors.New("something went wrong 1")),

					m.Cmd.EXPECT().Exec(&runcmd.ExecParams{
						PWD:  "/my/root",
						CMD:  "go",
						Args: []string{"build", "-trimpath", "-o", "out/dir/lambdas/path2", "./lambdas/path2"},

						EnvVars: map[string]string{
							"GOOS":   "linux",
							"GOARCH": "amd64",
						},
					}).Return("", nil),
					m.Zip.EXPECT().ZipFile("out/dir/lambdas/path2", "path2").Return(errors.New("something went wrong 2")),
				}
			},
		},
	}

	ensure.Run("when parallel mode disabled", func(ensure ensurepkg.Ensure) {
		ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
			entry := table[i]
			entry.Subject.Logger = log.New(ioutil.Discard, "", 0)
			entry.Config.DisableParallelBuild = true
			gomock.InOrder(entry.AssembleMocks(entry.Mocks)...)

			err := entry.Subject.BuildBinaries(entry.Config)
			ensure(err).IsError(err)
		})
	})

	ensure.Run("when parallel mode enabled", func(ensure ensurepkg.Ensure) {
		ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
			entry := table[i]
			entry.Subject.Logger = log.New(ioutil.Discard, "", 0)
			entry.Config.DisableParallelBuild = false
			entry.AssembleMocks(entry.Mocks)

			err := entry.Subject.BuildBinaries(entry.Config)
			ensure(err).IsError(err)
		})
	})
}
