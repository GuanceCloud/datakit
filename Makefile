.PHONY: default test

default: local

# 正式环境
RELEASE_DOWNLOAD_ADDR = zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit

# 测试环境
TEST_DOWNLOAD_ADDR = zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit

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
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build  \
		 -release $(1) -pub-dir $(PUB_DIR) -archs $(2) -cgo $(3)
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	#@tar czf $(PUB_DIR)/release/$(NAME)-$(VERSION).tar.gz autostart agent -C build .
	tree -Csh $(PUB_DIR)
endef

local:
	$(call build,test, "linux/amd64", 1)

test:
	$(call build,test, "all", 1)

release:
	$(call build,release,"all", 0)

define pub
	echo "publish $(1) $(NAME) ..."
	go run make.go -pub -release $(1) -pub-dir $(PUB_DIR) -name $(NAME) -download-addr $(2) -archs $(3) -os $(4)
endef

pub_test:
	$(call pub,test,$(TEST_DOWNLOAD_ADDR),"linux/amd64","linux")

pub_release:
	$(call pub,release,$(RELEASE_DOWNLOAD_ADDR),"linux/amd64","linux")

clean:
	rm -rf build/*
	rm -rf $(PUB_DIR)/*
