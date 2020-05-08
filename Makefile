.PHONY: default test

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
ENTRY = main.go

VERSION := $(shell git describe --always --tags)

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
	@export GO111MODULE=off
	@mkdir -p build $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo 'package git; const (BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build  \
		 -release $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)
	tree -Csh build pub
endef

define pub
	echo "publish $(1) $(NAME) ..."
	go run make.go -pub -release $(1) -pub-dir $(PUB_DIR) -name $(NAME) -download-addr $(2) -archs $(3)
endef

local:
	$(call build,local, "windows/amd64|linux/amd64", $(LOCAL_DOWNLOAD_ADDR))

test:
	$(call build,test, "all", $(TEST_DOWNLOAD_ADDR))

release:
	$(call build,release,"all", $(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),"linux/amd64|windows/amd64")

pub_test:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),"all")

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),"all")

define build_agent
	@echo "==== build telegraf... ===="
	cd telegraf && go mod download
	cd telegraf && GOOS=darwin  GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/darwin-amd64/agent      ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-amd64/agent       ./cmd/telegraf  
	cd telegraf && GOOS=linux   GOARCH=386   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-386/agent         ./cmd/telegraf  
	cd telegraf && GOOS=freebsd GOARCH=386   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-386/agent       ./cmd/telegraf  
	cd telegraf && GOOS=freebsd GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-amd64/agent     ./cmd/telegraf  
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
