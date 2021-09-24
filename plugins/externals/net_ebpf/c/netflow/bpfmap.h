#ifndef __BPFMAP_H
#define __BPFMAP_H

#include "bpf_helpers.h"
#include "conn_stats.h"
// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

struct bpf_map_def SEC("maps/bpfmap_conn_stats") bpfmap_conn_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info),
    .value_size = sizeof(struct connection_stats),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_conn_tcp_stats") bpfmap_conn_tcp_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info), // pid shoud be set to 0
    .value_size = sizeof(struct connection_tcp_stats),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_closed_event") bpfmap_closed_event = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),   // smp_processor_id
    .value_size = sizeof(__u32), // perf file fd
    .max_entries = 0,
};

// Temporarily store the pid_tgid(key, u64) and port(value, u16) when inet_bind(v4/v6) is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_inetbind") bpfmap_tmp_inetbind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(__u16),
    .max_entries = 65536,
};

// map key: struct port_bind
// map value: PORT_CLOSED or PORT_LISTENING
struct bpf_map_def SEC("maps/bpfmap_port_bind") bpfmap_port_bind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct port_bind),
    .value_size = sizeof(__u8),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_udp_port_bind") bpfmap_udp_port_bind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct port_bind),
    .value_size = sizeof(__u8),
    .max_entries = 65536,
};
struct udp_revcmsg_tmp
{
    struct sock *sk;
    struct msghdr *msg;
};

struct bpf_map_def SEC("maps/bpf_map_tmp_udprecvmsg") bpf_map_tmp_udprecvmsg = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct udp_revcmsg_tmp),
    .max_entries = 65536,
};

// Temporarily store the pid_tgid(key, u64) and sockfd(value, u32) when sockfd_lookup_light is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_sockfdlookuplight") bpfmap_tmp_sockfdlookuplight = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(__u32),
    .max_entries = 65536,
};

// key: struct pid_fd, value: struct sock pointer
struct bpf_map_def SEC("maps/bpfmap_sockfd") bpfmap_sockfd = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct pid_fd),
    .value_size = sizeof(struct sock *),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_sockfd_inverted") bpfmap_sockfd_inverted = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct sock *),
    .value_size = sizeof(struct pid_fd),
    .max_entries = 65536,
};

#endif // !__BPFMAP_H