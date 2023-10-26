#ifndef __OFFSET_BPFMAP_H
#define __OFFSET_BPFMAP_H

#include "bpf_helpers.h"
#include "offset.h"
// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

struct bpf_map_def SEC("maps/bpfmap_offset_guess") bpfmap_offset_guess = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct offset_guess),
    .max_entries = 1,
};

struct bpf_map_def SEC("maps/bpfmap_tcpv6conn") bpfmap_tcpv6conn = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(char *),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_offset_httpflow") bpf_map_offset_httpflow = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct offset_httpflow),
    .max_entries = 1,
};

struct bpf_map_def SEC("maps/bpfmap_file_ptr") bpf_map_file_ptr = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(void *),
    .max_entries = 1024,
};

struct bpf_map_def SEC("maps/bpfmap_sock_common_getsockopt_arg") bpf_map_sock_common_getsockopt_arg = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct comm_getsockopt_arg),
    .max_entries = 1024,
};

#endif // !__BPFMAP_H