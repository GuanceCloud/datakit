#ifndef __GOID_TO_TID_H__
#define __GOID_TO_TID_H__

#include "bpf_helpers.h"

// pid_goid -> pid_tid
BPF_HASH_MAP(bmap_goid2tid, u64, u64, 12800);

// pid_tgid -> pid_goid
BPF_HASH_MAP(bmap_tid2goid, u64, u64, 12800);

#endif
