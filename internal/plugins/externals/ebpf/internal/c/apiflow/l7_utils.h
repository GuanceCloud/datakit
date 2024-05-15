#ifndef __L7_UTILS_
#define __L7_UTILS_

#define KEEPPACKET -1
#define DROPPACKET 0

#include <linux/fdtable.h>
#include <linux/socket.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/in.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/ipv6.h>
#include <uapi/linux/tcp.h>
#include <uapi/linux/udp.h>

#include "../netflow/netflow_utils.h"
#include "../conntrack/maps.h"
#include "../process_sched/goid2tid.h"
#include "bpfmap_l7.h"

#include "bpf_helpers.h"
#include "l7_stats.h"
#include "tp_arg.h"

enum MSG_RW
{
    MSG_READ = 0b01,
    MSG_WRITE = 0b10,
};

enum
{
    HTTP_METHOD_UNKNOWN = 0x00,
    HTTP_METHOD_GET,
    HTTP_METHOD_POST,
    HTTP_METHOD_PUT,
    HTTP_METHOD_DELETE,
    HTTP_METHOD_HEAD,
    HTTP_METHOD_OPTIONS,
    HTTP_METHOD_PATCH,

    // TODO: parse such HTTP data.
    HTTP_METHOD_CONNECT,
    HTTP_METHOD_TRACE
};

static __always_inline void fill_conn_uni_id(conn_uni_id_t *id, void *sk, __u64 k_time)
{
    id->ktime = (__u32)k_time;
    id->sk = (__u64)sk;
    id->prandom = bpf_get_prandom_u32();
}

static __always_inline void get_conn_uni_id(void *sk, __u64 pid_tgid, conn_uni_id_t *dst)
{
    conn_uni_id_t *id = NULL;
    id = bpf_map_lookup_elem(&mp_conn_uni_id, &sk);
    if (id == NULL)
    {
        __u64 k_time = bpf_ktime_get_ns();
        fill_conn_uni_id(dst, sk, k_time);
        bpf_map_update_elem(&mp_conn_uni_id, &sk, dst, BPF_NOEXIST);
        id = bpf_map_lookup_elem(&mp_conn_uni_id, &sk);
        if (id != NULL)
        {
            __builtin_memcpy(dst, id, sizeof(conn_uni_id_t));
        }
    }
    else
    {
        __builtin_memcpy(dst, id, sizeof(conn_uni_id_t));
    }
}

static __always_inline void del_conn_uni_id(void *sk)
{
    bpf_map_delete_elem(&mp_conn_uni_id, &sk);
}

static __always_inline __u32
get_sock_buf_index(void *sk)
{
    __u32 i = 0;
    __u32 *idx = bpf_map_lookup_elem(&mp_sock_buf_index, &sk);
    if (idx == NULL)
    {
        bpf_map_update_elem(&mp_sock_buf_index, &sk, &i, BPF_NOEXIST);
        __u32 *idx = bpf_map_lookup_elem(&mp_sock_buf_index, &sk);
        if (idx != NULL)
        {
            i = *idx;
        }
    }
    else
    {
        i = *idx;
    }
    i += 1;
    bpf_map_update_elem(&mp_sock_buf_index, &sk, &i, BPF_ANY);
    return i;
}

static __always_inline void del_sock_buf_index(void *sk)
{
    bpf_map_delete_elem(&mp_sock_buf_index, &sk);
}

// args: syscall_rw_arg_t, syscall_rw_v_arg_t; dst: netwrk_data_t
#define read_net_meta(args, pid_tgid, dst)                             \
    do                                                                 \
    {                                                                  \
        __u64 ts = bpf_ktime_get_ns();                                 \
        if (!conn_info_from_skt(args->skt, &dst->meta.conn, pid_tgid)) \
        {                                                              \
            goto cleanup;                                              \
        }                                                              \
                                                                       \
        struct sock *sk = NULL;                                        \
        enum sock_type sktype = 0;                                     \
        get_sock_from_skt(args->skt, &sk, &sktype);                    \
        get_conn_uni_id(sk, pid_tgid, &dst->meta.uni_id);              \
        dst->meta.index = get_sock_buf_index(sk);                      \
        dst->meta.ts = args->ts;                                       \
        dst->meta.ts_tail = ts;                                        \
        dst->meta.tid_utid = pid_tgid << 32;                           \
        dst->meta.tcp_seq = args->tcp_seq;                             \
        dst->meta.func_id = fn;                                        \
        dst->meta.fd = args->fd;                                       \
        dst->meta.act_size = ctx->ret;                                 \
                                                                       \
        __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);  \
        if (goid != NULL)                                              \
        {                                                              \
            dst->meta.tid_utid |= *goid;                               \
        }                                                              \
    } while (0)

static __always_inline bool proto_filter(__u64 pid_tgid, void *sock_ptr)
{
    pid_skptr_t key = {
        .pid = pid_tgid >> 32,
        .sk_ptr = (__u64)sock_ptr};

    __u8 *val = bpf_map_lookup_elem(&mp_protocol_filter, &key);
    if (val != NULL && *val == 1)
    {
        return true;
    }
    return false;
}

static __always_inline void clean_protocol_filter(__u64 pid_tgid, void *sock_ptr)
{
    pid_skptr_t key = {
        .pid = pid_tgid >> 32,
        .sk_ptr = (__u64)sock_ptr};

    bpf_map_delete_elem(&mp_protocol_filter, &key);
}

// ret 0: r, 1: w
static __always_inline int vfs_r_or_w(tp_syscalls_fn_t f)
{
    switch (f)
    {
    // syscalls
    case P_SYSCALL_WRITE:
        return P_GROUP_WRITE;
        break;
    case P_SYSCALL_READ:
        return P_GROUP_READ;
        break;
    case P_SYSCALL_SENDTO:
        return P_GROUP_WRITE;
        break;
    case P_SYSCALL_RECVFROM:
        return P_GROUP_READ;
        break;
    case P_SYSCALL_WRITEV:
        return P_GROUP_WRITE;
        break;
    case P_SYSCALL_READV:
        return P_GROUP_READ;
        break;
    case P_SYSCALL_SENDFILE:
        return P_GROUP_WRITE;
        break;

    // user
    case P_USR_SSL_READ:
        return P_GROUP_READ;
        break;
    case P_USR_SSL_WRITE:
        return P_GROUP_WRITE;
        break;
    default:
        return P_GROUP_UNKNOWN;
        break;
    }
}

static __always_inline int p_group_eq(tp_syscalls_fn_t src, tp_syscalls_fn_t dst)
{
    int s = vfs_r_or_w(src);
    int d = vfs_r_or_w(dst);
    if (s == d)
    {
        return 1;
    }
    return 0;
}

struct buf_iovec
{
    // we need to divide a large buffer into several small pieces
    __u8 data[BUF_IOVEC_LEN];
};

static __always_inline void read_network_data_from_vec(netwrk_data_t *dst, struct iovec *vec,
                                                       __u64 vlen, __s64 len_or_errno)
{
    if (len_or_errno <= 0)
    {
        return;
    }

    __s32 offset = 0;
#pragma unroll
    for (int i = 0; i < 5; i++)
    {
        if (i >= vlen)
        {
            break;
        }
        struct iovec v = {0};
        bpf_probe_read(&v, sizeof(v), vec + i);
        int iov_len = v.iov_len;
        if (iov_len > 0)
        {

            struct buf_iovec *buf = (struct buf_iovec *)((__u8 *)dst->payload + offset);
            if (offset + sizeof(buf->data) > sizeof(dst->payload))
            {
                break;
            }

            if (iov_len >= sizeof(buf->data))
            {
                bpf_probe_read(&buf->data, sizeof(buf->data), (__u8 *)v.iov_base);
                offset += sizeof(buf->data);
                // 不连续则丢弃后续的数据
                break;
            }
            else
            {
                iov_len = iov_len & (sizeof(buf->data) - 1);
                if (iov_len > 0)
                {
                    bpf_probe_read(&buf->data, iov_len, (__u8 *)v.iov_base);
                    offset += iov_len;
                }
            }
        }
    }

    dst->meta.buf_len = offset;
}

static __always_inline void read_netwrk_data(netwrk_data_t *dst, __u8 *buf, __s64 len_or_errno)
{
    if (len_or_errno <= 0)
    {
        return;
    }

    if (len_or_errno >= sizeof(dst->payload))
    {
        len_or_errno = sizeof(dst->payload);
    }
    else
    {
        len_or_errno = len_or_errno & (sizeof(dst->payload) - 1);
    }
    bpf_probe_read(&dst->payload, len_or_errno, buf);
    dst->meta.buf_len = len_or_errno;
}

static __always_inline struct socket *get_socket_from_fd(
    struct task_struct *task, int fd)
{
    struct files_struct *files = NULL;
    __u64 offset = 0;
    offset = load_offset_task_struct_files();

    bpf_probe_read(
        &files, sizeof(files),
        (__u8 *)task +
            offset); // bpf_probe_read(&files, sizeof(files), &task->files);

    if (files == NULL)
    {
        return NULL;
    }

    struct fdtable *fdt = NULL;
    offset = load_offset_files_struct_fdt();

    bpf_probe_read(
        &fdt, sizeof(fdt),
        (__u8 *)files +
            offset); // bpf_probe_read(&fdt, sizeof(fdt), &files->fdt);

    if (fdt == NULL)
    {
        return NULL;
    }

    struct file **farry = NULL;
    bpf_probe_read(&farry, sizeof(farry), &fdt->fd);
    if (farry == NULL)
    {
        return NULL;
    }

    struct file *skfile = NULL;
    bpf_probe_read(&skfile, sizeof(skfile), farry + fd);
    if (skfile == NULL)
    {
        return NULL;
    }

    // TODO: check file ops
    // if (skfile->f_op == &socket_file_ops) {
    //}

    struct socket *skt = NULL;
    offset = load_offset_file_private_data();

    bpf_probe_read(
        &skt, sizeof(skt),
        (__u8 *)skfile +
            offset); // bpf_probe_read(&skt, sizeof(skt), &skfile->private_data);
    if (skt == NULL)
    {
        return NULL;
    }

    // check is socket
    struct file *file_addr = NULL;
    offset = load_offset_socket_file();
    bpf_probe_read(&file_addr, sizeof(file_addr),
                   (__u8 *)skt + offset); // bpf_probe_read(&file_addr,
                                          // sizeof(file_addr), &skt->file);
    if (file_addr != skfile)
    {
        return NULL;
    }

    return skt;
}

static __always_inline int get_sock_from_skt(struct socket *skt,
                                             struct sock **sk,
                                             enum sock_type *sktype)
{
    __u64 offset_socket_sk = load_offset_socket_sk();

    struct proto_ops *ops = NULL;
    bpf_probe_read(&ops, sizeof(ops),
                   (__u8 *)skt + offset_socket_sk + sizeof(void *));
    if (ops == NULL)
    {
        return -1;
    }

    bpf_probe_read(sktype, sizeof(short), &skt->type);

    bpf_probe_read(sk, sizeof(struct sock *), (__u8 *)skt + offset_socket_sk);

    return 0;
}

static __always_inline void init_ssl_sockfd(void *ssl_ctx, __u32 fd)
{
    bpf_map_update_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx, &fd, BPF_ANY);
}

static __always_inline bool conn_info_from_skt(
    struct socket *skt, struct connection_info *conn, __u64 pid_tgid)
{
    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0)
    {
        return false;
    }

    // tcp only
    switch (sktype)
    {
    case SOCK_STREAM:
        break;
    default:
        return false;
    }

    if (read_connection_info(sk, conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return false;
    }

    return true;
}

static __always_inline netwrk_data_t *get_netwrk_data_percpu()
{
    __s32 index = 0;
    netwrk_data_t *data = bpf_map_lookup_elem(&mp_netwrk_data_pool, &index);
    if (data == NULL)
    {
        return NULL;
    }
    __builtin_memset(&data->meta, 0, sizeof(data->meta));
    bpf_get_current_comm(&data->meta.comm, KERNEL_TASK_COMM_LEN);

    return data;
}

static __always_inline bool put_rw_args(tp_syscall_rw_args_t *ctx, void *bpf_map, enum MSG_RW rw)
{
    if ((ctx == NULL) || ctx->fd < 3)
    {
        return false;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return false;
    }

    syscall_rw_arg_t arg = {
        .buf = ctx->buf,
        .fd = ctx->fd,
        .skt = skt,
        .ts = bpf_ktime_get_ns(),
    };

    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0 || sktype != SOCK_STREAM)
    {
        return false;
    }

    if (proto_filter(pid_tgid, sk))
    {
        return false;
    }

    switch (rw)
    {
    case MSG_READ:
        arg.tcp_seq = read_copied_seq(sk);
        break;
    case MSG_WRITE:
        arg.tcp_seq = read_write_seq(sk);
        break;
    }

    bpf_map_update_elem(bpf_map, &pid_tgid, &arg, BPF_ANY);

    return true;
}

static __always_inline syscall_rw_arg_t *get_rw_args(void *bpf_map, __u64 *key)
{
    syscall_rw_arg_t *arg = (syscall_rw_arg_t *)bpf_map_lookup_elem(
        bpf_map, key);

    if (arg == NULL || arg->fd <= 2) // fd 0-2: stdin, stdout, stderr
    {
        return NULL;
    }

    return arg;
}

static __always_inline void del_rw_args(void *bpf_map, __u64 *key)
{
    bpf_map_delete_elem(bpf_map, key);
}

static __always_inline bool put_rw_v_args(tp_syscall_rw_v_args_t *ctx, void *bpf_map, enum MSG_RW rw)
{
    if ((ctx == NULL) || ctx->fd < 3)
    {
        return false;
    }
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct task_struct *task = bpf_get_current_task();

    struct socket *skt = get_socket_from_fd(task, ctx->fd);
    if (skt == NULL)
    {
        return false;
    }

    syscall_rw_v_arg_t arg = {
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
        return false;
    }

    if (proto_filter(pid_tgid, sk))
    {
        return false;
    }

    switch (rw)
    {
    case MSG_READ:
        arg.tcp_seq = read_copied_seq(sk);
        break;
    case MSG_WRITE:
        arg.tcp_seq = read_write_seq(sk);
        break;
    }

    bpf_map_update_elem(bpf_map, &pid_tgid, &arg, BPF_ANY);

    return true;
}

static __always_inline syscall_rw_v_arg_t *get_rw_v_args(void *bpf_map, __u64 *key)
{
    syscall_rw_v_arg_t *arg = (syscall_rw_v_arg_t *)bpf_map_lookup_elem(
        bpf_map, key);

    if (arg == NULL || arg->fd <= 2) // fd 0-2: stdin, stdout, stderr
    {
        return NULL;
    }

    return arg;
}

static __always_inline void del_rw_v_args(void *bpf_map, __u64 *key)
{
    bpf_map_delete_elem(bpf_map, key);
}

static __always_inline void read_rw_data(tp_syscall_exit_args_t *ctx, void *bpf_map, tp_syscalls_fn_t fn)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    if (ctx->ret <= 0)
    {
        goto cleanup;
    }

    syscall_rw_arg_t *rw_args = get_rw_args(bpf_map, &pid_tgid);
    if (rw_args == NULL)
    {
        goto cleanup;
    }

    netwrk_data_t *dst = get_netwrk_data_percpu();
    if (dst == NULL)
    {
        goto cleanup;
    }

    read_net_meta(rw_args, pid_tgid, dst);
    read_netwrk_data(dst, rw_args->buf, ctx->ret);

    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &mp_upload_netwrk_data, cpu, dst, sizeof(netwrk_data_t));

#ifdef __DKE_DEBUG_RW__
    bpf_printk("act len: %d %d\n", dst->meta.act_size, ctx->ret);
    bpf_printk("fn: %d, len %d, data: %s\n", fn, dst->meta.buf_len, dst->payload);
#endif

cleanup:
    del_rw_args(bpf_map, &pid_tgid);
}

static __always_inline void read_rw_v_data(tp_syscall_exit_args_t *ctx, void *bpf_map, tp_syscalls_fn_t fn)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    if (ctx->ret <= 0)
    {
        goto cleanup;
    }

    syscall_rw_v_arg_t *rwv_args = get_rw_v_args(bpf_map, &pid_tgid);
    if (rwv_args == NULL)
    {
        goto cleanup;
    }

    __u64 vlen = rwv_args->vlen;
    if (vlen == 0)
    {
        goto cleanup;
    }

    netwrk_data_t *dst = get_netwrk_data_percpu();
    if (dst == NULL)
    {
        goto cleanup;
    }

    read_net_meta(rwv_args, pid_tgid, dst);
    read_network_data_from_vec(dst, rwv_args->vec, vlen, ctx->ret);

    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &mp_upload_netwrk_data, cpu, dst, sizeof(netwrk_data_t));

#ifdef __DKE_DEBUG_RW_V__
    bpf_printk("act len: %d %d\n", dst->meta.act_size, ctx->ret);
    bpf_printk("fn: %d, len %d, data: %s\n", fn, dst->meta.buf_len, dst->payload);
#endif

cleanup:
    del_rw_v_args(bpf_map, &pid_tgid);
}

#endif // !__L7_UTILS_
