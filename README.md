# scafall

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

### Of `Overrides` and `DefaultValues`

When using `scafall` programmatically you may want to provide some constant values.  In `scafall` these are termed _overrides_.  An override may define `map[string]string{"PI": "3.14"}` any prompting for an alternative value to `PI` is skipped and the `3.14` values is used in templates.  This is particularly useful where the calling code calculates a value, such as a username, and does not want the end-user to be prompted to chage this value.

In all cases, overrides _can_ be provided in a `.override.toml` file.  The `.override.toml` file is intended to simplify testing and therefore the format is an implementation detail.  Because the format is an implementation detail, we do not document it here.

`DefaultValues` are useful where the calling code provides a sane default for the end-user.  For example, a template project may prompt the user to input a `http_proxy`, the default of which can be calculated in a calling application.  The `DefaultValues` feature is motivated by the need for a calling application to provide a range of choices to the end-user, without specifying a constant value as an override.

## Project Templates

Project templates are normal source code projects with the addition of a `prompts.toml` file.  The `prompts.toml` file defines questions to ask of the end-user.  The answers to the questions are available as template variables.  For example, suppose we have a project template to create a new Python project, we just need to ask the end-user which python interprer to use and how many python digits to generate:

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

## Template Collections

Unlike a template project a template collection must not contain a `prompts.toml` at the root project directory.  Any top-level `prompts.toml` will be silently ignored.  Instead, a template collection is a git repository that contains multiple template projects.

Given a template collection, the end-user is prompted to choose to create a project from one of the project templates.

```bash
$ tree .
├── go
│   ├── prompts.toml
│   └── {{.ProjectName}}
│       └── main.go
└── python
    ├── prompts.toml
    └── {{.ProjectName}}
        └── print_pi.py
```

Running `scafall` produces a default prompt to choose between project templates.  Project template specific prompts follow in end-user prompts.

```bash
$ scafall http://github.com/AidanDelaney/scafall-collection-eg/
Use the arrow keys to navigate: ↓ ↑ → ←
? Choose a project template:
  ▸ go
    python
```

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
