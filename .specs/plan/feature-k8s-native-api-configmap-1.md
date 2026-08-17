---
goal: Replace kubectl shell-out in loadK8s() with native Kubernetes client-go API calls
version: "1.0"
date_created: 2026-02-18
last_updated: 2026-02-18
owner: AB-Lindex/tmpl8
status: 'In Progress'
tags: [feature, kubernetes, refactor, dependency]
---

# Introduction

![Status: Planned](https://img.shields.io/badge/status-Planned-blue)

Replace the `kubectl` subprocess invocation in `load.go:loadK8s()` with a direct Kubernetes API call using `k8s.io/client-go`. This removes the runtime dependency on the `kubectl` binary, makes ConfigMap loading faster and more portable, and allows `exec.go` to be deleted entirely.

## 1. Requirements & Constraints

- **REQ-001**: The `k8s:<namespace>/<configmap>` input syntax must continue to work identically for end users — no CLI or behaviour change.
- **REQ-002**: Authentication must auto-detect in-cluster service account token first, then fall back to kubeconfig (`$KUBECONFIG` / `~/.kube/config`).
- **REQ-003**: All key/value pairs from `ConfigMap.data` must be returned as `[]entry` in the same format as today (`k8s:<ns>/<cm>/<key>`).
- **REQ-004**: `exec.go` must be deleted once no call sites remain.
- **REQ-005**: Error messages must include namespace and ConfigMap name for debuggability.
- **SEC-001**: No new credentials or credential-storage mechanisms are introduced; existing kubeconfig or in-cluster token is reused.
- **SEC-002**: Trivy scan of the resulting binary must produce zero high or critical vulnerabilities from new dependencies.
- **SEC-003**: Minimum RBAC for in-cluster use: a `Role` granting only `get` on `configmaps` in the target namespace.
- **CON-001**: Binary size will increase due to `k8s.io/client-go` and its transitive dependencies; this is accepted.
- **CON-002**: Go module version must remain compatible with `go 1.26.0` as declared in `go.mod`.
- **GUD-001**: Use `clientcmd.NewNonInteractiveDeferredLoadingClientConfig` with `rest.InClusterConfig()` fallback — the idiomatic `client-go` pattern.
- **GUD-002**: Create the `kubernetes.Clientset` once per `loadK8s` call for MVP; caching can be added later.
- **PAT-001**: Follow existing error-wrapping style in `load.go` using `fmt.Errorf("...: %w", err)`.

Update the status tag of each task as work progresses using the status tags defined in Section 2.

## 1.1. Repository Context

- **Repository Type**: Single-Product
- **PRD**: `/.specs/PRD.md` — not present; feature spec used as source of truth
- **Features**: `/.specs/features/k8s-native-api-configmap.md`
- **Technology Stack**: Go 1.26, `k8s.io/client-go`, `github.com/rs/zerolog`
- **Cross-Product Dependencies**: None

## 2. Implementation Steps

### Implementation Phase 1 — Add client-go dependency

- **GOAL-001**: Introduce `k8s.io/client-go` and its required companion modules into the Go module, ensuring the project compiles cleanly before any logic is changed.

- **TASK-001**: Add `k8s.io/client-go`, `k8s.io/api`, and `k8s.io/apimachinery` to `go.mod` and generate `go.sum`. `[✅ Completed: 2026-02-18]`
  - Files: `go.mod`, `go.sum`
  - Action: Run `go get k8s.io/client-go@latest` — this will pull in `k8s.io/api` and `k8s.io/apimachinery` as transitive dependencies automatically. Then run `go mod tidy`.
  - Estimated effort: 15 minutes
  - Note: `client-go` major version must align with the target cluster API version; use the latest stable release (e.g., `v0.32.x` for Kubernetes 1.32).

- **TASK-002**: Verify the project compiles after adding the dependency — `task build` must succeed with no errors. `[✅ Completed: 2026-02-18]`
  - Files: none (verification only)
  - Dependencies: TASK-001 must be completed first

### Implementation Phase 2 — Rewrite loadK8s()

- **GOAL-002**: Replace the `execCmd("kubectl", ...)` call in `load.go:loadK8s()` with a direct `client-go` API call, removing the `configmap` struct, and retaining identical return behaviour.

- **TASK-003**: Create a new file `k8s.go` in the package root with the following imports: `context`, `fmt`, `strings`, `k8s.io/client-go/kubernetes`, `k8s.io/client-go/rest`, `k8s.io/client-go/tools/clientcmd`, `k8s.io/apimachinery/pkg/apis/meta/v1`. `[✅ Completed: 2026-02-18]`
  - Files: `k8s.go` (new)
  - Dependencies: TASK-001
  - Note: All Kubernetes-specific logic is isolated in this new file; `load.go` retains only the `return loadK8s(fn[4:])` call site and is otherwise unchanged.

- **TASK-004**: Implement `loadK8s()` in `k8s.go` using the following exact logic: `[✅ Completed: 2026-02-18]`
  - Files: `k8s.go` (new)
  - Dependencies: TASK-003
  - Logic:
    1. Split `fn` on `/` and validate exactly 2 parts (unchanged).
    2. Call `rest.InClusterConfig()` to attempt in-cluster auth.
    3. On error, fall back to `clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()`.
    4. If both fail, return `fmt.Errorf("failed to build kubernetes client config: %w", err)`.
    5. Call `kubernetes.NewForConfig(config)` to create the clientset.
    6. Call `clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})`.
    7. On error, return `fmt.Errorf("failed to get configmap %s/%s: %w", namespace, name, err)`.
    8. Iterate `cm.Data` and build `[]entry` with name `fmt.Sprintf("k8s:%s/%s/%s", namespace, name, key)`.

- **TASK-005**: Remove the `configmap` struct and the `loadK8s()` function from `load.go` — both are superseded by `k8s.go`. `[✅ Completed: 2026-02-18]`
  - Files: `load.go`
  - Dependencies: TASK-004
  - Note: Also remove the `encoding/json` import from `load.go` if it is no longer used elsewhere in that file. The `return loadK8s(fn[4:])` call site in `load()` remains — it now resolves to the function defined in `k8s.go`.

### Implementation Phase 3 — Delete exec.go

- **GOAL-003**: Remove `exec.go` entirely now that `execCmd` has no call sites remaining.

- **TASK-006**: Verify `execCmd` and `execPipe` have zero remaining call sites — search the codebase with `grep -r "execCmd\|execPipe" .` before deleting. `[✅ Completed: 2026-08-17]`
  - Files: none (verification only)
  - Dependencies: TASK-004

- **TASK-007**: Delete `exec.go` from the repository. `[✅ Completed: 2026-08-17]`
  - Files: `exec.go` (deleted)
  - Dependencies: TASK-006
  - Action: `git rm exec.go`

- **TASK-008**: Run `task build` and confirm zero compile errors after deletion. `[✅ Completed: 2026-08-17]`
  - Dependencies: TASK-007

### Implementation Phase 4 — Testing & Documentation

- **GOAL-004**: Validate correctness, ensure existing tests pass, and update documentation to reflect that `kubectl` is no longer required.

- **TASK-009**: Run the full test suite (`tests/run_all.sh`) and confirm all tests pass. Pay particular attention to `tests/08_k8s_configmap/`. `[✅ Completed: 2026-08-17]`
  - Files: `tests/run_all.sh`
  - Dependencies: TASK-008
  - Note: `08_k8s_configmap` may require a live cluster or a mock; if the test currently skips without `kubectl`, verify it still skips (or passes) gracefully.

- **TASK-010**: Update `README.md` to remove any implication that `kubectl` must be installed, and add a note that `k8s:` loading now uses the native API. `[✅ Completed: 2026-08-17]`
  - Files: `README.md`
  - Target location: "Template argument forms" table row for `k8s:ns/name` — append note: "No `kubectl` required; uses in-cluster token or kubeconfig."

- **TASK-011**: Add the minimum RBAC `Role` manifest as a documentation example or in `examples/` for users running `tmpl8` in-cluster. `[✅ Completed: 2026-08-17]`
  - Files: `examples/README.md` or a new `examples/rbac-tmpl8.yaml`
  - Content:
    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    metadata:
      name: tmpl8-configmap-reader
    rules:
      - apiGroups: [""]
        resources: ["configmaps"]
        verbs: ["get"]
    ```

**Status Tags:**
- `[✅ Completed: YYYY-MM-DD]` - Task finished
- `[⏳ In Progress]` - Currently being worked on
- `[📋 Planned]` - Not yet started
- `[⚠️ Blocked: reason]` - Cannot proceed due to dependency
- `[❌ Cancelled: reason]` - Task no longer needed

## 3. Alternatives

- **ALT-001**: Keep `exec.go` but fix the hardcoded `'kubectl'` error messages to use `name` — rejected because it does not remove the runtime `kubectl` dependency, which is the primary problem.
- **ALT-002**: Use `dynamic.Interface` from `client-go` (schema-less client) instead of the typed `CoreV1().ConfigMaps()` — rejected because the typed client is simpler, avoids manual JSON unmarshalling, and the `ConfigMap` type is stable.
- **ALT-003**: Use the raw Kubernetes HTTP API without `client-go` to keep binary size small — rejected because it requires reimplementing authentication, TLS, and token refresh logic that `client-go` already provides correctly.
- **ALT-004**: Build a `--no-k8s` tag to compile `client-go` out optionally — rejected as premature optimisation; can be revisited if binary size becomes a concrete concern.

## 4. Dependencies

- **DEP-001**: `k8s.io/client-go` — Kubernetes Go client library (in-cluster + kubeconfig auth, typed API).
- **DEP-002**: `k8s.io/api` — Kubernetes API type definitions (pulled in transitively by `client-go`).
- **DEP-003**: `k8s.io/apimachinery` — Kubernetes API machinery (`metav1.GetOptions`, etc.; pulled in transitively).

## 5. Files

- **FILE-001**: `load.go` — `loadK8s()` function and `configmap` struct removed; `encoding/json` import removed if unused. The `return loadK8s(fn[4:])` call site is kept unchanged.
- **FILE-007**: `k8s.go` (new) — Contains `loadK8s()` rewritten with `client-go`; all Kubernetes-specific imports live here.
- **FILE-002**: `exec.go` — **Deleted** entirely.
- **FILE-003**: `go.mod` — New `require` entries for `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery`.
- **FILE-004**: `go.sum` — Updated with new dependency hashes.
- **FILE-005**: `README.md` — Minor wording update to `k8s:ns/name` entry.
- **FILE-006**: `examples/rbac-tmpl8.yaml` (new) — Minimum RBAC Role for in-cluster use.

## 6. Testing

- **TEST-001**: `task build` succeeds after each phase (TASK-002, TASK-008).
- **TEST-002**: `tests/run_all.sh` passes all existing tests without regression (TASK-009).
- **TEST-003**: `tests/08_k8s_configmap/` — if it tests the `k8s:` prefix, ensure it works against a live or mocked cluster after the rewrite.
- **TEST-004**: Manual smoke test: run `tmpl8 -i k8s:<ns>/<cm> '?{{ . | toYaml }}'` against a real cluster using both in-cluster (pod exec) and kubeconfig (local) authentication.
- **TEST-005**: Trivy scan of the compiled binary — `trivy fs --scanners vuln ./bin/tmpl8` — must return zero high or critical findings.

## 7. Risks & Assumptions

- **RISK-001**: Binary size increase from `client-go` transitive dependencies may be significant (10–30 MB). Mitigation: document the size change in the PR; revisit build tags if it becomes a hard constraint.
- **RISK-002**: `client-go` version may need to be pinned to match the oldest Kubernetes cluster version in use at Lindex. Mitigation: verify `client-go` compatibility matrix; `client-go` is backwards compatible with Kubernetes clusters within ±3 minor versions.
- **RISK-003**: `tests/08_k8s_configmap/` may rely on a live cluster not available in CI. Mitigation: the test can remain as a manual/integration-only step; add a skip guard if `kubectl` / cluster is unavailable.
- **ASSUMPTION-001**: The existing in-cluster auth (service account token) and kubeconfig auth paths in `client-go` are sufficient; no custom auth providers (e.g., exec plugins, OIDC) need to be supported at this time.
- **ASSUMPTION-002**: The target Kubernetes clusters run version 1.29 or newer, which is within the `client-go v0.32.x` compatibility window.

## 8. Related Specifications / Further Reading

- [Feature spec: k8s-native-api-configmap](/.specs/features/k8s-native-api-configmap.md)
- [client-go authentication documentation](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration)
- [client-go out-of-cluster configuration example](https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration)
- [client-go / Kubernetes version compatibility matrix](https://github.com/kubernetes/client-go#compatibility-matrix)
