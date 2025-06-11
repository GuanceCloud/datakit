.PHONY: default testing local deps prepare cspell

default: local


# ligai version notify settings
LIGAI_CUSTOMFIELD       ?=NOT_SET
LIGAI_AUTO_DEVOPS_TOKEN ?=NOT_SET
LIGAI_API               ?=NOT_SET

DOCKER_IMAGE_REPO      ?= NOT_SET
BRAND                  ?= NOT_SET
DIST_DIR               = dist
BIN                    = datakit
NAME                   = datakit
NAME_EBPF              = datakit-ebpf
ENTRY                  = cmd/datakit/main.go
LOCAL_ARCHS            = local
DEFAULT_ARCHS          = all
MAC_ARCHS              = darwin/amd64
DOCKER_IMAGE_ARCHS     = linux/arm64,linux/amd64
UOS_DOCKER_IMAGE_ARCHS = linux/arm64,linux/amd64
DCA_BUILD_ARCH         = linux/arm64,linux/amd64
GOLINT_BINARY         ?= golangci-lint
CGO_FLAGS              = "-Wno-undef-prefix -Wno-deprecated-declarations" # to disable warnings from gopsutil on macOS
HL                     = \033[0;32m # high light
NC                     = \033[0m    # no color
RED                    = \033[31m   # red
LOG_LEVEL             ?= "info"

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
VERSION                      ?= $(shell git describe --always --tags)
DCA_VERSION                  ?= NOT_SET
DATAWAY_URL                  ?= NOT_SET
GIT_BRANCH                   ?= $(shell git rev-parse --abbrev-ref HEAD)
DATAKIT_EBPF_ARCHS           ?= linux/arm64,linux/amd64
RACE_DETECTION               ?= off
PKGEBPF                      ?= 0
DLEBPF                       ?= 0
AUTO_FIX                     ?= true
UT_EXCLUDE                   ?= "-"
UT_ONLY                      ?= "-"
UT_PARALLEL                  ?= "0"
DOCKER_REMOTE_HOST           ?= "0.0.0.0" # default use localhost as docker server
DOCKER_IMAGE_PROJECT_PATH    ?= NOT_SET
DOCKERFILE_SUFFIX            ?= NOT_SET
HELM_CHART_DIR               ?= "charts/datakit"
MERGE_REQUEST_TARGET_BRANCH  ?= ""
ONLY_BUILD_INPUTS_EXTENTIONS ?= 0

# Generate 'internal/git' package
define GIT_INFO
// Package git used to define basic git info abount current version.
package git

//nolint
const (
	BuildAt    string = "$(DATE)"
	Version    string = "$(VERSION)"
	Golang     string = "$(GOVERSION)"
	Commit     string = "$(COMMIT)"
	Branch     string = "$(GIT_BRANCH)"
	Uploader   string = "$(UPLOADER)"
	DCAVersion string = "$(DCA_VERSION)"
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
		-log-level $(LOG_LEVEL) \
		-main $(ENTRY) \
		-binary $(BIN) \
		-name $(NAME) \
		-release $(1) \
		-dist-dir $(DIST_DIR) \
		-archs $(2) \
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

	@rm -rf $(DIST_DIR)/$(1)/*
	@mkdir -p $(DIST_DIR)/$(1)
	@echo "===== building $(BIN) $(1) ====="
	GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) go run \
		-tags with_inputs cmd/make/make.go      \
		-log-level $(LOG_LEVEL)                 \
		-release $(1)                           \
		-archs $(2)                             \
		-main $(ENTRY)                          \
		-binary $(BIN)                          \
		-name $(NAME)                           \
		-dist-dir $(DIST_DIR)                   \
		-race $(RACE_DETECTION)                 \
		-brand $(BRAND)                         \
		-docker-image-repo $(DOCKER_IMAGE_REPO) \
		-helm-chart-dir $(HELM_CHART_DIR)       \
		-pkg-ebpf $(PKGEBPF)                    \
		-only-external-inputs $(ONLY_BUILD_INPUTS_EXTENTIONS)
	@tree -Csh -L 3 $(DIST_DIR)
endef

# pub used to publish datakit version(for release/testing/local)
define publish
	@echo "===== publishing $(1) $(NAME) ====="
	GO111MODULE=off CGO_CFLAGS=$(CGO_FLAGS) go run \
		-tags with_inputs cmd/make/make.go      \
		-log-level $(LOG_LEVEL)                 \
		-release $(1)                           \
		-archs $(2)                             \
		-pub                                    \
		-enable-upload-aws                      \
		-dist-dir $(DIST_DIR)                   \
		-name $(NAME)                           \
		-brand $(BRAND)                         \
		-helm-chart-dir $(HELM_CHART_DIR)       \
		-docker-image-repo $(DOCKER_IMAGE_REPO) \
		-download-ebpf $(DLEBPF)
endef

define pub_ebpf
	@echo "===== publishing $(1) $(NAME_EBPF) ====="
	@GO111MODULE=off CGO_CFLAGS=$(CGO_FLAGS) go run \
		-tags with_inputs cmd/make/make.go \
		-log-level $(LOG_LEVEL) \
		-release $(1)           \
		-archs $(2)             \
		-pub-ebpf               \
		-dist-dir $(DIST_DIR)   \
		-name $(NAME_EBPF)
endef

define build_docker_image
	echo 'publishing to $(2)...';
	@if [ $(2) = "registry.jiagouyun.com" ]; then \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/datakit:$(VERSION) \
			-f dockerfiles/Dockerfile.$(DOCKERFILE_SUFFIX) . --push; \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/datakit-elinker:$(VERSION) \
			-f dockerfiles/Dockerfile_elinker.$(DOCKERFILE_SUFFIX) . --push; \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/logfwd:$(VERSION) \
			-f dockerfiles/Dockerfile_logfwd.$(DOCKERFILE_SUFFIX) . --push; \
	else \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/datakit:$(VERSION) \
			-f dockerfiles/Dockerfile.$(DOCKERFILE_SUFFIX) . --push; \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/datakit-elinker:$(VERSION) \
			-f dockerfiles/Dockerfile_elinker.$(DOCKERFILE_SUFFIX) . --push; \
		sudo docker buildx build --platform $(1) \
			--build-arg DIST_DIR=$(DIST_DIR) \
			-t $(2)/logfwd:$(VERSION) \
			-f dockerfiles/Dockerfile_logfwd.$(DOCKERFILE_SUFFIX) . --push; \
	fi
endef

define build_uos_image
	echo 'publishing to $(2)...';
	sudo docker buildx build --platform $(1) \
	--build-arg DIST_DIR=$(DIST_DIR) \
	-t $(2)/datakit:$(VERSION) \
	-f dockerfiles/Dockerfile.uos . --push; \
	sudo docker buildx build --platform $(1) \
	--build-arg DIST_DIR=$(DIST_DIR) \
	-t $(2)/datakit-elinker:$(VERSION) \
	-f dockerfiles/Dockerfile_elinker.uos . --push; \
	sudo docker buildx build --platform $(1) \
	--build-arg DIST_DIR=$(DIST_DIR) \
	-t $(2)/logfwd:$(VERSION) \
	-f dockerfiles/Dockerfile_logfwd.uos . --push;
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
	@GO111MODULE=off CGO_ENABLED=0 go run -tags with_inputs cmd/make/make.go -build-isp -log-level $(LOG_LEVEL)
endef

##############################################################################
# Rules in the Makefile
##############################################################################

local: deps
	$(call build_bin,local,$(LOCAL_ARCHS))

pub_local: deps
	$(call publish,local,$(LOCAL_ARCHS))

pub_ebpf_local: deps
	$(call build_bin,local,$(LOCAL_ARCHS))
	$(call pub_ebpf,local,$(LOCAL_ARCHS))

pub_ebpf_local_nobuild: deps
	$(call pub_ebpf,local,$(LOCAL_ARCHS))

pub_ebpf_testing: deps
	$(call build_bin,testing,$(DATAKIT_EBPF_ARCHS))
	$(call pub_ebpf,testing,$(DATAKIT_EBPF_ARCHS))

pub_ebpf_production: deps
	$(call build_bin,production,$(DATAKIT_EBPF_ARCHS))
	$(call pub_ebpf,production,$(DATAKIT_EBPF_ARCHS))

testing_notify: deps
	$(call notify_build,testing,$(DEFAULT_ARCHS))

testing: deps
	$(call build_bin,testing,$(DEFAULT_ARCHS))
	$(call publish,testing,$(DEFAULT_ARCHS))

testing_image:
	$(call build_docker_image,$(DOCKER_IMAGE_ARCHS),'registry.jiagouyun.com/datakit')
	# we also publishing testing image to public image repo
	$(call build_docker_image,$(DOCKER_IMAGE_ARCHS),$(DOCKER_IMAGE_REPO))

production_notify: deps
	$(call notify_build,production,$(DEFAULT_ARCHS))

production: deps # stable release
	$(call build_bin,production,$(DEFAULT_ARCHS))
	$(call publish,production,$(DEFAULT_ARCHS))

production_image:
	$(call build_docker_image,$(DOCKER_IMAGE_ARCHS),$(DOCKER_IMAGE_REPO))

uos_image_testing: deps
	$(call build_bin,testing,$(DOCKER_IMAGE_ARCHS))
	$(call build_uos_image,$(UOS_DOCKER_IMAGE_ARCHS),'registry.jiagouyun.com/uos-dataflux') # testing image always push to registry.jiagouyun.com
	$(call build_uos_image,$(UOS_DOCKER_IMAGE_ARCHS),$(DOCKER_IMAGE_REPO)) # we also publishing testing image to public image repo

uos_image_production: deps
	$(call build_bin,production,$(DOCKER_IMAGE_ARCHS))
	$(call build_uos_image,$(UOS_DOCKER_IMAGE_ARCHS),$(DOCKER_IMAGE_REPO))

production_mac: deps
	$(call build_bin,production,$(MAC_ARCHS))
	$(call publish,production,$(MAC_ARCHS))

ip2isp:
	$(call build_ip2isp)

build_dca_web:
	mkdir -p $(DIST_DIR)
	cd dca && npm ci --registry=http://registry.npmmirror.com --disturl=http://npmmirror.com/dist --unsafe-perm && \
	cd web && npm ci --registry=http://registry.npmmirror.com --disturl=http://npmmirror.com/dist --unsafe-perm && \
	npm run build \
  cd ../..

build_dca: deps build_dca_web
	@echo "===== building $(BRAND).dca ====="
	@mv dca/web/build $(DIST_DIR)/dca-web # move DCA web(build during build_dca_web) to $(DIST_DIR)
	@CGO_CFLAGS=$(CGO_FLAGS) GO111MODULE=off CGO_ENABLED=0 \
		go run cmd/make/make.go -dca \
		-log-level $(LOG_LEVEL) \
		-archs $(DCA_BUILD_ARCH) \
		-dist-dir $(DIST_DIR) \
		-dca-version $(DCA_VERSION) \
		-brand $(BRAND)
	@echo "===== building $(BRAND).dca done ====="

build_dca_image:
	sudo docker buildx build \
		--platform $(DOCKER_IMAGE_ARCHS) \
		--build-arg DIST_DIR=$(DIST_DIR) \
		-t $(DOCKER_IMAGE_REPO):$(DCA_VERSION) \
		-f dca/Dockerfile.$(DOCKERFILE_SUFFIX) . --push;

deps: prepare gofmt 

# ignore files under vendor/.git/git
gofmt:
	@GO111MODULE=off gofmt -w -l $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/git/\|.*_y.go\|packed-packr.go")

vet:
	@go vet ./...

ut: deps
	CGO_CFLAGS=$(CGO_FLAGS) GO111MODULE=off CGO_ENABLED=1 \
	REMOTE_HOST=$(DOCKER_REMOTE_HOST) \
	go run cmd/make/make.go \
	--log-level $(LOG_LEVEL) \
	-ut -ut-exclude $(UT_EXCLUDE) -ut-only $(UT_ONLY) -ut-parallel $(UT_PARALLEL) \
	-log-level $(LOG_LEVEL) \
	-dataway-url $(DATAWAY_URL); \
		if [ $$? != 0 ]; then \
			exit 1; \
		else \
			echo "######################"; \
		fi

code_lint: deps copyright_check
	@$(GOLINT_BINARY) --version
ifeq ($(AUTO_FIX),true)
		@printf "$(HL)lint with auto fix...\n$(NC)"; \
			$(GOLINT_BINARY) run --fix --allow-parallel-runners;
else
		@printf "$(HL)lint without auto fix...\n$(NC)"; \
			$(GOLINT_BINARY) run --allow-parallel-runners;
endif

	@if [ $$? != 0 ]; then \
		printf "$(RED)[FAIL] lint failed\n$(NC)"; \
		exit -1; \
	fi

	# check print/printf
	go run scripts/disable-funcs/main.go -config scripts/disable-funcs/conf.toml;
	@if [ $$? != 0 ]; then \
		printf "$(RED)[FAIL] lint on print/printf failed\n$(NC)"; \
		exit -1; \
	fi

# lint code and document
lint: code_lint sample_conf_lint md_lint

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
	@echo 'check markdown files under $(1)...'

	# custom keys checking
	if ./scripts/check-custom-key.sh $(1); then \
		echo "custom key check passing..."; \
	else \
		exit 1; \
	fi

	# spell checking
	cspell lint --show-suggestions \
		-c scripts/cspell.json \
		--no-progress $(1)/**/*.md | tee $(DIST_DIR)/cspell.lint

	@if [ -s $(DIST_DIR)/cspell.lint ]; then \
		printf "$(RED) [FAIL] $(DIST_DIR)/cspell.lint not empty \n$(NC)"; \
		cat $(DIST_DIR)/cspell.lint; \
		exit -1; \
	fi

  # check markdown style
	# markdownlint install: https://github.com/igorshubovych/markdownlint-cli
	@echo 'version of markdownlint: $(shell markdownlint --version)'
	@truncate -s 0 $(DIST_DIR)/md-lint.json
	@if markdownlint -c scripts/markdownlint.yml -j -o $(DIST_DIR)/md-lint.json $(1); then \
		printf "markdownlint check ok\n"; \
	else \
		printf "$(RED) [FAIL] $(DIST_DIR)/md-lint.json not empty \n$(NC)"; \
		cat $(DIST_DIR)/md-lint.json; \
		exit -1; \
	fi
endef

exportdir=$(DIST_DIR)/export
# only check ZH docs, EN docs too many errors
# template generated real markdown files
docs_dir=$(exportdir)/guance-doc/docs
# all markdown template files
docs_template_dir=internal/export/doc

md_lint: md_export
	# Check on generated docs
	# Disable autofix on checking generated documents.
	# Also disable section check on generated docs(there are sections that rended in measurement name)
	@GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) \
		go run cmd/make/make.go \
		--log-level $(LOG_LEVEL) \
		--mdcheck $(docs_dir) \
		--mdcheck-no-section-check --mdcheck-no-autofix $(AUTO_FIX)
	$(call check_docs,$(docs_dir))

md_export:
	@GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) \
		go run cmd/make/make.go \
		--log-level $(LOG_LEVEL) \
		--mdcheck $(docs_template_dir) \
		--mdcheck-no-autofix $(AUTO_FIX);
	@rm -rf $(exportdir) && mkdir -p $(exportdir)
	@bash export.sh -D $(exportdir) -E -V 0.0.0

sample_conf_lint:
	@GO111MODULE=off CGO_ENABLED=0 CGO_CFLAGS=$(CGO_FLAGS) go run -tags with_inputs cmd/make/make.go \
		--sample-conf-check --log-level $(LOG_LEVEL)

project_words:
	cspell -c cspell/cspell.json --words-only --unique internal/export/doc/zh/** | sort --ignore-case >> project-words.txt

code_stat:
	cloc --exclude-dir=vendor,tests --exclude-lang=JSON,HTML .

# promlinter: show prometheuse metrics defined in Datakit.
# go install github.com/yeya24/promlinter/cmd/promlinter@latest
metrics:
	@promlinter list . --add-help -o md --with-vendor --add-position > dk.metrics

clean:
	@rm -rf build/*
	@rm -rf internal/io/parser/gram_y.go
	@rm -rf internal/io/parser/gram.y.go
	@rm -rf internal/pipeline/parser/parser.y.go
	@rm -rf internal/pipeline/parser/parser_y.go
	@rm -rf internal/pipeline/parser/gram.y.go
	@rm -rf internal/pipeline/parser/gram_y.go
	@rm -rf check.err
	@rm -rf $(DIST_DIR)/*

define check_mr_target_branch
	@if [ $1 = main -o $1 = master ]; then \
		printf "$(RED)[FAIL] merge request to branch '$1' disabled\n$(NC)"; \
		exit -1; \
	fi
endef

detect_mr_target_branch:
	$(call check_mr_target_branch,$(MERGE_REQUEST_TARGET_BRANCH))

push_ligai_version:
	@printf "$(HL)push new datakit version $(VERSION) to ligai...\n$(NC)";
	@curl -i -X POST \
		-H 'Content-Type: application/json' \
		-H "auto_devops_token: $(LIGAI_AUTO_DEVOPS_TOKEN)" \
		-d '{"version":"$(VERSION)","field_code":"$(LIGAI_CUSTOMFIELD)"}' \
		$(LIGAI_API)
	@if [ $$? != 0 ]; then \
		printf "$(RED) [WARN] push version to ligai failed"; \
	else \
		printf "[INFO] push version to ligai ok"; \
	fi
