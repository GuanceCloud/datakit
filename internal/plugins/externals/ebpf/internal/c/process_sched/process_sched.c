#include <net/net_namespace.h>

#include "bpf_helpers.h"
#include "process_sched.h"
#include "bpfmap.h"
#include "goidtid.h"

SEC("tracepoint/sched/sched_process_fork")
int tracepoint__sched_process_fork(struct tp_sched_process_fork_args *ctx)
{
    // syscall: clone, clone3, fork, vfork ...
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    rec_process_sched_status_t rec = {0};
    rec.status = REC_SCHED_FORK;
    rec.prv_pid = pid_tgid >> 32;
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
    rec.nxt_pid = rec.prv_pid;
    rec.status = REC_SCHED_EXEC;
    bpf_get_current_comm(&rec.comm, KERNEL_TASK_COMM_LEN);

    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &process_sched_event, cpu,
                          &rec, sizeof(rec_process_sched_status_t));

    return 0;
}

SEC("tracepoint/sched/sched_process_exit")
int tracepoint__sched_process_exit(struct tp_sched_process_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_delete_elem(&bmap_tid2goid, &pid_tgid);

    int pid = pid_tgid >> 32;
    int tid = (u32)pid_tgid;
    if (pid == tid) // process(all threads) exit
    {
        rec_process_sched_status_t rec = {0};
        rec.status = REC_SCHED_EXIT;
        rec.prv_pid = pid;
        rec.nxt_pid = tid;

        bpf_map_delete_elem(&bmap_procinject, &pid);
        bpf_map_delete_elem(&bmap_proc_filter, &pid);

        __u64 cpu = bpf_get_smp_processor_id();
        bpf_perf_event_output(ctx, &process_sched_event, cpu,
                              &rec, sizeof(rec_process_sched_status_t));
    }
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
    __u64 goid = 0;
    bpf_probe_read(&goid, sizeof(goid), g + pr->offset_go_runtime_g_goid);

    // bpf_printk("goid %d", pid_goid);

    bpf_map_update_elem(&bmap_tid2goid, &pid_tgid, &goid, BPF_ANY);

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
