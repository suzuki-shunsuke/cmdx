# cmdx

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/cmdx/status.svg)](https://cloud.drone.io/suzuki-shunsuke/cmdx)
[![codecov](https://codecov.io/gh/suzuki-shunsuke/cmdx/branch/master/graph/badge.svg)](https://codecov.io/gh/suzuki-shunsuke/cmdx)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/cmdx)](https://goreportcard.com/report/github.com/suzuki-shunsuke/cmdx)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/cmdx.svg)](https://github.com/suzuki-shunsuke/cmdx)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/cmdx/master/LICENSE)

Task runner.

## Project Status

This project status is alpha.
The API is unstable and the test are poor.

## Alternative

* Make
* [npm scripts](https://docs.npmjs.com/misc/scripts)
* [Task](https://taskfile.dev)
* [tj/robo](https://github.com/tj/robo)
* [mumoshu/variant](https://github.com/mumoshu/variant)

## Overview

`cmdx` is the project specific task runner.
Using `cmdx` you can manage tasks for your project such as test, build, format, lint, and release.

For example, This is the tasks for `cmdx` itself.

```console
$ cmdx -l
coverage - test a package
test - test
fmt - format the go code
vet - go vet
lint - lint the go code
release - release the new version
durl - check dead links
ci-local - run the Drone pipeline at localhost

$ cmdx help release
NAME:
   cmdx release - release the new version

USAGE:
   cmdx release [arguments...]

DESCRIPTION:
   release the new version
```

You can make the simple shell script rich with `cmdx`

* `cmdx` supports the parse of the flag and positional arguments
* `cmdx` provides useful help messages

`cmdx` searches the configuration file from the current directory to the root directory recursively and run the task at the directory where the configuration file exists so the result of task doesn't depend on the directory you run `cmdx`.

## Install

Download the binary from the [release page](https://github.com/suzuki-shunsuke/cmdx/releases).

## Getting Started

Create the configuration file.

```console
$ cmdx --init
$ cat .cmdx.yaml
```

Edit the configuration file and register the task `hello`.

```console
$ vi .cmdx.yaml
$ cat .cmdx.yaml
---
tasks:
- name: hello
  description: hello command
  usage: hello command
  flags:
  - name: source
    short: s
    usage: source file path
    required: true
    bind_envs:
    - NAME
  - name: switch
    type: bool
  args:
  - name: name
    usage: name
    default: bb
  environment:
    FOO: foo
  script: "echo hello {{.source}} $NAME {{if .switch}}on{{else}}off{{end}} {{.name}} $FOO" # use Go's text/template
```

Output the help.

```console
$ cmdx help
NAME:
   cmdx - task runner

USAGE:
   cmdx [global options] command [command options] [arguments...]

VERSION:
   0.2.2

AUTHOR:
   Shunsuke Suzuki

COMMANDS:
   hello     hello command
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  configuration file path
   --name value, -n value    configuration file name. The configuration file is searched from the current directory to the root directory recursively
   --init, -i                create the configuration file
   --list, -l                list tasks
   --help, -h                show help
   --version, -v             print the version
```

List tasks.


```console
$ cmdx -l
hello - hello command
```

Run the task `hello`.

```console
$ cmdx hello -s README
+ echo hello README $NAME off bb $FOO
hello README README off bb foo

$ cmdx hello -s README --switch
+ echo hello README $NAME on bb $FOO
hello README README on bb foo
```

```console
$ cmdx hello --target foo
target: foo
```

## Configuration

path | type | description | required | default
--- | --- | --- | --- | ---
.bind_envs | []string | default environment variable binding | false | []
.tasks | []task | the list of tasks | true |
task.name | string | the task name | true |
task.short | string | the task short name | false |
task.description | string | the task description | false | ""
task.usage | string | the task usage | false | ""
task.flags | []flag | the task flag arguments | false | []
task.args | []arg | the task positional arguments | false | []
task.bind_envs | []string | task level environment variable binding | false | []
task.environment | map[string]string | the task's environment variables | false | {}
task.script | string | the task command. This is run by `sh -c` | true |
flag.name | string | the flag name | true |
flag.short | string | the flag short name | false |
flag.usage | string | the flag usage | false | ""
flag.default | string | the flag argument's default value | false | ""
flag.bind_envs | []string | flag level environment variable binding | false | []
flag.type | string | the flag type. Either "string" or "bool" | false | "string"
flag.required | bool | whether the flag argument is required | false | false
arg.name | string | the positional argument name | true |
arg.usage | string | the positional argument usage | false | ""
arg.default | string | the positional argument's default value | false | ""
arg.bind_envs | []string | the positional argument level environment variable binding | false | []
arg.required | bool | whether the argument is required | false | false

### bind_envs

`cmdx` supports the bidirectional binding between the variable and the environment variable.

Let's see the following configuration.

```yaml
args:
- name: source
  bind_envs:
  - "foo"
```

By the above configuration, if the environment variable "FOO" is set then the variable "source" is set to the value of the environment variable "FOO".
And if the positional argument "source" is set and the environment variable "FOO" isn't set, the environment variable "FOO" is set to the value of the variable "source".

The element of `bind_envs` is parsed by Golang's text/template and the argument name is referred by `{{.name}}`.

`bind_envs` can be defined at the following levels.

1. flag or arg level
2. task level
3. root level

The priority is `flag or arg level` > `task level` > `root level`.

```yaml
---
bind_envs:
- "{{.name}}"
tasks:
- name: foo
  bind_envs:
  - "{{.name}}"
  args:
  - name: source
    bind_envs:
    - "{{.name}}"
  flags:
  - name: id
    bind_envs:
    - "{{.name}}"
```


### script

`task.script` is the task command.
This is parsed by Golang's [text/template](https://golang.org/pkg/text/template/) package.
The value of the flag and positional argument can be referred by the argument name.

For example,

```yaml
# refer the value of the argument "source"
script: "echo {{.source}}"
```

If the positional argument is optional and the argument isn't passed and the default value isn't set,
the value is an empty string `""`.

And some special variables are defined.

name | type | description
--- | --- | ---
`_builtin.args` | []string | the list of positional arguments which aren't defined by the configuration `args`
`_builtin.args_string` | string | the string which joins _builtin.args by the space " "
`_builtin.all_args` | []string | the list of all positional arguments
`_builtin.args_string` | string | the string which joins _builtin.all_args by the space " "

### Example

```yaml
---
tasks:
- name: hello
  flags:
  - name: source
    short: s
    description: source file path
    default: config.json
    required: false
  args:
  - name: id
    description: id
    required: true
    bind_envs:
    - USER_ID
  environment:
    TOKEN: "*****"
  script: "bash scripts/hello.sh ${source}"
```

## License

[MIT](LICENSE)
