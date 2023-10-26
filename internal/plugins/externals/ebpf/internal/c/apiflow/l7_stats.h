#ifndef __HTTP_STATS_H
#define __HTTP_STATS_H

#include <linux/types.h>
#include "../netflow/conn_stats.h"

enum
{
#define L7_BUFFER_LEFT_SHIFT 11
    L7_BUFFER_SIZE = (1 << L7_BUFFER_LEFT_SHIFT), // 2^10
#define L7_BUFFER_SIZE L7_BUFFER_SIZE
    L7_BUFFER_SIZE_MASK = (L7_BUFFER_SIZE - 1), // need to be larger than buffer size
#define L7_BUFFER_SIZE_MASK L7_BUFFER_SIZE_MASK

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

    P_USR_SSL_READ,
    P_USR_SSL_WRITE

} k_u_func_t;

#define P_GROUP_UNKNOWN 0
#define P_GROUP_READ 1
#define P_GROUP_WRITE 2

// Need to associate payload and conn info.
struct payload_id
{
    __u64 ktime;
    __u64 pid_tid;
    __u16 cpuid;
    __u16 prandomhalf;
    __u32 prandom;
};

#define PROTO_HTTP 1

// typedef struct proto_data
// {
//     __u8 proto;
//     __u8 req_resp;

//     __u16 _pad0;
//     __u32 fd;

//     proto_buffer_t *buf;

// } proto_data_t;

typedef struct l7_buffer
{
    struct payload_id thr_trace_id;
    struct payload_id id;

    __s64 isentry;

    __s32 len;
    __u8 payload[L7_BUFFER_SIZE];
    __u64 req_ts;
    __u8 cmd[KERNEL_TASK_COMM_LEN];
} proto_buffer_t;

// struct http_stats
// {
//     // struct connection_info conn_info;

//     __u8 payload[HTTP_PAYLOAD_SIZE];
//     __u8 method;
//     __u16 status_code;
//     __u32 http_version;
//     __u64 req_ts;
//     __u64 resp_ts;
// };

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

struct layer7_http
{
    struct payload_id req_payload_id;
    __u32 direction;
    __u32 method;

    __u64 req_ts;

    __u32 http_version;
    __u32 status_code;

    __u64 resp_ts;

    __be32 nat_daddr[4]; // dst ip address
    __u16 nat_dport;     // dst port
    __u16 _pad0;

    // __u64 pid_tid;

    __u32 req_seq;
    __u32 resp_seq;

    __u32 req_func;
    __u32 resp_func;

    __s32 sent_bytes;
    __s32 recv_bytes;

    // __u8 cmd[64];
};

struct span_info
{
    struct payload_id req_payload_id;

    __u32 pid;
    __u32 tid;

    __u32 k_u_fn;
    __s32 bytes;

    __u64 start;
    __u64 end;

    __u64 parent_span;
    __u64 span;
};

struct http_req_finished
{
    struct connection_info conn_info;
    struct layer7_http http;
};

struct ssl_read_args
{
    void *ctx;
    void *buf;
    __s32 num;
    __s32 _pad0;

    __u32 copied_seq;
    __u32 write_seq;

    void *skt;

    __u64 ts;
};

struct syscall_read_write_arg
{
    __u64 fd;
    void *buf;
    struct socket *skt;
    __u32 w_size;
    __u32 _pad;
    __u64 ts;

    __u32 copied_seq;
    __u32 write_seq;
};

struct syscall_readv_writev_arg
{
    __u64 fd;
    struct iovec *vec;
    __u64 vlen;

    struct socket *skt;
    __u64 ts;

    __u32 copied_seq;
    __u32 write_seq;
};

struct syscall_sendfile_arg
{
    __u64 fd;
    struct socket *skt;
    __u64 ts;

    __u32 copied_seq;
    __u32 write_seq;
};

#endif // !__HTTP_STATS_H