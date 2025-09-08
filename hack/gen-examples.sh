#!/bin/bash

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(dirname "$SCRIPT_DIR")"

# Delete and recreate output directories
rm -rf examples/cluster/groups examples/cluster/projects examples/namespaced/projects examples/namespaced/groups
mkdir -p examples/cluster/groups examples/cluster/projects examples/namespaced/projects examples/namespaced/groups

# Generate CRD examples directly to target directories
find package/crds -name "*.yaml" | while read -r crd_file; do
    group=$(yq eval '.spec.group' "$crd_file")
    
    if [[ "$group" == "groups.gitlab.crossplane.io" ]] || [[ "$group" == "gitlab.crossplane.io" ]]; then
        cty generate crd -c "$crd_file" -o examples/cluster/groups/
    elif [[ "$group" == "projects.gitlab.crossplane.io" ]]; then
        cty generate crd -c "$crd_file" -o examples/cluster/projects/
    elif [[ "$group" == "groups.gitlab.m.crossplane.io" ]] || [[ "$group" == "gitlab.m.crossplane.io" ]]; then
        cty generate crd -c "$crd_file" -o examples/namespaced/groups/
    elif [[ "$group" == "projects.gitlab.m.crossplane.io" ]]; then
        cty generate crd -c "$crd_file" -o examples/namespaced/projects/
    fi
done

# Remove status field and add metadata to all generated files
for dir in examples/cluster examples/namespaced examples/cluster/groups examples/cluster/projects examples/namespaced/projects examples/namespaced/groups; do
    find "$dir" -maxdepth 1 -name "*.yaml" -o -name "*.yml" | while read -r file; do
        yq eval 'del(.status)' -i "$file"
        
        # Add metadata name for cluster resources
        if [[ "$dir" == *"cluster"* ]]; then
            kind=$(yq eval '.kind' "$file" | tr '[:upper:]' '[:lower:]')
            yq eval ".metadata.name = \"example-$kind\"" -i "$file"
        fi
        
        # Add namespace for namespaced resources
        if [[ "$dir" == *"namespaced"* ]]; then
            yq eval '.metadata.namespace = "example-namespace"' -i "$file"
        fi
    done
done
