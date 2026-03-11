#!/bin/bash
# Build DataKit external collector binaries for specified architectures
# Uses Docker Buildx for reliable cross-compilation
#
# Usage:
#   ./scripts/build-external-collector.sh --input <collector-name>
#   ./scripts/build-external-collector.sh --input journald --arch amd64,arm64
#   ./scripts/build-external-collector.sh -i journald -o journald-collector
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

set -e

# Note: Uses Docker Buildx for cross-compilation which provides better stability
# and caching compared to raw QEMU emulation. Buildx handles platform emulation
# more reliably and supports multi-arch builds in a single command.

# Default values
COLLECTOR_NAME=""
OUTPUT_NAME=""
ARCHS=${BUILD_ARCH:-"amd64"}
CGO_CFLAGS=${CGO_CFLAGS:-"-Wno-undef-prefix"}
LDFLAGS=${LDFLAGS:--w -s}
VERBOSE=${VERBOSE:-false}
PROJECT_ROOT=$(pwd)
HOST_ARCH=$(uname -m)
# Default CGO_ENABLED=1, but can be overridden via --cgo-enabled flag
CGO_ENABLED="1"
# Force buildx usage (no fallback to legacy build)
FORCE_BUILDX=${FORCE_BUILDX:-false}

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
  -i, --input <name>        Collector name (e.g., journald, syslog)
                            Maps to: internal/plugins/externals/<name>/main.go

Optional:
   -o, --output <name>       Output binary name (default: same as --input)
   -a, --arch <archs>        Architectures to build, comma-separated (default: amd64)
                             Can also use BUILD_ARCH environment variable
   --cgo-enabled <0|1>       Enable CGO (default: 1)
   --cgo-flags <flags>       CGO compiler flags (default: -Wno-undef-prefix)
   --ldflags <flags>         Go linker flags (default: -w -s)
   --force-buildx            Force use of buildx (no fallback to legacy build)
   -v, --verbose             Enable verbose build output
   -h, --help                Show this help message

Environment Variables:
  BUILD_ARCH               Alternative to --arch
  CGO_ENABLED              Alternative to --cgo-enabled
  CGO_CFLAGS               Alternative to --cgo-flags
  LDFLAGS                  Alternative to --ldflags
  VERBOSE                  Alternative to --verbose

Examples:
  # Build journald collector for current architecture
  $0 --input journald

  # Build for multiple architectures
  BUILD_ARCH=amd64,arm64 $0 --input journald

  # Build with custom settings
  $0 --input journald --output journald-collector --arch arm64 --verbose

  # Build without CGO (if collector supports it)
  CGO_ENABLED=0 $0 --input journald

Output:
  Binaries are placed in: dist/datakit-linux-<arch>/externals/<output-name>
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
        --force-buildx)
            FORCE_BUILDX=true
            shift
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

# Set output name if not specified
if [ -z "${OUTPUT_NAME}" ]; then
    OUTPUT_NAME="${COLLECTOR_NAME}"
fi

# Set verbose flag
if [ "${VERBOSE}" = "true" ]; then
    set -x
    VERBOSE_FLAG="-x"
else
    VERBOSE_FLAG=""
fi

# Source and output paths
SOURCE_PATH="internal/plugins/externals/${COLLECTOR_NAME}/main.go"
OUTPUT_DIR="dist/datakit-linux-\${ARCH}/externals"
OUTPUT_BINARY="${OUTPUT_NAME}"

echo "=========================================="
echo "DataKit External Collector Builder"
echo "=========================================="
echo "Collector: ${COLLECTOR_NAME}"
echo "Output: ${OUTPUT_NAME}"
echo "Architectures: ${ARCHS}"
echo "Host Architecture: ${HOST_ARCH}"
echo "CGO Enabled: ${CGO_ENABLED}"
echo "CGO CFLAGS: ${CGO_CFLAGS}"
echo "Project Root: ${PROJECT_ROOT}"
echo "Script called from: $(pwd)"
echo "Environment check:"
echo "  CGO_ENABLED (env): ${CGO_ENABLED}"
echo "  CGO_CFLAGS (env): ${CGO_CFLAGS}"
echo

# Verify source file exists
if [ ! -f "${PROJECT_ROOT}/${SOURCE_PATH}" ]; then
    echo "✗ Source file not found: ${PROJECT_ROOT}/${SOURCE_PATH}"
    echo "  Expected location: ${SOURCE_PATH}"
    exit 1
fi

echo "✓ Source file found: ${SOURCE_PATH}"
echo

# Check if Docker is available
DOCKER_AVAILABLE=false
BUILDX_AVAILABLE=false
if command -v docker &>/dev/null; then
    if sudo docker info &>/dev/null; then
        DOCKER_AVAILABLE=true
        if sudo docker buildx version &>/dev/null; then
            BUILDX_AVAILABLE=true
            echo "Docker: Available with Buildx (recommended for cross-compilation)"
        else
            echo "Docker: Available (will use legacy build for cross-compilation)"
        fi
    else
        echo "Docker: Installed but not running (will use native builds where possible)"
    fi
else
    echo "Docker: Not available (will use native builds where possible)"
fi
echo

# Parse architectures
IFS=',' read -ra ARCH_ARRAY <<<"${ARCHS}"

for ARCH in "${ARCH_ARRAY[@]}"; do
    ARCH=$(echo "${ARCH}" | xargs) # Trim whitespace
    echo "=========================================="
    echo "Building for linux/${ARCH}"
    echo "=========================================="

    # Create output directory
    mkdir -p "${PROJECT_ROOT}/dist/datakit-linux-${ARCH}/externals"

    # Determine build method
    if [ "${ARCH}" = "${HOST_ARCH}" ]; then
        # Native build
        echo "Build Method: Native (host architecture matches)"
        if [ "${CGO_ENABLED}" = "1" ]; then
            echo "Note: CGO_ENABLED=1 (required for collectors using CGO)"
        fi
        echo
        echo "Step 1: Building ${OUTPUT_NAME} binary..."
        
        export CGO_ENABLED="${CGO_ENABLED}"
        export CGO_CFLAGS="${CGO_CFLAGS}"
        export GOOS=linux
        export GOARCH="${ARCH}"
        
        go build ${VERBOSE_FLAG} \
            -o "dist/datakit-linux-${ARCH}/externals/${OUTPUT_BINARY}" \
            -ldflags "${LDFLAGS}" \
            "${SOURCE_PATH}"

    elif [ "${DOCKER_AVAILABLE}" = "true" ]; then
        # Docker build for cross-compilation with QEMU
        echo "Build Method: Docker with QEMU (cross-compilation)"
        echo
        echo "Step 1: Building Docker image for ${ARCH}..."
        
        # Use buildx if available or forced
        USE_BUILDX=false
        
        if [ "${BUILDX_AVAILABLE}" = "true" ] || [ "${FORCE_BUILDX}" = "true" ]; then
            # Ensure binfmt is installed for QEMU
            echo "  Checking QEMU binfmt support..."
            if ! ls /proc/sys/fs/binfmt_misc/ | grep -q qemu; then
                echo "  Installing QEMU binfmt support..."
                sudo docker run --privileged --rm tonistiigi/binfmt --install all 2>/dev/null || \
                    echo "  Warning: Could not install binfmt, will try anyway..."
            fi
            
            if [ "${FORCE_BUILDX}" = "true" ]; then
                echo "  Building with Docker Buildx (forced, no fallback)..."
            else
                echo "  Building with Docker Buildx..."
            fi
            
            # Build image using buildx (uses local moby/buildkit image, no network pull needed)
            if sudo docker buildx build \
                --platform linux/${ARCH} \
                --build-arg TARGETARCH=${ARCH} \
                -f "${PROJECT_ROOT}/dockerfiles/Dockerfile_externals" \
                -t dk-external-collector-builder:${ARCH} \
                --load \
                "${PROJECT_ROOT}"; then
                USE_BUILDX=true
            else
                if [ "${FORCE_BUILDX}" = "true" ]; then
                    echo "✗ Buildx failed and --force-buildx is set, aborting..."
                    exit 1
                else
                    echo "  Buildx failed, falling back to legacy build..."
                fi
            fi
        fi
        
        # Fallback to legacy build if buildx failed or not available
        if [ "${USE_BUILDX}" != "true" ]; then
            echo "  Using legacy docker build..."
            sudo docker build \
                --platform linux/${ARCH} \
                --build-arg TARGETARCH=${ARCH} \
                -f "${PROJECT_ROOT}/dockerfiles/Dockerfile_externals" \
                -t dk-external-collector-builder:${ARCH} \
                "${PROJECT_ROOT}" || {
                echo "✗ Docker build failed for ${ARCH}"
                echo "  Note: Cross-compilation requires QEMU support"
                echo "  Try: sudo pacman -S qemu-user-static"
                exit 1
            }
        fi

        echo "Step 2: Building ${OUTPUT_NAME} binary in container..."
        echo "  Docker env vars: CGO_ENABLED=${CGO_ENABLED}, CGO_CFLAGS=${CGO_CFLAGS}, GOOS=linux, GOARCH=${ARCH}"
        
        # Retry logic for QEMU instability (GCC sometimes crashes under emulation)
        MAX_RETRIES=3
        RETRY_COUNT=0
        BUILD_SUCCESS=false
        
        while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
            RETRY_COUNT=$((RETRY_COUNT + 1))
            echo "  Build attempt $RETRY_COUNT of $MAX_RETRIES..."
            
            if sudo docker run --rm \
                -v "${PROJECT_ROOT}:/root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit" \
                -w /root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit \
                -e CGO_ENABLED="${CGO_ENABLED}" \
                -e CGO_CFLAGS="${CGO_CFLAGS}" \
                -e GOOS=linux \
                -e GOARCH="${ARCH}" \
                dk-external-collector-builder:${ARCH} \
                bash -c "go build ${VERBOSE_FLAG} \
                    -o \"dist/datakit-linux-${ARCH}/externals/${OUTPUT_BINARY}\" \
                    -ldflags \"${LDFLAGS}\" \
                    \"${SOURCE_PATH}\""; then
                BUILD_SUCCESS=true
                echo "  ✓ Build succeeded on attempt $RETRY_COUNT"
                break
            else
                BUILD_EXIT_CODE=$?
                echo "  ✗ Build failed on attempt $RETRY_COUNT (exit code: $BUILD_EXIT_CODE)"
                
                if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
                    echo "  Retrying in 2 seconds..."
                    sleep 2
                fi
            fi
        done
        
        if [ "$BUILD_SUCCESS" = false ]; then
            echo "✗ Build failed after $MAX_RETRIES attempts"
            echo "  This is likely due to QEMU emulation instability (GCC crashes under emulation)"
            echo "  Consider building on native ARM64 hardware or in CI with ARM64 runners"
            exit 1
        fi

        # Fix permissions
        sudo chown $(whoami):$(whoami) "dist/datakit-linux-${ARCH}/externals/${OUTPUT_BINARY}" 2>/dev/null || true
    else
        echo "✗ Cannot build for ${ARCH}: Docker required for cross-compilation"
        echo "  Host: ${HOST_ARCH}, Target: ${ARCH}"
        echo "  Solution: Install and start Docker, or build on native hardware"
        exit 1
    fi

    # Verify binary
    echo
    echo "Step 3: Verifying binary..."
    BINARY_PATH="${PROJECT_ROOT}/dist/datakit-linux-${ARCH}/externals/${OUTPUT_BINARY}"

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

    # Step 4: Clean up Docker image to save disk space (only for Docker builds)
    if [ "${DOCKER_AVAILABLE}" = "true" ] && [ "${ARCH}" != "${HOST_ARCH}" ]; then
        echo
        echo "Step 4: Cleaning up Docker image..."
        sudo docker rmi dk-external-collector-builder:${ARCH} 2>/dev/null &&
            echo "✓ Removed Docker image: dk-external-collector-builder:${ARCH}" ||
            echo "  Note: Docker image cleanup skipped (may be in use or already removed)"
    fi

    echo "✓ Build complete for linux/${ARCH}"
    echo
done

# Clean up buildx builder if it was created
if [ "${BUILDX_AVAILABLE}" = "true" ] || [ "${FORCE_BUILDX}" = "true" ]; then
    echo "Cleaning up Docker Buildx builder..."
    sudo docker buildx rm datakit-builder 2>/dev/null &&
        echo "✓ Removed buildx builder: datakit-builder" ||
        echo "  Note: Buildx builder cleanup skipped (may not exist)"
fi

echo "=========================================="
echo "✓ All builds completed successfully!"
echo "=========================================="
echo
echo "Built binaries:"
for ARCH in "${ARCH_ARRAY[@]}"; do
    ARCH=$(echo "${ARCH}" | xargs)
    echo "  - dist/datakit-linux-${ARCH}/externals/${OUTPUT_BINARY}"
done
echo
echo "Next steps:"
echo "  1. Test binary: ./dist/datakit-linux-<arch>/externals/${OUTPUT_BINARY} --help"
echo "  2. Build Docker image with the binary"
echo "  3. Deploy to target environment"
echo
