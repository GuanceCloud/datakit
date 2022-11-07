.PHONY: default testing local deps prepare plparser_disable_line

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
LOCAL_OSS_ADDR?="not-set"
LOCAL_DOWNLOAD_ADDR=${LOCAL_OSS_ADDR}

PUB_DIR = dist
BUILD_DIR = dist

BIN = datakit
NAME = datakit
NAME_EBPF = datakit-ebpf
ENTRY = cmd/datakit/main.go

UNAME_S:=$(shell uname -s)
UNAME_M:=$(shell uname -m | sed -e s/x86_64/x86_64/ -e s/aarch64.\*/arm64/)
LOCAL_ARCHS:="local"
DEFAULT_ARCHS:="all"
MAC_ARCHS:="darwin/amd64"
NOT_SET="not-set"
VERSION?=$(shell git describe --always --tags)
DATE:=$(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION:=$(shell go version)
COMMIT:=$(shell git rev-parse --short HEAD)
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)
COMMITER:=$(shell git log -1 --pretty=format:'%an')
UPLOADER:=$(shell hostname)/${USER}/${COMMITER}
DOCKER_IMAGE_ARCHS:="linux/arm64,linux/amd64"
DATAKIT_EBPF_ARCHS?="linux/arm64,linux/amd64"
IGN_EBPF_INSTALL_ERR?=0
RACE_DETECTION?="off"

GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GO_PATCH_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f3)
MINIMUM_SUPPORTED_GO_MAJOR_VERSION = 1
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 16
GO_VERSION_VALIDATION_ERR_MSG = Golang version is not supported, please update to at least $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION).$(MINIMUM_SUPPORTED_GO_MINOR_VERSION)
BUILDER_GOOS_GOARCH=$(shell go env GOOS)-$(shell go env GOARCH)

GOLINT_BINARY = golangci-lint
GOLINT_VERSION = "$(shell $(GOLINT_BINARY) --version | cut -c 27- | cut -d' ' -f1)"
SUPPORTED_GOLINT_VERSION = "1.46.2"
SUPPORTED_GOLINT_VERSION_ANOTHER = "v1.46.2"
GOLINT_VERSION_VALIDATION_ERR_MSG = golangci-lint version($(GOLINT_VERSION)) is not supported, please use version $(SUPPORTED_GOLINT_VERSION)

#####################
# Large strings
#####################

define GIT_INFO
// Package git used to define basic git info abount current version.
package git

//nolint
const (
	BuildAt  string = "$(DATE)"
	Version  string = "$(VERSION)"
	Golang   string = "$(GOVERSION)"
	Commit   string = "$(COMMIT)"
	Branch   string = "$(GIT_BRANCH)"
	Uploader string = "$(UPLOADER)"
);
endef
export GIT_INFO

define notify_build
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		exit 0 ; \
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	fi
	@echo "===== notify $(BIN) $(1) ===="
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go \
		-main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir $(BUILD_DIR) \
		-release $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3) \
		-notify-only
endef

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
		-release $(1) -pub-dir $(PUB_DIR) -archs $(2) -download-addr $(3) -race $(RACE_DETECTION)
	@tree -Csh -L 3 $(BUILD_DIR)
endef

define pub
	@echo "publish $(1) $(NAME) ..."
	@GO111MODULE=off go run cmd/make/make.go \
		-pub -release $(1) -pub-dir $(PUB_DIR) \
		-name $(NAME) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
endef

define pub_ebpf
	@echo "publish $(1) $(NAME_EBPF) ..."
	@GO111MODULE=off go run cmd/make/make.go \
		-pub-ebpf -release $(1) -pub-dir $(PUB_DIR) \
		-name $(NAME_EBPF) -download-addr $(2) \
		-build-dir $(BUILD_DIR) -archs $(3)
endef

define build_docker_image
	@if [ $(2) = "registry.jiagouyun.com" ]; then \
		echo 'publish to $(2)...'; \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/datakit:$(VERSION) . --push --build-arg IGN_EBPF_INSTALL_ERR=$(IGN_EBPF_INSTALL_ERR); \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/logfwd:$(VERSION) -f Dockerfile_logfwd . --push ; \
	else \
		echo 'publish to $(2)...'; \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/datakit:$(VERSION) \
			-t $(2)/dataflux/datakit:$(VERSION) \
			-t $(2)/dataflux-prev/datakit:$(VERSION) . --push --build-arg IGN_EBPF_INSTALL_ERR=$(IGN_EBPF_INSTALL_ERR); \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/logfwd:$(VERSION) \
			-t $(2)/dataflux/logfwd:$(VERSION) \
			-t $(2)/dataflux-prev/logfwd:$(VERSION) -f Dockerfile_logfwd . --push; \
	fi
endef

define build_k8s_charts
	@helm repo ls
	@echo `echo $(VERSION) | cut -d'-' -f1`
	@sed -e "s,{{tag}},$(VERSION),g" -e "s,{{repository}},$(2)/datakit/datakit,g" charts/values.yaml > charts/datakit/values.yaml
	@helm package charts/datakit --version `echo $(VERSION) | cut -d'-' -f1` --app-version `echo $(VERSION)`
	@if [ $$((`echo $(VERSION) | awk -F . '{print $$2}'`%2)) -eq 0 ];then \
        helm cm-push datakit-`echo $(VERSION) | cut -d'-' -f1`.tgz $(1); \
     else \
			  printf "\033[31m [FAIL] unstable version not allowed\n\033[0m"; \
        exit 1; \
     fi

	@rm -f datakit-`echo $(VERSION) | cut -d'-' -f1`.tgz
endef

define show_poor_logs
  # 没有传参的日志，我们认为其日志信息是不够完整的，日志的意义也相对不大
	@grep --color=always --exclude-dir=vendor --exclude-dir=.git --exclude=*.html -nr '\.Debugf(\|\.Debug(\|\.Infof(\|\.Info(\|\.Warnf(\|\.Warn(\|\.Errorf(\|\.Error(' . | grep -vE ","
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

local: deps
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

pub_local: deps
	$(call pub, local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_ebpf_local: deps
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))
	$(call pub_ebpf, local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_epbf_testing: deps
	$(call build,testing, $(DATAKIT_EBPF_ARCHS), $(TESTING_DOWNLOAD_ADDR))
	$(call pub_ebpf, testing,$(TESTING_DOWNLOAD_ADDR),$(DATAKIT_EBPF_ARCHS))

pub_ebpf_production: deps
	$(call build,production, $(DATAKIT_EBPF_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))
	$(call pub_ebpf, production,$(PRODUCTION_DOWNLOAD_ADDR),$(DATAKIT_EBPF_ARCHS))

testing_notify: deps
	$(call notify_build,testing, $(DEFAULT_ARCHS), $(TESTING_DOWNLOAD_ADDR))

testing: deps
	$(call build, testing, $(DEFAULT_ARCHS), $(TESTING_DOWNLOAD_ADDR))
	$(call pub, testing,$(TESTING_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

testing_image:
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'registry.jiagouyun.com')
	# we also publish testing image to public image repo
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'pubrepo.jiagouyun.com')
	$(call build_k8s_charts, 'datakit-testing', registry.jiagouyun.com)

production_notify: deps
	$(call notify_build,production, $(DEFAULT_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))

production: deps # stable release
	$(call build, production, $(DEFAULT_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))
	$(call pub, production, $(PRODUCTION_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

production_image:
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'pubrepo.jiagouyun.com')
	$(call build_k8s_charts, 'datakit', pubrepo.guance.com)

production_mac: deps
	$(call build, production, $(MAC_ARCHS), $(PRODUCTION_DOWNLOAD_ADDR))
	$(call pub,production,$(PRODUCTION_DOWNLOAD_ADDR),$(MAC_ARCHS))

testing_mac: deps
	$(call build, testing, $(MAC_ARCHS), $(TESTING_DOWNLOAD_ADDR))
	$(call pub, testing,$(TESTING_DOWNLOAD_ADDR),$(MAC_ARCHS))

# not used
pub_testing_win_img:
	@mkdir -p embed/windows-amd64
	@wget --quiet -O - "https://$(TESTING_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t registry.jiagouyun.com/datakit/datakit-win:$(VERSION) -f ./Dockerfile_win .
	@sudo docker push registry.jiagouyun.com/datakit/datakit-win:$(VERSION)


# not used
pub_testing_charts:
	@helm package ${CHART_PATH%/*} --version $(VERSION) --app-version $(VERSION)
	@helm helm cm-push ${TEMP\#\#*/}-$TAG.tgz datakit-test-chart

# not used
pub_release_win_img:
	# release to pub hub
	@mkdir -p embed/windows-amd64
	@wget --quiet -O - "https://$(PRODUCTION_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t pubrepo.jiagouyun.com/datakit/datakit-win:$(VERSION) -f ./Dockerfile_win .
	@sudo docker push pubrepo.jiagouyun.com/datakit/datakit-win:$(VERSION)



# Config samples should only be published by production release,
# because config samples in multiple testing releases may not be compatible to each other.
pub_conf_samples:
	@echo "upload config samples to oss..."
	@go run cmd/make/make.go -dump-samples -release production

# testing/production downloads config samples from different oss bucket.
check_testing_conf_compatible:
	@go run cmd/make/make.go -download-samples -release testing
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-config --config-dir samples
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-sample

check_production_conf_compatible:
	@go run cmd/make/make.go -download-samples -release production
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-config --config-dir samples
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit --check-sample

shame_logging:
	$(call show_poor_logs)

define build_ip2isp
	rm -rf china-operator-ip
	git clone -b ip-lists https://github.com/gaoyifan/china-operator-ip.git
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -build-isp
endef

define do_lint
	$(GOLINT_BINARY) --version
	GOARCH=$(1) GOOS=$(2) $(GOLINT_BINARY) run --fix --allow-parallel-runners
endef

ip2isp:
	$(call build_ip2isp)

deps: prepare gofmt lfparser_disable_line plparser_disable_line

# ignore files under vendor/.git/git
gofmt:
	@GO111MODULE=off gofmt -w -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/git/\|.*_y.go\|packed-packr.go")

vet:
	@go vet ./...

ut: deps
	@GO111MODULE=off CGO_ENABLED=1 go run cmd/make/make.go -ut

# all testing

all_test: deps
	@truncate -s 0 test.output
	@echo "#####################" | tee -a test.output
	@echo "#" $(DATE) | tee -a test.output
	@echo "#" $(VERSION) | tee -a test.output
	@echo "#####################" | tee -a test.output
	i=0; \
	for pkg in `go list ./... | grep -vE 'datakit/git'`; do \
		echo "# testing $$pkg..." | tee -a test.output; \
		GO111MODULE=off CGO_ENABLED=1 LOGGER_PATH=nul go test -timeout 1m -cover $$pkg; \
		if [ $$? != 0 ]; then \
			printf "\033[31m [FAIL] %s\n\033[0m" $$pkg; \
			i=`expr $$i + 1`; \
		else \
			echo "######################"; \
			fi \
	done; \
	if [ $$i -gt 0 ]; then \
		printf "\033[31m %d case failed.\n\033[0m" $$i; \
		exit 1; \
	else \
		printf "\033[32m all testinig passed.\n\033[0m"; \
	fi

test_deps: prepare gofmt lfparser_disable_line plparser_disable_line vet

lint: deps check_man copyright_check
	$(call check_golint_version)
	if [ $(UNAME_S) != Darwin ] && [ $(UNAME_M) != arm64 ]; then \
		echo '============== lint under amd64/linux ==================='; \
		$(GOLINT_BINARY) --version; \
		GOARCH=amd64 GOOS=linux $(GOLINT_BINARY) run --fix --allow-parallel-runners ; \
	fi
	@echo '============== lint under amd64/darwin==================='
	$(call do_lint,amd64,darwin)
	@echo '============== lint under 386/windows ==================='
	$(call do_lint,386,windows)
	@echo '============== lint under amd64/windows ==================='
	$(call do_lint,amd64,windows)
	@echo '============== lint under arm/linux ==================='
	$(call do_lint,arm,linux)
	@echo '============== lint under arm64/linux ==================='
	$(call do_lint,arm64,linux)
	@echo '============== lint under 386/linux ==================='
	$(call do_lint,386,linux)

lfparser_disable_line:
	@rm -rf io/parser/gram_y.go
	@rm -rf io/parser/parser_y.go
	@goyacc -l -o io/parser/gram_y.go io/parser/gram.y # use -l to disable `//line`

plparser_disable_line:
	@rm -rf pipeline/parser/gram_y.go
	@rm -rf pipeline/parser/parser.y.go

	@rm -rf pipeline/core/parser/gram_y.go
	@rm -rf pipeline/core/parser/parser.y.go
	@goyacc -l -o pipeline/core/parser/gram_y.go pipeline/core/parser/gram.y # use -l to disable `//line`

prepare:
	@mkdir -p git
	@echo "$$GIT_INFO" > git/git.go

copyright_check:
	@python3 copyright.py --dry-run && \
		{ echo "copyright check ok"; exit 0; } || \
		{ echo "copyright check failed"; exit -1; }

copyright_check_auto_fix:
	@python3 copyright.py --fix

md_lint:
	# markdownlint install: https://github.com/igorshubovych/markdownlint-cli
	@markdownlint man/manuals 2>&1 > md.lint
	@if [ $$? != 0 ]; then \
		cat md.lint; \
		exit -1; \
	fi

# 要求所有文档的章节必须带上指定的标签（历史原因，先忽略 changelog.md）
check_man:
	@grep --color=always --exclude man/manuals/changelog.md -nr '^##' man/manuals/* | grep -vE ' {#' | grep -vE '{{' && \
		{ echo "[E] some bad docs"; exit -1; } || \
		{ echo "all docs ok"; exit 0; }

code_stat:
	cloc --exclude-dir=vendor,tests --exclude-lang=JSON,HTML .

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
