package runcmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ExecParams struct {
	PWD  string
	CMD  string
	Args []string

	EnvVars map[string]string
}

type RunnerAPI interface {
	Exec(params *ExecParams) (string, error)
}

type Runner struct{}

var _ RunnerAPI = &Runner{}

// Exec the command defined in the provided params.
func (*Runner) Exec(params *ExecParams) (string, error) {
	//nolint:gosec
	c := exec.Command(params.CMD, params.Args...)
	c.Dir = params.PWD
	c.Env = buildEnv(params.EnvVars)

	out, err := c.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			//nolint:goerr113
			return "", errors.New(string(out))
		}

		return "", err
	}

	return string(out), err
}

func buildEnv(envVars map[string]string) []string {
	env := os.Environ()

	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}
