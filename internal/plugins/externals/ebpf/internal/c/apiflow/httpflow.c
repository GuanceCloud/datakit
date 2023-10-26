#include "bpf_helpers.h"
#include "bpfmap_l7.h"
#include "../netflow/conn_stats.h"
#include "l7_stats.h"
#include "l7_utils.h"
#include "tp_arg.h"

// ============================= syscall =========================

SEC("tracepoint/syscalls/sys_enter_write")
int tracepoint__sys_enter_write(struct tp_syscall_read_write_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_read_write_arg arg = {
        .buf = ctx->buf,
        .fd = ctx->fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),
    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_write_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_write")
int tracepoint__sys_exit_write(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    if (arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    {
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf,
                                    &conn, &stats, ctx->ret);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_WRITE);
    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer, P_SYSCALL_WRITE);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_WRITE, P_SYSCALL_WRITE, arg->fd);

clean:
    http_try_upload(ctx, &conn, 0, buf_size, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_WRITE);
    bpf_map_delete_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_read")
int tracepoint__sys_enter_read(struct tp_syscall_read_write_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_read_write_arg arg = {
        .buf = ctx->buf,
        .fd = ctx->fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),
    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_read_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_read")
int tracepoint__sys_exit_read(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    if (arg == NULL || arg->fd <= 2) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    {
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf,
                                    &conn, &stats, ctx->ret);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }
    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_READ);
    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer, P_SYSCALL_READ);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_READ, P_SYSCALL_READ, arg->fd);

clean:
    http_try_upload(ctx, &conn, buf_size, 0, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_READ);
    bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    return 0;
}

// used by curl ...
SEC("tracepoint/syscalls/sys_enter_sendto")
int tracepoint__sys_enter_sendto(struct tp_syscall_read_write_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_read_write_arg arg = {
        .buf = ctx->buf,
        .fd = ctx->fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),

    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_write_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_sendto")
int tracepoint__sys_exit_sendto(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    if (arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_write_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    {
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf,
                                    &conn, &stats, ctx->ret);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }
    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_WRITE);
    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer, P_SYSCALL_SENDTO);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_WRITE, P_SYSCALL_SENDTO, arg->fd);

clean:
    http_try_upload(ctx, &conn, 0, buf_size, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_SENDTO);

    bpf_map_delete_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_recvfrom")
int tracepoint__sys_enter_recvfrom(struct tp_syscall_read_write_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_read_write_arg arg = {
        .buf = ctx->buf,
        .fd = ctx->fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),

    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_read_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_recvfrom")
int tracepoint__sys_exit_recvfrom(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    if (arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    {
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf,
                                    &conn, &stats, ctx->ret);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }
    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_READ);
    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer, P_SYSCALL_RECVFROM);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_READ, P_SYSCALL_RECVFROM, arg->fd);

clean:
    http_try_upload(ctx, &conn, buf_size, 0, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_RECVFROM);
    bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_writev")
int tracepoint__sys_enter_writev(struct tp_syscall_writev_readv_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_readv_writev_arg arg = {
        .fd = ctx->fd,
        .vec = ctx->vec,
        .vlen = ctx->vlen,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),

    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_writev_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_writev")
int tracepoint__sys_exit_writev(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_readv_writev_arg *arg = (struct syscall_readv_writev_arg *)bpf_map_lookup_elem(&bpfmap_syscall_writev_arg, &pid_tgid);
    if (arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    { // FIN, error, ...
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    if (arg->vlen == 0)
    {
        goto clean;
    }
    struct iovec vec = {0};
    bpf_probe_read(&vec, sizeof(vec), arg->vec);

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, vec.iov_base,
                                    &conn, &stats, vec.iov_len);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }
    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_WRITE);

    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_iovec(arg->vec, arg->vlen, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_WRITE, P_SYSCALL_WRITEV, arg->fd);

clean:
    http_try_upload(ctx, &conn, 0, buf_size, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_WRITEV);

    bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_readv")
int tracepoint__sys_enter_readv(struct tp_syscall_writev_readv_args *ctx)
{
    if (ctx->fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_readv_writev_arg arg = {
        .fd = ctx->fd,
        .vec = ctx->vec,
        .vlen = ctx->vlen,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),
    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_readv_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_readv")
int tracepoint__sys_exit_readv(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    __s64 buf_size = ctx->ret;
    struct syscall_readv_writev_arg *arg = (struct syscall_readv_writev_arg *)bpf_map_lookup_elem(&bpfmap_syscall_readv_arg, &pid_tgid);
    if (arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
        return 0;
    }

    if (buf_size <= 0)
    {
        goto clean;
    }

    // arg->ts = bpf_ktime_get_ns();

    if (arg->vlen == 0)
    {
        goto clean;
    }
    struct iovec vec = {0};
    bpf_probe_read(&vec, sizeof(vec), arg->vec);

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, vec.iov_base,
                                    &conn, &stats, vec.iov_len);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }
    rec_seq(&stats, arg->copied_seq, arg->write_seq, req_resp, MSG_READ);

    upate_req_payload_id(&stats, pid_tgid, arg->ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_iovec(arg->vec, arg->vlen, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp,
                 MSG_READ, P_SYSCALL_READV, arg->fd);

clean:
    http_try_upload(ctx, &conn, buf_size, 0, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_READV);

    bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    return 0;
}

// tcp_close
SEC("kprobe/tcp_close")
int kprobe__tcp_close(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);

    if (sk == NULL)
    {
        return 0;
    }

    {
        __u8 dd[4];
        bpf_get_current_comm(dd, 4);

        if (dd[0] == 'm' && dd[1] == 'a' && dd[2] == 'i')
        {
            __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);
            // if (goid != NULL)
            // {
            //     bpf_printk("close %d", *goid);
            // }
            // else
            // {
            //     bpf_printk("close %d %d", pid_tgid >> 32, (__u32)pid_tgid);
            // }
        }
    }

    struct connection_info conn = {0};

    if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return 0;
    }

    http_try_upload(ctx, &conn, 0, 0, 0, bpf_ktime_get_ns(), P_SYSCALL_CLOSE);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_sendfile64")
int tracepoint__sys_enter_sendfile64(struct tp_syscall_sendfile_arg *ctx)
{
    if (ctx->out_fd < 3)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->out_fd);
    if (skt == NULL)
    {
        return 0;
    }

    struct syscall_sendfile_arg arg = {
        .fd = ctx->out_fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),
    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0)
    {
        return 0;
    }

    // tcp only
    switch (sktype)
    {
    case SOCK_STREAM:
        break;
    default:
        return 0;
    }

    arg.copied_seq = read_copied_seq(sk);
    arg.write_seq = read_write_seq(sk);
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_sendfile_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_sendfile64")
int tracepoint__sys_exit_sendfile64(struct tp_syscall_exit_args *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct connection_info conn = {0};

    struct syscall_sendfile_arg *arg = (struct syscall_sendfile_arg *)bpf_map_lookup_elem(
        &bpfmap_syscall_sendfile_arg, &pid_tgid);

    if (arg == NULL || arg->fd < 3)
    {
        goto end_run;
    }

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(arg->skt, &sk, &sktype) != 0)
    {
        goto end_run;
    }

    // tcp only
    switch (sktype)
    {
    case SOCK_STREAM:
        break;
    default:
        goto end_run;
    }

    if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        goto end_run;
    }

    http_try_upload(ctx, &conn, 0, ctx->ret, arg->ts, bpf_ktime_get_ns(), P_SYSCALL_SENDFILE);
end_run:
    bpf_map_delete_elem(&bpfmap_syscall_sendfile_arg, &pid_tgid);
    return 0;
}

// SEC("tracepoint/syscalls/sys_enter_sendmmsg")
// int tracepoint__sys_enter_sendmmsg(struct tp_syscall_exit)

// ---------------------------------------- uprobe --------------------------------------------

SEC("uprobe/SSL_set_fd")
int uprobe__SSL_set_fd(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    __u32 fd = (__u32)PT_REGS_PARM2(ctx);

    init_ssl_sockfd(ssl_ctx, fd);

    return 0;
}

SEC("uprobe/BIO_new_socket")
int uprobe__BIO_new_socket(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 fd = PT_REGS_PARM1(ctx);
    bpf_map_update_elem(&bpf_map_bio_new_socket_args, &pid_tgid, &fd, BPF_ANY);
    return 0;
}

SEC("uretprobe/BIO_new_socket")
int uretprobe__BIO_new_socket(struct pt_regs *ctx)
{
    __u64 pid_tgid = (__u64)bpf_get_current_pid_tgid();
    __u32 *fd_map_value = (__u32 *)bpf_map_lookup_elem(&bpf_map_bio_new_socket_args, &pid_tgid);
    if (fd_map_value == NULL)
    {
        goto cleanup;
    }

    void *bio = (void *)PT_REGS_RC(ctx);
    if (bio == NULL)
    {
        goto cleanup;
    }

    __u32 fd = *fd_map_value;
    bpf_map_update_elem(&bpf_map_ssl_bio_fd, &bio, &fd, BPF_ANY);

cleanup:
    bpf_map_delete_elem(&bpf_map_bio_new_socket_args, &pid_tgid);
    return 0;
}

SEC("uprobe/SSL_set_bio")
int uprobe__SSL_set_bio(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    void *bio = (void *)PT_REGS_PARM2(ctx);

    __u32 *fd = bpf_map_lookup_elem(&bpf_map_ssl_bio_fd, &bio);
    if (fd == NULL)
    {
        goto cleanup;
    }
    init_ssl_sockfd(ssl_ctx, *fd);

cleanup:
    bpf_map_delete_elem(&bpf_map_ssl_bio_fd, &bio);
    return 0;
}

SEC("uprobe/SSL_read")
int uprobe__SSL_read(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct ssl_read_args args = {0};
    args.ctx = (void *)PT_REGS_PARM1(ctx);
    args.buf = (void *)PT_REGS_PARM2(ctx);
    args.num = (__s32)PT_REGS_PARM3(ctx);

    void *ssl_ctx = args.ctx;
    __u64 *fd_ptr = (__u64 *)bpf_map_lookup_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    if (fd_ptr == NULL)
    {
        return 0;
    }
    struct task_struct *task = bpf_get_current_task();
    struct socket *skt = get_socket_from_fd(task, *fd_ptr);
    if (skt == NULL)
    {
        return 0;
    }

    // socket addr
    args.skt = skt;

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    args.copied_seq = read_copied_seq(sk);
    args.write_seq = read_write_seq(sk);

    args.ts = bpf_ktime_get_ns();

    bpf_map_update_elem(&bpfmap_ssl_read_args, &pid_tgid, &args, BPF_ANY);
    return 0;
}

SEC("uretprobe/SSL_read")
int uretprobe__SSL_read(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct connection_info conn = {0};

    struct ssl_read_args *args = bpf_map_lookup_elem(&bpfmap_ssl_read_args, &pid_tgid);
    if (args == NULL)
    {
        return 0;
    }

    if (args->num <= 0)
    {
        goto clean;
    }

    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTPS(args->skt, args->buf,
                                     &conn, &stats, args->num);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    rec_seq(&stats, args->copied_seq, args->write_seq, req_resp, MSG_READ);

    upate_req_payload_id(&stats, pid_tgid, args->ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(args->buf, args->num, &stats.req_payload_id, l7buffer, P_USR_SSL_READ);
    parse_http1x(ctx, l7buffer, args->ts, &conn, &stats, req_resp, MSG_READ, P_USR_SSL_READ, -1);

clean:
    http_try_upload(ctx, &conn, 0, 0, 0, bpf_ktime_get_ns(), P_USR_SSL_READ);
    bpf_map_delete_elem(&bpfmap_ssl_read_args, &pid_tgid);
    return 0;
}

SEC("uprobe/SSL_write")
int uprobe__SSL_write(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    void *write_buf = (void *)PT_REGS_PARM2(ctx);
    __s32 num = (__s32)PT_REGS_PARM3(ctx);

    __u64 *fd_ptr = (__u64 *)bpf_map_lookup_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    if (fd_ptr == NULL)
    {
        return 0;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, *fd_ptr);

    if (skt == NULL)
    {
        return 0;
    }

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return 0;
    }

    __u64 ts = bpf_ktime_get_ns();
    struct connection_info conn = {0};
    struct layer7_http stats = {0};

    if (num <= 0)
    {
        goto clean;
    }

    req_resp_t req_resp = checkHTTPS(skt, write_buf,
                                     &conn, &stats, num);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    __u32 copied_seq = read_copied_seq(sk);
    __u32 write_seq = read_write_seq(sk);

    rec_seq(&stats, copied_seq, write_seq, req_resp, MSG_WRITE);

    upate_req_payload_id(&stats, pid_tgid, ts);

    // If it is resp, this buffer is not used.
    struct l7_buffer *l7buffer = get_l7_buffer_percpu();
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(write_buf, num, &stats.req_payload_id, l7buffer, P_USR_SSL_WRITE);
    parse_http1x(ctx, l7buffer, ts, &conn, &stats, req_resp, MSG_WRITE, P_USR_SSL_WRITE, -1);

clean:
    http_try_upload(ctx, &conn, 0, 0, 0, ts, P_USR_SSL_WRITE);

    return 0;
}

SEC("uprobe/SSL_shutdown")
int uprobe__SSL_shutdown(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    bpf_map_delete_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
