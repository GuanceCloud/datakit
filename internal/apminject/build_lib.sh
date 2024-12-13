#!/bin/sh

# set -x
docker run --privileged --rm pubrepo.jiagouyun.com/ebpf-dev/binfmt:qemu-v7.0.0 \
    --install all

repo_name="gitlab.jiagouyun.com/cloudcare-tools/datakit"
container_repo_dir=/root/go/src/$repo_name

target_arch=$1
if [ -z "$target_arch" ]; then
    target_arch=$(uname -m | sed -e s/x86_64/amd64/ \
        -e s/aarch64.\*/arm64/)
fi

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

    docker run --rm --platform "$arch" \
        -v "$(go env GOPATH)"/src/$repo_name:$container_repo_dir \
        -w$container_repo_dir/internal/apminject \
        $image make $target DIST_DIR="$container_repo_dir"/"$dist_rela_dir" || exit $?
}

make -f "$(go env GOPATH)"/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/Makefile \
    rewriter DIST_DIR="$(go env GOPATH)"/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/"$dist_rela_dir" \
    ARCH="$target_arch" REPO_PATH="$(go env GOPATH)"/src/gitlab.jiagouyun.com/cloudcare-tools/datakit || exit $?

docker_build "$target_arch" glibc
docker_build "$target_arch" musl
