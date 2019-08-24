# cmdx

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/cmdx/status.svg)](https://cloud.drone.io/suzuki-shunsuke/cmdx)
[![codecov](https://codecov.io/gh/suzuki-shunsuke/cmdx/branch/master/graph/badge.svg)](https://codecov.io/gh/suzuki-shunsuke/cmdx)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/cmdx)](https://goreportcard.com/report/github.com/suzuki-shunsuke/cmdx)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/cmdx.svg)](https://github.com/suzuki-shunsuke/cmdx)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/cmdx/master/LICENSE)

Task runner.

## Alternative

* Make
* [npm scripts](https://docs.npmjs.com/misc/scripts)
* [Task](https://taskfile.dev)
* [tj/robo](https://github.com/tj/robo)
* [mumoshu/variant](https://github.com/mumoshu/variant)

## Overview

`cmdx` is the task runner.
Using `cmdx` you can manage tasks for your project such as test, build, format, lint, and release.

For example, This is the tasks for `cmdx` itself.

```console
$ cmdx -l
init, i - setup git hooks
coverage, c - test a package
test, t - test
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
* `cmds` supports the interactive prompt by [AlecAivazis/survey](https://github.com/AlecAivazis/survey)

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
    input_envs:
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
.timeout | timeout | the task command timeout | false |
.input_envs | []string | default environment variable binding | false | []
.script_envs | []string | default environment variable binding | false | []
.tasks | []task | the list of tasks | true |
task.name | string | the task name | true |
task.short | string | the task short name | false |
task.description | string | the task description | false | ""
task.usage | string | the task usage | false | ""
task.flags | []flag | the task flag arguments | false | []
task.args | []arg | the task positional arguments | false | []
task.input_envs | []string | task level environment variable binding | false | []
task.script_envs | []string | task level environment variable binding | false | []
task.environment | map[string]string | the task's environment variables | false | {}
task.script | string | the task command. This is run by `sh -c` | true |
task.timeout | timeout | the task command timeout | false |
timtout.duration | int | the task command timeout (second) | false | 36000 (10 hours)
timtout.kill_after | int | the duration the kill signal is sent after `timeout.duration` | false | 0, which means the command isn't killed
flag.name | string | the flag name | true |
flag.short | string | the flag short name | false |
flag.usage | string | the flag usage | false | ""
flag.default | string | the flag argument's default value | false | ""
flag.input_envs | []string | flag level environment variable binding | false | []
flag.script_envs | []string | flag level environment variable binding | false | []
flag.type | string | the flag type. Either "string" or "bool" | false | "string"
flag.required | bool | whether the flag argument is required | false | false
flag.prompt | prompt | prompt | false | prompt is disabled
prompt.type | string | prompt type | true |
prompt.message | string | prompt message | false | `flag.name` or `arg.name`
prompt.help | string | prompt help | false |
prompt.options | []string | entries of `select` or `multi_select` prompt | true if the prompt type is `select` or `multi_select` |
arg.name | string | the positional argument name | true |
arg.usage | string | the positional argument usage | false | ""
arg.default | string | the positional argument's default value | false | ""
arg.input_envs | []string | the positional argument level environment variable binding | false | []
arg.script_envs | []string | the positional argument level environment variable binding | false | []
arg.required | bool | whether the argument is required | false | false
arg.prompt | prompt | prompt | false | prompt is disabled

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

### input_envs, script_envs, environment

`input_envs` is a list of environment variables which are bound to the variable.

```yaml
tasks:
- name: foo
  script: "echo {{.source}}"
  args:
  - name: source
    input_envs:
    - command_env
```

```console
$ COMMAND_ENV=zzz cmdx foo
+ echo zzz
zzz
```

`script_envs` is a list of environment variables which the variable is bound to.

```yaml
tasks:
- name: foo
  script: "echo $COMMAND_ENV"
  args:
  - name: source
    script_envs:
    - command_env
```

```console
$ cmdx foo zzz
+ echo $COMMAND_ENV
zzz
```

### timout

`cmdx` supports the configuration about the timeout of the task.

1. send SIGINT after `timeout.duration` seconds (default 36,000 seconds)
2. if `timeout.kill_after` isn't 0, send SIGKILL after `timeout.duration + timeout.kill_after` seconds. By default `timeout.kill_after` is 0 so SIGKILL isn't sent

For example, the following task `foo`'s timeout is 3 seconds.

```yaml
tasks:
- name: foo
  script: sleep 100
  timeout:
    duration: 3
```

```console
$ cmdx foo
+ sleep 100
the command is timeout: 3 sec
```

The task timeout configuration inherits the top level timeout configuration.

```yaml
timeout:
  duration: 3
tasks:
- name: foo # the timeout.duration is 3
  script: sleep 100
```

## prompt

`cmdx` supports the interactive prompt by [AlecAivazis/survey](https://github.com/AlecAivazis/survey).
`cmdx` supports the following prompt types.

* input
* multiline
* password
* confirm
* select
* multi_select
* editor

About prompt type, please see [AlecAivazis/survey's document](https://github.com/AlecAivazis/survey#prompts).

## value source priority

1. command line arguments
2. environment variable (input_envs)
3. prompt (prompt isn't launched if the value is set by command line argument or environment variable)
4. default value

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
    input_envs:
    - USER_ID
  environment:
    TOKEN: "*****"
  script: "bash scripts/hello.sh ${source}"
```

## Contributing

Please see the [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
