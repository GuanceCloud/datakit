#ifndef __DKE_COMMON_H
#define __DKE_COMMON_H

#include "bpf_helpers.h"

#define FN_NAME(a, b, arg...) a##b(arg)

#define FN_TP_SYSCALL(fn, arg...)   \
    SEC("tracepoint/syscalls/" #fn) \
    int FN_NAME(tracepoint__, fn, arg)

#define FN_KPROBE(fn)  \
    SEC("kprobe/" #fn) \
    int FN_NAME(kprobe__, fn, struct pt_regs *ctx)

#define FN_KRETPROBE(fn)  \
    SEC("kretprobe/" #fn) \
    int FN_NAME(kretprobe__, fn, struct pt_regs *ctx)

#define FN_UPROBE(fn)  \
    SEC("uprobe/" #fn) \
    int FN_NAME(uprobe__, fn, struct pt_regs *ctx)

#define FN_URETPROBE(fn)  \
    SEC("uretprobe/" #fn) \
    int FN_NAME(uretprobe__, fn, struct pt_regs *ctx)

#endif