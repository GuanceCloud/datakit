#ifndef __HTTP_STATS_H
#define __HTTP_STATS_H

#include <linux/types.h>
#include "../netflow/conn_stats.h"

enum
{
    L7_BUFFER_SIZE = 1 << 10, // 2^10
#define L7_BUFFER_SIZE L7_BUFFER_SIZE
    L7_BUFFER_SIZE_MASK = 0xFFF // or 0x7FFFFFFF, need to be larger than buffer size
#define L7_BUFFER_SIZE_MASK L7_BUFFER_SIZE_MASK
};

typedef enum
{
    HTTP_REQ_UNKNOWN = 0b00,
    HTTP_REQ_REQ = 0b01,
    HTTP_REQ_RESP = 0b10
} req_resp_t;

// Need to associate payload and conn info.
struct payload_id
{
    __u64 pid_tid;
    __u64 ktime;
    __u32 cpuid;
    __u32 prandom;
};

struct l7_buffer
{
    struct payload_id id;
    __s32 len;
    __u8 payload[L7_BUFFER_SIZE];
    __u64 req_ts;
};

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
    __u32 _pad1;
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
};

struct syscall_read_write_arg
{
    __u64 fd;
    void *buf;
    struct socket *skt;
    __u32 w_size;
    __u32 _pad;
    __u64 ts;
};

struct syscall_readv_writev_arg
{
    __u64 fd;
    struct iovec *vec;
    __u64 vlen;

    struct socket *skt;
    __u64 ts;
};

#endif // !__HTTP_STATS_H