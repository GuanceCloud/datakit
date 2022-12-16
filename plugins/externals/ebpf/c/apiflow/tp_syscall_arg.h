#ifndef __TP_SYSCALL_ARG_H_
#define __TP_SYSCALL_ARG_H_

#include <linux/types.h>
#include <linux/socket.h>

struct tp_syscall_exit_args
{
    __u64 _common;

    __s32 _syscall_nr;
    __u32 _pad0;

    __s64 ret; // offset:16, size:8(signed)
};

struct tp_syscall_send_recv_msg_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;                        // offset:16, size:8
    struct user_msghdr *user_msghdr; // offset:24, size:8
    __u64 flags;                     // offset:32, size:8
};

// for syscall read, write, sendto, recvfrom
// ！ sendto 和 recvfrom 后几个参数的 addr 通常与 tcp_sock 中的一致
struct tp_syscall_read_write_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;              // offset:16, size:8
    void *buf;             // offset:24, size:8
    __kernel_size_t count; // offset:32, size:8 (64bit arch)
};

struct tp_syscall_writev_readv_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;          // offset:16, size:8
    struct iovec *vec; // offset:24, size:8
    __u64 vlen;        // offset:32, size:8
};

#endif