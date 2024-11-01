#ifndef __HTTP_STATS_H
#define __HTTP_STATS_H

#include <linux/types.h>
#include "../netflow/conn_stats.h"

enum
{
#define L7_BUFFER_LEFT_SHIFT 11

    L7_BUFFER_SIZE = (1 << (L7_BUFFER_LEFT_SHIFT)), // 2^10
#define L7_BUFFER_SIZE L7_BUFFER_SIZE

#define IOVEC_LEFT_SHIFT 10

    BUF_IOVEC_LEN = (1 << (IOVEC_LEFT_SHIFT)),
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
// cpu id | ktime | id(auto increment)
typedef struct id_generator
{
    __u8 init;
    __u8 _pad;
    __u16 cpu_id;
    __u32 id;
    __u64 ktime;
} id_generator_t;

typedef struct sk_inf
{
    id_generator_t uni_id;
    __u64 index;
    __u64 skptr;
    conn_inf_t conn;
} sk_inf_t;

typedef struct netdata_meta
{
    __u64 ts;
    __u64 ts_tail;
    __u64 tid_utid;
    __u8 comm[KERNEL_TASK_COMM_LEN];

    sk_inf_t sk_inf;

    __u32 tcp_seq;

    __u16 _pad0;
    __u16 func_id;

    __s32 original_size;
    __s32 capture_size;
} netdata_meta_t;

// TODO: 考虑暂存此对象减少上报次数
typedef struct netwrk_data
{
    netdata_meta_t meta;
    __u8 payload[L7_BUFFER_SIZE];
} net_data_t;

typedef struct event_rec
{
    __u32 num;
    __u32 bytes;
} event_rec_t;

enum
{
    L7_EVENT_SIZE = (L7_BUFFER_SIZE * 4 - sizeof(event_rec_t)),
#define L7_EVENT_SIZE L7_EVENT_SIZE
};

typedef struct network_events
{
    event_rec_t rec;
    __u8 payload[L7_EVENT_SIZE];
} network_events_t;

typedef struct net_event_comm
{
    event_rec_t rec;
    netdata_meta_t meta;
} net_event_comm_t;

typedef struct
{
    net_event_comm_t event_comm;
    __u8 payload[L7_BUFFER_SIZE];
} net_event_t;

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
