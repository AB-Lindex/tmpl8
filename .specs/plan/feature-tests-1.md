---
- **GOAL-003**: Verify the suite runs green and update the README.
version: 1.0
status: 'In progress'
tags: [feature, tests, validation]
---
# Introduction

![Status: In progress](https://img.shields.io/badge/status-In%20progress-yellow)

Create the fixture files (inputs, templates, expected outputs) for each of the 8 test cases documented in `tests/README.md`, plus a `run_all.sh` script that executes every case and validates its output — exiting non-zero on any failure.

## 1. Requirements & Constraints

- **REQ-001**: Each test case must be fully self-contained inside its own sub-folder under `tests/`.
- **REQ-002**: Every test must have an `expected.txt` or `expected.yaml` file whose content is the exact expected output.
- **REQ-003**: `run_all.sh` must print a `PASS` / `FAIL` line per test and a final summary count.
- **REQ-004**: `run_all.sh` must exit with code `0` only when all tests pass.
- **REQ-005**: Test 05 (fail guard) must validate exit code, not stdout content.
- **CON-001**: No external tools beyond `tmpl8`, `sh`, and `diff` are required to run the suite.
- **CON-002**: Template and input files must match exactly what is documented in `tests/README.md`.

## 1.1. Repository Context

- **Repository Type**: Single-Product
- **PRD**: `/.specs/PRD.md` (not yet present — no PRD found)
- **Features**: `/.specs/features/`
- **Technology Stack**: Go (binary), POSIX shell (test runner)

## 2. Implementation Steps

### Implementation Phase 1 — Fixture files

- **GOAL-001**: Create all input, template, and expected-output files for each of the 8 test cases.

- **TASK-001**: Create `tests/01_inline_smoke/` fixture `[✅ Completed: 2026-02-18]`
  - No input or template files needed (inline-only test)
  - Files: `tests/01_inline_smoke/expected.txt`
  - Content of `expected.txt`: `Hello, world!`

- **TASK-002**: Create `tests/02_json_to_text/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/02_json_to_text/input.json`, `tests/02_json_to_text/template.tmpl`, `tests/02_json_to_text/expected.txt`

- **TASK-003**: Create `tests/03_yaml_to_yaml/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/03_yaml_to_yaml/input.yaml`, `tests/03_yaml_to_yaml/template.tmpl`, `tests/03_yaml_to_yaml/expected.yaml`

- **TASK-004**: Create `tests/04_no_input/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/04_no_input/template.tmpl`, `tests/04_no_input/expected.txt`
  - Note: expected output contains a literal `{{ version }}` placeholder — the script must compare only the non-version lines, or skip the version line with a pattern match.

- **TASK-005**: Create `tests/05_fail_on_invalid/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/05_fail_on_invalid/template.tmpl`
  - No `expected.txt` — validation checks exit code only (must be non-zero).

- **TASK-006**: Create `tests/06_readfile_and_conversions/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/06_readfile_and_conversions/input.json`, `tests/06_readfile_and_conversions/data.json`, `tests/06_readfile_and_conversions/template.tmpl`, `tests/06_readfile_and_conversions/expected.yaml`
  - Note: `tmpl8` must be run from inside the test folder so `readfile` can resolve `data.json` by relative path.

- **TASK-007**: Create `tests/07_chained_templates/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/07_chained_templates/input.yaml`, `tests/07_chained_templates/header.tmpl`, `tests/07_chained_templates/body.tmpl`, `tests/07_chained_templates/footer.tmpl`, `tests/07_chained_templates/expected.txt`

- **TASK-008**: Create `tests/08_k8s_configmap/` fixtures `[✅ Completed: 2026-02-18]`
  - Files: `tests/08_k8s_configmap/tree.json`, `tests/08_k8s_configmap/kustomization.tmpl`, `tests/08_k8s_configmap/expected.yaml`

### Implementation Phase 2 — Validation script

- **GOAL-002**: Write `tests/run_all.sh` that runs each test case, compares output to expected, and reports results.

- **TASK-009**: Write `tests/run_all.sh` `[✅ Completed: 2026-02-18]`
  - Files: `tests/run_all.sh`
  - The script must:
    1. Resolve `tmpl8` from `$PATH`, falling back to `../bin/tmpl8`.
    2. Define a `run_test NAME CMD EXPECTED_FILE` helper that runs `CMD`, diffs stdout against `EXPECTED_FILE`, and prints `PASS` or `FAIL [diff]`.
    3. Define a `run_fail_test NAME CMD` helper that runs `CMD` and checks that the exit code is non-zero.
    4. For test 04, strip or ignore the version line before diffing (use `grep -v "^tmpl8 version:"` on both actual and expected).
    5. Print a final summary: `N/8 tests passed`.
    6. Exit `0` only if all tests passed.
  - Each test must `cd` into its own folder before invoking `tmpl8` so relative paths (`readfile`) work correctly.
  - Make the file executable (`chmod +x`).

### Implementation Phase 3 — Smoke-run & documentation update


- **GOAL-003**: Verify the suite runs green and update the README.

- **TASK-010**: Run `tests/run_all.sh` and confirm all 8 tests pass `[✅ Completed: 2026-08-17]`
  - Prerequisites: TASK-001 through TASK-009 complete, `tmpl8` binary available.
  - Testing: all 8 lines must read `PASS`.

- **TASK-011**: Remove the "will be placed" placeholder sentence from `tests/README.md` now that `run_all.sh` exists `[✅ Completed: 2026-08-17]`
  - Files: `tests/README.md`

## 3. Alternatives

- **ALT-001**: Use a Go test harness (`go test`) instead of a shell script — rejected to keep the suite dependency-free and easy to run without a Go toolchain.
- **ALT-002**: Golden-file framework (e.g., `cupaloy`) — rejected for the same reason; `diff` is sufficient.

## 4. Dependencies

- **DEP-001**: `tmpl8` binary — built from source (`go build`) or downloaded from GitHub Releases.
- **DEP-002**: POSIX `diff` and `sh` — available on all target platforms.

## 5. Files

- **FILE-001**: `tests/run_all.sh` — main validation script
- **FILE-002**: `tests/01_inline_smoke/expected.txt`
- **FILE-003**: `tests/02_json_to_text/{input.json,template.tmpl,expected.txt}`
- **FILE-004**: `tests/03_yaml_to_yaml/{input.yaml,template.tmpl,expected.yaml}`
- **FILE-005**: `tests/04_no_input/{template.tmpl,expected.txt}`
- **FILE-006**: `tests/05_fail_on_invalid/template.tmpl`
- **FILE-007**: `tests/06_readfile_and_conversions/{input.json,data.json,template.tmpl,expected.yaml}`
- **FILE-008**: `tests/07_chained_templates/{input.yaml,header.tmpl,body.tmpl,footer.tmpl,expected.txt}`
- **FILE-009**: `tests/08_k8s_configmap/{tree.json,kustomization.tmpl,expected.yaml}`

## 6. Testing

- **TEST-001**: `run_all.sh` exits `0` with 8/8 PASS lines on a clean checkout with `tmpl8` on `$PATH`.
- **TEST-002**: Deleting one `expected.txt` causes the corresponding test to FAIL and the script to exit non-zero.
- **TEST-003**: Passing a corrupt input to test 05 confirms the non-zero exit code is caught correctly.

## 7. Risks & Assumptions

- **RISK-001**: The `version` output in test 04 changes between builds — mitigated by filtering that line in the diff.
- **RISK-001**: `readfile` resolves paths relative to the working directory, not the template file — mitigated by `cd`-ing into each test folder before running `tmpl8`.
- **ASSUMPTION-001**: `tmpl8` is either on `$PATH` or available at `../bin/tmpl8` relative to the `tests/` folder.

## 8. Related Specifications / Further Reading

- [tests/README.md](../../tests/README.md) — full description of each test case with inputs, templates, and expected outputs
- [examples/README.md](../../examples/README.md) — walkthrough that tests 01 and 08 are based on
- [Go text/template docs](https://pkg.go.dev/text/template)
- [Sprig function reference](https://masterminds.github.io/sprig/)
