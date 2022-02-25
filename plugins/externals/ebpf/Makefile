KERNEL_RELEASE := $(shell uname -r)

ARCH ?= $(shell uname -m | sed -e s/x86_64/x86_64/ \
				  -e s/aarch64.*/arm64/)

MACHINE_ARCH := $(ARCH)
GO_ARCH := $(MACHINE_ARCH)

# not support 32bit arch

ifeq ($(MACHINE_ARCH),x86_64)
        MACHINE_ARCH := x86
		GO_ARCH := amd64
endif
ifeq ($(MACHINE_ARCH),amd64)
        MACHINE_ARCH := x86
		GO_ARCH := amd64
endif

SRCPATH ?= .
SRC_PATH := $(SRCPATH)

BIN_PATH := $(SRC_PATH)/c/bin/$(GO_ARCH)

OUTPATH ?= $(SRC_PATH)/dist/$(GO_ARCH)/
OUT_PATH := $(OUTPATH)

$(shell mkdir -p $(BIN_PATH))

KERNELSRC ?= /usr/src/linux-headers-$(KERNEL_RELEASE)
KERNEL_SRC := $(KERNELSRC)

KERNEL_UAPI_INCLUDE := -isystem$(KERNEL_SRC)/arch/$(MACHINE_ARCH)/include/uapi \
		-isystem$(KERNEL_SRC)/arch/$(MACHINE_ARCH)/include/generated/uapi \
		-isystem$(KERNEL_SRC)/include/uapi \
		-isystem$(KERNEL_SRC)/include/generated/uapi

KERNEL_INCLUDE := -isystem$(KERNEL_SRC)/arch/$(MACHINE_ARCH)/include \
		-isystem$(KERNEL_SRC)/arch/$(MACHINE_ARCH)/include/generated \
		-isystem$(KERNEL_SRC)/include \
		$(KERNEL_UAPI_INCLUDE)

BPF_INCLUDE := $(KERNEL_INCLUDE) \
		-include linux/kconfig.h \
		-I$(SRC_PATH)/c/common \
		-include asm_goto_workaround.h

BUILD_TAGS := -D__KERNEL__ -D__BPF_TRACING__ \
		-fno-stack-protector -g \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm

all: go_bindata buildexec

httpflow.o: 
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(SRC_PATH)/c/netflow/httpflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(BIN_PATH)/httpflow.o

netflow.o: httpflow.o
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(SRC_PATH)/c/netflow/netflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(BIN_PATH)/netflow.o

offset_guess.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(SRC_PATH)/c/offset_guess/offset_guess.c \
		-o - | llc -march=bpf -filetype=obj -o $(BIN_PATH)/offset_guess.o

bash_history.o : 
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-c $(SRC_PATH)/c/bash_history/bash_history.c \
		-o - | llc -march=bpf -filetype=obj -o $(BIN_PATH)/bash_history.o

bindata: offset_guess.o netflow.o bash_history.o
	llvm-strip $(BIN_PATH)/*.o --no-strip-all -R .BTF

go_bindata: bindata
	go-bindata -pkg ebpf -prefix "$(BIN_PATH)/" -o "$(SRC_PATH)/c/bindata_linux_$(GO_ARCH).go" \
		 "$(BIN_PATH)/offset_guess.o" "$(BIN_PATH)/netflow.o" "$(BIN_PATH)/httpflow.o" "$(BIN_PATH)/bash_history.o"

buildexec:
	go build -tags="ebpf" -o $(OUT_PATH) $(SRC_PATH)/datakit-ebpf.go

clean:
	rm -r $(OUT_PATH)
	rm -r $(BIN_PATH)