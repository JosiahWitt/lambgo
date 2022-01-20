# Lambgo

[![CI](https://github.com/JosiahWitt/lambgo/workflows/CI/badge.svg)](https://github.com/JosiahWitt/lambgo/actions?query=branch%3Amaster+workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/JosiahWitt/lambgo)](https://goreportcard.com/report/github.com/JosiahWitt/lambgo)
[![codecov](https://codecov.io/gh/JosiahWitt/lambgo/branch/master/graph/badge.svg)](https://codecov.io/gh/JosiahWitt/lambgo)

## Go Version Support
Only the last two minor versions of Go are officially supported.

## Table of Contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Install](#install)
- [About](#about)
- [Configuring Lambgo](#configuring-lambgo)
- [Examples](#examples)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

---

## Install
```bash
# Requires Go 1.16+
$ go install github.com/JosiahWitt/lambgo/cmd/lambgo@latest
```


## About
Lambgo is a simple framework for building AWS Lambdas in Go.
It currently consists of a CLI to build paths listed in the [`.lambgo.yml` file](#configuring-lambgo).


## Configuring Lambgo
Lambgo is configured using a `.lambgo.yml` file which is located in the root of your Go Module (next to the `go.mod` file).

Here is an example `.lambgo.yml` file:

```yaml
# Directory to use as the root for build artifacts.
# Optional, defaults to tmp.
outDirectory: tmp

# File name to use for all zipped binaries.
# Useful when using provided.al2 instead of go1.x for the Lambda runtime.
# Optional, defaults to the name of the Lambda's directory.
# zippedFileName: bootstrap

# Additional build flags passed to "go build"
# For example, if you want to provide extra compiler or linker options
# buildFlags: -tags extra,tags -ldflags="-linker -flags"

# Paths to build into Lambda zip files.
# Each path should contain a main package.
# The artifacts are built to: <outDirectory>/<buildPath>.zip
buildPaths:
  - lambdas/hello_world
```

Using the above example file would cause `lambgo build` to build `lambdas/hello_world` to `tmp/lambdas/hello_world.zip`.


## Examples
See the [`examples` directory](./examples) for examples.
