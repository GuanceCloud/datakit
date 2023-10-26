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
#include "print_apiflow.h"

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

static __always_inline void rec_seq(struct layer7_http *stats, __u32 copied_seq, __u32 write_seq, req_resp_t req_resp, enum MSG_RW rw)
{
    if (req_resp == HTTP_REQ_REQ)
    {
        if (rw == MSG_READ)
        { // server
            stats->req_seq = copied_seq;
        }
        else if (rw == MSG_WRITE)
        {
            stats->req_seq = write_seq;
        }
    }
    else if (req_resp == HTTP_REQ_RESP)
    {
        if (rw == MSG_READ)
        { // client
            stats->resp_seq = copied_seq;
        }
        else if (rw == MSG_WRITE)
        {
            stats->resp_seq = write_seq;
        }
    }
}

// ret 0: r, 1: w
static __always_inline int vfs_r_or_w(k_u_func_t f)
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

static __always_inline int p_group_eq(k_u_func_t src, k_u_func_t dst)
{
    int s = vfs_r_or_w(src);
    int d = vfs_r_or_w(dst);
    if (s == d)
    {
        return 1;
    }
    return 0;
}

// modifying this macro requires modifying the function copy_data_from_iovec at the same time
#define BUF_IOVEC_LEN (2 << (L7_BUFFER_LEFT_SHIFT - 1))
#define BUF_IOVEC_LEN_MASK (BUF_IOVEC_LEN - 1)

struct buf_iovec
{
    // we need to divide a large buffer into several small pieces
    __u8 data[BUF_IOVEC_LEN];
};

static __always_inline int copy_data_from_iovec(struct iovec *vec, __u64 vlen,
                                                struct payload_id *id,
                                                struct l7_buffer *to_buf)
{
    to_buf->len = 0;

    __s32 count = 0;
#pragma unroll
    for (int i = 0; i < 10; i++)
    {
        if (i > vlen)
        {
            break;
        }
        struct iovec v = {0};
        bpf_probe_read(&v, sizeof(v), vec + i);
        int iov_len = v.iov_len;
        if (iov_len > 0)
        {
            count &= L7_BUFFER_SIZE_MASK;
            if (count + BUF_IOVEC_LEN > sizeof(to_buf->payload))
            {
                break;
            }

            struct buf_iovec *buf = (struct buf_iovec *)((__u8 *)to_buf->payload + count);

            if (v.iov_len > BUF_IOVEC_LEN)
            {
                bpf_probe_read(buf->data, BUF_IOVEC_LEN, v.iov_base);
                count += BUF_IOVEC_LEN;

                int diff = v.iov_len - BUF_IOVEC_LEN;
                if (diff < 0)
                {
                    break;
                }

                if (diff > BUF_IOVEC_LEN)
                {
                    diff = BUF_IOVEC_LEN;
                }
                else
                {
                    diff &= BUF_IOVEC_LEN_MASK;
                }

                if (count + BUF_IOVEC_LEN > sizeof(to_buf->payload))
                {
                    break;
                }

                buf = (struct buf_iovec *)((__u8 *)to_buf->payload + count);

                bpf_probe_read(buf->data, diff, v.iov_base + BUF_IOVEC_LEN);
                count += diff;
            }
            else
            {
                if (iov_len > BUF_IOVEC_LEN)
                {
                    iov_len = BUF_IOVEC_LEN;
                }
                else
                {
                    iov_len &= BUF_IOVEC_LEN_MASK;
                }
                bpf_probe_read(buf->data, iov_len, v.iov_base);
                count += iov_len;
            }
        }
    }

    to_buf->len = count;

    __builtin_memcpy(&to_buf->id, id, sizeof(struct payload_id));

    return 0;
}

static __always_inline int copy_data_from_buffer(__u8 *buffer, __u64 size,
                                                 struct payload_id *id,
                                                 struct l7_buffer *l7buffer,
                                                 k_u_func_t fn)
{
    int l = 0;
    if (size >= L7_BUFFER_SIZE)
    {
        l = L7_BUFFER_SIZE;
    }
    else
    {
        l = size & L7_BUFFER_SIZE_MASK;
    }

    bpf_probe_read(&l7buffer->payload, l, buffer);

    __builtin_memset(&l7buffer->thr_trace_id, 0, sizeof(struct payload_id));

    __builtin_memset(&l7buffer->cmd, 0, KERNEL_TASK_COMM_LEN);
    bpf_get_current_comm(&l7buffer->cmd, KERNEL_TASK_COMM_LEN);

    l7buffer->len = l;

    __builtin_memcpy(&l7buffer->id, id, sizeof(struct payload_id));

    return 0;
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

static __always_inline req_resp_t
parse_layer7_http1(__u8 *buffer, struct layer7_http *stats)
{
    switch (buffer[0])
    {
    case 'G':
        if (buffer[1] == 'E' && buffer[2] == 'T' && buffer[3] == ' ') // HTTP GET
        {
            stats->method = HTTP_METHOD_GET;
            return HTTP_REQ_REQ;
        }
        break;
    case 'P':
        switch (buffer[1])
        {
        case 'O':
            if (buffer[2] == 'S' && buffer[3] == 'T' &&
                buffer[4] == ' ') // HTTP POST
            {
                stats->method = HTTP_METHOD_POST;
                return HTTP_REQ_REQ;
            }
            break;
        case 'U':
            if (buffer[2] == 'T' && buffer[3] == ' ') // HTTP PUT
            {
                stats->method = HTTP_METHOD_PUT;
                return HTTP_REQ_REQ;
            }
            break;
        case 'A':
            if (buffer[2] == 'T' && buffer[3] == 'C' && buffer[4] == 'H' &&
                buffer[5] == ' ') // HTTP PATCH
            {
                stats->method = HTTP_METHOD_PATCH;
                return HTTP_REQ_REQ;
            }
            break;
        default:
            break;
        }
    case 'D':
        if (buffer[1] == 'E' && buffer[2] == 'L' && buffer[3] == 'E' &&
            buffer[4] == 'T' && buffer[5] == 'E' &&
            buffer[6] == ' ') // HTTP DELETE
        {
            stats->method = HTTP_METHOD_DELETE;
            return HTTP_REQ_REQ;
        }
        break;
    case 'H':
        if (buffer[1] == 'T' && buffer[2] == 'T' &&
            buffer[3] == 'P') // response payload
        {
            goto HTTPRESPONSE;
        }
        else if (buffer[1] == 'E' && buffer[2] == 'A' && buffer[3] == 'D' &&
                 buffer[4] == ' ') // HTTP HEAD
        {
            stats->method = HTTP_METHOD_HEAD;
            return HTTP_REQ_REQ;
        }
        break;
    case 'O':
        if (buffer[1] == 'P' && buffer[2] == 'T' && buffer[3] == 'I' &&
            buffer[4] == 'O' && buffer[5] == 'N' && buffer[6] == 'S' &&
            buffer[7] == ' ') // HTTP OPTIONS
        {
            stats->method = HTTP_METHOD_OPTIONS;
            return HTTP_REQ_REQ;
        }
        break;
    // case 'C':
    // if (buffer[1] == 'O' && buffer[2] == 'N' && buffer[3] == 'N' &&
    // buffer[4] == 'E' && buffer[5] == 'C' && buffer[6] == 'T') // HTTP
    // CONNECTION
    // {
    // l7http->method = HTTP_METHOD_CONNECT;
    // return HTTP_REQ_REQ;
    // }
    // break;
    // case 'T':
    // if (buffer[1] == 'R' && buffer[2] == 'A' && buffer[3] == 'C' && buffer[4]
    // == 'E') { // HTTP TRACE l7http->method = HTTP_METHOD_TRACE; return
    // HTTP_REQ_REQ;
    // }
    // break;
    default:
        break;
    }

    return HTTP_REQ_UNKNOWN;

HTTPRESPONSE:
    if (buffer[4] != '/' || buffer[6] != '.' || buffer[8] != ' ')
    {
        return HTTP_REQ_UNKNOWN;
    }
    stats->http_version = ((buffer[5] - '0') << 16) + (buffer[7] - '0');
    stats->status_code =
        (buffer[9] - '0') * 100 + (buffer[10] - '0') * 10 + (buffer[11] - '0');
    return HTTP_REQ_RESP;
}

static __always_inline void init_ssl_sockfd(void *ssl_ctx, __u32 fd)
{
    bpf_map_update_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx, &fd, BPF_ANY);
}

static __always_inline int record_http_req(void *ctx,
                                           struct connection_info *conn,
                                           struct layer7_http *stats,
                                           struct l7_buffer *l7buffer,
                                           enum MSG_RW rw, k_u_func_t ku_fn)
{
    // set payload id
    __builtin_memcpy(&l7buffer->id, &stats->req_payload_id,
                     sizeof(struct payload_id));

    // payload used
    l7buffer->req_ts = stats->req_ts;

    switch (rw)
    {
    case MSG_READ:
        stats->direction = CONN_DIRECTION_INCOMING;
        break;
    case MSG_WRITE:
        stats->direction = CONN_DIRECTION_OUTGOING;
        do_dnapt(conn, stats->nat_daddr, &stats->nat_dport);
        break;
    }

    stats->req_func = ku_fn;
    // stats->span_id = bpf_get_prandom_u32() << 32 | bpf_get_prandom_u32();

    // send data to datakit-ebpf agent
    __u64 cpu = bpf_get_smp_processor_id();

    bpf_perf_event_output(ctx, &bpfmap_l7_buffer_out, cpu,
                          l7buffer, sizeof(struct l7_buffer));

    bpf_map_update_elem(&bpfmap_http_stats, conn, stats, BPF_NOEXIST);

    return 0;
}

static __always_inline int record_http_resp(void *ctx,
                                            struct connection_info *conn,
                                            struct layer7_http *stats,
                                            enum MSG_RW rw, k_u_func_t ku_fn)
{
    struct layer7_http *http_info =
        bpf_map_lookup_elem(&bpfmap_http_stats, conn);
    if (http_info == NULL)
    {
        return 0;
    }

    http_info->resp_seq = stats->resp_seq;
    http_info->status_code = stats->status_code;
    http_info->http_version = stats->http_version;
    http_info->resp_ts = stats->resp_ts;
    http_info->resp_func = ku_fn;

    return 0;
}

static __always_inline int http_try_upload_keep_alive_fin(void *ctx,
                                                          struct connection_info *conn, k_u_func_t ku_fn)
{
    struct layer7_http *http_info =
        bpf_map_lookup_elem(&bpfmap_http_stats, conn);

    if (http_info == NULL || http_info->resp_func == P_UNKNOWN)
    {
        return 0;
    }

    if (p_group_eq(http_info->resp_func, ku_fn) != 1)
    {
        struct http_req_finished http_finished = {0};

        __builtin_memcpy(&http_finished.conn_info, conn,
                         sizeof(struct connection_info));
        __builtin_memcpy(&http_finished.http, http_info,
                         sizeof(struct layer7_http));

        // delete cached stats
        bpf_map_delete_elem(&bpfmap_http_stats, conn);

        // bpf_printk("finished req %u resp %u", http_finished.http.req_seq, http_finished.http.rsp_seq);

        __u64 cpu = bpf_get_smp_processor_id();
        bpf_perf_event_output(ctx, &bpfmap_httpreq_fin_event, cpu, &http_finished,
                              sizeof(struct http_req_finished));
    }
    return 0;
}

static __always_inline int http_try_upload(void *ctx, struct connection_info *conn,
                                           int r, int w, __u64 start, __u64 end, k_u_func_t ku_fn)
{
    struct layer7_http *http_info =
        bpf_map_lookup_elem(&bpfmap_http_stats, conn);

    if (http_info == NULL)
    {
        return 0;
    }

    // if you want to handle the conn for http keep-alive here, don't do it
    if (r > 0)
    {
        __sync_fetch_and_add(&http_info->recv_bytes, r);
    }
    else if (w > 0)
    {
        __sync_fetch_and_add(&http_info->sent_bytes, w);
    }

    // http request only
    if (http_info->resp_func == P_UNKNOWN)
    {

        if (ku_fn == P_SYSCALL_CLOSE) // no response, but the connection is closed
        {
            bpf_map_delete_elem(&bpfmap_http_stats, conn);
        }

        return 0;
    }

    // a response, but the connection is not closed
    if (ku_fn != P_SYSCALL_CLOSE)
    {
        // for the http protocol, the connection may be kept alive if the func groups are not equal
        if (p_group_eq(http_info->resp_func, ku_fn) == 1)
        {
            http_info->resp_ts = end;
            return 0;
        }
        else
        {
            // We have to do something ahead, see function http_try_upload_keep_alive_fin.
            // Upload request-response data, if function groups are not equal.
            // Do not update time and byte count.
            return 0;
        }
    }

    // tcp closed

    struct http_req_finished http_finished = {0};

    __builtin_memcpy(&http_finished.conn_info, conn,
                     sizeof(struct connection_info));
    __builtin_memcpy(&http_finished.http, http_info,
                     sizeof(struct layer7_http));

    // delete cached stats
    bpf_map_delete_elem(&bpfmap_http_stats, conn);

    // bpf_printk("finished req %u resp %u", http_finished.http.req_seq, http_finished.http.rsp_seq);
    __u64 cpu = bpf_get_smp_processor_id();
    bpf_perf_event_output(ctx, &bpfmap_httpreq_fin_event, cpu, &http_finished,
                          sizeof(struct http_req_finished));

    return 0;
}

static __always_inline req_resp_t checkHTTP(struct socket *skt, __u8 *buf,
                                            struct connection_info *conn,
                                            struct layer7_http *stats,
                                            __s64 buf_size)
{
    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0)
    {
        return HTTP_REQ_UNKNOWN;
    }

    // tcp only
    switch (sktype)
    {
    case SOCK_STREAM:
        break;
    default:
        return HTTP_REQ_UNKNOWN;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    if (read_connection_info(sk, conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return HTTP_REQ_UNKNOWN;
    }

    // skip https
    if (conn->sport == 443 || conn->dport == 443)
    {
        return HTTP_REQ_UNKNOWN;
    }

    // bpf_printk("r byte: %d, r %s", buf_size, tmp_buf);

    __u8 tmp_buffer[32] = {0};
    int size_copy = 0;
    if (buf_size > 31)
    {
        size_copy = 31;
    }
    else
    {
        size_copy = buf_size & 0x1F;
    }

    bpf_probe_read(&tmp_buffer, size_copy, buf);

    // Determine request/response and whether it is a server.
    return parse_layer7_http1(tmp_buffer, stats);
}

static __always_inline req_resp_t checkHTTPS(struct socket *skt, __u8 *buf,
                                             struct connection_info *conn,
                                             struct layer7_http *stats,
                                             __s64 buf_size)
{
    struct sock *sk = NULL;
    enum sock_type sktype = 0;

    if (get_sock_from_skt(skt, &sk, &sktype) != 0)
    {
        return HTTP_REQ_UNKNOWN;
    }

    // tcp only
    switch (sktype)
    {
    case SOCK_STREAM:
        break;
    default:
        return HTTP_REQ_UNKNOWN;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    if (read_connection_info(sk, conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return HTTP_REQ_UNKNOWN;
    }

    // bpf_printk("r byte: %d, r %s", buf_size, tmp_buf);

    __u8 tmp_buffer[32] = {0};

    int size_copy = 0;
    if (buf_size > 31)
    {
        size_copy = 31;
    }
    else
    {
        size_copy = buf_size & 0x1F;
    }
    bpf_probe_read(&tmp_buffer, size_copy, buf);

    // Determine request/response and whether it is a server.
    return parse_layer7_http1(tmp_buffer, stats);
}

static __always_inline void upate_req_payload_id(struct layer7_http *stats,
                                                 __u64 pid_tgid, __u64 k_time)
{
    stats->req_payload_id.ktime = k_time;
    stats->req_payload_id.pid_tid = pid_tgid;
    stats->req_payload_id.prandomhalf = bpf_get_prandom_u32() >> 16;
    stats->req_payload_id.cpuid = bpf_get_smp_processor_id();
    stats->req_payload_id.prandom = bpf_get_prandom_u32();
}

static __always_inline int parse_http1x(void *ctx, struct l7_buffer *l7buffer,
                                        __u64 k_time, struct connection_info *conn,
                                        struct layer7_http *stats, req_resp_t req_type,
                                        enum MSG_RW rw, k_u_func_t ku_fn, int fd)
{
    __u64 pid_tgid = l7buffer->id.pid_tid;

    pidfd_t pfd = {
        .fd = fd,
        .pid = pid_tgid >> 32,
    };

    // Determine request/response and whether it is a server.
    switch (req_type)
    {
    case HTTP_REQ_REQ:
        http_try_upload_keep_alive_fin(ctx, conn, ku_fn);

        __u64 pid_tig_goid = pid_tgid;
        __u64 *goid = bpf_map_lookup_elem(&bmap_tid2goid, &pid_tgid);
        if (goid != NULL)
        {
            pid_tig_goid = *goid;
            // bpf_printk("insert tid_goid %d, tid %d  %d", (__u32)pid_tig_goid, (__u32)pid_tgid, conn->sport);
        }

        struct payload_id *thrid = bpf_map_lookup_elem(&bpfmap__pidtidgoid_thr_traceid, &pid_tig_goid);
        if (thrid == NULL)
        {
            // bpf_printk("req fd %d, pid %d, tid %d", fd, pid_tgid >> 32, (__u32)pid_tgid);
            __builtin_memcpy(&l7buffer->thr_trace_id, &l7buffer->id, sizeof(struct payload_id));
            l7buffer->thr_trace_id.pid_tid = pid_tig_goid;

            l7buffer->isentry = 1;
            __u32 tid_goid = (__u32)pid_tig_goid;
            bpf_map_update_elem(&bpfmap_pidfd_tidgoid, &pfd, &tid_goid, BPF_NOEXIST);
            struct payload_id thrid_tmp = {0};
            __builtin_memcpy(&thrid_tmp, &l7buffer->thr_trace_id, sizeof(struct payload_id));

            bpf_map_update_elem(&bpfmap__pidtidgoid_thr_traceid, &pid_tig_goid, &thrid_tmp, BPF_NOEXIST);
        }
        else
        {
            // bpf_printk("req x fd %d, pid %d, tid %d", fd, pid_tgid >> 32, (__u32)pid_tgid);
            __builtin_memcpy(&l7buffer->thr_trace_id, thrid, sizeof(struct payload_id));
        }

        stats->req_ts = k_time;
        record_http_req(ctx, conn, stats, l7buffer, rw, ku_fn);
        break;

    case HTTP_REQ_RESP:;
        __u32 *tid_goid = bpf_map_lookup_elem(&bpfmap_pidfd_tidgoid, &pfd);
        if (tid_goid != NULL)
        {
            // if (*tid_goid < 100)
            // {
            //     bpf_printk("del port %d tid %d goid %d", conn->sport, (__u32)pid_tgid, *tid_goid);
            // }
            bpf_map_delete_elem(&bpfmap_pidfd_tidgoid, &pfd);
            pidtid_t t = {
                .tid = *tid_goid,
                .pid = pfd.pid,
            };
            bpf_map_delete_elem(&bpfmap__pidtidgoid_thr_traceid, &t);
        }

        stats->resp_ts = bpf_ktime_get_ns();
        record_http_resp(ctx, conn, stats, rw, ku_fn);
        break;
    default:
        return -1;
    }
    return 0;
}

// static __always_inline int process_proto(void *ctx)
#endif // !__L7_UTILS_
