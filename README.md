# tmpl8

Generic (and Kubernetes-friendly) Templating Engine using the Go `text/template` engine and Sprig functions.

![Go version](https://img.shields.io/github/go-mod/go-version/AB-Lindex/tmpl8)
![Latest release](https://img.shields.io/github/v/release/AB-Lindex/tmpl8)
![License](https://img.shields.io/github/license/AB-Lindex/tmpl8)

## Features

* Supports **JSON**, **YAML**, and **TOML** as input/output formats
* Uses the Go [text/template](https://pkg.go.dev/text/template) engine
* Includes the full [Masterminds/sprig](https://github.com/Masterminds/sprig) function library
* Additional built-in template functions (see [Template Functions](#template-functions))
* Import templates and data directly from a Kubernetes ConfigMap
* Chain multiple templates and multiple input objects
* `define`/`block` helpers shared across all template sources

## Installation

Download a pre-built binary for your platform from the [GitHub Releases](https://github.com/AB-Lindex/tmpl8/releases) page (Linux, macOS, and Windows — amd64/arm64).

Or install from source with Go:

```sh
go install github.com/AB-Lindex/tmpl8@latest
```

## Usage

```
tmpl8 [--input INPUT] [--output FILE] [-v] [-t] [--raw] [--split] [-z] TEMPLATE [TEMPLATE ...]
```

Supported input formats: JSON, YAML

### Options

| Option              | Description |
| ------------------- | ----------- |
| `-i` / `--input`    | Input file(s) to import. Use `-` for stdin or `?inlinedata` for inline data. Can be specified multiple times. |
| `-z`                | Add an empty object as input (for template-only processing) |
| `-o` / `--output`   | Write output to a file instead of stdout |
| `-v` / `--verbose`  | Verbose output |
| `-t` / `--trace`    | Trace output (implies verbose) |
| `-r` / `--raw`      | Do not ensure each template block ends with a newline (not recommended for YAML output) |
| `-s` / `--split`    | Split a JSON array input into separate objects (similar to YAML `---` separators) |
| `--left-delimiter`  | Template opening delimiter (default: `{{`) |
| `--right-delimiter` | Template closing delimiter (default: `}}`) |
| `--version`         | Display current version |

### Template argument forms

| Form              | Description |
| ----------------- | ----------- |
| `file.tmpl`       | A single template file |
| `@filelist.lst`   | A text file containing a list of template filenames (one per line) |
| `k8s:ns/name`     | All values from a Kubernetes ConfigMap, imported as templates. No `kubectl` required; uses in-cluster token or kubeconfig. |
| `?inline`         | Inline template text (no file needed) |

## Examples

```sh
# Read input.json and apply it to t1.yaml and t2.yaml
$ tmpl8 <input.json t1.yaml t2.yaml
```

### Equivalent commands

```sh
$ tmpl8 <input.json t1.yaml t2.yaml >output.txt
$ tmpl8 -i input.json t1.yaml t2.yaml -o output.txt
$ cat input.json | tmpl8 @t.lst >output.txt
```

where `t.lst` contains:
```
t1.yaml
t2.yaml
```

See the [`examples/`](examples/) folder for more detailed walkthroughs.

## Inline data

Prefix an argument with `?` to use the remaining text as raw data instead of a filename. Works for both `-i` inputs and template arguments.

```sh
$ tmpl8 -i '?name: alpha' '?{{ .name | upper }}'
ALPHA
```

### Custom delimiters

Use `--left-delimiter` and `--right-delimiter` when template sources must preserve Go-style `{{` and `}}` expressions for another tool.

```sh
$ tmpl8 --left-delimiter '[[' --right-delimiter ']]' -i '?name: alpha' '?[[ .name | upper ]]'
ALPHA
```

## Template Functions

In addition to the full [Sprig](https://masterminds.github.io/sprig/) library, tmpl8 provides these built-in functions:

| Function         | Description |
| ---------------- | ----------- |
| `toYaml`         | Marshal a value to a YAML string |
| `fromYaml`       | Parse a YAML string into a map |
| `fromYamlArray`  | Parse a YAML array string into a slice |
| `toJson`         | Marshal a value to a JSON string |
| `fromJson`       | Parse a JSON string into a map |
| `fromJsonArray`  | Parse a JSON array string into a slice |
| `toToml`         | Marshal a value to a TOML string |
| `readfile`       | Read the contents of a file as a string |
| `version`        | Return the current tmpl8 version string |
| `isfunc`         | Check whether a given function name is available in the template context |

## define/block support

Any `define`/`block` helpers defined in a template are available in all subsequent template sources and can be overridden by later templates.

## Kubernetes support

To import data from a Kubernetes ConfigMap, use a filename of the form `k8s:namespace/configmap`.
All values within the ConfigMap will be imported as templates (or as inputs when used with `-i`).

```sh
$ tmpl8 k8s:default/my-templates input-data.yaml
```