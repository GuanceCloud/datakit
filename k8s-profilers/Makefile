# Unless explicitly stated otherwise all files in this repository are licensed
# under the MIT License.
# This product includes software developed at Guance Cloud (https://www.guance.com/).
# Copyright 2021-present Guance, Inc.

PY_SPY_IMAGE_VER=0.1.0
ASYNC_PROFILER_IMAGE_VER=0.1.0
PPROF_IMAGE_VER=0.1.0
DOCKER_IMAGE_REPO?="not_set"

TARGET=async-profiler go-pprof py-spy
#TARGET=async-profiler py-spy go-pprof
#TARGET=go-pprof
.PHONY: async-profiler go-pprof py-spy

build: $(TARGET)

async-profiler:
	cd async-profiler && docker buildx build --platform linux/arm64,linux/amd64 \
		-t $(DOCKER_IMAGE_REPO)/async-profiler:$(ASYNC_PROFILER_IMAGE_VER) \
		-t $(DOCKER_IMAGE_REPO)/async-profiler:latest \
		-f Dockerfile --push .

go-pprof:
	cd go-pprof && docker buildx build --platform linux/arm64,linux/amd64 \
		-t $(DOCKER_IMAGE_REPO)/go-pprof:$(PPROF_IMAGE_VER) \
		-t $(DOCKER_IMAGE_REPO)/go-pprof:latest \
		-f Dockerfile --push .

py-spy:
	cd py-spy && docker buildx build --platform linux/arm64,linux/amd64 \
        -t $(DOCKER_IMAGE_REPO)/py-spy:$(PY_SPY_IMAGE_VER) \
        -t $(DOCKER_IMAGE_REPO)/py-spy:latest \
        -f Dockerfile --push .
