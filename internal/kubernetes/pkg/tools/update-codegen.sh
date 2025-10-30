#!/usr/bin/env bash

# ====================================================================
# Kubernetes CRD code generation script
# Generates clientset, informers, listers and deepcopy methods
# Uses temporary directory to work around code-generator path issues
# Copies generated files to correct project locations
# ====================================================================

set -o errexit
set -o nounset
set -o pipefail

MODULE="gitlab.jiagouyun.com/cloudcare-tools/datakit"
APIS_PKG="internal/kubernetes/pkg/apis"
OUTPUT_PKG="internal/kubernetes/pkg/client"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../../" && pwd)"
CODEGEN_PKG="${PROJECT_ROOT}/vendor/k8s.io/code-generator"

echo "Generating Kubernetes CRD code..."

# Check if code-generator exists
if [ ! -d "${CODEGEN_PKG}" ]; then
    echo "Error: code-generator not found, run 'go mod vendor' first"
    exit 1
fi

# Setup temporary directory
TEMP_OUTPUT_BASE="${PROJECT_ROOT}/_tmp_codegen_output"

# Cleanup previous runs
if [ -d "${TEMP_OUTPUT_BASE}" ]; then
    rm -rf "${TEMP_OUTPUT_BASE}"
fi
if [ -d "${PROJECT_ROOT}/gitlab.jiagouyun.com" ]; then
    rm -rf "${PROJECT_ROOT}/gitlab.jiagouyun.com"
fi

mkdir -p "${TEMP_OUTPUT_BASE}"

# Generate code
echo "Running code generator..."
bash "${CODEGEN_PKG}"/generate-groups.sh "all" \
  "${MODULE}/${OUTPUT_PKG}" "${MODULE}/${APIS_PKG}" \
  "datakits:v1alpha1" \
  --go-header-file "${SCRIPT_DIR}"/boilerplate.go.txt \
  --output-base "${TEMP_OUTPUT_BASE}"

# Copy generated files
echo "Copying generated files..."
GENERATED_ROOT="${TEMP_OUTPUT_BASE}/${MODULE}"

if [ -d "${GENERATED_ROOT}/${OUTPUT_PKG}" ]; then
    mkdir -p "${PROJECT_ROOT}/${OUTPUT_PKG}"
    cp -rf "${GENERATED_ROOT}/${OUTPUT_PKG}"/* "${PROJECT_ROOT}/${OUTPUT_PKG}/"
    echo "Client code copied to: ${PROJECT_ROOT}/${OUTPUT_PKG}"
else
    echo "Warning: Client code not found, generation may have failed"
fi

# Copy deepcopy files
DEEPCOPY_SRC="${GENERATED_ROOT}/${APIS_PKG}/datakits/v1alpha1/zz_generated.deepcopy.go"
DEEPCOPY_DEST="${PROJECT_ROOT}/${APIS_PKG}/datakits/v1alpha1"

if [ -f "${DEEPCOPY_SRC}" ]; then
    mkdir -p "${DEEPCOPY_DEST}"
    cp -f "${DEEPCOPY_SRC}" "${DEEPCOPY_DEST}/"
    echo "Deepcopy file copied to: ${DEEPCOPY_DEST}"
else
    echo "Warning: Deepcopy file not found"
fi

# Cleanup
rm -rf "${TEMP_OUTPUT_BASE}"
echo "Code generation completed"
