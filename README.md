# scafall

[![Build results](https://github.com/buildpacks-community/scafall/workflows/build/badge.svg)](https://github.com/buildpacks-community/scafall/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/buildpacks-community/scafall)](https://goreportcard.com/report/github.com/buildpacks-community/scafall)
[![codecov](https://codecov.io/gh/buildpacks-community/scafall/branch/main/graph/badge.svg)](https://codecov.io/gh/buildpacks-community/scafall)
[![GoDoc](https://godoc.org/github.com/buildpacks-community/scafall?status.svg)](https://godoc.org/github.com/buildpacks-community/scafall)
[![GitHub license](https://img.shields.io/github/license/buildpacks-community/scafall)](https://github.com/buildpacks-community/scafall/blob/main/LICENSE)
[![Slack](https://img.shields.io/badge/slack-join-ff69b4.svg?logo=slack)](https://slack.cncf.io/)
[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/buildpacks-community/scafall)


A project scaffolding tool inspired by [cookiecutter](https://github.com/cookiecutter/cookiecutter).

## Problem

We needed a tool to create new source code projects from templates.  In addition, we needed the tool to be a library written in [Go](https://go.dev/).  Scafall takes project templates, asks the end-user some questions and produces an output folder.

Scafall differs from some other Go scaffolding/templating tools as it passes through unknown template subsitutions.  For example, if your input application source or documentation contains a `{{.Foo}}` template and no argument is provided (either programmatically or by the end-user) then the output file will contain the string `{{.Foo}}`.  This allows the generation of projects where the generated source contains templates.

## Installation and CLI

As a Go developer you can install `scafall` into your `GOBIN` directory.

```bash
$ go install github.com/buildpacks-community/scafall@latest
```

The `scafall` CLI should now be available for use

```bash
$ scafall http://github.com/AidanDelaney/scafall-python-eg.git
✔ Please input a project name: pyexample
? Which Python version to use: [Use arrows to move, type to filter]
  ▸ python3.10
    python3.9
    python3.8
How many digits of Pi to render: 3
$ cd pyexample
$ ./print_pi.py
```

## Programmatic Usage

The programmatic API is documented on [`pkg.go.dev`](https://pkg.go.dev/github.com/buildpacks/scafall), which contains more examples.  A basic example will prompt the end-user for any values the project scaffolding requires:

```go
package main

import (
  "fmt"

  scafall "github.com/buildpacks-community/scafall/pkg"
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

## Project Templates

Project templates are normal source code projects with the addition of a `prompts.toml` file.  The `prompts.toml` file defines questions to ask of the end-user.  The answers to the questions are available as template variables.  For example, suppose we have a project template to create a new Python project, we only need to ask the end-user which python interpreter to use and how many python digits to generate:

```bash
$ scafall http://github.com/AidanDelaney/scafall-python-eg.git
? Which Python version to use: [Use arrows to move, type to filter]
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
