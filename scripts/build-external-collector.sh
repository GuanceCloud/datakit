#!/bin/bash
# Build DataKit external collector binaries for specified architectures
# using a prebuilt docker build environment image.
#
# Usage:
#   ./scripts/build-external-collector.sh --input <collector-name>
#   ./scripts/build-external-collector.sh --input journald --arch amd64,arm64
#   ./scripts/build-external-collector.sh --input ebpf --output-dir dist/datakit-linux-arm64/externals --arch arm64
#
# Examples:
#   # Build journald collector for current architecture
#   ./scripts/build-external-collector.sh --input journald
#
#   # Build for multiple architectures
#   BUILD_ARCH=amd64,arm64 ./scripts/build-external-collector.sh --input journald
#
#   # Build with custom output name
#   ./scripts/build-external-collector.sh --input journald --output journald-bin
#
#   # Build standalone datakit-ebpf
#   ./scripts/build-external-collector.sh --input ebpf --arch arm64

set -euo pipefail

# Default values
COLLECTOR_NAME=""
OUTPUT_NAME=""
ARCHS=${BUILD_ARCH:-"amd64"}
CGO_CFLAGS=${CGO_CFLAGS:-"-Wno-undef-prefix"}
LDFLAGS=${LDFLAGS:--w -s}
GOFLAGS_VALUE=${GOFLAGS:-"-mod=mod"}
VERBOSE=${VERBOSE:-false}
PROJECT_ROOT=$(pwd)
HOST_ARCH=$(uname -m)
CGO_ENABLED="1"
OUTPUT_DIR=""
BUILD_ENV_IMAGE=${DK_BUILD_ENV_IMAGE:-"pubrepo.jiagouyun.com/ebpf-dev/dk_build_env:latest"}
DOCKER_CMD=${DK_BUILD_DOCKER_CMD:-"sudo docker"}
read -r -a DOCKER_CMD_ARR <<< "${DOCKER_CMD}"
LEGACY_SYSROOT_BASE=${DK_LEGACY_SYSROOT_BASE:-"/opt/sysroots/debian10"}
GO_BUILD_TAGS=""

# Convert host architecture to GOARCH format
if [ "${HOST_ARCH}" = "x86_64" ]; then
    HOST_ARCH="amd64"
elif [ "${HOST_ARCH}" = "aarch64" ]; then
    HOST_ARCH="arm64"
fi

# Print usage
usage() {
    cat << EOF
Usage: $0 --input <collector-name> [OPTIONS]

Build DataKit external collector binaries for specified architectures.

Required:
  -i, --input <name>        Collector name (e.g., journald, db2, oracle, logfwd, ebpf)

Optional:
   -o, --output <name>       Output binary name (default: same as --input)
   -a, --arch <archs>        Architectures to build, comma-separated (default: amd64)
                             Can also use BUILD_ARCH environment variable
   --output-dir <path>       Output directory for the binary
   --cgo-enabled <0|1>       Enable CGO (default: 1)
   --cgo-flags <flags>       CGO compiler flags (default: -Wno-undef-prefix)
   --ldflags <flags>         Go linker flags (default: -w -s)
   --image <name>            Build environment image (default: pubrepo.jiagouyun.com/ebpf-dev/dk_build_env:latest)
   --docker-cmd <cmd>        Docker runner command (default: sudo docker)
   -v, --verbose             Enable verbose build output
   -h, --help                Show this help message

Environment Variables:
  BUILD_ARCH               Alternative to --arch
  DK_BUILD_ENV_IMAGE       Alternative to --image
  DK_BUILD_DOCKER_CMD      Alternative to --docker-cmd
  CGO_ENABLED              Alternative to --cgo-enabled
  CGO_CFLAGS               Alternative to --cgo-flags
  LDFLAGS                  Alternative to --ldflags
  VERBOSE                  Alternative to --verbose

Examples:
  # Build journald collector for current architecture
  $0 --input journald

  # Build for multiple architectures
  BUILD_ARCH=amd64,arm64 $0 --input journald

  # Build datakit-ebpf to the regular externals directory
  $0 --input ebpf --arch arm64 --output-dir dist/datakit-linux-arm64/externals
EOF
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -i|--input)
            COLLECTOR_NAME="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_NAME="$2"
            shift 2
            ;;
        -a|--arch)
            ARCHS="$2"
            shift 2
            ;;
        --cgo-enabled)
            CGO_ENABLED="$2"
            shift 2
            ;;
        --cgo-flags)
            CGO_CFLAGS="$2"
            shift 2
            ;;
        --ldflags)
            LDFLAGS="$2"
            shift 2
            ;;
        --output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --image)
            BUILD_ENV_IMAGE="$2"
            shift 2
            ;;
        --docker-cmd)
            DOCKER_CMD="$2"
            read -r -a DOCKER_CMD_ARR <<< "${DOCKER_CMD}"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required parameters
if [ -z "${COLLECTOR_NAME}" ]; then
    echo "Error: --input is required"
    usage
fi

# Set verbose flag
if [ "${VERBOSE}" = "true" ]; then
    set -x
    VERBOSE_FLAG="-x"
else
    VERBOSE_FLAG=""
fi

ENTRY_PATH=""
COLLECTOR_KIND="go"
if [ -z "${OUTPUT_NAME}" ]; then
    OUTPUT_NAME="${COLLECTOR_NAME}"
fi

case "${COLLECTOR_NAME}" in
    ebpf)
        COLLECTOR_KIND="ebpf"
        OUTPUT_NAME="datakit-ebpf"
        ENTRY_PATH="internal/plugins/externals/ebpf/Makefile"
        ;;
    db2|journald|oracle)
        GO_BUILD_TAGS="netgo"
        ENTRY_PATH="internal/plugins/externals/${COLLECTOR_NAME}/main.go"
        ;;
    logfwd)
        ENTRY_PATH="internal/plugins/externals/logfwd/cmd/main.go"
        ;;
    *)
        ENTRY_PATH="internal/plugins/externals/${COLLECTOR_NAME}/main.go"
        ;;
esac

echo "=========================================="
echo "DataKit External Collector Builder"
echo "=========================================="
echo "Collector: ${COLLECTOR_NAME}"
echo "Output: ${OUTPUT_NAME}"
echo "Architectures: ${ARCHS}"
echo "Host Architecture: ${HOST_ARCH}"
echo "CGO Enabled: ${CGO_ENABLED}"
echo "CGO CFLAGS: ${CGO_CFLAGS}"
echo "GOFLAGS: ${GOFLAGS_VALUE}"
echo "Build env image: ${BUILD_ENV_IMAGE}"
echo "Docker command: ${DOCKER_CMD}"
echo "Legacy sysroot base: ${LEGACY_SYSROOT_BASE}"
echo "Go build tags: ${GO_BUILD_TAGS:-<none>}"
echo "Project Root: ${PROJECT_ROOT}"
echo "Script called from: $(pwd)"
echo "Environment check:"
echo "  CGO_ENABLED (env): ${CGO_ENABLED}"
echo "  CGO_CFLAGS (env): ${CGO_CFLAGS}"
echo

# Verify source file exists
if [ ! -f "${PROJECT_ROOT}/${ENTRY_PATH}" ]; then
    echo "✗ Source file not found: ${PROJECT_ROOT}/${ENTRY_PATH}"
    echo "  Expected location: ${ENTRY_PATH}"
    exit 1
fi

echo "✓ Source file found: ${ENTRY_PATH}"
echo

if ! "${DOCKER_CMD_ARR[@]}" info &>/dev/null; then
    echo "✗ Docker is not available via: ${DOCKER_CMD}"
    exit 1
fi

if ! "${DOCKER_CMD_ARR[@]}" image inspect "${BUILD_ENV_IMAGE}" &>/dev/null; then
    echo "! Build environment image not found locally, pulling: ${BUILD_ENV_IMAGE}"
    if ! "${DOCKER_CMD_ARR[@]}" pull "${BUILD_ENV_IMAGE}"; then
        echo "✗ Failed to pull build environment image: ${BUILD_ENV_IMAGE}"
        exit 1
    fi
fi

build_in_container() {
    local arch="$1"
    local output_path="$2"
    local output_path_container=""
    local container_cmd=""
    local cross_env=""
    local triplet=""
    local sysroot_env=""
    local go_tags_arg=""

    mkdir -p "$(dirname "$(abs_output_path "${output_path}")")"

    if [[ "${output_path}" = /* ]]; then
        output_path_container="${output_path/#${PROJECT_ROOT}/\/work}"
    else
        output_path_container="/work/${output_path}"
    fi

    case "${arch}" in
        amd64)
            triplet="x86_64-linux-gnu"
            ;;
        arm64)
            triplet="aarch64-linux-gnu"
            ;;
        *)
            triplet=""
            ;;
    esac

    if [ -n "${triplet}" ]; then
        sysroot_env=$(cat <<EOF
if [ -d '${LEGACY_SYSROOT_BASE}/${arch}' ]; then
export SYSROOT='${LEGACY_SYSROOT_BASE}/${arch}'
export PKG_CONFIG_SYSROOT_DIR="\${SYSROOT}"
export PKG_CONFIG_LIBDIR="\${SYSROOT}/usr/lib/${triplet}/pkgconfig:\${SYSROOT}/usr/lib/pkgconfig:\${SYSROOT}/usr/share/pkgconfig"
export PKG_CONFIG_PATH=
export PKG_CONFIG_ALLOW_CROSS=1
export CGO_CPPFLAGS="--sysroot=\${SYSROOT}"
export CGO_CFLAGS="--sysroot=\${SYSROOT} ${CGO_CFLAGS}"
export CGO_CXXFLAGS="--sysroot=\${SYSROOT}"
export CGO_LDFLAGS="--sysroot=\${SYSROOT} -L\${SYSROOT}/lib/${triplet} -L\${SYSROOT}/usr/lib/${triplet} -Wl,-rpath-link,\${SYSROOT}/lib/${triplet} -Wl,-rpath-link,\${SYSROOT}/usr/lib/${triplet}"
fi
EOF
)
    fi

    if [ -n "${GO_BUILD_TAGS}" ]; then
        go_tags_arg="-tags '${GO_BUILD_TAGS}'"
    fi

    if [ "${arch}" = "arm64" ]; then
        cross_env=$(cat <<EOF
export CC=aarch64-linux-gnu-gcc
export CXX=aarch64-linux-gnu-g++
EOF
)
    elif [ "${arch}" = "amd64" ]; then
        cross_env=$(cat <<EOF
export CC=gcc
export CXX=g++
EOF
)
    fi

    case "${COLLECTOR_KIND}" in
        ebpf)
            local kernel_headers="/usr/src/linux-headers-${arch}"
            local ebpf_args=""
            if [ "${arch}" = "arm64" ]; then
                ebpf_args="ARGS='--target=aarch64-linux-gnu'"
            fi
            container_cmd=$(cat <<EOF
set -euo pipefail
git config --global --add safe.directory /work
git config --global --add safe.directory /work/internal/plugins/externals/ebpf
cd /work/internal/plugins/externals/ebpf
export GOOS=linux
export GOARCH=${arch}
export CGO_ENABLED=1
unset GOFLAGS
${cross_env}
${sysroot_env}
make -j8 \
  SRCPATH=. \
  OUTPATH='${output_path_container}' \
  ARCH=${arch} \
  DK_BPF_KERNEL_SRC_PATH='${kernel_headers}' \
  ${ebpf_args}
EOF
)
            ;;
        *)
            container_cmd=$(cat <<EOF
set -euo pipefail
export CGO_ENABLED=${CGO_ENABLED}
export GOOS=linux
export GOARCH=${arch}
export GOFLAGS='${GOFLAGS_VALUE}'
${cross_env}
${sysroot_env}
go build ${VERBOSE_FLAG} \
  ${go_tags_arg} \
  -o '${output_path_container}' \
  -ldflags '${LDFLAGS}' \
  '${ENTRY_PATH}'
EOF
)
            ;;
    esac

    "${DOCKER_CMD_ARR[@]}" run --rm \
        -v "${PROJECT_ROOT}:/work" \
        -w /work \
        "${BUILD_ENV_IMAGE}" \
        bash -lc "${container_cmd}"

    "${DOCKER_CMD_ARR[@]}" run --rm \
        -v "${PROJECT_ROOT}:/work" \
        -w /work \
        "${BUILD_ENV_IMAGE}" \
        chown -R "$(id -u):$(id -g)" "$(dirname "${output_path}")" >/dev/null 2>&1 || true
}

abs_output_path() {
    local path="$1"
    if [[ "${path}" = /* ]]; then
        printf '%s\n' "${path}"
    else
        printf '%s\n' "${PROJECT_ROOT}/${path}"
    fi
}

# Parse architectures
IFS=',' read -ra ARCH_ARRAY <<<"${ARCHS}"

for ARCH in "${ARCH_ARRAY[@]}"; do
    ARCH=$(echo "${ARCH}" | xargs) # Trim whitespace
    CURRENT_OUTPUT_DIR="${OUTPUT_DIR}"
    if [ -z "${CURRENT_OUTPUT_DIR}" ]; then
        CURRENT_OUTPUT_DIR="dist/datakit-linux-${ARCH}/externals"
    fi

    echo "=========================================="
    echo "Building for linux/${ARCH}"
    echo "=========================================="

    echo "Build Method: Docker image ${BUILD_ENV_IMAGE}"
    echo
    echo "Step 1: Building ${OUTPUT_NAME} binary in container..."

    BINARY_PATH="$(abs_output_path "${CURRENT_OUTPUT_DIR}/${OUTPUT_NAME}")"
    build_in_container "${ARCH}" "${CURRENT_OUTPUT_DIR}/${OUTPUT_NAME}"

    # Verify binary
    echo
    echo "Step 2: Verifying binary..."

    if [ ! -f "${BINARY_PATH}" ]; then
        echo "✗ Binary not found: ${BINARY_PATH}"
        exit 1
    fi

    echo "Binary info:"
    ls -lh "${BINARY_PATH}"
    file "${BINARY_PATH}"

    # Check if binary has correct architecture
    if [ "${ARCH}" = "amd64" ]; then
        if file "${BINARY_PATH}" | grep -q "x86-64"; then
            echo "✓ Architecture verification passed (x86-64)"
        else
            echo "✗ Wrong architecture detected!"
            exit 1
        fi
    elif [ "${ARCH}" = "arm64" ]; then
        if file "${BINARY_PATH}" | grep -q "aarch64"; then
            echo "✓ Architecture verification passed (aarch64)"
        else
            echo "✗ Wrong architecture detected!"
            exit 1
        fi
    fi



    echo "✓ Build complete for linux/${ARCH}"
    echo
done



echo "=========================================="
echo "✓ All builds completed successfully!"
echo "=========================================="
echo
echo "Built binaries:"
for ARCH in "${ARCH_ARRAY[@]}"; do
    ARCH=$(echo "${ARCH}" | xargs)
    if [ -n "${OUTPUT_DIR}" ]; then
        echo "  - ${OUTPUT_DIR}/${OUTPUT_NAME}"
    else
        echo "  - dist/datakit-linux-${ARCH}/externals/${OUTPUT_NAME}"
    fi
done
echo
echo "Next steps:"
echo "  1. Test binary: ./dist/datakit-linux-<arch>/externals/${OUTPUT_NAME} --help"
echo "  2. Build Docker image with the binary"
echo "  3. Deploy to target environment"
echo
