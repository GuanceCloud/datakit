#ifndef __BPFMAP_L7_H
#define __BPFMAP_L7_H

#include "bpf_helpers.h"
#include "conn_stats.h"
#include "l7_stats.h"

#define MAPCANSAVEREQNUM 4
#define DEFAULTCPUNUM 256

// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

struct bpf_map_def SEC("maps/bpfmap_conn_stats") bpfmap_http_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info),
    .value_size = sizeof(struct http_stats),
    .max_entries = 65536,
};

// 使每个 cpu 可以存 MAPCANSAVEREQNUM 个 HTTP 请求
struct bpf_map_def SEC("maps/bpfmap_httpreq_finished") bpfmap_httpreq_finished = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u16),
    .value_size = sizeof(struct http_req_finished_info),
    .max_entries = MAPCANSAVEREQNUM * DEFAULTCPUNUM,
};


struct bpf_map_def SEC("maps/bpfmap_httpreq_fin_event") bpfmap_httpreq_fin_event = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),
    .value_size = sizeof(__u32),
    .max_entries = 0,
};

#endif // !__BPFMAP_L7_H