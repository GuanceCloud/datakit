#ifndef __TP_ARG_H_
#define __TP_ARG_H_

#include <linux/types.h>
#include <linux/socket.h>

typedef struct tp_syscall_exit_args
{
    __u64 _common;

    __s32 _syscall_nr;
    __u32 _pad0;

    __s64 ret; // offset:16, size:8(signed)
} tp_syscall_exit_args_t;

typedef struct tp_syscall_close_args
{
    __u64 _common;

    __s32 _syscall_nr;
    __u32 _pad0;

    __u64 fd; // offset:16, size:8(unsigned)
} tp_syscall_close_args_t;

typedef struct tp_syscall_send_recv_msg_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;                        // offset:16, size:8
    struct user_msghdr *user_msghdr; // offset:24, size:8
    __u64 flags;                     // offset:32, size:8
} tp_syscall_send_recv_msg_args_t;

// For syscall read, write, sendto, recvfrom,
// the addr of the last few parameters of sendto and recvfrom is usually consistent with that in tcp_sock.
typedef struct tp_syscall_rw_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;              // offset:16, size:8
    void *buf;             // offset:24, size:8
    __kernel_size_t count; // offset:32, size:8 (64bit arch)
} tp_syscall_rw_args_t;

typedef struct tp_syscall_rw_v_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __u64 fd;          // offset:16, size:8
    struct iovec *vec; // offset:24, size:8
    __u64 vlen;        // offset:32, size:8
} tp_syscall_rw_v_args_t;

typedef struct tp_syscall_sendfile_args
{
    __u64 _common; // offset:8, size:2 + 1 + 1 + 4(signed)

    __s32 _sycall_nr; // offset:8, size:4(signed)
    __u32 _pad0;      // offset:12, size:4

    __s64 out_fd; // signed
    __s64 in_fd;

    __s64 offset;
    __s64 count;
} tp_syscall_sendfile_args_t;

#endif
