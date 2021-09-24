uname = $(shell uname -r)

$(shell mkdir -p $(BASEPATH)/dist $(BASEPATH)/c/bin)

all: netflow_c offset_guess_c bindata buildexec

netflow_c: 
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASEPATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-DKBUILD_MODNAME=\"dknettracer\" \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASEPATH)/c/netflow/netflow.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASEPATH)/c/bin/netflow.o
	llvm-strip $(BASEPATH)/c/bin/netflow.o --no-strip-all  -R .BTF

offset_guess_c:
	clang \
		-I/usr/src/linux-headers-$(uname)/arch/x86/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated \
		-I/usr/src/linux-headers-$(uname)/include -I/usr/src/linux-headers-$(uname)/arch/x86/include/uapi -I/usr/src/linux-headers-$(uname)/arch/x86/include/generated/uapi \
		-I/usr/src/linux-headers-$(uname)/include/uapi -I/usr/src/linux-headers-$(uname)/include/generated/uapi \
		-fno-stack-protector -g \
		-include linux/kconfig.h \
		-I$(BASEPATH)/c/common \
		-include asm_goto_workaround.h \
		-D__KERNEL__ -D__BPF_TRACING__ \
		-DKBUILD_MODNAME=\"dknettracer\" \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wno-gnu-variable-sized-type-not-at-end \
		-Wno-address-of-packed-member \
		-Wno-tautological-compare\
		-Wno-unknown-warning-option \
		-O2 -emit-llvm \
		-c $(BASEPATH)/c/offset_guess/offset_guess.c \
		-o - | llc -march=bpf -filetype=obj -o $(BASEPATH)/c/bin/offset_guess.o
	llvm-strip $(BASEPATH)/c/bin/offset_guess.o --no-strip-all -R .BTF

bindata:
	$(shell mkdir -p $(BASEPATH)/c/bin)
	go-bindata -pkg ebpf -prefix "$(BASEPATH)/c/bin/" -o "$(BASEPATH)/c/ebpf_bindata.go" "$(BASEPATH)/c/bin/netflow.o" "$(BASEPATH)/c/bin/offset_guess.o"

debug: all
	sudo $(OUTPATH) --log-level debug --log /dev/stdout --interval 10s

debug_go: buildexec
	sudo $(OUTPATH) --log-level debug --log /dev/stdout --interval 10s

# /dist/net_ebpf
buildexec:
	$(shell mkdir -p $(BASEPATH)/dist )
	go build -o $(OUTPATH) $(BASEPATH)/net_ebpf.go

# /dist/offset_guess
build_guess:
	$(shell mkdir -p $(BASEPATH)/dist )
	go build -o $(OUTPATH) $(BASEPATH)/net_ebpf.go

clean:
	rm dist/*
	rm c/bin/*