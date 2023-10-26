#ifndef __FILTER_H__
#define __FILTER_H__

#include "bpf_helpers.h"
#include "offset.h"

static __always_inline int skipConn(__u8 *process_name, __u64 pidtgid)
{
    char actual[KERNEL_TASK_COMM_LEN] = {};
    bpf_get_current_comm(&actual, KERNEL_TASK_COMM_LEN);
    for (int i = 0; i < KERNEL_TASK_COMM_LEN - 1; i++)
    {
        if (actual[i] != process_name[i])
        {
            return 1;
        }
    }
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    if (pid_tgid != pidtgid)
    {
        return 1;
    }
    return 0;
}

#endif