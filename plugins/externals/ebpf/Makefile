uname = $(shell uname -r)


BASE_PATH = $(BASEPATH)
OUT_PATH = $(OUTPATH)

$(shell mkdir -p $(BASE_PATH)/dist $(BASE_PATH)/c/bin)

all: netflow_c offset_guess_c bash_history_c bindata buildexec

netflow_c: 
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASE_PATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-DKBUILD_MODNAME=\"dktracer\" \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASE_PATH)/c/netflow/netflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASE_PATH)/c/bin/netflow.o
	llvm-strip $(BASE_PATH)/c/bin/netflow.o --no-strip-all  -R .BTF

offset_guess_c:
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASE_PATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-DKBUILD_MODNAME=\"dktracer\" \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASE_PATH)/c/offset_guess/offset_guess.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASE_PATH)/c/bin/offset_guess.o
	llvm-strip $(BASE_PATH)/c/bin/offset_guess.o --no-strip-all -R .BTF

dns_filter_c:
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASE_PATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-DKBUILD_MODNAME=\"dktracer\" \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASE_PATH)/c/socket_filter/dns_filter.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASE_PATH)/c/bin/dns_filter.o
	llvm-strip $(BASE_PATH)/c/bin/dns_filter.o --no-strip-all -R .BTF

bash_history_c : 
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASE_PATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASE_PATH)/c/bash_history/bash_history.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASE_PATH)/c/bin/bash_history.o
	llvm-strip $(BASE_PATH)/c/bin/bash_history.o --no-strip-all  -R .BTF

bindata:
	go-bindata -pkg ebpf -prefix "$(BASE_PATH)/c/bin/" -o "$(BASE_PATH)/c/ebpf_bindata.go" \
		 "$(BASE_PATH)/c/bin/offset_guess.o" "$(BASE_PATH)/c/bin/netflow.o" "$(BASE_PATH)/c/bin/bash_history.o"

# /dist/ebpf
buildexec:
	go build -tags="ebpf" -o $(OUT_PATH) $(BASE_PATH)/datakit-ebpf.go

clean:
	rm dist/*
	rm c/bin/*