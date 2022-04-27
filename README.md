# tmpl8
Generic (and Kubernetes-friendly) Templating Engine using the go text/template and Sprig functions

### Features
* Import multiple JSON and/or YAML documents
* Using the Go [text/template](https://pkg.go.dev/text/template) engine
* Supports the [Mastermind/sprig](https://github.com/Masterminds/sprig) package for extra template functions
* Can import templates (and data) from Kuberentes ConfigMap

## Usage / Help
Supported input formats: JSON and YAML

```
tmpl8 [--input INPUT] [--output FILE] [--verbose] [--trace] [--raw] [--split] [--noinput] TEMPLATE [TEMPLATE ...]
```

### Options
| Option             | Description |
| -----------------  | ----------- |
| `-i` / `--input`   | Manual assign where to read input object(s) from, use '`-`' to use stdin, multiple inputs supported |
| `-z` / `--noinput` | Add an empty object as input (for template-only processing) |
| `-o` / `--output`  | Set output filename (instead of using `>`)             |
| `-r` / `--raw`     | Don't ensure each template ends with a newline (not recommended for YAML output)|
| `-s` / `--split`   | Split JSON input objects arrays to multiple objects parsed each (similar to YAMLs with `---` separators) |

## define/block support
Any `define`/`block` helpers are useable in templates in all following template sources (and can be overridden)

## Examples

```sh
$ tmpl8 <input.json t1.yaml t2.yaml
# Reads input.json and apply that to the templates t1.yaml and t2.yaml (in that order)
```

### Equivalent commands
```sh
$ tmpl8 <input.json t1.yaml t2.yaml >output.txt
$ tmpl8 -i input.json t1.yaml t2.yaml -o output.txt
$ cat input.json | tmpl8 @t.lst  >output.txt
```

where `t.lst` contains
```
t1.yaml
t2.yaml
```

## Inline data
Adding a '`?`' character to an input or output will use the remaining text as raw data.

```sh
$ tmpl8 -i '?name: alpha' '?{{ .name | upper }}'
ALPHA
```

## Kubernetes support
To import data from a Kubernetes Configmap you only need to use a filename like '`k8s:namespace/configmap`'
All values within will be imported as templates (or inputs if used with '`-i`')