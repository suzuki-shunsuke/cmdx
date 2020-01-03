# Example

In order to understand `cmdx`, it is good to execute `cmdx` command actually.
We prepare the sample configuration file for you to execute `cmdx`.

[cmdx.yaml](cmdx.yaml)

```console
$ cmdx -l
```

```console
$ cmdx help
```

```console
$ cmdx help i
```

```console
# read the top level environment variable
$ cmdx simple
+ echo $FOO
foo
```

### environment

```console
$ cmdx env
+ echo $FOO $BAR $ZOO
foo barbar zoo
```

### positional argument

```console
# positional argument default value
$ cmdx hello
+ echo "Hello, Bob"
Hello, Bob
```

```console
# positional argument
$ cmdx hello foo
+ echo "Hello, foo"
Hello, foo
```

### flag argument

```console
# flag default value
$ cmdx i
+ echo $NAME
Bob
```

```console
# flag
$ cmdx i --name foo
+ echo $NAME
foo
```

### input_envs, script_envs

```yaml
  script: "echo $INSTALL_NAME"
  flags:
  - name: name
    usage: your name
    default: Bob
    input_envs:
    - "{{.name}}"
    script_envs:
    - "INSTALL_{{.name}}"
```

```console
# NAME -(input_envs)-> .name -(script_envs)-> INSTALL_NAME
$ NAME=zzz cmdx i
+ echo $INSTALL_NAME
zzz
```

### timeout

```console
$ cmdx timeout
+ sleep 10
the command is timeout: 3 sec
```

### interactive input

```console
$ cmdx input --name foo
+ echo foo
foo
```

```console
$ cmdx input
? name foo
+ echo foo
foo
```

### interactive input's message and help

```yaml
prompt:
  type: input
  message: What's your name?
  help: Please input your name.
```

```console
$ cmdx input-help
? What's your name? [? for help]
```

Type "?".

```console
$ cmdx input-help
? Please input your name.
? What's your name?
```

### interactive password

```console
$ cmdx password
? password ****
+ echo ffff
ffff
```

### interactive select

```yaml
prompt:
  type: select
  options:
  - red
  - green
  - blue
```

```console
$ cmdx select
? select  [Use arrows to move, type to filter]
> red
  green
  blue
```

Select "green".

```console
$ cmdx select
? select green
+ echo green
green
```

### multiline

```yaml
prompt:
  type: multiline
```

```console
$ cmdx multiline
? multiline [Enter 2 empty lines to finish]
fff
bbb

aaa
```

```console
? multiline
fff
bbb

aaa
+ echo 'fff
bbb

aaa'
fff
bbb

aaa
```

### editor

```yaml
prompt:
  type: editor
```

```console
$ cmdx editor
? profile [Enter to launch editor]
```

### confirm

```yaml
- name: confirm
  script: "echo '{{if .confirm}}Good morning{{else}}Good afternoon{{end}}' $CONFIRM"
  flags:
  - name: confirm
    prompt:
      type: confirm
    script_envs:
    - CONFIRM
```

```console
$ cmdx confirm
? confirm (y/N)
```

Type "N".

```console
$ cmdx confirm
? confirm No
+ echo 'Good afternoon' $CONFIRM
Good afternoon false
```

### multi select

```console
$ cmdx multi-select
? select  [Use arrows to move, space to select, type to filter]
> [ ]  red
  [ ]  green
  [ ]  blue
```

Select "red" and "blue".

```console
$ cmdx multi-select
? select  [Use arrows to move, space to select, type to filter]
  [x]  red
  [ ]  green
> [x]  blue
```

```console
$ cmdx multi-select
? select red, blue
+ echo [red blue] $COLORS
[red blue] red,blue
```

`{{.select}}` is array `["red", "blue"]` and the environment variable `COLORS` is `"red,blue"`.
