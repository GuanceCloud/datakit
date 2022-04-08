#ifndef __HTTP_STATS_H
#define __HTTP_STATS_H

enum
{
    HTTP_PAYLOAD_MAXSIZE = 69 // ((8 * 8) + 5), no pad
#define HTTP_PAYLOAD_MAXSIZE HTTP_PAYLOAD_MAXSIZE 
};

#include <linux/types.h>
#include "conn_stats.h"

struct http_stats
{
    // struct connection_info conn_info;

    __u8 payload[HTTP_PAYLOAD_MAXSIZE];
    __u8 req_method;
    __u16 resp_code;
    __u32 http_version;
    __u64 req_ts;
    __u64 resp_ts;
};

struct http_req_finished_info
{
    struct connection_info conn_info;
    struct http_stats http_stats;
};


struct layer7_http
{
    __u32 method;
    __u32 http_version;
    __u32 status_code;
    __u32 req_status; // request | response
};

#endif // !__HTTP_STATS_H