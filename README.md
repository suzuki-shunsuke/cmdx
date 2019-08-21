# cmdx

Task runner.

## Alternative

* Make
* [npm scripts](https://docs.npmjs.com/misc/scripts)
* [tj/robo](https://github.com/tj/robo)
* [mumoshu/variant](https://github.com/mumoshu/variant)

## Features

* the rich argument parser with validation
* the rich help message
* traverse configuration file and the task is run at the directory the configuration file exists
* support the variables and environment variables

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

```yaml
---
commands:
- name: hello
  flags:
  - name: source
    short: s
    description: source file path
    default: .drone.jsonnet
    required: false
  args:
  - name: name
    description: source file path
    required: true
    binding: NAME
  either:
  -
    - foo
    - bar
  both:
  -
    - foo
    - bar
  environment:
    FOO: foo
  script: "bash scripts/foo.sh ${source} ${name}"
```

## License

[MIT](LICENSE)
