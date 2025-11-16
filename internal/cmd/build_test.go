package cmd_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensuring"
	"github.com/JosiahWitt/lambgo/internal/cmd"
	"github.com/JosiahWitt/lambgo/internal/lambgofile"
	"github.com/JosiahWitt/lambgo/internal/mocks/mock_builder"
	"github.com/JosiahWitt/lambgo/internal/mocks/mock_lambgofile"
)

func TestBuild(t *testing.T) {
	ensure := ensure.New(t)

	type Mocks struct {
		LambgoFileLoader *mock_lambgofile.MockLoaderAPI
		Builder          *mock_builder.MockLambdaBuilderAPI
	}

	exampleError := errors.New("something went wrong")
	defaultWd := func() (string, error) {
		return "/test", nil
	}

	table := []struct {
		Name          string
		ExpectedError error
		Flags         []string

		Getwd      func() (string, error)
		Mocks      *Mocks
		SetupMocks func(*Mocks)
		Subject    *cmd.App
	}{
		{
			Name:  "with valid execution",
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 3,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:  "with valid execution: disable parallel generation",
			Flags: []string{"--disable-parallel"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 1,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: disable parallel generation when --num-parallel is provided",
			Flags: []string{"--disable-parallel", "--num-parallel=12"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 1,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:  "with valid execution: filter using --only flag",
			Flags: []string{"--only", "abc/123", "--only", "xyz/456"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("first/0", nil),
							makeLambda("abc/123", nil),
							makeLambda("xyz/456", nil),
							makeLambda("qwerty/789", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 2,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("abc/123", nil),
							makeLambda("xyz/456", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:  "with valid execution: filter using --only flag with directory filter",
			Flags: []string{"--only", "nested/", "--only", "xyz/456"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("first/0", nil),
							makeLambda("abc/123", nil),
							makeLambda("xyz/456", nil),
							makeLambda("qwerty/789", nil),
							makeLambda("nested/one", nil),
							makeLambda("nested/two", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 3,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("nested/one", nil),
							makeLambda("nested/two", nil),
							makeLambda("xyz/456", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:  "with valid execution: setting --num-parallel to all",
			Flags: []string{"--num-parallel=all"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 3,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: configuring --num-parallel to a static value",
			Flags: []string{"--num-parallel=2"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 2,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: configuring --num-parallel based on the number of CPUs",
			Flags: []string{"--num-parallel=1.5x"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: int(1.5 * float64(runtime.NumCPU())),
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:  "with valid execution: filter using --only flag with per-lambda buildFlags",
			Flags: []string{"--only", "lambdas/api"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod"}),
							makeLambda("lambdas/worker", []string{"-ldflags", "-s"}),
							makeLambda("lambdas/simple", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 1,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod"}),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: filter using --only flag with directory and mixed buildFlags",
			Flags: []string{"--only", "lambdas/"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod"}),
							makeLambda("lambdas/worker", []string{"-ldflags", "-s"}),
							makeLambda("lambdas/simple", nil),
							makeLambda("functions/other", []string{"-tags", "dev"}),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 3,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod"}),
							makeLambda("lambdas/simple", nil),
							makeLambda("lambdas/worker", []string{"-ldflags", "-s"}),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: filter multiple lambdas with different buildFlags",
			Flags: []string{"--only", "lambdas/api", "--only", "functions/other"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
							makeLambda("lambdas/worker", nil),
							makeLambda("functions/other", []string{"-tags", "dev"}),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 2,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("functions/other", []string{"-tags", "dev"}),
							makeLambda("lambdas/api", []string{"-tags", "prod", "-ldflags=-s -w"}),
						},
					}).
					Return(nil)
			},
		},
		{
			Name:  "with valid execution: filter preserves empty buildFlags override",
			Flags: []string{"--only", "lambdas/worker"},
			Getwd: defaultWd,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/api", []string{"-tags", "prod"}),
							makeLambda("lambdas/worker", nil),
						},
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						NumParallel: 1,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("lambdas/worker", nil),
						},
					}).
					Return(nil)
			},
		},

		{
			Name:          "when error loading working directory",
			Getwd:         func() (string, error) { return "", exampleError },
			ExpectedError: exampleError,
		},

		{
			Name:          "when cannot load config",
			Getwd:         defaultWd,
			ExpectedError: exampleError,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().LoadConfig("/test").Return(nil, exampleError)
			},
		},

		{
			Name:          "when cannot filter a build path with --only",
			Flags:         []string{"--only", "abc/123", "--only", "xyz"}, // xyz doesn't end in a /, thus it should not prefix match
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrCannotFilterBuildPaths,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						NumParallel: 2,
						RootPath:    "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("abc/123", nil),
							makeLambda("xyz/456", nil),
						},
					}, nil)
			},
		},

		{
			Name:          "when --num-parallel has a non-number value before x",
			Flags:         []string{"--num-parallel=nox"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel has a small number before x",
			Flags:         []string{"--num-parallel=0.0001x"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel has a negative number before x",
			Flags:         []string{"--num-parallel=-1x"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel is not a number",
			Flags:         []string{"--num-parallel=no"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel is a float",
			Flags:         []string{"--num-parallel=1.5"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel is zero",
			Flags:         []string{"--num-parallel=0"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},
		{
			Name:          "when --num-parallel is a negative number",
			Flags:         []string{"--num-parallel=-1"},
			Getwd:         defaultWd,
			ExpectedError: cmd.ErrInvalidNumParallel,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
						Lambdas: []*lambgofile.Lambda{
							makeLambda("path1", nil),
							makeLambda("path2", nil),
							makeLambda("path3", nil),
						},
					}, nil)
			},
		},

		{
			Name:          "when cannot generate mocks",
			Getwd:         defaultWd,
			ExpectedError: exampleError,
			SetupMocks: func(m *Mocks) {
				m.LambgoFileLoader.EXPECT().
					LoadConfig("/test").
					Return(&lambgofile.Config{
						RootPath: "/some/root/path",
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						RootPath: "/some/root/path",
					}).
					Return(exampleError)
			},
		},
	}

	ensure.RunTableByIndex(table, func(ensure ensuring.E, i int) {
		entry := table[i]
		entry.Subject.Getwd = entry.Getwd

		err := entry.Subject.Run(append([]string{"lambgo", "build"}, entry.Flags...))
		ensure(err).IsError(entry.ExpectedError)
	})
}

func makeLambda(path string, buildFlags []string) *lambgofile.Lambda {
	return &lambgofile.Lambda{
		Path:       path,
		BuildFlags: buildFlags,
	}
}
