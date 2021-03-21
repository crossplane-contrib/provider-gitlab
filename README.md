# provider-gitlab

## Overview

`provider-gitlab` is the Crossplane infrastructure provider for
[Gitlab](https://gitlab.com/). The provider that is built from the source code
in this repository can be installed into a Crossplane control plane and adds the
following new functionality:

* Custom Resource Definitions (CRDs) that model Gitlab resources
* Controllers to provision these resources in Github based on the users desired
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
    secretRef:
      namespace: crossplane-system
      name: gitlab-credentials
      key: token
```
```bash
kubectl apply -f examples/providerconfig/provider.yaml
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