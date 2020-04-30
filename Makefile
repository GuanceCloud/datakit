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

all: test release preprod local

define build
	@echo "===== $(BIN) $(1) ===="
	@rm -rf $(PUB_DIR)/$(1)/*
	@export GO111MODULE=off
	@mkdir -p build $(PUB_DIR)/$(1)
	@mkdir -p git
	@echo 'package git; const (BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build  \
		 -release $(1) -pub-dir $(PUB_DIR) -archs $(2) -cgo $(3) -download-addr $(4)
	tree -Csh build pub
endef

define pub
	echo "publish $(1) $(NAME) ..."
	go run make.go -pub -release $(1) -pub-dir $(PUB_DIR) -name $(NAME) -download-addr $(2) -archs $(3)
endef

local:
	$(call build,local, "windows/amd64|linux/amd64", 1, $(LOCAL_DOWNLOAD_ADDR))

test:
	$(call build,test, "all", 1, $(TEST_DOWNLOAD_ADDR))

release:
	$(call build,release,"all", 1, $(RELEASE_DOWNLOAD_ADDR))

pub_local:
	$(call pub,local,$(LOCAL_DOWNLOAD_ADDR),"linux/amd64|windows/amd64")

pub_test:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),"all")

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),"all")

clean:
	rm -rf build/*
	rm -rf $(PUB_DIR)/*
