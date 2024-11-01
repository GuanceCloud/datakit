#ifndef __GOID_TID_H__
#define __GOID_TID_H__

#include <linux/types.h>
#include "bpf_helpers.h"

#include "process_sched.h"

// pid_tgid -> goid
BPF_HASH_MAP(bmap_tid2goid, u64, u64, 12800);

BPF_HASH_MAP(bmap_proc_filter, u32, proc_filter_info_t, 12800);

static __always_inline bool need_filter_proc(u32 *pid)
{
    proc_filter_info_t *inf = NULL;
    inf = bpf_map_lookup_elem(&bmap_proc_filter, pid);
    if (inf == NULL)
    {
        return false;
    }

    if (inf->disable)
    {
        return true;
    }
    return false;
}

#endif
