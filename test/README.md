# Testing

We have created chainsaw tests for testing most resources. The tests can be executed with:

## Namespaced Resources

```bash
chainsaw test --test-dir test/namespaced --namespace chainsaw --values - <<EOF
token: glpat-....
baseUrl: https://gitlab.com
key: $(ssh-keygen -t ecdsa -b 256 -f /tmp/temp_key -N "" -C "just a test" >/dev/null 2>&1 && cat /tmp/temp_key.pub && rm /tmp/temp_key /tmp/temp_key.pub)
EOF
```

## Cluster-scoped Resources

```bash
chainsaw test --test-dir test/cluster --values - <<EOF
token: glpat-....
baseUrl: https://gitlab.com
key: $(ssh-keygen -t ecdsa -b 256 -f /tmp/temp_key -N "" -C "just a test" >/dev/null 2>&1 && cat /tmp/temp_key.pub && rm /tmp/temp_key /tmp/temp_key.pub)
EOF
```

## Using kwok for Testing

You can use [kwok](https://kwok.sigs.k8s.io/) to create lightweight Kubernetes clusters for testing. Kwok seems to be perfect for testing Crossplane providers, as long as you don't have to run any pods.

### Quick Start with kwok

Create a test cluster:
```bash
kwokctl create cluster
```

Delete the cluster when done:
```bash
kwokctl delete cluster
```

For more information and advanced configuration options, visit the [kwok documentation](https://kwok.sigs.k8s.io/).