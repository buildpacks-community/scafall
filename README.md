# scafall

A project scaffolding tool inspired by [cookiecutter](https://github.com/cookiecutter/cookiecutter).

## Problem

We needed a tool to create new source code projects from templates.  In addition, we needed the tool to be a libaray written in [Go](https://go.dev/).  Scafall takes project templates, asks the end-user some questions and produces an output folder.

## Project Templates

Project templates are normal source code projects with the addition of a `prompts.toml` file.  The `prompts.toml` file defines questions to ask of the end-user.  The answers to the questions are available as template variables.  For example, suppose we have a project template to create a new Python project, we just need to ask the end-user which python interprer to use and how many python digits to generate:

```bash
$ scafall http://github.com/AidanDelaney/scafall-python-eg.git python-pi
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

