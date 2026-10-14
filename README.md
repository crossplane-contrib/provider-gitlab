# provider-gitlab

## Overview

`provider-gitlab` is the Crossplane infrastructure provider for
[Gitlab](https://gitlab.com/). The provider that is built from the source code
in this repository can be installed into a Crossplane control plane and adds the
following new functionality:

* Custom Resource Definitions (CRDs) that model Gitlab resources
* Controllers to provision these resources in Gitlab based on the users desired
  state captured in CRDs they create
* Implementations of Crossplane's portable resource
  abstractions, enabling
  Gitlab resources to fulfill a user's general need for Gitlab configurations

## Getting Started and Documentation

Create a [Personal Access Token](https://gitlab.com/-/profile/personal_access_tokens) on your GitLab instance with the scope set to `api` and fill in the corresponding Kubernetes secret:

```bash
kubectl create secret generic gitlab-credentials -n crossplane-system --from-literal=token="<PERSONAL_ACCESS_TOKEN>"
```

Configure a `ProviderConfig` with a baseURL pointing to your GitLab instance:

```yaml
apiVersion: gitlab.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: gitlab-provider
spec:
  baseURL: https://gitlab.com/
  credentials:
    source: Secret
    method: PersonalAccessToken
    secretRef:
      namespace: crossplane-system
      name: gitlab-credentials
      key: token
```

```bash
kubectl apply -f examples/providerconfig/provider.yaml
```

### Self-rotating service account tokens

The namespaced `groups.gitlab.m.crossplane.io/v1alpha1` `ServiceAccountAccessToken`
resource can keep a short-lived token alive by rotating it before it expires.

It supports two modes, selected automatically:

* **Owner mode** (default): the `ProviderConfig` is a group owner. The token is
  created, observed, rotated and revoked through the group service-account
  endpoints.
* **Self-managed mode**: the `ProviderConfig` authenticates with the *very token*
  this resource manages — i.e. the `ProviderConfig.credentials.secretRef` points
  at the same secret in the ProviderConfig's own namespace (name and `token` key) that the resource writes
  via `writeConnectionSecretToRef`. The provider then acts as the service account
  itself and uses the self endpoints (`GET/POST /personal_access_tokens/self`).
  This lets a short-lived token reconcile a whole group and keep itself alive by
  self-rotating. A `SelfManaged` status condition reports the active mode.

> **Bootstrap secret type.** In self-managed mode the rotated token is written
> back into the secret the `ProviderConfig` reads. Crossplane only writes
> connection secrets it controls, so a **hand-created** bootstrap secret must use
> the connection secret type `connection.crossplane.io/v1alpha1` — a default
> `Opaque` secret (e.g. from `kubectl create secret generic`, which has no
> `--type` flag) is rejected with `refusing to modify uncontrolled secret of type
> "Opaque"`. Create it from a manifest:
>
> ```yaml
> apiVersion: v1
> kind: Secret
> metadata:
>   name: gitlab-self-rotating-token
>   namespace: default
> type: connection.crossplane.io/v1alpha1
> stringData:
>   # a service account PAT with at least the `api` and `self_rotate` scopes
>   token: "<PERSONAL_ACCESS_TOKEN>"
> ```
>
> Alternatively, bootstrap in owner mode first (Crossplane then creates and owns
> a correctly typed connection secret), and switch `providerConfigRef` to the
> self `ProviderConfig` afterwards.

See [`examples/groups/serviceaccountaccesstoken.yaml`](examples/groups/serviceaccountaccesstoken.yaml)
for full examples of both modes.

#### Project service account tokens

The namespaced `projects.gitlab.m.crossplane.io/v1alpha1` `ServiceAccountAccessToken`
resource provides the same capability for **project** service accounts (also
available with at least a Premium license). It is keyed by `projectId` +
`serviceAccountId`, and supports the same two modes:

* **Owner mode** (default): the token is created, observed, rotated and revoked
  through the project service-account endpoints. This mode requires a
  `ProviderConfig` that is a project owner/maintainer.
* **Self-managed mode**: identical to the group resource — once bootstrapped, the
  self-rotating loop runs on the service account's own credential. Detection and
  the `SelfManaged` status condition work the same way.

A companion `projects.gitlab.m.crossplane.io/v1alpha1` `ServiceAccount` resource
manages the project service account itself. The bootstrap-secret note above
applies unchanged. See
[`examples/projects/serviceaccountaccesstoken.yaml`](examples/projects/serviceaccountaccesstoken.yaml)
for full examples of both modes.

#### Instance service account tokens

The namespaced `instance.gitlab.m.crossplane.io/v1alpha1` `ServiceAccountAccessToken`
resource provides the same capability for **instance** service accounts (only
available on self-managed GitLab with at least a Premium license). It is keyed by
`serviceAccountId` instead of a group, and supports the same two modes:

* **Owner mode** (default): instance service-account tokens have no dedicated
  endpoints, so the token is created, observed, rotated and revoked through the
  **admin** personal access tokens API. This mode therefore requires a
  `ProviderConfig` that authenticates with an **instance-admin** token.
* **Self-managed mode**: identical to the group resource — once bootstrapped, the
  self-rotating loop runs on the service account's own credential and needs **no
  admin token**. Detection and the `SelfManaged` status condition work the same
  way.

The bootstrap-secret note above applies unchanged. See
[`examples/instance/serviceaccountaccesstoken.yaml`](examples/instance/serviceaccountaccesstoken.yaml)
for full examples of both modes.

### Self-rotating group and project access tokens

The namespaced `groups.gitlab.m.crossplane.io/v1alpha1` and
`projects.gitlab.m.crossplane.io/v1alpha1` `AccessToken` resources support the
same self-rotation capability. A group or project access token is backed by a
bot-user personal access token, so the generic self endpoints
(`GET/POST/DELETE /personal_access_tokens/self`) apply. Both modes are selected
automatically:

* **Owner mode** (default): the `ProviderConfig` is a group/project owner or
  maintainer. The token is created, observed, rotated and revoked through the
  group/project access-token endpoints. The `groupId`/`projectId` scopes every
  call server-side, so there is no cross-scope reach and no additional identity
  check is required.
* **Self-managed mode**: the `ProviderConfig` authenticates with the *very token*
  this resource manages — i.e. its `credentials.secretRef` points at the same
  secret in the ProviderConfig's own namespace (name and `token` key) that the resource writes via
  `writeConnectionSecretToRef`. The provider then acts as the token's bot user
  and uses the self endpoints, keeping the token alive by self-rotating before
  expiry. A `SelfManaged` status condition reports the active mode.

> **Identity guard.** In self-managed mode the provider adopts whatever token the
> credential authenticates as. The self-inform response does not expose the
> token's group/project binding, so the only spec-derived identity signal is the
> token name: the resource refuses to manage a token whose name does not match
> `spec.forProvider.name`. Include the `self_rotate` scope (plus `api`) so the
> token can rotate itself.

The bootstrap-secret note above applies unchanged. See
[`examples/groups/accesstoken.yaml`](examples/groups/accesstoken.yaml) and
[`examples/projects/accesstoken.yaml`](examples/projects/accesstoken.yaml) for
full examples of both modes.

### Delete a namespace that holds managed resources

When you delete a namespace, Kubernetes deletes all of its content at once. That
includes a namespaced `ProviderConfig`, its credentials `Secret`, and the
`ProviderConfigUsage` objects that keep the `ProviderConfig` in use. A managed
resource that must still delete its GitLab resource then has no credentials. It
reports `cannot get referenced ProviderConfig ... not found` or
`Cannot find referenced secret` and keeps its finalizer. The namespace stays in
`Terminating`, and the GitLab resource can remain. See
[crossplane/crossplane-runtime#1049](https://github.com/crossplane/crossplane-runtime/issues/1049).

[crossplane/crossplane-runtime#1113](https://github.com/crossplane/crossplane-runtime/pull/1113)
keeps the `ProviderConfig` until its managed resources are gone. The change is
merged but not released. This provider uses crossplane-runtime v2.4.2, which
does not include it. The change also does not protect the credentials `Secret`.
[crossplane/crossplane#7801](https://github.com/crossplane/crossplane/issues/7801)
tracks protection for the `Secret`.

To delete a namespace with its managed resources, use one of these options:

* Reference a `ClusterProviderConfig` from managed resources in a namespace that
  you delete as a whole. A managed resource without `providerConfigRef` uses the
  `ClusterProviderConfig` named `default`. Keep its credentials `Secret` in a
  namespace that you do not delete, for example `crossplane-system`. A
  terminating namespace refuses to create a new `ProviderConfigUsage`. The
  provider ignores that refusal while the managed resource is being deleted. See
  [crossplane/crossplane-runtime#1150](https://github.com/crossplane/crossplane-runtime/issues/1150).
  [#450](https://github.com/crossplane-contrib/provider-gitlab/issues/450)
  tracks removing this workaround once the provider uses a crossplane-runtime
  release that includes
  [crossplane/crossplane-runtime#1113](https://github.com/crossplane/crossplane-runtime/pull/1113).
* Delete the managed resources first and wait until they are gone. Then delete
  the namespace.

If a namespace is already stuck, point each stuck managed resource at a
`ClusterProviderConfig` whose token can manage its GitLab resource. The managed
resource then deletes its GitLab resource and releases the namespace. If the
token cannot see the project or group, GitLab answers 404. The managed resource
then removes its finalizer and leaves the GitLab resource in place.

Patch each resource by the full name that the first command prints, for example
`variable.projects.gitlab.m.crossplane.io/my-variable`. A short name such as
`variable` matches several API groups.

```bash
kubectl get managed -n <namespace> -o name | grep '\.gitlab\.m\.crossplane\.io/'
kubectl patch -n <namespace> <full-name> --type=merge \
  -p '{"spec":{"providerConfigRef":{"kind":"ClusterProviderConfig","name":"<cluster-provider-config>"}}}'
```

If you have no `ClusterProviderConfig`, remove the finalizer of each stuck
managed resource by hand as a last resort. The namespace then finishes deleting.

> [!WARNING]
> Kubernetes deletes the managed resource, but the GitLab resource stays. Before
> you patch, note the `crossplane.io/external-name` annotation. Then delete the
> GitLab resource by hand.

```bash
kubectl patch -n <namespace> <full-name> --type=merge \
  -p '{"metadata":{"finalizers":null}}'
```

## Contributing

provider-gitlab is a community driven project and we welcome contributions. See
the Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-gitlab/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Governance and Owners

provider-gitlab is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

provider-gitlab adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-gitlab is under the Apache 2.0 license.

[![FOSSA
Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-gitlab.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-gitlab?ref=badge_large)
