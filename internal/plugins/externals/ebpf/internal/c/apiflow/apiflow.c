#include "common.h"
#include "bpf_helpers.h"
#include "bpfmap_l7.h"
#include "../netflow/conn_stats.h"
#include "l7_stats.h"
#include "l7_utils.h"
#include "tp_arg.h"

FN_TP_SYSCALL(sys_enter_write, tp_syscall_rw_args_t *ctx)
{
    put_rw_args(ctx, &mp_syscall_rw_arg, MSG_WRITE);
    return 0;
}

FN_TP_SYSCALL(sys_exit_write, tp_syscall_exit_args_t *ctx)
{
    read_rw_data(ctx, &mp_syscall_rw_arg, P_SYSCALL_WRITE);
    return 0;
}

FN_TP_SYSCALL(sys_enter_read, tp_syscall_rw_args_t *ctx)
{
    put_rw_args(ctx, &mp_syscall_rw_arg, MSG_READ);
    return 0;
}

FN_TP_SYSCALL(sys_exit_read, tp_syscall_exit_args_t *ctx)
{
    read_rw_data(ctx, &mp_syscall_rw_arg, P_SYSCALL_READ);
    return 0;
}

FN_TP_SYSCALL(sys_enter_sendto, tp_syscall_rw_args_t *ctx)
{
    put_rw_args(ctx, &mp_syscall_rw_arg, MSG_WRITE);
    return 0;
}

FN_TP_SYSCALL(sys_exit_sendto, tp_syscall_exit_args_t *ctx)
{
    read_rw_data(ctx, &mp_syscall_rw_arg, P_SYSCALL_SENDTO);
    return 0;
}

FN_TP_SYSCALL(sys_enter_recvfrom, tp_syscall_rw_args_t *ctx)
{
    put_rw_args(ctx, &mp_syscall_rw_arg, MSG_READ);
    return 0;
}

FN_TP_SYSCALL(sys_exit_recvfrom, tp_syscall_exit_args_t *ctx)
{
    read_rw_data(ctx, &mp_syscall_rw_arg, P_SYSCALL_RECVFROM);
    return 0;
}

FN_TP_SYSCALL(sys_enter_writev, tp_syscall_rw_v_args_t *ctx)
{
    put_rw_v_args(ctx, &mp_syscall_rw_v_arg, MSG_WRITE);
    return 0;
}

FN_TP_SYSCALL(sys_exit_writev, tp_syscall_exit_args_t *ctx)
{
    read_rw_v_data(ctx, &mp_syscall_rw_v_arg, P_SYSCALL_WRITEV);
    return 0;
}

FN_TP_SYSCALL(sys_enter_readv, tp_syscall_rw_v_args_t *ctx)
{
    put_rw_v_args(ctx, &mp_syscall_rw_v_arg, MSG_READ);
    return 0;
}

FN_TP_SYSCALL(sys_exit_readv, tp_syscall_exit_args_t *ctx)
{
    read_rw_v_data(ctx, &mp_syscall_rw_v_arg, P_SYSCALL_READV);
    return 0;
}

FN_TP_SYSCALL(sys_enter_sendfile64, tp_syscall_sendfile_args_t *ctx)
{
    return 0;
}

FN_TP_SYSCALL(sys_exit_sendfile64, tp_syscall_exit_args_t *ctx)
{
    return 0;
}

FN_KPROBE(tcp_close)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);

    if (sk == NULL)
    {
        return 0;
    }

    net_data_t *dst = get_net_data_percpu();
    if (dst == NULL)
    {
        return 0;
    }

    __u8 found = 0;
    found = get_sk_inf(sk, &dst->meta.sk_inf, 0);
    if (found == 0)
    {
        return 0;
    }

    del_sk_inf(sk);

    dst->meta.func_id = P_SYSCALL_CLOSE;
    dst->meta.tid_utid = pid_tgid << 32;
    __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);
    if (goid != NULL)
    {
        dst->meta.tid_utid |= *goid;
    }

    try_upload_net_events(ctx, dst);

    clean_protocol_filter(pid_tgid, sk);

    return 0;
}

FN_UPROBE(SSL_set_fd)
{
    return 0;
}

FN_UPROBE(BIO_new_socket)
{
    return 0;
}

FN_URETPROBE(BIO_new_socket)
{
    return 0;
}

FN_UPROBE(SSL_read)
{
    return 0;
}

FN_URETPROBE(SSL_read)
{
    return 0;
}

FN_UPROBE(SSL_write)
{
    return 0;
}

FN_UPROBE(SSL_shutdown)
{
    return 0;
}

FN_KPROBE(sched_getaffinity)
{
    __u64 cpu = bpf_get_smp_processor_id();
    __s32 index = 0;
    network_events_t *events = bpf_map_lookup_elem(&mp_network_events, &index);
    if (events == NULL)
    {
        return 0;
    }

    if (events->pos.num > 0)
    {
        bpf_perf_event_output(ctx, &mp_upload_netwrk_events, cpu, events, sizeof(network_events_t));
        events->pos.len = 0;
        events->pos.num = 0;
    }

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
