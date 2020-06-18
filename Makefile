.PHONY: default test local

default: local

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit

# 本地环境
LOCAL_DOWNLOAD_ADDR = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit

PUB_DIR = pub
BIN = datakit
NAME = datakit
ENTRY = cmd/datakit/main.go

LOCAL_ARCHS = "linux/amd64|linux/386"
#LOCAL_ARCHS = "all"
DEFAULT_ARCHS = "all"

VERSION := $(shell git describe --always --tags)
DATE := $(shell date +'%Y-%m-%d %H:%M:%S')
GOVERSION := $(shell go version)
COMMIT := $(shell git rev-parse --short HEAD)
UPLOADER:= ${USER}

###################################
# Detect telegraf update info
###################################
TELEGRAF_VERSION := $(shell cd telegraf && git describe --exact-match --tags 2>/dev/null)
TELEGRAF_BRANCH := $(shell cd telegraf && git rev-parse --abbrev-ref HEAD)
TELEGRAF_COMMIT := $(shell cd telegraf && git rev-parse --short HEAD)
TELEGRAF_LDFLAGS := $(LDFLAGS) -w -s -X main.commit=$(TELEGRAF_COMMIT) -X main.branch=$(TELEGRAF_BRANCH) # use -w -s to strip the binary size
ifdef TELEGRAF_VERSION
	TELEGRAF_LDFLAGS += -X main.version=$(TELEGRAF_VERSION)
endif

all: test release preprod local

define build
	@echo "===== $(BIN) $(1) ===="
	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p build $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo 'package git; const (BuildAt string="$(DATE)"; Version string="$(VERSION)"; Golang string="$(GOVERSION)"; Sha1 string="$(COMMIT)"; Uploader string="$(UPLOADER)");' > git/git.go
	@go run cmd/make/make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build  \
		 -release $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)
	tree -Csh build pub
endef

define pub
	echo "publish $(1) $(NAME) ..."
	go run cmd/make/make.go -pub -release $(1) -pub-dir $(PUB_DIR) -name $(NAME) -download-addr $(2) -archs $(3)
endef

check:
	@golangci-lint run --timeout 1h # https://golangci-lint.run/usage/install/#local-installation
	@go vet ./...

local:
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

test:
	$(call build,test, $(DEFAULT_ARCHS), $(TEST_DOWNLOAD_ADDR))

release:
	$(call build,release, $(DEFAULT_ARCHS), $(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_test:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_image:
	@sudo docker build --tag registry.jiagouyun.com/datakit/datakit:$(VERSION) -f internal-dk.Dockerfile .
	@sudo docker push registry.jiagouyun.com/datakit/datakit:$(VERSION)

define build_agent
	-git submodule add -f https://github.com/influxdata/telegraf.git

	@echo "==== build telegraf... ===="
	cd telegraf && go mod download

	# Linux
	cd telegraf && GOOS=linux   GOARCH=amd64   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-amd64/agent     ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=386     GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-386/agent       ./cmd/telegraf
	#cd telegraf && GOOS=linux  GOARCH=s390x   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-s390x/agent    ./cmd/telegraf
	#cd telegraf && GOOS=linux  GOARCH=ppc64le GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-ppc64le/agent  ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=arm     GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-arm/agent      ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=arm64   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-arm64/agent    ./cmd/telegraf

	# Mac
	cd telegraf && GOOS=darwin  GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/darwin-amd64/agent      ./cmd/telegraf

	# FreeBSD
	#cd telegraf && GOOS=freebsd GOARCH=386   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-386/agent      ./cmd/telegraf
	#cd telegraf && GOOS=freebsd GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-amd64/agent    ./cmd/telegraf

	# Windows
	cd telegraf && GOOS=windows GOARCH=386   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/windows-386/agent.exe   ./cmd/telegraf
	cd telegraf && GOOS=windows GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/windows-amd64/agent.exe ./cmd/telegraf

	tree -Csh embed
endef

.PHONY: agent
agent:
	$(call build_agent)

clean:
	rm -rf build/*
	rm -rf $(PUB_DIR)/*
	rm -rf embed/*
