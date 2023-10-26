#ifndef __BPFMAP_L7_H
#define __BPFMAP_L7_H

#include "../netflow/conn_stats.h"
#include "bpf_helpers.h"
#include "l7_stats.h"

#define MAPCANSAVEREQNUM 4
#define DEFAULTCPUNUM 256

// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

// Temporarily store tcp payload data.
struct bpf_map_def SEC("maps/bpfmap_l7_buffer") bpfmap_l7_buffer = {
    .type = BPF_MAP_TYPE_PERCPU_ARRAY,
    .key_size = sizeof(__s32),
    .value_size = sizeof(struct l7_buffer),
    .max_entries = 1,
};

static __always_inline struct l7_buffer *get_l7_buffer_percpu()
{
    __s32 index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        return NULL;
    }

    l7buffer->len = 0;
    l7buffer->req_ts = 0;
    l7buffer->isentry = 0;

    return l7buffer;
}

// Upload tcp payload data to user mode agent program.
struct bpf_map_def SEC("maps/bpfmap_l7_buffer_out") bpfmap_l7_buffer_out = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),
    .value_size = sizeof(__u32),
    .max_entries = 0,
};

// Need to clean up manually.
struct bpf_map_def SEC("maps/bpfmap_http_stats") bpfmap_http_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info),
    .value_size = sizeof(struct layer7_http),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_httpreq_fin_event")
    bpfmap_httpreq_fin_event = {
        .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
        .key_size = sizeof(__u32),
        .value_size = sizeof(__u32),
        .max_entries = 0,
};

// 需要对比 fd 是否被其他 tid 操作，如果是则出现了请求与响应不在同一线程/进程，
// 此时大概率跟丢了，可能出现异常的链路子 span

// 获取 tid 用于清理 thr_traceid
// pid + fd ->  tid
// pid + tid -> (entry) thr_traceid

BPF_HASH_MAP(bpfmap_pidfd_tidgoid, pidfd_t, __u32, 65536);
BPF_HASH_MAP(bpfmap__pidtidgoid_thr_traceid, pidtid_t, struct payload_id, 65536);

struct bpf_map_def SEC("maps/bpfmap_ssl_read_args") bpfmap_ssl_read_args = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct ssl_read_args),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_bio_new_socket_args")
    bpf_map_bio_new_socket_args = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),   // pid_tgid
        .value_size = sizeof(__u32), // fd
        .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_ssl_ctx_sockfd") bpfmap_ssl_ctx_sockfd = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(void *),
    .value_size = sizeof(__u64),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpf_map_ssl_bio_fd") bpf_map_ssl_bio_fd = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(void *),
    .value_size = sizeof(__u32),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_ssl_pidtgid_ctx") bpfmap_ssl_pidtgid_ctx = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(void *),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_syscall_read_arg")
    bpfmap_syscall_read_arg = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),
        .value_size = sizeof(struct syscall_read_write_arg),
        .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_syscall_write_arg")
    bpfmap_syscall_write_arg = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),
        .value_size = sizeof(struct syscall_read_write_arg),
        .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_syscall_readv_arg")
    bpfmap_syscall_readv_arg = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),
        .value_size = sizeof(struct syscall_readv_writev_arg),
        .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_syscall_writev_arg")
    bpfmap_syscall_writev_arg = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),
        .value_size = sizeof(struct syscall_readv_writev_arg),
        .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_syscall_sendfile_arg")
    bpfmap_syscall_sendfile_arg = {
        .type = BPF_MAP_TYPE_HASH,
        .key_size = sizeof(__u64),
        .value_size = sizeof(struct syscall_sendfile_arg),
        .max_entries = 1024,
};

#endif // !__BPFMAP_L7_H
