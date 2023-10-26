#ifndef __PROCESS_SCHED_BPFMAP_H_
#define __PROCESS_SCHED_BPFMAP_H_

#include "bpf_helpers.h"
#include "process_sched.h"

BPF_PERF_EVENT_MAP(process_sched_event);

BPF_HASH_MAP(bmap_procinject, u32, proc_inject_t, 4096);

#endif
