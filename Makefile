.PHONY: default testing local deps prepare cspell

default: local

# For production release, download address point to CDN, upload address point to aliyun OSS
PRODUCTION_UPLOAD_ADDR  ?= zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit
PRODUCTION_DOWNLOAD_CDN ?= static.guance.com/datakit
TESTING_UPLOAD_ADDR     ?= zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit # For testing: same download/upload address
TESTING_DOWNLOAD_CDN    ?= zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit
LOCAL_OSS_ADDR          ?= "not-set" # you should export these env in your make environment.
LOCAL_UPLOAD_ADDR       ?= ${LOCAL_OSS_ADDR}
LOCAL_DOWNLOAD_CDN      ?= ${LOCAL_OSS_ADDR} # CDN set as the same OSS bucket

# Local envs to publish local testing binaries.
# export LOCAL_OSS_ACCESS_KEY = '<your-oss-AK>'
# export LOCAL_OSS_SECRET_KEY = '<your-oss-SK>'
# export LOCAL_OSS_BUCKET     = '<your-oss-bucket>'
# export LOCAL_OSS_HOST       = 'oss-cn-hangzhou.aliyuncs.com'
# export LOCAL_OSS_ADDR       = '<your-oss-bucket>.oss-cn-hangzhou.aliyuncs.com/datakit'

PUB_DIR            = dist
BUILD_DIR          = dist
BIN                = datakit
NAME               = datakit
NAME_EBPF          = datakit-ebpf
ENTRY              = cmd/datakit/main.go
LOCAL_ARCHS        = local
DEFAULT_ARCHS      = all
MAC_ARCHS          = darwin/amd64
DOCKER_IMAGE_ARCHS = linux/arm64,linux/amd64
GOLINT_BINARY      = golangci-lint
CGO_FLAGS          = "-Wno-undef-prefix -Wno-deprecated-declarations" # to disable warnings from gopsutil on macOS
HL                 = \033[0;32m # high light
NC                 = \033[0m    # no color
RED                = \033[31m   # red

SUPPORTED_GOLINT_VERSION         = 1.46.2
SUPPORTED_GOLINT_VERSION_ANOTHER = v1.46.2

MINIMUM_GO_MAJOR_VERSION = 1
MINIMUM_GO_MINOR_VERSION = 16
GO_VERSION_ERR_MSG := Golang version is not supported, please update to at least $(MINIMUM_GO_MAJOR_VERSION).$(MINIMUM_GO_MINOR_VERSION)

# Make them evaluate(expand) only once
UNAME_S                := $(shell uname -s)
UNAME_M                := $(shell uname -m | sed -e s/x86_64/x86_64/ -e s/aarch64.\*/arm64/)
DATE                   := $(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION              := $(shell go version)
COMMIT                 := $(shell git rev-parse --short HEAD)
COMMITER               := $(shell git log -1 --pretty=format:'%an')
UPLOADER               := $(shell hostname)/${USER}/${COMMITER}
GO_MAJOR_VERSION       := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION       := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GO_PATCH_VERSION       := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f3)
BUILDER_GOOS_GOARCH    := $(shell go env GOOS)-$(shell go env GOARCH)
GOLINT_VERSION         := $(shell $(GOLINT_BINARY) --version | cut -c 27- | cut -d' ' -f1)
GOLINT_VERSION_ERR_MSG := golangci-lint version($(GOLINT_VERSION)) is not supported, please use version $(SUPPORTED_GOLINT_VERSION)

# These can be override at runtime by make variables
VERSION              ?= $(shell git describe --always --tags)
DATAWAY_URL          ?= "not-set"
GIT_BRANCH           ?= $(shell git rev-parse --abbrev-ref HEAD)
DATAKIT_EBPF_ARCHS   ?= linux/arm64,linux/amd64
IGN_EBPF_INSTALL_ERR ?= 0
RACE_DETECTION       ?= "off"
PKGEBPF              ?= false
AUTO_FIX             ?= on
UT_EXCLUDE           ?= "not-set"
DOCKER_REMOTE_HOST   ?= "0.0.0.0" # default use localhost as docker server

ifneq ($(PKGEBPF), false)
	PKGEBPF_FLAG = -pkg-ebpf
endif

# Generate 'internal/git' package
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

##############################################################################
# Functions used within the Makefile
##############################################################################

define notify_build
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_GO_MAJOR_VERSION) ]; then \
		exit 0 ; \
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_ERR_MSG)';\
		exit 1; \
	fi
	@echo "===== notify $(BIN) $(1) ===="
	GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) go run -tags with_inputs cmd/make/make.go \
		-main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir $(BUILD_DIR) \
		-release $(1) -pub-dir $(PUB_DIR) -archs $(2) -upload-addr $(3) -download-cdn $(4) \
		-notify-only
endef

# build used to compile datakit binary and related dists
define build_bin
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_GO_MAJOR_VERSION) ]; then \
		exit 0 ; \
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_ERR_MSG)';\
		exit 1; \
	fi

	@rm -rf $(PUB_DIR)/$(1)/*
	@mkdir -p $(BUILD_DIR) $(PUB_DIR)/$(1)
	@echo "===== building $(BIN) $(1) ====="
	GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) go run -tags with_inputs cmd/make/make.go \
		-release $(1)             \
		-archs $(2)               \
		-upload-addr $(3)         \
		-download-cdn $(4)        \
		-main $(ENTRY)            \
		-binary $(BIN)            \
		-name $(NAME)             \
		-build-dir $(BUILD_DIR)   \
		-pub-dir $(PUB_DIR)       \
		-race $(RACE_DETECTION)
	@tree -Csh -L 3 $(BUILD_DIR)
endef

# pub used to publish datakit version(for release/testing/local)
define publish
	@echo "===== publishing $(1) $(NAME) ====="
	GO111MODULE=off CGO_CFLAGS=$(CGO_FLAGS) go run -tags with_inputs cmd/make/make.go \
		-release $(1)            \
		-upload-addr $(2)        \
		-download-cdn $(3)       \
		-pub                     \
		-pub-dir $(PUB_DIR)      \
		-name $(NAME)            \
		-build-dir $(BUILD_DIR)  \
		-archs $(4)              \
		$(PKGEBPF_FLAG)
endef

define pub_ebpf
	@echo "===== publishing $(1) $(NAME_EBPF) ====="
	@GO111MODULE=off CGO_CFLAGS=$(CGO_FLAGS) go run -tags with_inputs cmd/make/make.go \
		-release $(1)             \
		-upload-addr $(2)         \
		-archs $(3)               \
		-pub-ebpf                 \
		-pub-dir $(PUB_DIR)       \
		-name $(NAME_EBPF)        \
		-build-dir $(BUILD_DIR)
endef

define build_docker_image
	@if [ $(2) = "registry.jiagouyun.com" ]; then \
		echo 'publishing to $(2)...'; \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/datakit:$(VERSION) . --push --build-arg IGN_EBPF_INSTALL_ERR=$(IGN_EBPF_INSTALL_ERR); \
		sudo docker buildx build --platform $(1) \
			-t $(2)/datakit/logfwd:$(VERSION) -f Dockerfile_logfwd . --push ; \
	else \
		echo 'publishing to $(2)...'; \
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
	@sed -e "s,{{repository}},$(2)/datakit/datakit,g" charts/values.yaml > charts/datakit/values.yaml
	@helm package charts/datakit --version `echo $(VERSION) | cut -d'-' -f1` --app-version `echo $(VERSION)`
	@helm cm-push datakit-`echo $(VERSION) | cut -d'-' -f1`.tgz $(1)
	@rm -f datakit-`echo $(VERSION) | cut -d'-' -f1`.tgz
endef

define show_poor_logs
  # 没有传参的日志，我们认为其日志信息是不够完整的，日志的意义也相对不大
	@grep --color=always \
		--exclude-dir=vendor \
		--exclude="*_test.go" \
		--exclude-dir=.git \
		--exclude=*.html \
		-nr '\.Debugf(\|\.Debug(\|\.Infof(\|\.Info(\|\.Warnf(\|\.Warn(\|\.Errorf(\|\.Error(' . | grep -vE "\"|uhttp|\`" && \
		{ echo "[E] some bad loggings in code"; exit -1; } || \
		{ echo "all loggings ok"; exit ; }
endef

define check_golint_version
	@case $(GOLINT_VERSION) in \
	$(SUPPORTED_GOLINT_VERSION)) \
	;; \
	$(SUPPORTED_GOLINT_VERSION_ANOTHER)) \
	;; \
	*) \
		echo '$(GOLINT_VERSION_ERR_MSG)'; \
		exit 1; \
	esac;
endef

define build_ip2isp
	rm -rf china-operator-ip
	git clone -b ip-lists https://github.com/gaoyifan/china-operator-ip.git
	@GO111MODULE=off CGO_ENABLED=0 go run -tags with_inputs cmd/make/make.go -build-isp
endef

##############################################################################
# Rules in the Makefile
##############################################################################

local: deps
	$(call build_bin, local, $(LOCAL_ARCHS), $(LOCAL_UPLOAD_ADDR), $(LOCAL_DOWNLOAD_CDN))

pub_local: deps
	$(call publish, local, $(LOCAL_UPLOAD_ADDR), $(LOCAL_DOWNLOAD_CDN), $(LOCAL_ARCHS))

pub_ebpf_local: deps
	$(call build_bin, local, $(LOCAL_ARCHS), $(LOCAL_UPLOAD_ADDR), $(LOCAL_DOWNLOAD_CDN))
	$(call pub_ebpf, local, $(LOCAL_DOWNLOAD_CDN), $(LOCAL_ARCHS))

pub_epbf_testing: deps
	$(call build_bin, testing, $(DATAKIT_EBPF_ARCHS), $(TESTING_UPLOAD_ADDR), $(TESTING_DOWNLOAD_CDN))
	$(call pub_ebpf, testing, $(TESTING_DOWNLOAD_CDN), $(DATAKIT_EBPF_ARCHS))

pub_ebpf_production: deps
	$(call build_bin, production, $(DATAKIT_EBPF_ARCHS), $(PRODUCTION_DOWNLOAD_CDN), $(TESTING_DOWNLOAD_CDN))
	$(call pub_ebpf, production, $(PRODUCTION_DOWNLOAD_CDN), $(DATAKIT_EBPF_ARCHS))

testing_notify: deps
	$(call notify_build, testing, $(DEFAULT_ARCHS), $(TESTING_UPLOAD_ADDR), $(TESTING_DOWNLOAD_CDN))

testing: deps
	$(call build_bin, testing, $(DEFAULT_ARCHS), $(TESTING_UPLOAD_ADDR), $(TESTING_DOWNLOAD_CDN))
	$(call publish, testing, $(TESTING_UPLOAD_ADDR), $(TESTING_DOWNLOAD_CDN), $(DEFAULT_ARCHS))

testing_image:
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'registry.jiagouyun.com')
	# we also publishing testing image to public image repo
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'pubrepo.guance.com')
	$(call build_k8s_charts, 'datakit-testing', registry.jiagouyun.com)

production_notify: deps
	$(call notify_build,production, $(DEFAULT_ARCHS), $(PRODUCTION_DOWNLOAD_CDN), $(TESTING_DOWNLOAD_CDN))

production: deps # stable release
	$(call build_bin, production, $(DEFAULT_ARCHS), $(PRODUCTION_UPLOAD_ADDR), $(PRODUCTION_DOWNLOAD_CDN))
	$(call publish, production, $(PRODUCTION_UPLOAD_ADDR), $(PRODUCTION_DOWNLOAD_CDN), $(DEFAULT_ARCHS))

production_image:
	$(call build_docker_image, $(DOCKER_IMAGE_ARCHS), 'pubrepo.guance.com')
	$(call build_k8s_charts, 'datakit', pubrepo.guance.com)

production_mac: deps
	$(call build_bin, production, $(MAC_ARCHS), $(PRODUCTION_UPLOAD_ADDR), $(PRODUCTION_DOWNLOAD_CDN))
	$(call publish, production, $(PRODUCTION_UPLOAD_ADDR), $(PRODUCTION_DOWNLOAD_CDN), $(MAC_ARCHS))

# not used
pub_testing_win_img:
	@mkdir -p embed/windows-amd64
	@wget --quiet -O - "https://$(TESTING_UPLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
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
	@wget --quiet -O - "https://$(PRODUCTION_UPLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t pubrepo.guance.com/datakit/datakit-win:$(VERSION) -f ./Dockerfile_win .
	@sudo docker push pubrepo.guance.com/datakit/datakit-win:$(VERSION)

# Config samples should only be published by production release,
# because config samples in multiple testing releases may not be compatible to each other.
pub_conf_samples:
	@echo "upload config samples to oss..."
	@CGO_CFLAGS=$(CGO_FLAGS) go run cmd/make/make.go -dump-samples -release production

# testing/production downloads config samples from different oss bucket.
check_testing_conf_compatible:
	@CGO_CFLAGS=$(CGO_FLAGS) go run cmd/make/make.go -download-samples -release testing
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit check --config --config-dir samples
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit check --sample

check_production_conf_compatible:
	@CGO_CFLAGS=$(CGO_FLAGS) go run cmd/make/make.go -download-samples -release production
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit check --config --config-dir samples
	@LOGGER_PATH=nul ./dist/datakit-$(BUILDER_GOOS_GOARCH)/datakit check --sample

# 没有传参的日志，我们认为其日志信息是不够完整的，日志的意义也相对不大
shame_logging:
	@grep --color=always \
		--exclude-dir=vendor \
		--exclude="*_test.go" \
		--exclude-dir=.git \
		--exclude=*.html \
		-nr '\.Debugf(\|\.Debug(\|\.Infof(\|\.Info(\|\.Warnf(\|\.Warn(\|\.Errorf(\|\.Error(' . | grep -vE "\"|uhttp|\`" && \
		{ echo "[E] some bad loggings in code"; exit -1; } || { echo "all loggings ok"; exit 0; }

ip2isp:
	$(call build_ip2isp)

deps: prepare gofmt 

# ignore files under vendor/.git/git
gofmt:
	@GO111MODULE=off gofmt -w -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/git/\|.*_y.go\|packed-packr.go")

vet:
	@go vet ./...

ut: deps
	CGO_CFLAGS=$(CGO_FLAGS) GO111MODULE=off CGO_ENABLED=1 \
						 REMOTE_HOST=$(DOCKER_REMOTE_HOST) \
						 go run cmd/make/make.go -ut -ut-exclude $(UT_EXCLUDE) \
						 -dataway-url $(DATAWAY_URL); \
		if [ $$? != 0 ]; then \
			exit 1; \
		else \
			echo "######################"; \
		fi

code_lint: deps copyright_check
ifeq ($(AUTO_FIX),on)
		@printf "$(HL)lint with auto fix...\n$(NC)"; \
			$(GOLINT_BINARY) run --fix --allow-parallel-runners;
else
		@printf "$(HL)lint without auto fix...\n$(NC)"; \
			$(GOLINT_BINARY) run --allow-parallel-runners;
endif

	@if [ $$? != 0 ]; then \
		printf "$(RED)[FAIL] lint failed\n$(NC)" $$pkg; \
		exit -1; \
	fi

# lint code and document
lint: code_lint md_lint

prepare:
	@mkdir -p internal/git
	@echo "$$GIT_INFO" > internal/git/git.go

copyright_check:
	@python3 copyright.py --dry-run && \
		{ echo "copyright check ok"; exit 0; } || \
		{ echo "copyright check failed"; exit -1; }

copyright_check_auto_fix:
	@python3 copyright.py --fix

define check_docs
	# check spell on docs
	@echo 'version of cspell: $(shell cspell --version)'
	cspell lint --show-suggestions -c scripts/cspell.json --no-progress $(1)/**/*.md | tee dist/cspell.lint

  # check markdown style
	# markdownlint install: https://github.com/igorshubovych/markdownlint-cli
	@echo 'version of markdownlint: $(shell markdownlint --version)'
	@truncate -s 0 dist/md-lint.json
	markdownlint -c scripts/markdownlint.yml -j -o dist/md-lint.json $(1) 

	@if [ -s dist/md-lint.json ]; then \
		printf "$(RED) [FAIL] dist/md-lint.json not empty \n$(NC)"; \
		exit -1; \
	fi

	@if [ -s dist/cspell.lint ]; then \
		printf "$(RED) [FAIL] dist/cspell.lint not empty \n$(NC)"; \
		exit -1; \
	fi
endef

exportdir=dist/export
# only check ZH docs, EN docs too many errors
docs_dir=$(exportdir)/guance-doc/docs/zh
docs_template_dir=internal/export/doc/zh

md_lint:
	@GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) \
		go run cmd/make/make.go \
		--mdcheck $(docs_template_dir) \
		--mdcheck-autofix=$(AUTO_FIX) # check doc templates
	@rm -rf $(exportdir) && mkdir -p $(exportdir)
	@bash export.sh -D $(exportdir) -E -V 0.0.0
	@GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) \
		go run cmd/make/make.go -mdcheck $(docs_dir) \
		--mdcheck-autofix off # disable autofix on checking generated documents
	$(call check_docs,$(docs_dir))

project_words:
	cspell -c cspell/cspell.json --words-only --unique internal/export/doc/zh/** | sort --ignore-case >> project-words.txt

code_stat:
	cloc --exclude-dir=vendor,tests --exclude-lang=JSON,HTML .

# promlinter: show prometheuse metrics defined in Datakit.
# go install github.com/yeya24/promlinter/cmd/promlinter@latest
show_metrics:
	@promlinter list . --add-help -o md --with-vendor

clean:
	@rm -rf build/*
	@rm -rf internal/io/parser/gram_y.go
	@rm -rf internal/io/parser/gram.y.go
	@rm -rf internal/pipeline/parser/parser.y.go
	@rm -rf internal/pipeline/parser/parser_y.go
	@rm -rf internal/pipeline/parser/gram.y.go
	@rm -rf internal/pipeline/parser/gram_y.go
	@rm -rf check.err
	@rm -rf $(PUB_DIR)/*
