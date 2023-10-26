#ifndef __TP_ARG_H_
#define __TP_ARG_H_

#include <linux/types.h>
#include <linux/socket.h>

struct tp_syscall_exit_args
{
    __u64 _common;

    __s32 _syscall_nr;
    __u32 _pad0;

    __s64 ret; // offset:16, size:8(signed)
};

struct tp_syscall_close_args
{
    __u64 _common;

    __s32 _syscall_nr;
    __u32 _pad0;

    __u64 fd; // offset:16, size:8(unsigned)
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

// For syscall read, write, sendto, recvfrom,
// the addr of the last few parameters of sendto and recvfrom is usually consistent with that in tcp_sock.
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

struct tp_syscall_sendfile_arg
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __s64 out_fd;   // signed
    __s64 in_fd;

    __s64 offset;
    __s64 count;
};

#endif