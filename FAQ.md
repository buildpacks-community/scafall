# How Do I

## Define a Template Project

A project template is a collection of files.  These can have any folder structure and the files can contain any text content.  It is customary to have a top-level `{{.ProjectName}}` folder and to include a prompt for `ProjectName`.  This allows the end-user to scaffold the project in a sub directory.

First, create a `git` repository.  Within that repository create `prompts.toml` and a `{{.ProjectName}}` folder.  In `{{.ProjectName}}` create` the source structure that you want in an output project.  Filenames and directory paths can contain template variables.  

Where filenames, directory paths or source text files contain a template variable, eg: `{{.Foo}}`, then it is usual to add a prompt in the top-level `prompts.toml` file.  The `prompts.toml` file allows the end-user to provide a value for `Foo`.  If a template project uses a `prompts.toml` file it must be included in the project root directory.

## Create a Templated Directory

In a project template we can create a directory such as `pkg/{{.PackageName}}`.  In an output project `{{.PackageName}}`  will be replaces with the value of `PackageName` variable.

## Format a Template Variable

There is often a need to read a variale from a user prompt and apply some processing to it.  For example we may need to read a `PackageName` from the user and ensure that it contains no spaces or `-` characters.  Scafall supports all [sprig](http://masterminds.github.io/sprig/) functions that can be used for such processing.

The expression `{{.PackageName | snakecase}}` removes spaces from `PackageName` and replaces all occurances of hyphen with underscore.  The expression `{{.PackageName | snakecase}}` is a valid filename and directory path.  It can also be used internally in text files.

## Use `scafall` Behind a Proxy

Export both `HTTP_PROXY` and `HTTPS_PROXY` environment variables and these will be used by `scafall`.
