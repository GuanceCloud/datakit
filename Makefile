.PHONY: default test local man

default: local

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit

# 本地环境: 需配置环境变量，便于完整测试采集器的发布、更新等流程
# export LOCAL_OSS_ACCESS_KEY='<your-oss-AK>'
# export LOCAL_OSS_SECRET_KEY='<your-oss-SK>'
# export LOCAL_OSS_BUCKET='<your-oss-bucket>'
# export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com' # 一般都是这个地址
# export LOCAL_OSS_ADDR='<your-oss-bucket>.oss-cn-hangzhou.aliyuncs.com/datakit'
# 如果只是编译，LOCAL_OSS_ADDR 这个环境变量可以随便给个值
LOCAL_DOWNLOAD_ADDR = "${LOCAL_OSS_ADDR}"

PUB_DIR = dist
BUILD_DIR = dist

BIN = datakit
NAME = datakit
ENTRY = cmd/datakit/main.go

LOCAL_ARCHS = "local"
DEFAULT_ARCHS = "all"
MAC_ARCHS = "darwin/amd64"
GIT_VERSION := $(shell git describe --always --tags | sed 's/-/./g')
DATE := $(shell date -u +'%Y-%m-%d %H:%M:%S')
GOVERSION := $(shell go version)
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMITER := $(shell git log -1 --pretty=format:'%an')
UPLOADER:= $(shell hostname)/${USER}/${COMMITER}

NOTIFY_MSG_RELEASE:=$(shell echo '{"msgtype": "text","text": {"content": "$(UPLOADER) 发布了 DataKit 新版本($(GIT_VERSION))"}}')
NOTIFY_MSG_TEST:=$(shell echo '{"msgtype": "text","text": {"content": "$(UPLOADER) 发布了 DataKit 测试版($(GIT_VERSION))"}}')
NOTIFY_CI:=$(shell echo '{"msgtype": "text","text": {"content": "$(COMMITER)正在执行 DataKit CI，此刻请勿在CI分支($(BRANCH))提交代码，以免 CI 任务失败"}}')

define GIT_INFO
//nolint
package git

const (
	BuildAt  string = "$(DATE)"
	Version  string = "$(GIT_VERSION)"
	Golang   string = "$(GOVERSION)"
	Commit   string = "$(COMMIT)"
	Branch   string = "$(BRANCH)"
	Uploader string = "$(UPLOADER)"
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
endef

lint:
	@golangci-lint run --timeout 1h | tee check.err # https://golangci-lint.run/usage/install/#local-installation

vet:
	@go vet ./...

test:
	@GO111MODULE=off go test ./...

local: man gofmt
	$(call build,local, $(LOCAL_ARCHS), $(LOCAL_DOWNLOAD_ADDR))

testing: man
	$(call build,test, $(DEFAULT_ARCHS), $(TEST_DOWNLOAD_ADDR))

release: man
	$(call build,release, $(DEFAULT_ARCHS), $(RELEASE_DOWNLOAD_ADDR))

release_mac: man
	$(call build,release, $(MAC_ARCHS), $(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_ARCHS))

pub_local_mac:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),$(MAC_ARCHS))

pub_testing:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_testing_mac:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),$(MAC_ARCHS))

pub_testing_img:
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(TEST_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t registry.jiagouyun.com/datakit/datakit:$(GIT_VERSION) .
	@sudo docker push registry.jiagouyun.com/datakit/datakit:$(GIT_VERSION)

pub_release_img:
	# release to pub hub
	@mkdir -p embed/linux-amd64
	@wget --quiet -O - "https://$(RELEASE_DOWNLOAD_ADDR)/iploc/iploc.tar.gz" | tar -xz -C .
	@sudo docker build -t pubrepo.jiagouyun.com/datakit/datakit:$(GIT_VERSION) .
	@sudo docker push pubrepo.jiagouyun.com/datakit/datakit:$(GIT_VERSION)

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(DEFAULT_ARCHS))

pub_release_mac:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),$(MAC_ARCHS))

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

define build_ip2isp
	rm -rf china-operator-ip
	git clone -b ip-lists https://github.com/gaoyifan/china-operator-ip.git
	@GO111MODULE=off CGO_ENABLED=0 go run cmd/make/make.go -build-isp
endef

.PHONY: agent
agent:
	$(call build_agent)

ip2isp:
	$(call build_ip2isp)

man:
	@packr2 clean
	@packr2

gofmt:
	@GO111MODULE=off go fmt ./...

clean:
	rm -rf build/*
	rm -rf $(PUB_DIR)/*
