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

    conn_uni_id_t uni_id = {0};
    get_conn_uni_id(sk, pid_tgid, &uni_id);
    del_conn_uni_id(sk);
    __u32 index = get_sock_buf_index(sk);
    del_sock_buf_index(sk);
    if (sk == NULL)
    {
        return 0;
    }

    netwrk_data_t *dst = get_netwrk_data_percpu();
    if (dst != NULL)
    {
        if (read_connection_info(sk, &dst->meta.conn, pid_tgid, CONN_L4_TCP) == 0)
        {
            dst->meta.index = index;
            dst->meta.func_id = P_SYSCALL_CLOSE;
            dst->meta.tid_utid = pid_tgid << 32;
            __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);
            if (goid != NULL)
            {
                dst->meta.tid_utid |= *goid;
            }

            __builtin_memcpy(&dst->meta.uni_id, &uni_id, sizeof(conn_uni_id_t));
            __u64 cpu = bpf_get_smp_processor_id();
            bpf_perf_event_output(ctx, &mp_upload_netwrk_data, cpu, dst, sizeof(netwrk_data_t));
        }
    }

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

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
