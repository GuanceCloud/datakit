#!/bin/sh

set -e

repo_name="gitlab.jiagouyun.com/cloudcare-tools/datakit"
script_dir=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd)
container_repo_dir=/root/go/src/$repo_name

target_arch=$1
if [ -z "$target_arch" ]; then
    target_arch=$(uname -m | sed -e s/x86_64/amd64/ \
        -e s/aarch64.\*/arm64/)
fi

host_arch=$(uname -m | sed -e s/x86_64/amd64/ \
    -e s/aarch64.\*/arm64/)

docker_mode=""

resolve_docker_mode() {
    if [ -n "$docker_mode" ]; then
        return 0
    fi

    if docker info >/dev/null 2>&1; then
        docker_mode="docker"
        return 0
    fi

    if sudo -n docker info >/dev/null 2>&1; then
        docker_mode="sudo"
        return 0
    fi

    echo "docker is not accessible: neither docker nor sudo -n docker can connect to daemon" >&2
    return 1
}

docker_exec() {
    resolve_docker_mode || return 1

    if [ "$docker_mode" = "sudo" ]; then
        sudo -n docker "$@"
    else
        docker "$@"
    fi
}

dist_rela_dir=dist/datakit-apm-inject-linux-$target_arch
if [ -n "$2" ]; then
    dist_rela_dir=$2
fi

docker_build() {
    arch=$1
    libc=$2
    image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev:1.1
    target=launcher

    if [ "$libc" = "musl" ]; then
        image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev-musl:1.1
        target=launcher_musl
    fi

    docker_exec run --rm --platform "$arch" \
        -v "$repo_root":$container_repo_dir \
        -w$container_repo_dir/internal/apminject \
        $image make $target DIST_DIR="$container_repo_dir"/"$dist_rela_dir" || exit $?
}

docker_host_build() {
    arch=$1
    libc=$2
    cc=$3
    image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev:1.1
    target=launcher

    if [ "$libc" = "musl" ]; then
        image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev-musl:1.1
        target=launcher_musl
    fi

    docker_exec run --rm --platform "linux/$host_arch" \
        -v "$repo_root":$container_repo_dir \
        -w$container_repo_dir/internal/apminject \
        $image make "$target" \
        DIST_DIR="$container_repo_dir"/"$dist_rela_dir" \
        ARCH="$arch" REPO_PATH="$container_repo_dir" CC="$cc" || exit $?
}

resolve_local_cc() {
    arch=$1
    libc=$2

    case "$arch:$libc" in
        amd64:glibc)
            if command -v x86_64-linux-gnu-gcc >/dev/null 2>&1; then
                echo x86_64-linux-gnu-gcc
                return 0
            fi
            if [ "$host_arch" = "amd64" ] && command -v gcc >/dev/null 2>&1; then
                echo gcc
                return 0
            fi
            ;;
        amd64:musl)
            if command -v x86_64-linux-musl-gcc >/dev/null 2>&1; then
                echo x86_64-linux-musl-gcc
                return 0
            fi
            if [ "$host_arch" = "amd64" ] && command -v musl-gcc >/dev/null 2>&1; then
                echo musl-gcc
                return 0
            fi
            ;;
        arm64:glibc)
            if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
                echo aarch64-linux-gnu-gcc
                return 0
            fi
            if [ "$host_arch" = "arm64" ] && command -v gcc >/dev/null 2>&1; then
                echo gcc
                return 0
            fi
            ;;
        arm64:musl)
            if command -v aarch64-linux-musl-gcc >/dev/null 2>&1; then
                echo aarch64-linux-musl-gcc
                return 0
            fi
            if [ "$host_arch" = "arm64" ] && command -v musl-gcc >/dev/null 2>&1; then
                echo musl-gcc
                return 0
            fi
            ;;
    esac

    return 1
}

resolve_docker_cc() {
    arch=$1
    libc=$2
    image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev:1.1
    candidates=""

    case "$arch:$libc" in
        amd64:glibc)
            candidates="x86_64-linux-gnu-gcc"
            if [ "$host_arch" = "amd64" ]; then
                candidates="$candidates gcc"
            fi
            ;;
        amd64:musl)
            image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev-musl:1.1
            candidates="x86_64-linux-musl-gcc"
            if [ "$host_arch" = "amd64" ]; then
                candidates="$candidates musl-gcc gcc"
            fi
            ;;
        arm64:glibc)
            candidates="aarch64-linux-gnu-gcc"
            if [ "$host_arch" = "arm64" ]; then
                candidates="$candidates gcc"
            fi
            ;;
        arm64:musl)
            image=pubrepo.jiagouyun.com/ebpf-dev/apm-inject-dev-musl:1.1
            candidates="aarch64-linux-musl-gcc"
            if [ "$host_arch" = "arm64" ]; then
                candidates="$candidates musl-gcc gcc"
            fi
            ;;
    esac

    docker_exec run --rm --platform "linux/$host_arch" "$image" sh -c '
for cc in '"$candidates"'; do
    if command -v "$cc" >/dev/null 2>&1; then
        echo "$cc"
        exit 0
    fi
done
exit 1
' 2>/dev/null
}

build_local() {
    arch=$1
    libc=$2
    cc=$(resolve_local_cc "$arch" "$libc") || return 1

    target=launcher
    if [ "$libc" = "musl" ]; then
        target=launcher_musl
    fi

    make -f "$repo_root"/internal/apminject/Makefile \
        "$target" DIST_DIR="$repo_root"/"$dist_rela_dir" \
        ARCH="$arch" REPO_PATH="$repo_root" CC="$cc" || exit $?
}

build_in_host_container() {
    arch=$1
    libc=$2
    cc=$(resolve_docker_cc "$arch" "$libc") || return 1
    docker_host_build "$arch" "$libc" "$cc"
}

ensure_binfmt() {
    docker_exec run --privileged --rm pubrepo.jiagouyun.com/ebpf-dev/binfmt:qemu-v7.0.0 \
        --install all
}

check_launcher_arch() {
    so_path=$1
    arch=$2

    if command -v readelf >/dev/null 2>&1; then
        machine=$(readelf -h "$so_path" | awk -F: '/Machine:/ {gsub(/^[ \t]+/, "", $2); print $2}')
        case "$arch" in
            amd64)
                echo "$machine" | grep -q "X86-64" && return 0
                ;;
            arm64)
                echo "$machine" | grep -q "AArch64" && return 0
                ;;
        esac

        echo "unexpected ELF machine for $so_path: $machine, target arch: $arch" >&2
        return 1
    fi

    if command -v file >/dev/null 2>&1; then
        desc=$(file "$so_path")
        case "$arch" in
            amd64)
                echo "$desc" | grep -qi "x86-64" && return 0
                ;;
            arm64)
                echo "$desc" | grep -qi "aarch64" && return 0
                ;;
        esac

        echo "unexpected file type for $so_path: $desc, target arch: $arch" >&2
        return 1
    fi

    echo "cannot verify ELF arch for $so_path: neither readelf nor file is available" >&2
    return 1
}

make -f "$repo_root"/internal/apminject/Makefile \
    rewriter dkrunc DIST_DIR="$repo_root"/"$dist_rela_dir" \
    ARCH="$target_arch" REPO_PATH="$repo_root" || exit $?

if build_local "$target_arch" glibc; then
    :
elif build_in_host_container "$target_arch" glibc; then
    :
else
    if [ "$target_arch" != "$host_arch" ]; then
        ensure_binfmt
    fi
    docker_build "$target_arch" glibc
fi
check_launcher_arch "$repo_root"/"$dist_rela_dir"/apm_launcher.so "$target_arch"

if build_local "$target_arch" musl; then
    :
elif build_in_host_container "$target_arch" musl; then
    :
else
    if [ "$target_arch" != "$host_arch" ]; then
        ensure_binfmt
    fi
    docker_build "$target_arch" musl
fi
check_launcher_arch "$repo_root"/"$dist_rela_dir"/apm_launcher_musl.so "$target_arch"
