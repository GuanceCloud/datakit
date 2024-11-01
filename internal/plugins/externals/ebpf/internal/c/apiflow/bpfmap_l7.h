#ifndef __BPFMAP_L7_H
#define __BPFMAP_L7_H

#include "../netflow/conn_stats.h"
#include "bpf_helpers.h"
#include "l7_stats.h"

#define MAPCANSAVEREQNUM 4
#define DEFAULTCPUNUM 256

// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

BPF_HASH_MAP(mp_syscall_rw_arg, __u64, syscall_rw_arg_t, 2048)
BPF_HASH_MAP(mp_syscall_rw_v_arg, __u64, syscall_rw_v_arg_t, 2048)

// Limit the number of connections tracked to 4k conns
BPF_HASH_MAP(mp_sk_inf, void *, sk_inf_t, 40960)

BPF_PERCPU_ARRAY(mp_uni_id_per_cpu, id_generator_t)

BPF_PERCPU_ARRAY(mp_network_data_per_cpu, net_data_t)

BPF_PERCPU_ARRAY(mp_network_events_per_cpu, network_events_t)

static __always_inline network_events_t *get_net_events()
{
    __s32 index = 0;
    network_events_t *events = bpf_map_lookup_elem(&mp_network_events_per_cpu, &index);
    return events;
}

BPF_PERF_EVENT_MAP(mp_upload_netwrk_events)

BPF_HASH_MAP(bpfmap_ssl_read_args, __u64, ssl_read_args_t, 2048)

BPF_HASH_MAP(bpfmap_bio_new_socket_args, __u64, __u32, 2048) // k: pid_tgid v: sockfd

BPF_HASH_MAP(bpfmap_ssl_ctx_sockfd, void *, __u64, 2048)

BPF_HASH_MAP(bpfmap_ssl_bio_fd, void *, __u32, 2048)

BPF_HASH_MAP(bpfmap_ssl_pidtgid_ctx, __u64, void *, 2048)

BPF_HASH_MAP(bpfmap_syscall_sendfile_arg, __u64, syscall_sendfile_arg_t, 2048)

BPF_HASH_MAP(mp_protocol_filter, void *, __u8, 65536)

#endif // !__BPFMAP_L7_H
