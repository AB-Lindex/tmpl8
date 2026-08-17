---
type: "feature"
feature: "k8s-native-api-configmap"
repository_type: "single-product"
status: "proposed"
priority: "medium"
complexity: "standard"
technology_stack: ["go"]
cloud_provider: "azure"
onpremise_integration: true
ci_cd_system: "azure-devops"
deployment_targets: ["onprem-k8s", "azure-aks"]
lindex_components:
  authentication: "in-cluster serviceaccount / kubeconfig"
  api_gateway: "none"
  security_scanning: "trivy"
  api_docs: "none"
  service_catalog: "backstage"
  team_docs: "confluence"
infrastructure: ["kubernetes"]
external_services: []
on_premises_services: []
kubernetes_multi_instance: "compatible"
observability: "basic"
related_prd: "none"
cross_product_dependencies: []
---

# Feature: Native Kubernetes API for ConfigMap Loading

## Problem Statement

`tmpl8` currently loads Kubernetes ConfigMaps by shelling out to `kubectl` via `exec.go`. This approach has several drawbacks:

- `kubectl` must be installed and available in `$PATH` wherever `tmpl8` runs (CI runners, containers, etc.)
- It spawns a child process and parses stdout/stderr, which is fragile and slow
- Error messages in `exec.go` are hardcoded to `'kubectl'` despite the generic function signature
- The `exec.go` file and `execCmd`/`execPipe` helpers serve no other purpose, adding dead weight

Replacing the shell-out with the official Kubernetes Go client (`client-go`) removes the `kubectl` runtime dependency and makes ConfigMap loading more robust, faster, and portable.

## User Stories

- As a tmpl8 user running inside a Kubernetes Pod, I want ConfigMap loading to work without `kubectl` being present in the container image so that my image stays minimal.
- As a tmpl8 user on a developer workstation, I want ConfigMap loading to continue working via my local kubeconfig so that there is no behaviour change in daily use.
- As a CI/CD pipeline operator, I want tmpl8 to authenticate to Kubernetes using the pipeline's service account token so that no extra tooling needs to be installed.

## Requirements

### Functional

1. `k8s:<namespace>/<configmap>` input syntax continues to work identically for end users.
2. Authentication automatically selects the best available mechanism:
   - In-cluster service account (`/var/run/secrets/kubernetes.io/serviceaccount/`) when running inside a Pod.
   - kubeconfig file (default `~/.kube/config`, or `$KUBECONFIG`) when running outside a Kubernetes cluster.
3. All key/value entries from `ConfigMap.data` are returned as `[]entry`, exactly as today.
4. `exec.go` is removed from the project once no longer needed.
5. Error messages clearly identify the namespace and ConfigMap name on failure.

### Non-Functional

- **Performance**: Direct API call eliminates the process-spawn overhead of `kubectl`.
- **Portability**: No `kubectl` binary required at runtime.
- **Security**:
  - No new credentials are introduced; the existing kubeconfig or in-cluster token is used.
  - Trivy container scanning must pass with no high/critical vulnerabilities after new dependencies are added.
  - Service account used inside a Pod should have only `get` on `configmaps` in the relevant namespace (principle of least privilege).
- **Kubernetes Multi-Instance Support**: Stateless read-only operation; safe for concurrent Pod instances.
- **Backwards Compatibility**: No change to CLI flags, input syntax, or output format.

## Technical Design

- **Architecture**: Replace `loadK8s()` in `load.go` to call the Kubernetes API directly instead of via `execCmd`.
- **Technology Stack**: Go, `k8s.io/client-go`
- **New Dependency**: Add `k8s.io/client-go` (and its transitive deps `k8s.io/api`, `k8s.io/apimachinery`) to `go.mod`.
- **Authentication**: Use `client-go`'s `clientcmd.BuildConfigFromFlags` with in-cluster fallback via `rest.InClusterConfig()`. The standard helper `clientcmd.NewNonInteractiveDeferredLoadingClientConfig` handles both cases automatically.
- **ConfigMap retrieval**: Use `clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})` — returns a typed `*v1.ConfigMap` struct; access `.Data` directly, no JSON unmarshalling needed.
- **Removal of exec.go**: Once `loadK8s` no longer calls `execCmd`, `exec.go` (and the unused `outOk`/`errOk` variables) can be deleted entirely.
- **Error handling**: Wrap Kubernetes API errors with namespace/name context using `fmt.Errorf`.

### Sketch of updated `loadK8s` in load.go

```go
func loadK8s(fn string) ([]entry, error) {
    parts := strings.Split(fn, "/")
    if len(parts) != 2 {
        return nil, fmt.Errorf("k8s-format is 'k8s:namespace/configname', not 'k8s:%s'", fn)
    }
    namespace, name := parts[0], parts[1]

    config, err := rest.InClusterConfig()
    if err != nil {
        // Fall back to kubeconfig
        loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
        configOverrides := &clientcmd.ConfigOverrides{}
        config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
        if err != nil {
            return nil, fmt.Errorf("failed to build kubernetes client config: %w", err)
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
    }

    cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get configmap %s/%s: %w", namespace, name, err)
    }

    var result []entry
    for key, data := range cm.Data {
        result = append(result, entry{fmt.Sprintf("k8s:%s/%s/%s", namespace, name, key), data})
    }
    return result, nil
}
```

### Files Changed

| File | Change |
|------|--------|
| `load.go` | Rewrite `loadK8s()` to use `client-go`; remove `configmap` struct; add new imports |
| `exec.go` | **Delete** — no longer needed |
| `go.mod` / `go.sum` | Add `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery` |

### Minimum Required RBAC (when running in-cluster)

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

## Implementation Phases

1. **MVP**:
   - Add `client-go` dependency.
   - Rewrite `loadK8s()` with in-cluster + kubeconfig fallback auth.
   - Delete `exec.go`.
   - Update `go.mod` / `go.sum`.
   - Verify existing tests pass (`tests/08_k8s_configmap/` or equivalent).

2. **Enhancement** (optional, future):
   - Cache the `kubernetes.Clientset` across multiple `k8s:` loads in the same run to avoid recreating it per call.
   - Support `k8s:<namespace>/<configmap>/<key>` to load a single key rather than all keys.
   - Add a `--kubeconfig` / `--kube-context` CLI flag to override the default config path.

## Integration

- **README Impact**: The "Template argument forms" table entry for `k8s:ns/name` remains unchanged. A note can be added that `kubectl` is no longer required.
- **Dependencies**: `k8s.io/client-go` is a large module; binary size will increase. Evaluate whether a slimmed build tag or lazy init is appropriate.
- **Cross-Product**: None — tmpl8 is a single standalone CLI tool.
