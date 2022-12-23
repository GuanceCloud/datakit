#include "bpf_helpers.h"
#include "bpfmap_l7.h"
#include "../netflow/conn_stats.h"
#include "l7_stats.h"
#include "l7_utils.h"
#include "tp_syscall_arg.h"
#include "print_apiflow.h"

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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_write_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_write")
int tracepoint__sys_exit_write(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf, arg->ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_WRITE);

clean:
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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_read_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_read")
int tracepoint__sys_exit_read(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd <= 2) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf, arg->ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_READ);

clean:
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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_write_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_sendto")
int tracepoint__sys_exit_sendto(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_write_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf, arg->ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_WRITE);

clean:
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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_read_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_recvfrom")
int tracepoint__sys_exit_recvfrom(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_read_write_arg *arg = (struct syscall_read_write_arg *)bpf_map_lookup_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd <= 2) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, arg->buf, arg->ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(arg->buf, ctx->ret, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_READ);

clean:
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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_writev_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_writev")
int tracepoint__sys_exit_writev(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_readv_writev_arg *arg = (struct syscall_readv_writev_arg *)bpf_map_lookup_elem(&bpfmap_syscall_writev_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    if (arg->vlen == 0)
    {
        goto clean;
    }
    struct iovec vec = {0};
    bpf_probe_read(&vec, sizeof(vec), arg->vec);

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, vec.iov_base, arg->ts, &conn, &stats);
    if ((req_resp != HTTP_REQ_REQ) && (req_resp != HTTP_REQ_RESP))
    {
        goto clean;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_iovec(arg->vec, arg->vlen, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_WRITE);

clean:
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

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_map_update_elem(&bpfmap_syscall_readv_arg, &pid_tgid, &arg, BPF_ANY);

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_readv")
int tracepoint__sys_exit_readv(struct tp_syscall_exit_args *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    __s64 buf_size = ctx->ret;
    struct syscall_readv_writev_arg *arg = (struct syscall_readv_writev_arg *)bpf_map_lookup_elem(&bpfmap_syscall_readv_arg, &pid_tgid);
    if (buf_size <= 0 || arg == NULL || arg->fd < 3) // fd 0-2: stdin, stdout, stderr
    {
        goto clean;
    }

    arg->ts = bpf_ktime_get_ns();

    if (arg->vlen == 0)
    {
        goto clean;
    }
    struct iovec vec = {0};
    bpf_probe_read(&vec, sizeof(vec), arg->vec);

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(arg->skt, vec.iov_base, arg->ts, &conn, &stats);
    if ((req_resp != HTTP_REQ_REQ) && (req_resp != HTTP_REQ_RESP))
    {
        goto clean;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_iovec(arg->vec, arg->vlen, &stats.req_payload_id, l7buffer);

    parse_http1x(ctx, l7buffer, arg->ts, &conn, &stats, req_resp, MSG_READ);

clean:
    bpf_map_delete_elem(&bpfmap_syscall_read_arg, &pid_tgid);
    return 0;
}

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

    bpf_map_update_elem(&bpfmap_ssl_read_args, &pid_tgid, &args, BPF_ANY);
    return 0;
}

SEC("uretprobe/SSL_read")
int uretprobe__SSL_read(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct ssl_read_args *args = bpf_map_lookup_elem(&bpfmap_ssl_read_args, &pid_tgid);
    if (args == NULL)
    {
        goto clean;
    }

    void *ssl_ctx = args->ctx;
    __u64 *fd_ptr = (__u64 *)bpf_map_lookup_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    if (fd_ptr == NULL)
    {
        goto clean;
    }

    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, *fd_ptr);

    if (skt == NULL)
    {
        goto clean;
    }

    __u64 ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(skt, args->buf, ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        goto clean;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        goto clean;
    }

    copy_data_from_buffer(args->buf, args->num, &stats.req_payload_id, l7buffer);
    parse_http1x(ctx, l7buffer, ts, &conn, &stats, req_resp, MSG_READ);

clean:
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

    __u64 ts = bpf_ktime_get_ns();

    struct connection_info conn = {0};
    struct layer7_http stats = {0};
    req_resp_t req_resp = checkHTTP(skt, write_buf, ts, &conn, &stats);
    if (req_resp == HTTP_REQ_UNKNOWN)
    {
        return 0;
    }

    // 如果是 resp，此 buffer 不使用
    int index = 0;
    struct l7_buffer *l7buffer = bpf_map_lookup_elem(&bpfmap_l7_buffer, &index);
    if (l7buffer == NULL)
    {
        return 0;
    }

    copy_data_from_buffer(write_buf, num, &stats.req_payload_id, l7buffer);
    parse_http1x(ctx, l7buffer, ts, &conn, &stats, req_resp, MSG_WRITE);

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
