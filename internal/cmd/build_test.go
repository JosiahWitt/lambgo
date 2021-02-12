package cmd_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
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
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						RootPath: "/some/root/path",
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
					}, nil)

				m.Builder.EXPECT().
					BuildBinaries(&lambgofile.Config{
						DisableParallelBuild: true,
						RootPath:             "/some/root/path",
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

	ensure.RunTableByIndex(table, func(ensure ensurepkg.Ensure, i int) {
		entry := table[i]
		entry.Subject.Getwd = entry.Getwd

		err := entry.Subject.Run(append([]string{"lambgo", "build"}, entry.Flags...))
		ensure(err).IsError(entry.ExpectedError)
	})
}
