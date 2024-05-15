#ifndef __HTTP_STATS_H
#define __HTTP_STATS_H

#include <linux/types.h>
#include "../netflow/conn_stats.h"

enum
{
#define L7_BUFFER_LEFT_SHIFT 12
    L7_BUFFER_SIZE = (1 << L7_BUFFER_LEFT_SHIFT), // 2^10
#define L7_BUFFER_SIZE L7_BUFFER_SIZE
#define IOVEC_LEFT_SHIFT 11

    BUF_IOVEC_LEN = (1 << IOVEC_LEFT_SHIFT),
#define BUF_IOVEC_LEN BUF_IOVEC_LEN
};

typedef enum
{
    HTTP_REQ_UNKNOWN = 0b00,
    HTTP_REQ_REQ = 0b01,
    HTTP_REQ_RESP = 0b10
} req_resp_t;

typedef enum
{
    P_UNKNOWN,

    P_SYSCALL_WRITE,
    P_SYSCALL_READ,
    P_SYSCALL_SENDTO,
    P_SYSCALL_RECVFROM,
    P_SYSCALL_WRITEV,
    P_SYSCALL_READV,
    P_SYSCALL_SENDFILE,

    P_SYSCALL_CLOSE,

    P_USR_SSL_WRITE,
    P_USR_SSL_READ

} tp_syscalls_fn_t;

#define P_GROUP_UNKNOWN 0
#define P_GROUP_READ 1
#define P_GROUP_WRITE 2

typedef struct pidfd
{
    __s32 fd;
    __u32 pid;
} pidfd_t;

typedef struct pidtid
{
    __u32 tid;
    __u32 pid;
} pidtid_t;

// 由于数据乱序上传，我们需要使用一个唯一值标示连接
typedef struct conn_uni_id
{
    __u64 sk;
    __u32 ktime;
    __u32 prandom;
} conn_uni_id_t;

typedef struct netdata_meta
{
    __u64 ts;
    __u64 ts_tail;
    __u64 tid_utid;
    __u8 comm[KERNEL_TASK_COMM_LEN];

    conn_uni_id_t uni_id;

    struct connection_info conn;
    __u32 tcp_seq;

    __u16 _pad0;
    __u16 func_id;

    __s32 fd;
    __s32 buf_len;
    __s32 act_size;
    __u32 index;
} netdata_meta_t;

// TODO: 考虑暂存此对象减少上报次数
typedef struct netwrk_data
{
    netdata_meta_t meta;
    __u8 payload[L7_BUFFER_SIZE];
} netwrk_data_t;

typedef struct ssl_read_args
{
    void *ctx;
    void *buf;
    __s32 num;
    __s32 _pad0;

    __u32 copied_seq;
    __u32 write_seq;

    void *skt;

    __u64 ts;
} ssl_read_args_t;

typedef struct syscall_rw_arg
{
    __u64 fd;
    void *buf;
    struct socket *skt;
    __u64 ts;
    __u32 _pad0;
    __u32 tcp_seq;
} syscall_rw_arg_t;

typedef struct syscall_rw_v_arg
{
    __u64 fd;
    struct iovec *vec;
    __u64 vlen;
    struct socket *skt;
    __u64 ts;
    __u32 _pad0;
    __u32 tcp_seq;
} syscall_rw_v_arg_t;

typedef struct syscall_sendfile_arg
{
    __u64 fd;
    struct socket *skt;
    __u64 ts;

    __u32 copied_seq;
    __u32 write_seq;
} syscall_sendfile_arg_t;

typedef struct pid_skptr
{
    __u32 pid;
    __u32 _pad0;
    __u64 sk_ptr;
} pid_skptr_t;

#endif // !__HTTP_STATS_H
