# Kubernetes Operator Build Notes

## Status

The Kubernetes operator code is present but requires code generation before it can be compiled.

## Required Steps

1. **Install controller-gen**:
   ```bash
   go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
   ```

2. **Generate DeepCopy methods**:
   ```bash
   cd k8s/operator
   controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
   ```

3. **Generate CRD manifests**:
   ```bash
   controller-gen crd:trivialVersions=true rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
   ```

4. **Uncomment registration** in `api/v1/types.go`:
   ```go
   func init() {
       SchemeBuilder.Register(&Agent{}, &AgentList{})
       SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
       SchemeBuilder.Register(&Policy{}, &PolicyList{})
   }
   ```

## Current State

- ✅ Controller logic is complete
- ✅ CRD type definitions are complete
- ✅ Kubernetes dependencies added
- ❌ DeepCopy methods need generation
- ❌ CRD YAML manifests need generation

## For GitHub Release

The operator directory can be included as-is with a note that users need to run code generation. Alternatively:
- Add a `Makefile` with generate targets
- Add a `.gitignore` entry for `zz_generated.deepcopy.go`
- Document in README that operator requires code generation step

## Build Without Operator

To build everything except the operator:

```bash
# Build core packages
go build $(go list ./... | grep -v '/k8s/operator')

# Or use build tags
go build -tags='!operator' ./...
```
