#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CRD_DIR="${ROOT_DIR}/test/crds"
TMP_DIR="${ROOT_DIR}/.tmp/crds"

# Versions
CAPI_VERSION="v1.12.3"
APP_CRD_VERSION="v0.6.2"
CAPZ_VERSION="v1.19.3"
CAPA_VERSION="v2.7.2"
CAPG_VERSION="v1.10.0"

rm -rf "${CRD_DIR}" "${TMP_DIR}"
mkdir -p "${CRD_DIR}" "${TMP_DIR}"

# Function to extract CRD documents from a multi-document YAML,
# stripping conversion webhook configs (incompatible with bare kube-apiserver)
# and making v1beta1 the storage version (our Go code uses v1beta1 types and
# without a conversion webhook v1beta2 storage would prune v1beta1-only fields).
extract_crds() {
    local input_file="$1"
    local output_file="$2"
    python3 -c "
import sys, re

content = open(sys.argv[1]).read()
docs = content.split('\n---\n')
crds = [d for d in docs if 'kind: CustomResourceDefinition' in d]

cleaned = []
for crd in crds:
    # Remove conversion webhook section
    crd = re.sub(r'\n  conversion:\n    strategy: Webhook\n.*?(?=\n  [a-z]|\n---|\Z)', '\n  conversion:\n    strategy: None', crd, flags=re.DOTALL)

    # Make v1beta1 the storage version instead of v1beta2 when v1beta1
    # exists. Without a conversion webhook the API server uses the storage
    # version's schema for pruning. Our Go types target v1beta1 so we need
    # v1beta1 as storage to avoid data loss. CRDs without v1beta1 keep
    # their original storage version.
    lines = crd.split('\n')
    # Check if this CRD has a v1beta1 version at the top-level indent.
    has_v1beta1 = any(re.match(r'^    name: v1beta1$', l) for l in lines)
    if has_v1beta1:
        out = []
        in_version = None
        for line in lines:
            m = re.match(r'^    name: (v1\S+)', line)
            if m:
                in_version = m.group(1)
            if re.match(r'^    storage:', line):
                if in_version == 'v1beta1':
                    line = '    storage: true'
                else:
                    line = '    storage: false'
            out.append(line)
        crd = '\n'.join(out)

    cleaned.append(crd)

print('\n---\n'.join(cleaned))
" "${input_file}" > "${output_file}"
}

echo "Downloading CAPI CRDs..."
curl -sL "https://github.com/kubernetes-sigs/cluster-api/releases/download/${CAPI_VERSION}/cluster-api-components.yaml" -o "${TMP_DIR}/capi-components.yaml"
extract_crds "${TMP_DIR}/capi-components.yaml" "${CRD_DIR}/capi.yaml"

echo "Downloading Application CRDs..."
APP_CRD_BASE="https://raw.githubusercontent.com/giantswarm/apiextensions-application/${APP_CRD_VERSION}/config/crd"
curl -sL "${APP_CRD_BASE}/application.giantswarm.io_apps.yaml" -o "${CRD_DIR}/application.giantswarm.io_apps.yaml"
curl -sL "${APP_CRD_BASE}/application.giantswarm.io_catalogs.yaml" -o "${CRD_DIR}/application.giantswarm.io_catalogs.yaml"
curl -sL "${APP_CRD_BASE}/application.giantswarm.io_appcatalogentries.yaml" -o "${CRD_DIR}/application.giantswarm.io_appcatalogentries.yaml"

echo "Downloading CAPZ CRDs..."
curl -sL "https://github.com/kubernetes-sigs/cluster-api-provider-azure/releases/download/${CAPZ_VERSION}/infrastructure-components.yaml" -o "${TMP_DIR}/capz-components.yaml"
extract_crds "${TMP_DIR}/capz-components.yaml" "${CRD_DIR}/capz.yaml"

echo "Downloading CAPA CRDs..."
curl -sL "https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/download/${CAPA_VERSION}/infrastructure-components.yaml" -o "${TMP_DIR}/capa-components.yaml"
extract_crds "${TMP_DIR}/capa-components.yaml" "${CRD_DIR}/capa.yaml"

echo "Downloading CAPG CRDs..."
curl -sL "https://github.com/kubernetes-sigs/cluster-api-provider-gcp/releases/download/${CAPG_VERSION}/infrastructure-components.yaml" -o "${TMP_DIR}/capg-components.yaml"
extract_crds "${TMP_DIR}/capg-components.yaml" "${CRD_DIR}/capg.yaml"

# Copy locally generated CRDs if they exist, skipping broken ones
# (CAPO and CAPZ v1alpha4 generate incomplete CRDs with missing group names)
if [ -d "${ROOT_DIR}/config/crd" ]; then
    echo "Copying local CRDs..."
    for f in "${ROOT_DIR}/config/crd/"*.yaml; do
        basename="$(basename "$f")"
        # Skip CRDs with missing group prefix (broken generation)
        if [[ "$basename" == _* ]]; then
            echo "  Skipping broken CRD: $basename"
            continue
        fi
        cp "$f" "${CRD_DIR}/"
    done
fi

# Cleanup
rm -rf "${TMP_DIR}"

echo "CRDs downloaded to ${CRD_DIR}"
ls -la "${CRD_DIR}/"
