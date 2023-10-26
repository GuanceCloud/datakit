ARCH ?= $(shell uname -m | sed -e s/x86_64/x86_64/ \
				  -e s/aarch64.\*/arm64/)

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

ARGS ?= ""

SRCPATH ?= .
SRC_PATH := $(SRCPATH)

INTERNAL_PATH := $(SRC_PATH)/internal

EBPF_BIN_PATH := $(INTERNAL_PATH)/c/elf/linux_$(GO_ARCH)

OUTPATH ?= $(SRC_PATH)/dist/$(GO_ARCH)/
OUT_PATH := $(OUTPATH)

$(shell mkdir -p $(EBPF_BIN_PATH))

DK_BPF_KERNEL_SRC_PATH ?= /usr/src/linux-headers-$(shell uname -r)

KERNEL_UAPI_INCLUDE := -isystem$(DK_BPF_KERNEL_SRC_PATH)/arch/$(MACHINE_ARCH)/include/uapi \
		-isystem$(DK_BPF_KERNEL_SRC_PATH)/arch/$(MACHINE_ARCH)/include/generated/uapi \
		-isystem$(DK_BPF_KERNEL_SRC_PATH)/include/uapi \
		-isystem$(DK_BPF_KERNEL_SRC_PATH)/include/generated/uapi

KERNEL_INCLUDE := -isystem$(DK_BPF_KERNEL_SRC_PATH)/arch/$(MACHINE_ARCH)/include \
		-isystem$(DK_BPF_KERNEL_SRC_PATH)/arch/$(MACHINE_ARCH)/include/generated \
		-isystem$(DK_BPF_KERNEL_SRC_PATH)/include \
		$(KERNEL_UAPI_INCLUDE)

BPF_INCLUDE := $(KERNEL_INCLUDE) \
		-include linux/kconfig.h \
		-I$(INTERNAL_PATH)/c/common \
		-include asm_goto_workaround.h

BUILD_TAGS := -D__KERNEL__ -D__BPF_TRACING__ $(ARGS) \
		-fno-stack-protector -g \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm

all: build

httpflow.o: 
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/apiflow/httpflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/httpflow.o

netflow.o: httpflow.o
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/netflow/netflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/netflow.o

process_sched.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/process_sched/process_sched.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/process_sched.o

conntrack.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/conntrack/conntrack.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/conntrack.o

offset_guess.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/offset_guess/offset_guess.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/offset_guess.o

offset_httpflow.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/offset_guess/offset_httpflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/offset_httpflow.o

offset_conntrack.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/offset_guess/offset_conntrack.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/offset_conntrack.o

offset_tcp_seq.o:
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-DKBUILD_MODNAME=\"datatkit-ebpf\" \
		-c $(INTERNAL_PATH)/c/offset_guess/offset_tcp_seq.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/offset_tcp_seq.o

bash_history.o: 
	clang $(BPF_INCLUDE) $(BUILD_TAGS) \
		-c $(INTERNAL_PATH)/c/bash_history/bash_history.c \
		-o - | llc -march=bpf -filetype=obj -o $(EBPF_BIN_PATH)/bash_history.o

bindata: offset_guess.o offset_httpflow.o offset_conntrack.o offset_tcp_seq.o \
			netflow.o bash_history.o conntrack.o process_sched.o
	llvm-strip $(EBPF_BIN_PATH)/*.o --no-strip-all -R .BTF  
	llvm-strip $(EBPF_BIN_PATH)/*.o --no-strip-all -R .BTF.ext 
	llvm-strip $(EBPF_BIN_PATH)/*.o --no-strip-all -R .rel.BTF
	llvm-strip $(EBPF_BIN_PATH)/*.o --no-strip-all -R .rel.BTF.ext

build: bindata
	go build -tags="ebpf" -o $(OUT_PATH) $(SRC_PATH)/datakit-ebpf.go 
	# go build -tags="ebpf" -ldflags "-w -s" -o $(OUT_PATH) $(SRC_PATH)/datakit-ebpf.go 



clean:
	rm -r $(OUT_PATH)
	rm -r $(EBPF_BIN_PATH)
