# scafall

[![Build results](https://github.com/buildpacks/scafall/workflows/build/badge.svg)](https://github.com/buildpacks/scafall/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/buildpacks/scafall)](https://goreportcard.com/report/github.com/buildpacks/scafall)
[![codecov](https://codecov.io/gh/buildpacks/scafall/branch/main/graph/badge.svg)](https://codecov.io/gh/buildpacks/scafall)
[![GoDoc](https://godoc.org/github.com/buildpacks/scafall?status.svg)](https://godoc.org/github.com/buildpacks/scafall)
[![GitHub license](https://img.shields.io/github/license/buildpacks/scafall)](https://github.com/buildpacks/scafall/blob/main/LICENSE)
[![Slack](https://img.shields.io/badge/slack-join-ff69b4.svg?logo=slack)](https://slack.cncf.io/)
[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/buildpacks/scafall)


A project scaffolding tool inspired by [cookiecutter](https://github.com/cookiecutter/cookiecutter).

## Problem

We needed a tool to create new source code projects from templates.  In addition, we needed the tool to be a libaray written in [Go](https://go.dev/).  Scafall takes project templates, asks the end-user some questions and produces an output folder.

## Installation and CLI

As a Go developer you can install `scafall` into your `GOBIN` directory.

```bash
$ go install github.com/AidanDelaney/scafall@latest
```

The `scafall` CLI should now be available for use

```bash
$ scafall http://github.com/AidanDelaney/scafall-python-eg.git
✔ Please input a project name: pyexample
Use the arrow keys to navigate: ↓ ↑ → ←
? Which Python version to use:
  ▸ python3.10
    python3.9
    python3.8
How many digits of Pi to render: 3
$ cd pyexample
$ ./print_pi.py
```

## Programmatic Usage

The programmatic API is documented on [`pkg.go.dev`](https://pkg.go.dev/github.com/AidanDelaney/scafall), which contains more examples.  A basic example will prompt the end-user for any values the project scaffolding requires:

```go
package main

import (
  "fmt"

  scafall "github.com/AidanDelaney/scafall/pkg"
)

func main() {
  s := scafall.NewScafall(scafall.WithOutputfolder("python-pi"))
  err := s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git")
  if err != nil {
    fmt.Printf("scaffolding failed: %s", err)
  }
}
```

### Of `Arguments`

When using `scafall` programmatically you may want to provide values for template variables.  In `scafall` these are termed _arguments_.  An argument may define `map[string]string{"PI": "3.14"}` any prompting for an alternative value to `PI` is skipped and the `3.14` values is used in templates.  This is particularly useful where the calling code calculates a value, such as a username, and does not want the end-user to be prompted to chage this value.

In all cases, arguments _can_ be provided in a `.override.toml` file.  The `.override.toml` file is intended to simplify testing and therefore the format is an implementation detail.  Because the format is an implementation detail, we do not document it here.

## Project Templates

Project templates are normal source code projects with the addition of a `prompts.toml` file.  The `prompts.toml` file defines questions to ask of the end-user.  The answers to the questions are available as template variables.  For example, suppose we have a project template to create a new Python project, we only need to ask the end-user which python interpreter to use and how many python digits to generate:

```bash
$ scafall http://github.com/AidanDelaney/scafall-python-eg.git
? Which Python version to use:
  ▸ python3.10
    python3.9
    python3.8
✔ How many digits of Pi to render: 3
2022/04/06 20:28:41     create  /print_pi.py
```

The values for the python interpreter and number of digits to render are available as `{{.PythonVersion}}` and `{{.NumDigits}}` respectively.  Thus the input template

```python
#!env -- {{.PythonVersion}}
from math import pi

print("%.{{.NumDigits}}f" % pi)
```

is generated as

```python
#!env -- python3.10
from math import pi

print("%.3f" % pi)
```

A project template containing a `prompts.toml` file will produce a generated project that omits the `prompts.toml` file.  In addition, any root-level `README.md` file in the project template is not propagated to the generated project.  This allows the project template to contain a `README.md` to explain usage of the project template.

## Prompts.toml Format

The `prompts.toml` file is a sequence of `[[prompt]]` which must each deine a `name` and `prompt`.  A minimal example is

```toml
[[prompt]]
name = "NumDigits"
prompt = "How many digits of Pi to render"
```

An example with two prompts is

```toml
[[prompt]]
name = "PythonVersion"
prompt = "Which Python version to use"
required = true
choices = ["python3.10", "python3.9", "python3.8"]

[[prompt]]
name = "NumDigits"
prompt = "How many digits of Pi to render"
default = "3"
```

The `choices` and `default` fields are mutually exclusive.  In the case that both `choices` and `default` are used, the `default` is silently ignored and the first of `choices` becomes the default.
