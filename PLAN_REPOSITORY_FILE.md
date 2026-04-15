# RepositoryFile Managed Resource — Implementierungsplan

## Problem

Dateien in GitLab-Repos über Crossplane verwalten. Hauptprobleme:
1. Standard-Reconcile-Loop flutet GitLab API mit GetFile-Calls
2. Manche Dateien sollen einmalig erstellt und nie wieder angefasst werden
3. Verschiedene Dateien brauchen verschiedene Reconcile-Frequenzen (1h vs 8h)

## Entscheidungen

| Frage | Entscheidung |
|---|---|
| Reconcile-Kontrolle | Custom `reconcileInterval` per Resource + `createOnly` Flag (setzt intern managementPolicies) |
| Content-Quelle | Inline `content` + `contentSecretRef` (Secret als Kubernetes-Objekt für Dateiinhalt) |
| Project-Referenz | `projectIdRef`/`projectIdSelector` wie bei allen anderen Resources im Provider |
| Commit Messages | Getrennt pro Action: `createCommitMessage`, `updateCommitMessage`, `deleteCommitMessage` |
| Große Dateien | 1MB etcd-Limit reicht. S3-Referenz wäre Future Work |
| createOnly | Bool-Flag `createOnly: true` das intern `managementPolicies: ["Create", "Delete", "Observe"]` setzt — User-freundliche Abstraktion über managementPolicies |
| Konfliktregel | Wenn `createOnly` und explizite `managementPolicies` widersprechen: bevorzugt Validation Error, sonst gewinnen `managementPolicies` |

## Reconcile-Kontrolle im Detail

### Per-Resource `reconcileInterval`

Feld `spec.forProvider.reconcileInterval` (z.B. `"1h"`, `"8h"`).

Implementierung im `Observe()`:
- Status-Feld `status.atProvider.lastObserveTime` speichert Zeitpunkt des letzten echten API-Calls
- Wenn `now - lastObserveTime < reconcileInterval` → return `ResourceExists: true, ResourceUpToDate: true` ohne API-Call
- Erst wenn Interval abgelaufen → tatsächlich `GetFile` aufrufen
- Default: Controller-globales Poll-Interval (wenn Feld nicht gesetzt)

### `createOnly` Flag

Feld `spec.forProvider.createOnly` (bool, default false).

Implementierung:
- Wenn `createOnly: true` → Controller verhält sich wie `managementPolicies: ["Create", "Delete", "Observe"]`
- Konkret: `Observe()` prüft nur Existenz (kein Content-Vergleich), `Update()` wird nie aufgerufen
- User muss managementPolicies nicht kennen
- Kann mit `reconcileInterval` kombiniert werden (z.B. `createOnly: true` + `reconcileInterval: "24h"` = einmal am Tag prüfen ob Datei noch existiert)

### Konfliktregel: `createOnly` vs `managementPolicies`

Wenn ein User sowohl `createOnly: true` als auch explizite `managementPolicies` setzt, die nicht zu
`["Create", "Delete", "Observe"]` passen:

- bevorzugt: Validation Error im CRD / Admission-Pfad
- fallback falls saubere Validation zu aufwendig ist: explizite `managementPolicies` gewinnen
- nie stillschweigend `managementPolicies` überschreiben

## CRD Design

```yaml
apiVersion: projects.gitlab.m.crossplane.io/v1alpha1
kind: RepositoryFile
metadata:
  name: my-readme
spec:
  forProvider:
    # --- Identifikation ---
    projectId: 12345
    projectIdRef:                         # Cross-Resource-Referenz auf Project CR
      name: my-project
    projectIdSelector:                    # Label-basierte Selektion
      matchLabels:
        team: platform

    filePath: "README.md"                 # +immutable
    branch: "main"                        # +immutable

    # --- Content ---
    content: "# My Project"              # inline, mutually exclusive mit contentSecretRef
    contentSecretRef:                     # Content aus Secret laden
      name: my-file-secret
      key: readme-content
    encoding: "text"                      # "text" (default) | "base64"

    # --- Git Commit Metadata (pro Action) ---
    createCommitMessage: "feat: initial file creation by crossplane"
    updateCommitMessage: "chore: update file content via crossplane"
    deleteCommitMessage: "chore: remove crossplane-managed file"
    authorEmail: "crossplane@example.com"
    authorName: "Crossplane"
    executeFilemode: false

    # --- Reconcile-Kontrolle ---
    reconcileInterval: "1h"              # optional, per-resource poll interval
    createOnly: true                     # optional, default false — erstellt Datei einmalig, updated nie

  providerConfigRef:
    name: gitlab-config

status:
  atProvider:
    filePath: "README.md"
    blobId: "abc123"
    commitId: "def456"
    lastCommitId: "ghi789"
    contentSha256: "e3b0c44..."
    size: 42
    lastObserveTime: "2026-03-31T12:00:00Z"
```

## Go Types (Entwurf)

```go
// RepositoryFileParameters define desired state of a GitLab Repository File
type RepositoryFileParameters struct {
    // ProjectID is the ID of the project.
    // +optional
    // +immutable
    ProjectID *int64 `json:"projectId,omitempty"`

    // ProjectIDRef is a reference to a project to retrieve its projectId.
    // +optional
    // +immutable
    ProjectIDRef *xpv1.NamespacedReference `json:"projectIdRef,omitempty"`

    // ProjectIDSelector selects reference to a project to retrieve its projectId.
    // +optional
    ProjectIDSelector *xpv1.NamespacedSelector `json:"projectIdSelector,omitempty"`

    // FilePath is the path of the file in the repository.
    // +immutable
    // +kubebuilder:validation:MinLength=1
    FilePath string `json:"filePath"`

    // Branch is the name of the branch to commit to.
    // +immutable
    // +kubebuilder:validation:MinLength=1
    Branch string `json:"branch"`

    // Content is the file content. Mutually exclusive with ContentSecretRef.
    // +optional
    Content *string `json:"content,omitempty"`

    // ContentSecretRef references a Secret key containing the file content.
    // Mutually exclusive with Content.
    // +optional
    ContentSecretRef *xpv1.LocalSecretKeySelector `json:"contentSecretRef,omitempty"`

    // Encoding is the file encoding: "text" (default) or "base64".
    // +optional
    // +kubebuilder:validation:Enum=text;base64
    // +kubebuilder:default="text"
    Encoding *string `json:"encoding,omitempty"`

    // CreateCommitMessage is the commit message used when creating the file.
    // +optional
    CreateCommitMessage *string `json:"createCommitMessage,omitempty"`

    // UpdateCommitMessage is the commit message used when updating the file.
    // +optional
    UpdateCommitMessage *string `json:"updateCommitMessage,omitempty"`

    // DeleteCommitMessage is the commit message used when deleting the file.
    // +optional
    DeleteCommitMessage *string `json:"deleteCommitMessage,omitempty"`

    // AuthorEmail is the commit author email.
    // +optional
    AuthorEmail *string `json:"authorEmail,omitempty"`

    // AuthorName is the commit author name.
    // +optional
    AuthorName *string `json:"authorName,omitempty"`

    // ExecuteFilemode enables the executable flag on the file.
    // +optional
    ExecuteFilemode *bool `json:"executeFilemode,omitempty"`

    // ReconcileInterval controls how often this resource is reconciled against
    // the GitLab API. Examples: "5m", "1h", "8h". If unset, uses the controller
    // default poll interval.
    // +optional
    ReconcileInterval *string `json:"reconcileInterval,omitempty"`

    // CreateOnly when true, creates the file once and never updates it.
    // The file is still deleted from GitLab when the CR is deleted.
    // Internally sets managementPolicies to ["Create", "Delete", "Observe"].
    // +optional
    // +kubebuilder:default=false
    CreateOnly *bool `json:"createOnly,omitempty"`
}

// RepositoryFileObservation represents observed state of a GitLab Repository File
type RepositoryFileObservation struct {
    FilePath     string      `json:"filePath,omitempty"`
    BlobID       string      `json:"blobId,omitempty"`
    CommitID     string      `json:"commitId,omitempty"`
    LastCommitID string      `json:"lastCommitId,omitempty"`
    SHA256       string      `json:"sha256,omitempty"`
    Size         int64       `json:"size,omitempty"`
    LastObserveTime *metav1.Time `json:"lastObserveTime,omitempty"`
}
```

## Dateien die erstellt/geändert werden

### Neue Dateien (nur namespaced — cluster wird via `make generate` erzeugt)

| # | Datei | Beschreibung |
|---|---|---|
| 1 | `apis/namespaced/projects/v1alpha1/repositoryfile_types.go` | CRD Types |
| 2 | `pkg/namespaced/clients/projects/repositoryfile.go` | GitLab Client Wrapper + Helpers |
| 3 | `pkg/namespaced/clients/projects/repositoryfile_test.go` | Client Helper Tests |
| 4 | `pkg/namespaced/controller/projects/repositoryfiles/controller.go` | Reconciler |
| 5 | `pkg/namespaced/controller/projects/repositoryfiles/controller_test.go` | Controller Tests |
| 6 | `examples/projects/repositoryfile.yaml` | Beispiel-Manifest |

### Geänderte Dateien

| # | Datei | Änderung |
|---|---|---|
| 7 | `apis/namespaced/projects/v1alpha1/register.go` | RepositoryFile + RepositoryFileList registrieren |
| 8 | `pkg/namespaced/clients/projects/fake/fake.go` | Mock-Methods für RepositoryFile |
| 9 | `pkg/namespaced/controller/projects/setup.go` | `repositoryfiles.SetupRepositoryFile` + Gated registrieren |

### Generierte Dateien (via `make generate`)

- `apis/cluster/projects/v1alpha1/zz_repositoryfile_types.go`
- `pkg/cluster/clients/projects/zz_repositoryfile.go`
- `pkg/cluster/controller/projects/repositoryfiles/zz_controller.go`
- `apis/*/projects/v1alpha1/zz_generated.deepcopy.go`
- `apis/*/projects/v1alpha1/zz_generated.managed.go`
- `package/crds/projects.gitlab.*.crossplane.io_repositoryfiles.yaml`

## Controller-Logik

### Observe

```
1. Prüfe projectID vorhanden
2. Prüfe reconcileInterval:
   - Parse reconcileInterval als time.Duration
   - Wenn status.atProvider.lastObserveTime + reconcileInterval > now:
     → return ResourceExists: true, ResourceUpToDate: true (KEIN API call)
3. GetFile(projectID, filePath, {Ref: branch})
   - 404 → ResourceExists: false
   - Error → return error
4. Wenn createOnly == true:
   → ResourceExists: true, ResourceUpToDate: true (Existenz reicht)
5. Sonst: Vergleiche content_sha256 aus GitLab Response mit SHA256 von spec content
   - Match → ResourceUpToDate: true
   - Mismatch → ResourceUpToDate: false
6. Update status.atProvider (blobId, commitId, sha256, size, lastObserveTime)
7. LateInitialize: encoding, executeFilemode
```

### Create

```
1. Resolve content (inline oder aus Secret via contentSecretRef)
2. CreateFile(projectID, filePath, {
     Branch, Content,
     CommitMessage: createCommitMessage (default: "crossplane: create <filePath>"),
     Encoding, AuthorEmail, AuthorName, ExecuteFilemode
   })
3. Set external-name annotation = filePath
```

### Update

```
1. Wenn createOnly == true → sollte nie aufgerufen werden (Observe returns UpToDate)
2. Resolve content
3. UpdateFile(projectID, filePath, {
     Branch, Content,
     CommitMessage: updateCommitMessage (default: "crossplane: update <filePath>"),
     Encoding, AuthorEmail, AuthorName, ExecuteFilemode,
     LastCommitID: status.atProvider.lastCommitId (optimistic locking)
   })
```

### Delete

```
1. DeleteFile(projectID, filePath, {
     Branch,
     CommitMessage: deleteCommitMessage (default: "crossplane: delete <filePath>"),
     AuthorEmail, AuthorName
   })
```

## External Name

`crossplane.io/external-name` = `filePath`. Unique Key = projectID + branch + filePath.

## GitLab Client Interface

```go
type RepositoryFileClient interface {
    GetFile(pid any, fileName string, opt *gitlab.GetFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error)
    CreateFile(pid any, fileName string, opt *gitlab.CreateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
    UpdateFile(pid any, fileName string, opt *gitlab.UpdateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
    DeleteFile(pid any, fileName string, opt *gitlab.DeleteFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}
```

Konstruktor: `NewRepositoryFileClient(cfg) → git.RepositoryFiles`

## Codebase Patterns zu beachten

- Alle Types in `apis/namespaced/` schreiben, `hack/generate-cluster-scope.go` erzeugt cluster-Variante
- `xpv2.ManagedResourceSpec` für namespaced (wird zu `xpv1.ResourceSpec` in cluster)
- `xpv1.NamespacedReference` für Refs (wird zu `xpv1.Reference` in cluster)
- `xpv1.LocalSecretKeySelector` für Secrets (wird zu `xpv1.SecretKeySelector` in cluster)
- Controller braucht `Setup` + `SetupGated` Funktionen
- `managed.WithManagementPolicies()` in Setup wenn Feature-Flag aktiv
- `managed.WithPollInterval(o.PollInterval)` für globales Interval
- Fake Client in `fake/fake.go` mit `Mock*` Feldern
- Tests nutzen `test.MockClient` und table-driven Tests

## Build & Verify

```bash
make generate   # generiert cluster-scope, deepcopy, CRDs, managed methodsets
make build      # kompiliert alles
make test       # unit tests
```
