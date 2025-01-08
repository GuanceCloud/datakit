.PHONY: all debug rewriter dkrunc test

ARCH ?= $(shell uname -m | sed -e s/x86_64/x86_64/ \
				  -e s/aarch64.\*/aarch64/)

ifeq ($(ARCH), x86_64)
    ARCH = amd64
endif
ifeq ($(ARCH), aarch64)
	ARCH = arm64
endif

REPO_PATH   ?= ${HOME}/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit
SOURCE_PATH := $(REPO_PATH)/internal/apminject
DIST_DIR    ?= $(REPO_PATH)/internal/apminject/dist/datakit-apm-inject-linux-$(ARCH)

DATAKIT_INJ_REWRITE_PROC ?= \"/usr/local/datakit/apm_inject/inject/rewriter\"

all: dkrunc rewriter launcher

rewriter:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o $(DIST_DIR)/rewriter \
		$(SOURCE_PATH)/rewriter/rewriter.go

dkrunc:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o $(DIST_DIR)/dkrunc \
		$(SOURCE_PATH)/dkrunc/dkrunc.go

launcher:
	mkdir -p $(DIST_DIR)
	gcc -DDATAKIT_INJ_REWRITE_PROC=${DATAKIT_INJ_REWRITE_PROC} \
		$(SOURCE_PATH)/apm_launcher.c -fPIC -shared \
		-o $(DIST_DIR)/apm_launcher.so

launcher_musl:
	mkdir -p $(DIST_DIR)
	gcc -DDATAKIT_INJ_REWRITE_PROC=${DATAKIT_INJ_REWRITE_PROC} \
		$(SOURCE_PATH)/apm_launcher.c -fPIC -shared \
		-o $(DIST_DIR)/apm_launcher_musl.so

launcher_debug:
	mkdir -p ${DIST_DIR}
	gcc -DDATAKIT_DEBUG -DDATAKIT_INJ_REWRITE_PROC=${DATAKIT_INJ_REWRITE_PROC} \
		$(SOURCE_PATH)/apm_launcher.c -fPIC -shared \
		-o $(DIST_DIR)/apm_launcher_debug.so

launcher_musl_debug:
	mkdir -p ${DIST_DIR}
	gcc -DDATAKIT_DEBUG -DDATAKIT_INJ_REWRITE_PROC=${DATAKIT_INJ_REWRITE_PROC} \
		$(SOURCE_PATH)/apm_launcher.c -fPIC -shared \
		-o $(DIST_DIR)/apm_launcher_musl_debug.so

test:
	mkdir -p ${DIST_DIR}
	gcc $(SOURCE_PATH)/launcher_test.c -o $(DIST_DIR)/launcher_test \
		&& $(DIST_DIR)/launcher_test && rm $(DIST_DIR)/launcher_test
