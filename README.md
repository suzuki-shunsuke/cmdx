# cmdx

Task runner.

## Project Status

This project status is alpha.
The API is unstable and the test and document are poor.

## Alternative

* Make
* [npm scripts](https://docs.npmjs.com/misc/scripts)
* [tj/robo](https://github.com/tj/robo)
* [mumoshu/variant](https://github.com/mumoshu/variant)

## Overview

cmdx is the project specific task runner.
Using cmdx you can manage tasks for your project such as test, build, format, lint, and release.

For example, This is the tasks for cmdx itself.

```console
$ cmdx -l
test - test
fmt - format the go code
vet - go vet
lint - lint the go code
release - release the new version
durl - check dead links
ci-local - run the Drone pipeline at localhost

$ cmdx help release
NAME:
   main release - release the new version

USAGE:
   main release [arguments...]

DESCRIPTION:
   release the new version
```

You can make the simple shell script rich with cmdx

* cmdx supports the parse of flag and positional arguments
* cmdx provides the useful help messages

cmdx searches the configuration file from the current directory to the root directory recursively and run the task at the directory where the configuration file exists so the result of task doesn't depend on the directory you run the task.

## Install

Download the binary from the [release page](https://github.com/suzuki-shunsuke/cmdx/releases).

## Getting Started

Create the configuration file.

```console
$ cmdx --init
$ cat .cmdx.yaml
```

Edit the configuration file.

```console
$ vi .cmdx.yaml
$ cat .cmdx.yaml
```

```console
$ cmdx help
```

Run the task

```console
$ cmdx hello
--target is required!

$ echo $?
1
```

```console
$ cmdx hello --target foo
target: foo
```

## Configuration

path | type | description | required | default
--- | --- | --- | --- | ---
.tasks | []task | the list of tasks | true |
task.name | string | the task name | true |
task.short | string | the task short name | false |
task.description | string | the task description | false | ""
task.usage | string | the task usage | false | ""
task.flags | []flag | the task flag arguments | false | []
task.args | []arg | the task positional arguments | false | []
task.environment | map[string]string | the task's environment variables | false | {}
task.script | string | the task command. This is run by `sh -c` | true |
flag.name | string | the flag name | true |
flag.short | string | the flag short name | false |
flag.usage | string | the flag usage | false | ""
flag.default | string | the flag argument's default value | false | ""
flag.env | string | the environment variable name which the flag value is set | false |
flag.type | string | the flag type. Either "string" or "bool" | false | "string"
flag.required | bool | whether the flag argument is required | false | false
arg.name | string | the positional argument name | true |
arg.usage | string | the positional argument usage | false | ""
arg.default | string | the positional argument's default value | false | ""
arg.env | string | the environment variable name which the argument value is set | false |
arg.required | bool | whether the argument is required | false | false

### script

`task.script` is the task command.
This is parsed by Golang's [text/template](https://golang.org/pkg/text/template/) package.
The value of the flag and positional argument can be refered by the argument name.

For example,

```yaml
# refer the value of the argument "source"
script: "echo {{.source}}"
```

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
    env: USER_ID
  environment:
    TOKEN: "*****"
  script: "bash scripts/hello.sh ${source}"
```

## License

[MIT](LICENSE)
