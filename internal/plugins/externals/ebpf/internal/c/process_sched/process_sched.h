#ifndef __PROCESS_SCHED_H__
#define __PROCESS_SCHED_H__

#include <linux/types.h>

#define KERNEL_TASK_COMM_LEN 16

#define FILENAME_LEN_MASK 0x7F
#define FILENAME_LEN (FILENAME_LEN_MASK + 1)

enum rec_sched_stat
{
    REC_SCHED_FORK = 0b1 << 0,
    REC_SCHED_EXEC = 0b1 << 1,
    REC_SCHED_EXIT = 0b1 << 2,
};

typedef struct rec_process_sched_status
{
    __s32 status;
    __s32 prv_pid; // parent_pid or tgid or old_pid
    __s32 nxt_pid;
    __s32 __pad;
    // __s32 cur_tgid;
    __u8 comm[KERNEL_TASK_COMM_LEN];
} rec_process_sched_status_t;

typedef struct rec_process_sched_status_with_filename
{
    rec_process_sched_status_t sched_status;
    __u64 filename_len;
    __u8 filename[FILENAME_LEN];
} rec_sched_with_fname_t;

struct tp_sched_process_fork_args
{
    __u64 _common;

    __u8 parent_comm[16];
    __s32 parent_pid;
    __u8 child_comm[16];
    __s32 child_pid;
};

struct tp_sched_process_exit_args
{
    __u64 _common;

    __u8 comm[16];
    __s32 pid;
    __s32 prio;
};

struct tp_sched_process_exec_args
{
    __u64 _common;

    __u32 filename;
    __s32 pid;
    __s32 old_pid;
};

typedef struct proc_inject
{
    __u64 offset_go_runtime_g_goid;
    __u64 go_use_register;
} proc_inject_t;
#endif