.PHONY: default test

default: local

# devops 测试环境
TEST_DOWNLOAD_ADDR = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/test
TEST_DOWNLOAD_ADDR_WIN = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/windows/test
TEST_SSL = 0
TEST_PORT = 10401

# 本地搭建的 kodo 测试
LOCAL_KODO_HOST = http://kodo-local.cloudcare.cn
LOCAL_DOWNLOAD_ADDR = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/ftagent/local
LOCAL_SSL = 0
LOCAL_PORT = 9527

# 预发环境
PREPROD_DOWNLOAD_ADDR = cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/preprod
PREPROD_SSL = 1
PREPROD_PORT = 443

# 正式环境
RELEASE_KODO_HOST = https://kodo.cloudcare.cn
RELEASE_CS_HOST = https://kodo-via-core-stone.cloudcare.cn
RELEASE_DOWNLOAD_ADDR = cloudcare-files.oss-cn-hangzhou.aliyuncs.com/ftagent/release
RELEASE_SSL = 1
RELEASE_PORT = 443

PUB_DIR = pub
BIN = datakit
NAME = datakit
ENTRY = main.go

VERSION := $(shell git describe --always --tags)

all: test release preprod local


local:
	$(call build,local,$(LOCAL_KODO_HOST),$(LOCAL_DOWNLOAD_ADDR),$(LOCAL_SSL),$(LOCAL_PORT))

preprod:
	@echo "===== $(BIN) preprod ===="
	@rm -rf $(PUB_DIR)/preprod
	@mkdir -p build $(PUB_DIR)/preprod
	@mkdir -p git
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build -archs "linux/amd64" \
		-kodo-host $(PREPROD_KODO_HOST) -download-addr $(PREPROD_DOWNLOAD_ADDR) -ssl $(PREPROD_SSL) -port $(PREPROD_PORT) \
		-release preprod -pub-dir $(PUB_DIR) -cs-host $(PREPROD_CS_HOST)  -cgo
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	@tar czf $(PUB_DIR)/preprod/$(NAME)-$(VERSION).tar.gz autostart -C build .
	tree -Csh $(PUB_DIR)

release:
	@echo "===== $(BIN) release ===="
	@rm -rf $(PUB_DIR)/release
	@mkdir -p build $(PUB_DIR)/release
	@mkdir -p git
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build -archs "linux/amd64" \
		-kodo-host $(RELEASE_KODO_HOST) -download-addr $(RELEASE_DOWNLOAD_ADDR) -ssl $(RELEASE_SSL) -port $(RELEASE_PORT) \
		-release release -pub-dir $(PUB_DIR) -cs-host $(RELEASE_CS_HOST) -cgo
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	@tar czf $(PUB_DIR)/release/$(NAME)-$(VERSION).tar.gz autostart agent -C build .
	tree -Csh $(PUB_DIR)

test:
	@echo "===== $(BIN) test ===="
	@rm -rf $(PUB_DIR)/test
	@mkdir -p build $(PUB_DIR)/test
	@mkdir -p git
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build -archs "linux/amd64" \
		 -download-addr $(TEST_DOWNLOAD_ADDR) -release test -pub-dir $(PUB_DIR)
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	#@tar czf $(PUB_DIR)/test/$(NAME)-$(VERSION).tar.gz autostart agent -C build .
	tree -Csh $(PUB_DIR)

test_win:
	@echo "===== $(BIN) test_win ===="
	@rm -rf $(PUB_DIR)/test_win
	@mkdir -p build $(PUB_DIR)/test_win
	@mkdir -p git
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build -archs "windows/amd64" \
		 -download-addr $(TEST_DOWNLOAD_ADDR_WIN) -release test -pub-dir $(PUB_DIR) -windows
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	#@tar czf $(PUB_DIR)/test_win/$(NAME)-$(VERSION).tar.gz -C windows agent.exe -C ../build .
	tree -Csh $(PUB_DIR)


test_mac:
	@echo "===== $(BIN) test_mac ===="
	@rm -rf $(PUB_DIR)/test_mac
	@mkdir -p build $(PUB_DIR)/test_mac
	@mkdir -p git
	@echo 'package git; const (Sha1 string=""; BuildAt string=""; Version string=""; Golang string="")' > git/git.go
	@go run make.go -main $(ENTRY) -binary $(BIN) -name $(NAME) -build-dir build -archs "darwin/amd64" \
		 -download-addr $(TEST_DOWNLOAD_ADDR_WIN) -release test -pub-dir $(PUB_DIR) -mac
	#@strip build/$(NAME)-linux-amd64/$(BIN)
	@tar czf $(PUB_DIR)/test_mac/$(NAME)-$(VERSION).tar.gz -C mac agent -C ../build .
	tree -Csh $(PUB_DIR)


pub_local:
	$(call pub,local)

pub_test:
	@echo "publish test ${BIN} ..."
	@go run make.go -pub -release test -pub-dir $(PUB_DIR) -name $(NAME)

pub_test_win:
	@echo "publish test windows ${BIN} ..."
	@go run make.go -pub -release test -pub-dir $(PUB_DIR) -archs "windows/amd64" -name $(NAME) -windows

pub_preprod:
	@echo "publish preprod ${BIN} ..."
	@go run make.go -pub -release preprod -pub-dir $(PUB_DIR) -name $(NAME)

pub_release:
	@echo "publish release ${BIN} ..."
	@go run make.go -pub -release release -pub-dir $(PUB_DIR) -name $(NAME)

clean:
	rm -rf build/*
	rm -rf $(PUB_DIR)/*
