#ifndef __BPFMAP_H
#define __BPFMAP_H

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

#endif // !__BPFMAP_H