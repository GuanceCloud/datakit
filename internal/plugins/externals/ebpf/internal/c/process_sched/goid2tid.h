#ifndef __GOID_TO_TID_H__
#define __GOID_TO_TID_H__

#include "bpf_helpers.h"

// pid_tgid -> goid
BPF_HASH_MAP(bmap_tid2goid, u64, u64, 12800);

#endif
