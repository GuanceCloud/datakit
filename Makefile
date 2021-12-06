.PHONY: default testing local man

default: local

# 正式环境
PRODUCTION_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit

# 测试环境
TESTING_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit

# 本地环境: 需配置环境变量，便于完整测试采集器的发布、更新等流程
# export LOCAL_OSS_ACCESS_KEY='<your-oss-AK>'
# export LOCAL_OSS_SECRET_KEY='<your-oss-SK>'
# export LOCAL_OSS_BUCKET='<your-oss-bucket>'
# export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com' # 一般都是这个地址
# export LOCAL_OSS_ADDR='<your-oss-bucket>.oss-cn-hangzhou.aliyuncs.com/datakit'
# 如果只是编译，LOCAL_OSS_ADDR 这个环境变量可以随便给个值
LOCAL_DOWNLOAD_ADDR=${LOCAL_OSS_ADDR}

PUB_DIR = dist
BUILD_DIR = dist

BIN = datakit
NAME = datakit
ENTRY = cmd/datakit/main.go

LOCAL_ARCHS:="local"
DEFAULT_ARCHS:="all"
MAC_ARCHS:="darwin/amd64"
DINGDING_TOKEN?="not-set"
DINGDING_DEBUG_HINT?=
GIT_VERSION?=$(shell git describe --always --tags)
DATE:=$(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION:=$(shell go version)
COMMIT:=$(shell git rev-parse --short HEAD)
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)
COMMITER:=$(shell git log -1 --pretty=format:'%an')
UPLOADER:=$(shell hostname)/${USER}/${COMMITER}

GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GO_PATCH_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f3)
MINIMUM_SUPPORTED_GO_MAJOR_VERSION = 1
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 16
GO_VERSION_VALIDATION_ERR_MSG = Golang version is not supported, please update to at least $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION).$(MINIMUM_SUPPORTED_GO_MINOR_VERSION)
BUILDER_GOOS_GOARCH="$(shell go env GOOS)-$(shell go env GOARCH)"

GOLINT_VERSION = "$(shell golangci-lint --version | cut -c 27- | cut -d' ' -f1)"
SUPPORTED_GOLINT_VERSION = "1.42.1"
SUPPORTED_GOLINT_VERSION_ANOTHER = "v1.42.1"
GOLINT_VERSION_VALIDATION_ERR_MSG = golangci-lint version($(GOLINT_VERSION)) is not supported, please use version $(SUPPORTED_GOLINT_VERSION)

#####################
# Large strings
#####################

define GIT_INFO
//nolint
package git

const (
	BuildAt  string = "$(DATE)"
	Version  string = "$(GIT_VERSION)"
	Golang   string = "$(GOVERSION)"
	Commit   string = "$(COMMIT)"
	Branch   string = "$(GIT_BRANCH)"
	Uploader string = "$(UPLOADER)"
);
endef
export GIT_INFO

# -------------
# 注意：这些 XXX_NOTIFY_MSG 只能定义在一行中，跨行会报错
# -------------

define LOCAL_NOTIFY_MSG
{ "msgtype": "text", "text": { "content": "$(UPLOADER) 「私自」发布了 DataKit 测试版($(GIT_VERSION))。\n\nLinux/Mac 安装：\nDK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$$(curl -L https://$(LOCAL_DOWNLOAD_ADDR)/install-$(GIT_VERSION).sh)\"\n\nWindows 安装：\n$$env:DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://$(LOCAL_DOWNLOAD_ADDR)/install-$(GIT_VERSION).ps1 -destination .install.ps1; powershell .install.ps1;\n\nLinux/Mac 升级：\nDK_UPGRADE=1 bash -c \"$$(curl -L https://$(LOCAL_DOWNLOAD_ADDR)/install-$(GIT_VERSION).sh)\"\n\nWindows 升级：\n$$env:DK_UPGRADE=\"1\"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://$(LOCAL_DOWNLOAD_ADDR)/install-$(GIT_VERSION).ps1 -destination .install.ps1; powershell .install.ps1;$(DINGDING_DEBUG_HINT)" }}
endef
export LOCAL_NOTIFY_MSG

define TESTING_NOTIFY_MSG
{ "msgtype": "text", "text": { "content": "$(UPLOADER) 发布了 DataKit 测试版($(GIT_VERSION))。\n\nLinux/Mac 安装：DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\" bash -c \"$$(curl -L https://$(TESTING_DOWNLOAD_ADDR)/install-$(GIT_VERSION).sh)\"\n\nWindows 安装：$$env:DK_DATAWAY=\"https://openway.guance.com?token=<TOKEN>\";Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://$(TESTING_DOWNLOAD_ADDR)/install-$(GIT_VERSION).ps1 -destination .install.ps1; powershell .install.ps1;\n\nLinux/Mac 升级：DK_UPGRADE=1 bash -c \"$$(curl -L https://$(TESTING_DOWNLOAD_ADDR)/install-$(GIT_VERSION).sh)\"\n\nWindows 升级：$$env:DK_UPGRADE=\"1\"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://$(TESTING_DOWNLOAD_ADDR)/install-$(GIT_VERSION).ps1 -destination .install.ps1; powershell .install.ps1;$(DINGDING_DEBUG_HINT)" } }
endef
export TESTING_NOTIFY_MSG

define CI_PASS_NOTIFY_MSG
{ "msgtype": "text", "text": { "content": "$(UPLOADER) 触发的 DataKit CI 通过$(DINGDING_DEBUG_HINT)" } }
endef
export CI_PASS_NOTIFY_MSG

define NOTIFY_MSG_RELEASE
{ "msgtype": "text", "text": { "content": "$(UPLOADER) 发布了 DataKit 新版本($(GIT_VERSION))$(DINGDING_DEBUG_HINT)" } }
endef
export NOTIFY_MSG_RELEASE

define NOTIFY_CI
{ "msgtype": "text", "text": { "content": "$(COMMITER)正在执行 DataKit CI，此刻请勿在CI分支[$(GIT_BRANCH)]提交代码，以免 CI 任务失败$(DINGDING_DEBUG_HINT)" }}
endef
export NOTIFY_CI

LINUX_RELEASE_VERSION = $(shell uname -r)

define build
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		exit 0 ; \
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	fi

	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p $(BUILD_DIR) $(PUB_DIR)/$(1)
	@echo "===== $(BIN) $(1) ===="
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go \
		-main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir $(BUILD_DIR) \
		-release $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3)
	@tree -Csh -L 3 $(BUILD_DIR)
endef

define pub
	@echo "publish $(1) $(NAME) ..."
	@GO111MODULE=off go run cmd/make/make.go \
		-pub -release $(1) -pub-dir $(PUB_DIR) \
		-name $(NAME) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
endef

define check_golint_version
	@case $(GOLINT_VERSION) in \
	$(SUPPORTED_GOLINT_VERSION)) \
	;; \
	$(SUPPORTED_GOLINT_VERSION_ANOTHER)) \
	;; \
	*) \
		echo '$(GOLINT_VERSION_VALIDATION_ERR_MSG)'; \
		exit 1; \
	esac;
endef

define dingding_notify
	@if [[ $(DINGDING_TOKEN) == "not-set" ]]; then \
		printf "\033[31m [FAIL] DINGDING_TOKEN not set%s\n\033[0m"; \
		exit 1; \
	fi
	@curl \
		'https://oapi.dingtalk.com/robot/send?access_token=$(DINGDING_TOKEN)' \
		-H 'Content-Type: application/json' \
		-d '$(1)'
endef

local: deps
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

build: prepare man gofmt lfparser_disable_line plparser_disable_line
	$(call build, testing, $(DEFAULT_ARCHS), $(TESTING_DOWNLOAD_ADDR))

testing: deps
	$(call build, testing, $(DEFAULT_ARCHS), $(TESTING_DOWNLOAD_ADDR))

production: deps
	$(call build, production, $(DEFAULT_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))

release_mac: deps
	$(call build, production, $(MAC_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))

testing_mac: deps
	$(call build, testing, $(MAC_ARCHS), $(TESTING_DOWNLOAD_ADDR))

pub_local:
	$(call pub, local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_testing:
	$(call pub, testing,$(TESTING_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_testing_mac:
	$(call pub, testing,$(TESTING_DOWNLOAD_ADDR),$(MAC_ARCHS))

pub_testing_win_img:
	@mkdir -p embed/windows-amd64
	@wget --quiet -O - "https://$(TESTING_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t registry.jiagouyun.com/datakit/datakit-win:$(GIT_VERSION) -f ./Dockerfile_win .
	@sudo docker push registry.jiagouyun.com/datakit/datakit-win:$(GIT_VERSION)

pub_testing_img:
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(TESTING_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker buildx build --platform linux/arm64,linux/amd64 \
		-t registry.jiagouyun.com/datakit/datakit:$(GIT_VERSION) . --push

pub_release_win_img:
	# release to pub hub
	@mkdir -p embed/windows-amd64
	@wget --quiet -O - "https://$(PRODUCTION_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t pubrepo.jiagouyun.com/datakit/datakit-win:$(GIT_VERSION) -f ./Dockerfile_win .
	@sudo docker push pubrepo.jiagouyun.com/datakit/datakit-win:$(GIT_VERSION)

pub_production_img:
	# release to pub hub
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(PRODUCTION_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker buildx build --platform linux/arm64,linux/amd64 -t \
		pubrepo.jiagouyun.com/datakit/datakit:$(GIT_VERSION) . --push

pub_production:
	$(call pub,production,$(PRODUCTION_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_release_mac:
	$(call pub,production,$(PRODUCTION_DOWNLOAD_ADDR),$(MAC_ARCHS))

ci_pass_notify:
	$(call dingding_notify, $(CI_PASS_NOTIFY_MSG))

test_notify:
	$(call dingding_notify, $(TESTING_NOTIFY_MSG))

local_notify:
	$(call dingding_notify, $(LOCAL_NOTIFY_MSG))

production_notify:
	$(call dingding_notify, $(NOTIFY_MSG_RELEASE))

ci_notify:
	$(call dingding_notify, $(NOTIFY_CI))

check_conf_compatible:
	./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-config --config-dir samples
	./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-sample

define build_ip2isp
	rm -rf china-operator-ip
	git clone -b ip-lists https://github.com/gaoyifan/china-operator-ip.git
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -build-isp
endef

ip2isp:
	$(call build_ip2isp)

deps: prepare man gofmt lfparser_disable_line plparser_disable_line lint_with_exit

man:
	@packr2 clean
	@packr2

# ignore files under vendor/.git/git
# install gofumpt: go install mvdan.cc/gofumpt@latest
gofmt:
	@GO111MODULE=off gofumpt -w -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/git/\|.*_y.go")

vet:
	@go vet ./...

# all testing
at: test_deps
	@truncate -s 0 test.output
	@echo "#####################" | tee -a test.output
	@echo "#" $(DATE) | tee -a test.output
	@echo "#" $(GIT_VERSION) | tee -a test.output
	@echo "#####################" | tee -a test.output
	for pkg in `go list ./...`; do \
		echo "# testing $$pkg..." | tee -a test.output; \
		GO111MODULE=off CGO_ENABLED=1 go test -race -timeout 30s \
			-cover -benchmem -bench . $$pkg | tee -a test.output; \
		echo "######################" | tee -a test.output; \
	done

test_deps: prepare man gofmt lfparser_disable_line plparser_disable_line vet

lint:
	$(call check_golint_version)
	@golangci-lint run --fix | tee lint.err

lint_with_exit:
	$(call check_golint_version)
	@golangci-lint run --fix

lfparser_disable_line:
	@rm -rf io/parser/gram_y.go
	@rm -rf io/parser/gram.y.go
	@rm -rf io/parser/parser.y.go
	@rm -rf io/parser/parser_y.go
	@goyacc -l -o io/parser/gram_y.go io/parser/gram.y # use -l to disable `//line`

plparser_disable_line:
	@rm -rf pipeline/parser/gram_y.go
	@rm -rf pipeline/parser/gram.y.go
	@rm -rf pipeline/parser/parser.y.go
	@rm -rf pipeline/parser/parser_y.go
	@goyacc -l -o pipeline/parser/gram_y.go pipeline/parser/gram.y # use -l to disable `//line`

prepare:
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go

clean:
	@rm -rf build/*
	@rm -rf io/parser/gram_y.go
	@rm -rf io/parser/gram.y.go
	@rm -rf pipeline/parser/parser.y.go
	@rm -rf pipeline/parser/parser_y.go
	@rm -rf pipeline/parser/gram.y.go
	@rm -rf pipeline/parser/gram_y.go
	@rm -rf check.err
	@rm -rf $(PUB_DIR)/*
