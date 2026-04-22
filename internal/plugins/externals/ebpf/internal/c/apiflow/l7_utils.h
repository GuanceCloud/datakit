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
#include "../process_sched/goidtid.h"
#include "bpfmap_l7.h"

#include "bpf_helpers.h"
#include "l7_stats.h"
#include "offset_read.h"
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

static __always_inline void get_uni_id(id_generator_t *dst)
{
    __u32 index = 0;
    id_generator_t *val = NULL;
    val = (id_generator_t *)bpf_map_lookup_elem(&mp_uni_id_per_cpu, &index);
    if (val == NULL)
    {
        return;
    }

    // initialize
    if (val->init == 0)
    {
        __u16 cpu_id = (__u16)bpf_get_smp_processor_id();
        val->cpu_id = cpu_id;
        val->init = 1;
    }

    __u64 ktime = bpf_ktime_get_ns();
    val->ktime = ktime;

    val->id++;

    __builtin_memcpy(dst, val, sizeof(id_generator_t));
}

static __always_inline __u8 get_sk_inf(void *sk, sk_inf_t *dst, __u8 force)
{
    sk_inf_t *inf = NULL;

    inf = (sk_inf_t *)bpf_map_lookup_elem(&mp_sk_inf, &sk);
    if (inf == NULL)
    {
        if (force != 0)
        {
            sk_inf_t i = {0};
            i.index = 1;
            get_uni_id(&i.uni_id);
            __u64 pid_tgid = bpf_get_current_pid_tgid();

            if (read_connection_info(sk, &i.conn, pid_tgid, CONN_L4_TCP) != 0)
            {
                return 0;
            }

            i.skptr = (u64)sk;
            // May fail due to exceeding the upper limit of the number of elements
            bpf_map_update_elem(&mp_sk_inf, &sk, &i, BPF_NOEXIST);

            inf = (sk_inf_t *)bpf_map_lookup_elem(&mp_sk_inf, &sk);
            if (inf != NULL)
            {
                __u8 val = 0;
                bpf_map_update_elem(&mp_protocol_filter, &sk, &val, BPF_NOEXIST);
            }
        }
    }

    if (inf != NULL)
    {
        __builtin_memcpy(dst, inf, sizeof(sk_inf_t));
        __sync_fetch_and_add(&inf->index, 1);
        return 1;
    }

    return 0;
}

static __always_inline void del_sk_inf(void *sk)
{
    bpf_map_delete_elem(&mp_sk_inf, &sk);
}

// args: syscall_rw_arg_t, syscall_rw_v_arg_t; dst: netwrk_data_t
#define read_net_meta(args, pid_tgid, dst)                            \
    do                                                                \
    {                                                                 \
        __u64 ts = bpf_ktime_get_ns();                                \
        struct sock *sk = NULL;                                       \
        if (!get_sk_with_typ(args->skt, &sk, SOCK_STREAM))            \
        {                                                             \
            goto cleanup;                                             \
        };                                                            \
        __u8 found = get_sk_inf(sk, &dst->meta.sk_inf, 1);            \
        if (found == 0)                                               \
        {                                                             \
            goto cleanup;                                             \
        }                                                             \
        dst->meta.ts = args->ts;                                      \
        dst->meta.ts_tail = ts;                                       \
        dst->meta.tid_utid = pid_tgid << 32;                          \
        dst->meta.tcp_seq = args->tcp_seq;                            \
        dst->meta.func_id = fn;                                       \
        dst->meta.original_size = ctx->ret;                           \
                                                                      \
        __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid); \
        if (goid != NULL)                                             \
        {                                                             \
            dst->meta.tid_utid |= *goid;                              \
        }                                                             \
    } while (0)

static __always_inline bool net_filtered(__u64 pid_tgid, void *sock_ptr)
{
    u32 pid = pid_tgid >> 32;
    if (need_filter_proc(&pid))
    {
        return false;
    }

    __u8 *val = bpf_map_lookup_elem(&mp_protocol_filter, &sock_ptr);
    if (val != NULL && *val == 1)
    {
        return false;
    }

    return true;
}

static __always_inline void clean_protocol_filter(void *sock_ptr)
{
    bpf_map_delete_elem(&mp_protocol_filter, &sock_ptr);
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

static __always_inline bool is_syscall_net_event(__u16 fn)
{
    switch (fn)
    {
    case P_SYSCALL_WRITE:
    case P_SYSCALL_READ:
    case P_SYSCALL_SENDTO:
    case P_SYSCALL_RECVFROM:
    case P_SYSCALL_WRITEV:
    case P_SYSCALL_READV:
    case P_SYSCALL_SENDFILE:
        return true;
    default:
        return false;
    }
}

struct buf_iovec
{
    // we need to divide a large buffer into several small pieces
    __u8 data[BUF_IOVEC_LEN];
};

enum
{
    BOUNDED_L7_CAPTURE_SIZE = 256,
};

static __always_inline __u32 bounded_capture_bucket(__s64 size)
{
    if (size <= 0)
    {
        return 0;
    }

    if (size >= BOUNDED_L7_CAPTURE_SIZE)
    {
        return BOUNDED_L7_CAPTURE_SIZE;
    }
    if (size >= 128)
    {
        return 128;
    }
    if (size >= 64)
    {
        return 64;
    }
    if (size >= 32)
    {
        return 32;
    }
    if (size >= 16)
    {
        return 16;
    }
    if (size >= 8)
    {
        return 8;
    }
    if (size >= 4)
    {
        return 4;
    }
    if (size >= 2)
    {
        return 2;
    }
    return 1;
}

#ifdef DK_LEGACY_APIFLOW_MINIMAL
static __always_inline __u32 legacy_capture_bucket(__s64 size)
{
    return bounded_capture_bucket(size);
}
#endif

#define LEGACY_COPY_PAYLOAD_CASES(dst, src, capture_size, done_stmt) \
    do                                                                \
    {                                                                 \
        if ((capture_size) == BOUNDED_L7_CAPTURE_SIZE)                \
        {                                                             \
            bpf_probe_read((dst), 256, (src));                        \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 128)                                    \
        {                                                             \
            bpf_probe_read((dst), 128, (src));                        \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 64)                                     \
        {                                                             \
            bpf_probe_read((dst), 64, (src));                         \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 32)                                     \
        {                                                             \
            bpf_probe_read((dst), 32, (src));                         \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 16)                                     \
        {                                                             \
            bpf_probe_read((dst), 16, (src));                         \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 8)                                      \
        {                                                             \
            bpf_probe_read((dst), 8, (src));                          \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 4)                                      \
        {                                                             \
            bpf_probe_read((dst), 4, (src));                          \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 2)                                      \
        {                                                             \
            bpf_probe_read((dst), 2, (src));                          \
            done_stmt;                                                \
        }                                                             \
        if ((capture_size) == 1)                                      \
        {                                                             \
            bpf_probe_read((dst), 1, (src));                          \
            done_stmt;                                                \
        }                                                             \
    } while (0)

#define LEGACY_MEMCPY_PAYLOAD_CASES(dst, src, capture_size, done_stmt) \
    do                                                                 \
    {                                                                  \
        if ((capture_size) == BOUNDED_L7_CAPTURE_SIZE)                 \
        {                                                              \
            __builtin_memcpy((dst), (src), 256);                       \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 128)                                     \
        {                                                              \
            __builtin_memcpy((dst), (src), 128);                       \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 64)                                      \
        {                                                              \
            __builtin_memcpy((dst), (src), 64);                        \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 32)                                      \
        {                                                              \
            __builtin_memcpy((dst), (src), 32);                        \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 16)                                      \
        {                                                              \
            __builtin_memcpy((dst), (src), 16);                        \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 8)                                       \
        {                                                              \
            __builtin_memcpy((dst), (src), 8);                         \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 4)                                       \
        {                                                              \
            __builtin_memcpy((dst), (src), 4);                         \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 2)                                       \
        {                                                              \
            __builtin_memcpy((dst), (src), 2);                         \
            done_stmt;                                                 \
        }                                                              \
        if ((capture_size) == 1)                                       \
        {                                                              \
            __builtin_memcpy((dst), (src), 1);                         \
            done_stmt;                                                 \
        }                                                              \
    } while (0)

static __always_inline void read_network_data_from_vec(net_data_t *dst, struct iovec *vec,
                                                       __u64 vlen, __s64 len_or_errno)
{
#ifdef DK_LEGACY_APIFLOW_MINIMAL
    (void)vec;
    (void)vlen;
    (void)len_or_errno;
    dst->meta.capture_size = 0;
    return;
#else
    const __s64 payload_limit = sizeof(dst->payload);

    if (len_or_errno <= 0)
    {
        dst->meta.capture_size = 0;
        return;
    }

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
            __s64 capture_len = (__s64)iov_len;
            __u32 capture_size = 0;
            if (capture_len <= 0)
            {
                continue;
            }
            if (capture_len > len_or_errno)
            {
                capture_len = len_or_errno;
            }
            if (capture_len > payload_limit)
            {
                capture_len = payload_limit;
            }
            capture_size = bounded_capture_bucket(capture_len);
            if (capture_size == 0)
            {
                continue;
            }
            LEGACY_COPY_PAYLOAD_CASES(dst->payload, (__u8 *)v.iov_base, capture_size,
                                      {
                                          dst->meta.capture_size = capture_size;
                                          return;
                                      });
        }
    }

    dst->meta.capture_size = 0;
#endif
}

static __always_inline void read_netwrk_data(net_data_t *dst, __u8 *buf, __s64 len_or_errno)
{
#ifdef DK_LEGACY_APIFLOW_MINIMAL
    __u32 capture_size = legacy_capture_bucket(len_or_errno);
    if (capture_size == 0)
    {
        dst->meta.capture_size = 0;
        return;
    }

    LEGACY_COPY_PAYLOAD_CASES(dst->payload, buf, capture_size,
                              {
                                  dst->meta.capture_size = capture_size;
                                  return;
                              });
    dst->meta.capture_size = 0;
#else
    __u32 capture_size = bounded_capture_bucket(len_or_errno);
    if (capture_size == 0)
    {
        dst->meta.capture_size = 0;
        return;
    }

    LEGACY_COPY_PAYLOAD_CASES(dst->payload, buf, capture_size,
                              {
                                  dst->meta.capture_size = capture_size;
                                  return;
                              });
    dst->meta.capture_size = 0;
#endif
}

static __always_inline struct socket *get_socket_from_fd(
    struct task_struct *task, int fd)
{
    if (task == NULL || fd < 0)
    {
        return NULL;
    }

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

    unsigned int max_fds = 0;
    bpf_probe_read(&max_fds, sizeof(max_fds), &fdt->max_fds);
    if ((__u32)fd >= max_fds)
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

    DK_READ_VALUE(skt, skfile, offset); // bpf_probe_read(&skt, sizeof(skt), &skfile->private_data);
    if (skt == NULL)
    {
        return NULL;
    }

    // check is socket
    struct file *file_addr = NULL;
    offset = load_offset_socket_file();
    DK_READ_VALUE(file_addr, skt, offset); // bpf_probe_read(&file_addr,
                                           // sizeof(file_addr), &skt->file);
    if (file_addr != skfile)
    {
        return NULL;
    }

    return skt;
}

static __always_inline int get_sk(struct socket *skt,
                                  struct sock **sk,
                                  enum sock_type *sktype)
{
    __u64 offset_socket_sk = load_offset_socket_sk();
    if (offset_socket_sk == 0)
    {
        return -1;
    }

    struct proto_ops *ops = NULL;
    DK_READ_VALUE(ops, skt, offset_socket_sk + sizeof(void *));
    if (ops == NULL)
    {
        return -1;
    }

    bpf_probe_read(sktype, sizeof(short), &skt->type);

    DK_READ_INTO(sk, skt, offset_socket_sk);

    return 0;
}

static __always_inline void init_ssl_sockfd(void *ssl_ctx, __u32 fd)
{
    bpf_map_update_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx, &fd, BPF_ANY);
}

static __always_inline bool get_sk_with_typ(struct socket *skt, struct sock **sk_ptr, enum sock_type sk_type)
{
    enum sock_type typ = 0;
    bool need_fallback = false;

    if (get_sk(skt, sk_ptr, &typ) != 0 || typ != sk_type)
    {
#ifdef DK_LEGACY_APIFLOW_MINIMAL
        need_fallback = true;
#endif
        if (!need_fallback)
        {
            return false;
        }
    }
    else if (*sk_ptr != NULL)
    {
        return true;
    }

#ifdef DK_LEGACY_APIFLOW_MINIMAL
    conn_inf_t conn = {};
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    bpf_probe_read(&typ, sizeof(short), &skt->type);
    if (typ != sk_type)
    {
        return false;
    }

#pragma unroll
    for (__u32 i = 0; i < 8; i++)
    {
        __u32 socket_sk_offset = i * sizeof(void *);
        struct sock *candidate = NULL;

        bpf_probe_read(&candidate, sizeof(candidate), (__u8 *)skt + socket_sk_offset);
        if (candidate == NULL)
        {
            continue;
        }

        if (read_connection_info(candidate, &conn, pid_tgid, CONN_L4_TCP) != 0)
        {
            continue;
        }

        *sk_ptr = candidate;
        return true;
    }
#endif
    return false;
}

static __always_inline net_data_t *get_net_data_percpu()
{
    __s32 index = 0;
    net_data_t *data = bpf_map_lookup_elem(&mp_network_data_per_cpu, &index);
    if (data == NULL)
    {
        return NULL;
    }
    __builtin_memset(&data->meta, 0, sizeof(data->meta));
    bpf_get_current_comm(&data->meta.comm, KERNEL_TASK_COMM_LEN);

    return data;
}

static __always_inline void try_upload_net_events(void *ctx, net_data_t *data)
{
    network_events_t *events = get_net_events();
    if (events == NULL)
    {
        return;
    }

    __u64 cpu = bpf_get_smp_processor_id();

    int capture_size = data->meta.capture_size;
    if (capture_size < 0)
    {
        capture_size = 0;
    }
    else if (capture_size > L7_BUFFER_SIZE)
    {
        capture_size = L7_BUFFER_SIZE;
    }

    if (data->meta.func_id != P_SYSCALL_CLOSE)
    {
        if (capture_size == 0)
        {
            return;
        }

        if (is_syscall_net_event(data->meta.func_id) &&
            apiflow_min_capture_size > 0 &&
            (__u64)capture_size < apiflow_min_capture_size)
        {
            return;
        }
    }

    net_event_t *net_event = (net_event_t *)(events->payload);
    events->rec.num = 1;
    events->rec.bytes = sizeof(net_event_comm_t) + capture_size;
    if (events->rec.bytes > L7_EVENT_SIZE)
    {
        events->rec.bytes = L7_EVENT_SIZE;
    }

    net_event->event_comm.rec.num = 1;
    net_event->event_comm.rec.bytes = capture_size;
    bpf_probe_read(&net_event->event_comm.meta, sizeof(data->meta), &data->meta);
    if (capture_size > 0)
    {
#ifdef DK_LEGACY_APIFLOW_MINIMAL
        LEGACY_MEMCPY_PAYLOAD_CASES(net_event->payload, data->payload, capture_size,
                                    {
                                        goto payload_copy_done;
                                    });
#else
        bpf_probe_read(&net_event->payload, capture_size, &data->payload);
#endif
payload_copy_done:
        ;
    }

    bpf_perf_event_output(ctx, &mp_upload_netwrk_events, cpu, events, sizeof(network_events_t));
    events->rec.bytes = 0;
    events->rec.num = 0;
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

    if (!get_sk_with_typ(skt, &sk, SOCK_STREAM))
    {
        return false;
    }

    if (!net_filtered(pid_tgid, sk))
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

    if (!get_sk_with_typ(skt, &sk, SOCK_STREAM))
    {
        return false;
    }

    if (!net_filtered(pid_tgid, sk))
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

    net_data_t *dst = get_net_data_percpu();
    if (dst == NULL)
    {
        goto cleanup;
    }

    read_net_meta(rw_args, pid_tgid, dst);

    read_netwrk_data(dst, rw_args->buf, ctx->ret);

    try_upload_net_events(ctx, dst);

#ifdef __DKE_DEBUG_RW__
    bpf_printk("cap len: %d %d\n", dst->meta.capture_size, ctx->ret);
    bpf_printk("fn: %d, len %d, data: %s\n", fn, dst->meta.original_size, dst->payload);
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

    net_data_t *dst = get_net_data_percpu();
    if (dst == NULL)
    {
        goto cleanup;
    }

    read_net_meta(rwv_args, pid_tgid, dst);

    read_network_data_from_vec(dst, rwv_args->vec, vlen, ctx->ret);

    try_upload_net_events(ctx, dst);

#ifdef __DKE_DEBUG_RW_V__
    bpf_printk("cap len: %d %d\n", dst->meta.capture_size, ctx->ret);
    bpf_printk("fn: %d, len %d, data: %s\n", fn, dst->meta.original_size, dst->payload);
#endif

cleanup:
    del_rw_v_args(bpf_map, &pid_tgid);
}

#endif // !__L7_UTILS_
