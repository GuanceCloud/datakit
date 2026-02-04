PY_SPY_IMAGE_VER=0.1.0
ASYNC_PROFILER_IMAGE_VER=0.6.1
PPROF_IMAGE_VER=0.1.0
DOCKER_IMAGE_REPO?="not_set"

TARGET=async-profiler-uos async-profiler async-profiler-alpine

.PHONY: async-profiler async-profiler-alpine async-profiler-uos go-pprof py-spy

build: $(TARGET)

async-profiler:
	cd async-profiler && docker buildx build --platform linux/arm64,linux/amd64 \
		-t pubrepo.guance.com/datakit-operator/async-profiler:$(ASYNC_PROFILER_IMAGE_VER) \
		-f Dockerfile --push .

async-profiler-alpine:
	cd async-profiler && docker buildx build --platform linux/arm64,linux/amd64 \
		-t pubrepo.guance.com/datakit-operator/async-profiler:$(ASYNC_PROFILER_IMAGE_VER)-alpine \
		-f alpine.Dockerfile --push .

async-profiler-uos:
	cd async-profiler && docker buildx build --platform linux/arm64,linux/amd64 \
		-t pubrepo.guance.com/datakit-operator/async-profiler:$(ASYNC_PROFILER_IMAGE_VER)-uos \
		-f uos.Dockerfile --push .

go-pprof:
	cd go-pprof && docker buildx build --platform linux/arm64,linux/amd64 \
		-t pubrepo.guance.com/datakit-operator/go-pprof:$(PPROF_IMAGE_VER) \
		-f Dockerfile --push .

py-spy:
	cd py-spy && docker buildx build --platform linux/arm64,linux/amd64 \
        -t pubrepo.guance.com/datakit-operator/py-spy:$(PY_SPY_IMAGE_VER) \
        -f Dockerfile --push .

