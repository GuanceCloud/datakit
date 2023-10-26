#include <net/net_namespace.h>

#include "bpf_helpers.h"
#include "process_sched.h"
#include "bpfmap.h"
#include "goid2tid.h"

SEC("tracepoint/sched/sched_process_fork")
int tracepoint__sched_process_fork(struct tp_sched_process_fork_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    rec_process_sched_status_t rec = {0};
    rec.status = REC_SCHED_FORK;
    rec.prv_pid = ctx->parent_pid;
    rec.nxt_pid = ctx->child_pid;

    bpf_probe_read(&rec.comm, sizeof(rec.comm), &ctx->child_comm);

    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &process_sched_event, cpu,
                          &rec, sizeof(rec_process_sched_status_t));

    // bpf_printk("fork, pid %d, tid %d\n", pid, tid);
    // bpf_printk("fork, parent %s %d, child pid %d\n",
    //    rec.comm, rec.prv_pid, rec.nxt_pid);
    return 0;
}

SEC("tracepoint/sched/sched_process_exec")
int tracepoint__sched_process_exec(struct tp_sched_process_exec_args *ctx)
{
    int offset = ctx->filename & 0xFFFF;
    int len = ctx->filename >> 16;

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s32 zero = 0;
    rec_process_sched_status_t rec = {0};

    // set status
    rec.prv_pid = pid_tgid >> 32;
    rec.nxt_pid = (__u32)pid_tgid;
    rec.status = REC_SCHED_EXEC;
    bpf_get_current_comm(&rec.comm, KERNEL_TASK_COMM_LEN);

    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &process_sched_event, cpu,
                          &rec, sizeof(rec_process_sched_status_t));

    // bpf_printk("exec comm %s, pid %d tid %d\n",
    //    rec->filename, rec->sched_status.prv_pid, rec->sched_status.nxt_pid);
    return 0;
}

SEC("tracepoint/sched/sched_process_exit")
int tracepoint__sched_process_exit(struct tp_sched_process_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __u64 *pidgoid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);
    bpf_map_delete_elem(&bmap_tid2goid, &pid_tgid);
    if (pidgoid != NULL)
    {
        __u64 pidgid = *pidgoid;
        bpf_map_delete_elem(&bmap_goid2tid, &pidgid);
    }

    if ((pid_tgid >> 32) == (__u32)pid_tgid) // process exit
    {
        rec_process_sched_status_t rec = {0};
        rec.status = REC_SCHED_EXIT;
        rec.prv_pid = pid_tgid >> 32;
        rec.nxt_pid = (__u32)pid_tgid;

        __u32 pid = pid_tgid >> 32;
        bpf_map_delete_elem(&bmap_procinject, &pid);

        __u64 cpu = bpf_get_smp_processor_id();
        bpf_perf_event_output(ctx, &process_sched_event, cpu,
                              &rec, sizeof(rec_process_sched_status_t));
    }

    // bpf_printk("exit comm %s, pid %d tid %d\n", ctx->comm, rec.prv_pid, rec.nxt_pid);
    return 0;
}

// ------------------------ uprobe: go sched ---------------
SEC("uprobe/runtime.execute")
int uprobe__go_runtime_execute(struct pt_regs *ctx)
{

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __u32 pid = pid_tgid >> 32;
    proc_inject_t *pr = bpf_map_lookup_elem(&bmap_procinject, &pid);
    if (pr == NULL)
    {
        return 0;
    }

    // bpf_printk("use_register %d offset %d", pr->go_use_register, pr->offset_go_runtime_g_goid);

    __u8 *g = NULL;

    if (pr->go_use_register != 0)
    {
        g = (__u8 *)PT_GO_REGS_PARAM1(ctx);
    }
    else
    {
        bpf_probe_read(&g, sizeof(g), (void *)(PT_REGS_SP(ctx) + 8));
    }
    if (g == NULL)
    {
        return 0;
    }

    // pid|goid
    __u64 pid_goid = 0;
    bpf_probe_read(&pid_goid, sizeof(pid_goid), g + pr->offset_go_runtime_g_goid);

    // bpf_printk("goid %d", pid_goid);

    pid_goid = (pid_tgid >> 32) << 32 | pid_goid;

    __u64 *pidtid = bpf_map_lookup_elem(&bmap_goid2tid, &pid_goid);
    if (pidtid != NULL)
    {
        __u64 pidtid_tmp = *pidtid;
        bpf_map_delete_elem(&bmap_tid2goid, &pidtid_tmp);
    }

    bpf_map_update_elem(&bmap_tid2goid, &pid_tgid, &pid_goid, BPF_ANY);
    bpf_map_update_elem(&bmap_goid2tid, &pid_goid, &pid_tgid, BPF_ANY);

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
