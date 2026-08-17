# Integration Testing

The integration tests connect to a real Kubernetes cluster. They are gated
behind the `integration` build tag so they never run during normal `go test ./...`.

## Running the tests

```bash
go test -v -tags integration ./...
```

The test (`TestLoadK8s_Integration`) will:
1. Connect to the cluster using your current kubeconfig (or in-cluster config).
2. Create (or update) the ConfigMap `default/tmpl8-test`.
3. Call `loadK8s` and verify the returned entries match.
4. Delete the ConfigMap on cleanup (unless it already existed before the test).

---

## Option A – Use an existing cluster

Make sure `kubectl` is configured and pointing at the right cluster:

```bash
kubectl cluster-info          # verify connectivity
kubectl get ns default        # verify the default namespace exists
```

Then run the tests:

```bash
go test -v -tags integration ./...
```

---

## Option B – Local cluster with kind

[kind](https://kind.sigs.k8s.io/) spins up a Kubernetes cluster inside Docker containers.

### Install kind

```bash
# macOS / Linux via Go
go install sigs.k8s.io/kind@latest

# or via Homebrew (macOS)
brew install kind

# or download the binary directly
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64
chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
```

### Create a cluster

```bash
kind create cluster --name tmpl8-test
```

This automatically updates your kubeconfig and sets the current context to
`kind-tmpl8-test`.

### Verify

```bash
kubectl cluster-info --context kind-tmpl8-test
```

### Run the integration tests

```bash
go test -v -tags integration ./...
```

### Tear down

```bash
kind delete cluster --name tmpl8-test
```

---

## Manually applying the ConfigMap (optional)

The test manages the ConfigMap itself, but if you want to pre-apply it:

```bash
kubectl apply -f tests/k8s/tmpl8-test-configmap.yaml
```

And verify:

```bash
kubectl get configmap tmpl8-test -n default -o yaml
```
