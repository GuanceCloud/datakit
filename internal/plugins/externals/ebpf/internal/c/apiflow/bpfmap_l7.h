#ifndef __BPFMAP_L7_H
#define __BPFMAP_L7_H

#include "../netflow/conn_stats.h"
#include "bpf_helpers.h"
#include "l7_stats.h"

#define MAPCANSAVEREQNUM 4
#define DEFAULTCPUNUM 256

// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

BPF_HASH_MAP(mp_syscall_rw_arg, __u64, syscall_rw_arg_t, 1024)
BPF_HASH_MAP(mp_syscall_rw_v_arg, __u64, syscall_rw_v_arg_t, 1024)

BPF_HASH_MAP(mp_sock_buf_index, void *, __u32, 65535)
BPF_HASH_MAP(mp_conn_uni_id, void *, conn_uni_id_t, 65535)

BPF_PERCPU_MAP(mp_netwrk_data_pool, netwrk_data_t)
BPF_PERF_EVENT_MAP(mp_upload_netwrk_data)

BPF_HASH_MAP(bpfmap_ssl_read_args, __u64, ssl_read_args_t, 1024);

BPF_HASH_MAP(bpfmap_bio_new_socket_args, __u64, __u32, 1024) // k: pid_tgid v: sockfd

BPF_HASH_MAP(bpfmap_ssl_ctx_sockfd, void *, __u64, 1024)

BPF_HASH_MAP(bpfmap_ssl_bio_fd, void *, __u32, 1024)

BPF_HASH_MAP(bpfmap_ssl_pidtgid_ctx, __u64, void *, 1024)

BPF_HASH_MAP(bpfmap_syscall_sendfile_arg, __u64, syscall_sendfile_arg_t, 1024)

// TODO: use it
BPF_HASH_MAP(mp_protocol_filter, pid_skptr_t, __u8, 65536)

#endif // !__BPFMAP_L7_H
