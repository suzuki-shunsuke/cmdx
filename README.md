# cmdx

[![Build Status](https://cloud.drone.io/api/badges/suzuki-shunsuke/cmdx/status.svg)](https://cloud.drone.io/suzuki-shunsuke/cmdx)
[![Go Report Card](https://goreportcard.com/badge/github.com/suzuki-shunsuke/cmdx)](https://goreportcard.com/report/github.com/suzuki-shunsuke/cmdx)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/cmdx.svg)](https://github.com/suzuki-shunsuke/cmdx)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/cmdx/main/LICENSE)

Task runner. It provides useful help messages and supports interactive prompts.

## Overview

`cmdx` is the task runner.
Using `cmdx` you can manage tasks for your project such as test, build, format, lint, and release.

For example, This is the tasks for `cmdx` itself.

```console
$ cmdx -l
init, i - setup git hooks
coverage, c - test a package (fzf is required)
test, t - test
fmt - format the go code
vet, v - go vet
lint, l - lint the go code
release, r - release the new version
durl - check dead links (durl is required)
ci-local - run the Drone pipeline at localhost (drone-cli is required)

$ cmdx help release
NAME:
   cmdx release - release the new version

USAGE:
   cmdx release <version>

DESCRIPTION:
   release the new version
ARGUMENTS:
   version
```

`cmdx` searches the configuration file from the current directory to the root directory recursively, and runs the task at the directory where the configuration file exists.

## Features

* Easy to install (one binary)
* Parse the flag and positional arguments
* Useful help messages
* Interactive prompt by [AlecAivazis/survey](https://github.com/AlecAivazis/survey)
* Validate requirements
* Validate flag and positional arguments
* Timeout
* Bash and Zsh completion
* Nested tasks (Sub tasks)

## Install

Download the binary from the [release page](https://github.com/suzuki-shunsuke/cmdx/releases).
you can install cmdx with [Homebrew](https://brew.sh/) too.

```console
$ brew install suzuki-shunsuke/cmdx/cmdx
```

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

## Example

In order to understand `cmdx`, it is good to execute `cmdx` command actually.
We prepare the sample configuration file for you to execute `cmdx`.

Please see [examples](examples).

## Configuration

path | type | description | required | default
--- | --- | --- | --- | ---
.timeout | timeout | the task command timeout | false |
.input_envs | []string | default environment variable binding | false | []
.script_envs | []string | default environment variable binding | false | []
.environment | map[string]string | top level environment variables | false | {}
.quiet | bool | Default configuration whether the content of script is outputted | false |
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
task.quiet | bool | task level default configuration whether the content of script is outputted | false |
task.shell | []string | shell command to run the script | `["sh", "-c"]`
task.timeout | timeout | the task command timeout | false |
task.require | require | requirement of task | false | {}
task.tasks | []task | sub tasks | false | `[]`
require.exec | []stringArray | required executable files | false | []
require.environment | []stringArray | required environment variables | false | []
stringArray | array whose element is string or array of string | |
timeout.duration | int | the task command timeout (second) | false | 36000 (10 hours)
timeout.kill_after | int | the duration the kill signal is sent after `timeout.duration` | false | 0, which means the command isn't killed
flag.name | string | the flag name | true |
flag.short | string | the flag short name | false |
flag.usage | string | the flag usage | false | ""
flag.default | string | the flag argument's default value | false | ""
flag.input_envs | []string | flag level environment variable binding | false | []
flag.script_envs | []string | flag level environment variable binding | false | []
flag.type | string | the flag type. Either "string" or "bool" | false | "string"
flag.required | bool | whether the flag argument is required | false | false
flag.validate | []validate | parameters to validate the value of flag | false | []
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
arg.validate | []validate | parameters to validate the value of arg | false | []
validate.type | string | value type (`email`, `url`, `int`) | false |
validate.regexp | string | the regular expression | false |
validate.min_length | int | the minimum string length | false |
validate.max_length | int | the maximum string length | false |
validate.prefix | string | the prefix | false |
validate.suffix | string | the suffix | false |
validate.contain | string | the string which the value should contain | false |
validate.enum | []string | enum | false |

### script

`task.script` is the task command.
This is parsed by Go's [text/template](https://golang.org/pkg/text/template/) package.
[sprig](http://masterminds.github.io/sprig/) functions can be used.
The value of the flag and positional argument can be referred by the argument name.

For example,

```yaml
# refer the value of the argument "source"
script: "echo {{.source}}"
```

Multiple lines

```yaml
script: |
  echo foo
  echo bar
```

If the positional argument is optional and the argument isn't passed and the default value isn't set,
the value is an empty string `""`.

And some special variables are defined.

name | type | description
--- | --- | ---
`_builtin.args` | []string | the list of positional arguments which aren't defined by the configuration `args`
`_builtin.args_string` | string | the string which joins `_builtin.args` by the space " "
`_builtin.all_args` | []string | the list of all positional arguments
`_builtin.args_string` | string | the string which joins `_builtin.all_args` by the space " "

### input_envs, script_envs

`input_envs` is a list of environment variables that are bound to the variable.

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

### timeout

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

### require

`task.require` is the requirement to run the task.

#### require.exec

For example, in the following example both of `curl` and `wget` is required.

```yaml
tasks:
- name: foo
  script: curl http://example.com
  require:
    exec:
    - curl
    - wget
```

If `curl` isn't installed, the task is failed.

```
$ cmdx foo
curl is required
```

Note that the shell's alias is ignored.
Internally, [exec.LookupPath](https://golang.org/pkg/os/exec/#LookPath) is used.

In the following example, either `curl` or `wget` is required.

```yaml
tasks:
- name: foo
  script: curl http://example.com
  require:
    exec:
    - - curl
      - wget
```

```
$ cmdx foo
one of the following is required: curl, wget
```

#### require.environment

`require.environment` is the required environment variables.
Note that if the value of the environment variable is an emtpy string, the environment variable is treated as unset.

```yaml
tasks:
- name: foo
  script: curl http://example.com
  require:
    environment:
    - GITHUB_TOKEN
```

```
$ cmdx foo
the environment variable 'GITHUB_TOKEN' is required
```

```yaml
tasks:
- name: foo
  script: curl http://example.com
  require:
    environment:
    - - GITHUB_TOKEN
      - GITHUB_ACCESS_TOKEN
```

```
$ cmdx foo
one of the following environment variables is required: GITHUB_TOKEN, GITHUB_ACCESS_TOKEN
```

## validation

`cmdx` supports to validate `args` and `flags`.

For example,

```yaml
# .cmdx.yaml
tasks:
- name: hello
  script: echo hello
  args:
  - name: age
    validate:
    - type: int
```

```
$ cmdx hello foo
age is invalid: must be int: foo
```

## quiet

By default `cmdx` outputs the content of task's `script` when the task is run.

In case of the following example, `+ echo hello` is outputted.

```yaml
# .cmdx.yaml
tasks:
- name: hello
  script: echo hello
```

```console
$ cmdx hello
+ echo hello
hello
```

You can suppress the output by `--quiet (-q)` option.

```console
# BAD: cmdx hello -q
$ cmdx -q hello
hello
```

And you can change the default configuration of `quiet` at the global level or the task level.

```yaml
# .cmdx.yaml
tasks:
- name: hello
  script: echo hello
  quiet: true  # task level
```

```yaml
# .cmdx.yaml
quiet: true # global level
tasks:
- name: hello
  script: echo hello
  quiet: true
```

The priority is

1. command line flag
2. task level configuration
3. global level configuration

If the quiet is enabled by configuration but you want to output `script`, please set the flag `-q=false`.

```
# "=" is needed
$ cmdx -q=false hello
+ echo hello
hello
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

## shell

**This is an advanced feature.**

By default task's `script` is run by the command `sh -c`.
You can change the command by the `shell` option.

For example, you can run Python script.

```yaml
- name: hello
  shell:
  - python
  - -c
  script: |
    print("hello")
```

And you can run the shell script in the container.

```yaml
- name: hello
  shell:
  - docker
  - exec
  - -ti
  - foo
  - sh
  - -c
  script: |
    whoami
    read -p "name?" name
    echo "name: $name"
```

## Bash (Zsh) Completion

https://github.com/urfave/cli/blob/477292c8d462a3f51cd18bc77c0542193a62274d/docs/v2/manual.md#bash-completion

`cmdx` supports Bash (Zsh) Completion powered by urfave/cli.

We test the completion with Zsh, but we don't test the completion with other shell.

To enable the completion, you have to load a shell script.
For detail, please see the [document of urfave/cli](https://github.com/urfave/cli/blob/477292c8d462a3f51cd18bc77c0542193a62274d/docs/v2/manual.md#bash-completion).

Please set `cmdx` to `PROG`

## Sub tasks

`cmdx` supports sub tasks.

For example,

```yaml
tasks:
- name: admin
  usage: administrator feature
  tasks:
  - name: cluster
    usage: manage clusters
    tasks:
    - name: create
      usage: create a cluster
      script: echo "create a cluster"
```

```
$ cmdx admin cluster create
+ echo "create a cluster"
create a cluster
```

The following attributes are inherited from the parent tasks.

* input_envs
* script_envs
* quiet
* environment
* timeout
* requires

For example,

```yaml
tasks:
- name: admin
  usage: administrator feature
  require:
    exec:
    - yamllint
  tasks:
  - name: cluster
    usage: manage clusters
    tasks:
    - name: create
      usage: create a cluster
      script: echo "create a cluster"
```

```
$ cmdx admin cluster create
yamllint is required
```

We can't set both `task.script` and `task.tasks`.

For example,

```yaml
tasks:
- name: hello
  script: echo hello
  tasks:
  - name: world
    script: echo "hello world"
```

```
$ cmdx hello
please fix the configuration file: the task `hello` is invalid. when sub tasks are set, 'script' can't b
e set
```

## Contributing

Please see the [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
