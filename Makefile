.PHONY: default test local

default: local

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit

# 预发环境
PRE_DOWNLOAD_ADDR = zhuyun-static-files-preprod.oss-cn-hangzhou.aliyuncs.com/datakit

# 本地环境
LOCAL_DOWNLOAD_ADDR = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit

PUB_DIR = dist
BUILD_DIR = dist

BIN = datakit
NAME = datakit
ENTRY = cmd/datakit/main.go

LOCAL_ARCHS = "local"
DEFAULT_ARCHS = "all"

VERSION := $(shell git describe --always --tags)
DATE := $(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION := $(shell go version)
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMITER := $(shell git log -1 --pretty=format:'%an')
UPLOADER:= $(shell hostname)/${USER}/${COMMITER}

NOTIFY_MSG_RELEASE:=$(shell echo '{"msgtype": "text","text": {"content": "$(UPLOADER) 发布了 DataKit 新版本($(VERSION))"}}')
NOTIFY_MSG_TEST:=$(shell echo '{"msgtype": "text","text": {"content": "$(UPLOADER) 发布了 DataKit 测试版($(VERSION))"}}')
NOTIFY_CI:=$(shell echo '{"msgtype": "text","text": {"content": "$(COMMITER)正在执行DataKit CI，此刻请勿在CI分支(dev/master)提交代码，避免CI任务失败[摊手]"}}')

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

define GIT_INFO
//nolint
package git
const (
	BuildAt string="$(DATE)"
	Version string="$(VERSION)"
	Golang string="$(GOVERSION)"
	Commit string="$(COMMIT)"
	Branch string="$(BRANCH)"
	Uploader string="$(UPLOADER)"
);
endef
export GIT_INFO

define build
	@echo "===== $(BIN) $(1) ===="
	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p $(BUILD_DIR) $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir $(BUILD_DIR) \
		 -env $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)
	@tree -Csh -L 3 $(BUILD_DIR)
endef

define pub
	@echo "publish $(1) $(NAME) ..."
	@GO111MODULE=off go run cmd/make/make.go -pub -env $(1) -pub-dir $(PUB_DIR) -name $(NAME) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
	@tree -Csh -L 3 $(PUB_DIR)
endef

lint:
	@golangci-lint run | tee lint.err # https://golangci-lint.run/usage/install/#local-installation

vet:
	@go vet ./...

local:
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

test:
	$(call build,test, $(DEFAULT_ARCHS), $(TEST_DOWNLOAD_ADDR))

preprod:
	$(call build,preprod, $(DEFAULT_ARCHS), $(PRE_DOWNLOAD_ADDR))

release:
	$(call build,release, $(DEFAULT_ARCHS), $(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_test:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_testing_img:
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(TEST_DOWNLOAD_ADDR)/telegraf/agent-linux-amd64.tar.gz" | tar -xz -C .
	@sudo docker build -t registry.jiagouyun.com/datakit/datakit:$(VERSION) .
	@sudo docker push registry.jiagouyun.com/datakit/datakit:$(VERSION)

pub_release_img:
	# release to pub hub
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(RELEASE_DOWNLOAD_ADDR)/telegraf/agent-linux-amd64.tar.gz" | tar -xz -C .
	@sudo docker build -t pubrepo.jiagouyun.com/dataflux/datakit:$(VERSION) .
	@sudo docker push pubrepo.jiagouyun.com/dataflux/datakit:$(VERSION)

pub_agent:
	@go run cmd/make/make.go -pub-agent -env local -pub-dir embed -download-addr $(LOCAL_DOWNLOAD_ADDR) -archs $(LOCAL_ARCHS)
	@go run cmd/make/make.go -pub-agent -env test -pub-dir embed -download-addr $(TEST_DOWNLOAD_ADDR) -archs $(DEFAULT_ARCHS)
	@go run cmd/make/make.go -pub-agent -env preprod -pub-dir embed -download-addr $(PRE_DOWNLOAD_ADDR) -archs $(DEFAULT_ARCHS)
	@go run cmd/make/make.go -pub-agent -env release -pub-dir embed -download-addr $(RELEASE_DOWNLOAD_ADDR) -archs $(DEFAULT_ARCHS)

pub_preprod:
	$(call pub,preprod,$(PRE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

test_notify:
	@curl \
		'https://oapi.dingtalk.com/robot/send?access_token=245327454760c3587f40b98bdd44f125c5d81476a7e348a2cc15d7b339984c87' \
		-H 'Content-Type: application/json' \
		-d '$(NOTIFY_MSG_TEST)'

release_notify:
	@curl \
		'https://oapi.dingtalk.com/robot/send?access_token=5109b365f7be669c45c5677418a1c2fe7d5251485a09f514131177b203ed785f' \
		-H 'Content-Type: application/json' \
		-d '$(NOTIFY_MSG_RELEASE)'

ci_notify:
	@curl \
		'https://oapi.dingtalk.com/robot/send?access_token=245327454760c3587f40b98bdd44f125c5d81476a7e348a2cc15d7b339984c87' \
		-H 'Content-Type: application/json' \
		-d '$(NOTIFY_CI)'

define build_agent
	@#git rm -rf telegraf
	@#- git submodule add -f https://github.com/influxdata/telegraf.git

	@echo "==== build telegraf... ===="
	@cd telegraf && go mod download

	# Linux
	cd telegraf && GOOS=linux   GOARCH=amd64   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-amd64/agent    ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=386     GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-386/agent      ./cmd/telegraf
	@#cd telegraf && GOOS=linux  GOARCH=s390x   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-s390x/agent    ./cmd/telegraf
	@#cd telegraf && GOOS=linux  GOARCH=ppc64le GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-ppc64le/agent  ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=arm     GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-arm/agent      ./cmd/telegraf
	cd telegraf && GOOS=linux   GOARCH=arm64   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/linux-arm64/agent    ./cmd/telegraf

	# Mac
	cd telegraf && GOOS=darwin  GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/darwin-amd64/agent      ./cmd/telegraf

	## FreeBSD
	##cd telegraf && GOOS=freebsd GOARCH=386   GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-386/agent      ./cmd/telegraf
	##cd telegraf && GOOS=freebsd GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(TELEGRAF_LDFLAGS)" -o ../embed/freebsd-amd64/agent    ./cmd/telegraf

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
