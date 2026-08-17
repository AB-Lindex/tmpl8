# tmpl8 – Test Suite

This folder contains self-contained test cases for `tmpl8`. Each test lives in its own numbered sub-folder and is fully reproducible from the command line.

The tests progress from the simplest possible invocation up to realistic pipeline use-cases, making them equally useful as a learning resource and as a regression suite.

---

## Running all tests

Run the helper script from the repository root:

```sh
sh tests/run_all.sh
```

Prerequisites:
- `tmpl8` is on your `$PATH` (or adjust the command to `../bin/tmpl8`)
- A POSIX shell (`bash`, `sh`, `zsh`, …)

---

## Test cases

### 01 – Inline data + inline template (smoke test)

**What it tests:** the simplest possible invocation — no files needed.

```sh
tmpl8 -i '?name: world' '?Hello, {{ .name }}!'
```

Expected output:
```
Hello, world!
```

**Why it matters:** confirms the binary is installed and the `?` inline syntax works for both data and template.

---

### 02 – JSON input → plain-text output

**What it tests:** reading a JSON file and rendering a simple text template.

Folder: `02_json_to_text/`

```sh
tmpl8 -i input.json template.tmpl
```

`input.json`
```json
{ "app": "my-service", "version": "1.2.3" }
```

`template.tmpl`
```
app:     {{ .app }}
version: {{ .version }}
```

Expected output:
```
app:     my-service
version: 1.2.3
```

**Why it matters:** the most common real-world use case — injecting build metadata into a config file.

---

### 03 – YAML input → YAML output

**What it tests:** YAML-to-YAML round-trip with field access and the `upper` Sprig filter.

Folder: `03_yaml_to_yaml/`

```sh
tmpl8 -i input.yaml template.tmpl
```

`input.yaml`
```yaml
name: alpha
namespace: staging
replicas: 2
```

`template.tmpl`
```yaml
metadata:
  name: {{ .name | upper }}
  namespace: {{ .namespace }}
spec:
  replicas: {{ .replicas }}
```

Expected output:
```yaml
metadata:
  name: ALPHA
  namespace: staging
spec:
  replicas: 2
```

**Why it matters:** shows that YAML is a first-class input format and that the full Sprig library is available.

---

### 04 – Empty input with `-z` flag

**What it tests:** generating output from a template that requires no input data.

Folder: `04_no_input/`

```sh
tmpl8 -z template.tmpl
```

`template.tmpl`
```
tmpl8 version: {{ version }}
generated-by: tmpl8
```

Expected output (version will vary):
```
tmpl8 version: v0.x.x
generated-by: tmpl8
```

**Why it matters:** demonstrates `-z` for template-only rendering and the built-in `version` function.

---

### 05 – `fail` on missing / invalid input

**What it tests:** that `fail` stops rendering with a useful error message when required data is absent.

Folder: `05_fail_on_invalid/`

```sh
# Should exit non-zero and print an error message
tmpl8 -z template.tmpl
```

`template.tmpl`
```
{{- if not . }}
{{- fail "Input is required" }}
{{- end }}
OK!
```

Expected: exits with a non-zero code and prints `Input is required`.

**Why it matters:** validates guard clauses used in production templates to catch misconfigured pipelines early.

---

### 06 – `readfile` + format conversions (`toYaml`, `toJson`)

**What it tests:** the `readfile`, `fromJson`, `toYaml`, and `toJson` built-in functions.

Folder: `06_readfile_and_conversions/`

```sh
tmpl8 -i input.json template.tmpl
```

`input.json`
```json
{ "payload": "data.json" }
```

`data.json` (read at render time by the template)
```json
{ "key": "value", "slice": [1, 2, 3] }
```

`template.tmpl`
```yaml
# read from disk and convert to YAML
output:
  {{- readfile .payload | fromJson | toYaml | nindent 2 }}
```

Expected output:
```yaml
# read from disk and convert to YAML
output:
  key: value
  slice:
    - 1
    - 2
    - 3
```

**Why it matters:** covers the most commonly used tmpl8-specific functions; also shows how templates can reach beyond the input object.

---

### 07 – Multiple templates chained together

**What it tests:** passing several template files in one invocation (they are rendered in order, sharing `define`/`block` helpers).

Folder: `07_chained_templates/`

```sh
tmpl8 -i input.yaml header.tmpl body.tmpl footer.tmpl
```

`input.yaml`
```yaml
title: Release Notes
version: 2.0.0
```

`header.tmpl`
```
=== {{ .title }} ===
```

`body.tmpl`
```
Version: {{ .version }}
```

`footer.tmpl`
```
=== end ===
```

Expected output:
```
=== Release Notes ===
Version: 2.0.0
=== end ===
```

**Why it matters:** exercises the multi-template chaining feature and the `@filelist` pattern, which are central to the Kubernetes ConfigMap workflow.

---

### 08 – Kubernetes ConfigMap generator (end-to-end)

**What it tests:** a realistic CI/CD scenario — generating a `kustomization.yaml` from a `tree`-style JSON input, mirroring the `examples/1_configmap` walkthrough.

Folder: `08_k8s_configmap/`

```sh
tree -J -L 2 --noreport > tree.json
tmpl8 -i tree.json kustomization.tmpl > kustomization.yaml
```

`tree.json` (output of `tree -J -L 2 --noreport`)
```json
[
  {"type":"directory","name":".","contents":[
    {"type":"directory","name":"app1","contents":[
      {"type":"file","name":"0_serviceaccount.yaml"},
      {"type":"file","name":"1_deployment.yaml"}
    ]},
    {"type":"directory","name":"worker1","contents":[
      {"type":"file","name":"0_serviceaccount.yaml"},
      {"type":"file","name":"1_cronjob.yaml"}
    ]}
  ]}
]
```

`kustomization.tmpl` (same template as in `examples/1_configmap`)
```gotmpl
namespace: default

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
{{- range $root := . }}
{{- if eq .name "." }}
{{- range $dirs := .contents }}
{{- if eq .type "directory" }}
  - name: {{.name}}
    files:
{{- range $file := .contents }}
{{- if eq .type "file" }}
    - {{ .name }}={{ $dirs.name }}/{{ .name }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

resources:
```

Expected output:
```yaml
namespace: default

generatorOptions:
  disableNameSuffixHash: true

configMapGenerator:
  - name: app1
    files:
    - 0_serviceaccount.yaml=app1/0_serviceaccount.yaml
    - 1_deployment.yaml=app1/1_deployment.yaml
  - name: worker1
    files:
    - 0_serviceaccount.yaml=worker1/0_serviceaccount.yaml
    - 1_cronjob.yaml=worker1/1_cronjob.yaml

resources:
```

**Why it matters:** a realistic CI/CD use-case — generating Kubernetes manifests from structured data inside a pipeline.

---

## Coverage summary

| # | Input type | Template feature | Relevance |
|---|-----------|-----------------|----------------------|
| 01 | inline `?` | basic rendering | smoke test |
| 02 | JSON file | field access | build metadata injection |
| 03 | YAML file | Sprig filters (`upper`) | manifest generation |
| 04 | `-z` (none) | `version` built-in | pipeline diagnostics |
| 05 | `-z` (none) | `fail` guard | pipeline safety |
| 06 | JSON + `readfile` | `fromJson`, `toYaml`, `nindent` | cross-file data wrangling |
| 07 | YAML file | multi-template chain | shared `define`/`block` helpers |
| 08 | JSON (tree) | range, nested loops | K8s ConfigMap CI/CD pipeline |

---

## Adding a new test

1. Create a new numbered sub-folder: `tests/NN_short_description/`
2. Add input file(s), template file(s), and an `expected.txt` (or `expected.yaml`) with the exact expected output
3. Document the test in this README following the pattern above
4. Optionally add an entry to `run_all.sh`
