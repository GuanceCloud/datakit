#ifndef __L7_UTILS_
#define __L7_UTILS_

#define KEEPPACKET -1
#define DROPPACKET 0

#include <linux/socket.h>
#include <linux/fdtable.h>

#include <uapi/linux/if_ether.h>
#include <uapi/linux/in.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/ipv6.h>
#include <uapi/linux/tcp.h>
#include <uapi/linux/udp.h>

#include "bpf_helpers.h"
#include "../netflow/netflow_utils.h"
#include "print_apiflow.h"
#include "l7_stats.h"

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

    // TODO 解析此类 HTTP 数据
    HTTP_METHOD_CONNECT,
    HTTP_METHOD_TRACE
};

static __always_inline int copy_data_from_iovec(struct iovec *vec, __u64 vlen, struct payload_id *id, struct l7_buffer *to_buf)
{
    __s32 count = 0;
#pragma unroll
    for (int i = 0; i < 10; i++)
    {
        if (i >= vlen || count >= sizeof(to_buf))
        {
            break;
        }

        struct iovec v = {0};
        bpf_probe_read(&v, sizeof(v), vec + i);

        count &= L7_BUFFER_SIZE_MASK; // 使其始终为正数

        if (count < sizeof(to_buf))
        {
            // 每次从 +offest 开始拷贝定长数据，长度为 buffer 的总长度，直到写满 buffer
            bpf_probe_read(to_buf->payload + count, L7_BUFFER_SIZE, v.iov_base);
        }

        count += v.iov_len;
    }

    to_buf->len = count;

    __builtin_memcpy(&to_buf->id, id, sizeof(struct payload_id));

    return 0;
}

static __always_inline int copy_data_from_buffer(__u8 *buffer, __u64 size, struct payload_id *id, struct l7_buffer *l7buffer)
{
    bpf_probe_read(l7buffer->payload, L7_BUFFER_SIZE, buffer);

    l7buffer->len = size;

    __builtin_memcpy(&l7buffer->id, id, sizeof(struct payload_id));

    return 0;
}

static __always_inline struct socket *get_socket_from_fd(struct task_struct *task, int fd)
{
    struct files_struct *files = NULL;
    __u64 offset = 0;
    offset = load_offset_task_struct_files();

    bpf_probe_read(&files, sizeof(files), (__u8 *)task + offset); // bpf_probe_read(&files, sizeof(files), &task->files);

    if (files == NULL)
    {
        return NULL;
    }

    struct fdtable *fdt = NULL;
    offset = load_offset_files_struct_fdt();

    bpf_probe_read(&fdt, sizeof(fdt), (__u8 *)files + offset); // bpf_probe_read(&fdt, sizeof(fdt), &files->fdt);

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

    bpf_probe_read(&skt, sizeof(skt), (__u8 *)skfile + offset); // bpf_probe_read(&skt, sizeof(skt), &skfile->private_data);
    if (skt == NULL)
    {
        return NULL;
    }

    // check is socket
    struct file *file_addr = NULL;
    offset = load_offset_socket_file();
    bpf_probe_read(&file_addr, sizeof(file_addr), (__u8 *)skt + offset); // bpf_probe_read(&file_addr, sizeof(file_addr), &skt->file);
    if (file_addr != skfile)
    {
        return NULL;
    }

    return skt;
}

static __always_inline int get_sock_from_skt(struct socket *skt, struct sock **sk, enum sock_type *sktype)
{
    __u64 offset_socket_sk = load_offset_socket_sk();

    struct proto_ops *ops = NULL;
    bpf_probe_read(&ops, sizeof(ops), (__u8 *)skt + offset_socket_sk + sizeof(void *));
    if (ops == NULL)
    {
        return -1;
    }

    bpf_probe_read(sktype, sizeof(short), &skt->type);

    bpf_probe_read(sk, sizeof(struct sock *), (__u8 *)skt + offset_socket_sk);

    return 0;
}

static __always_inline req_resp_t parse_layer7_http1(__u8 *buffer, struct layer7_http *stats)
{
    switch (buffer[0])
    {
    case 'G':
        if (buffer[1] == 'E' && buffer[2] == 'T') // HTTP GET
        {
            stats->method = HTTP_METHOD_GET;
            return HTTP_REQ_REQ;
        }
        break;
    case 'P':
        switch (buffer[1])
        {
        case 'O':
            if (buffer[2] == 'S' && buffer[3] == 'T') // HTTP POST
            {
                stats->method = HTTP_METHOD_POST;
                return HTTP_REQ_REQ;
            }
            break;
        case 'U':
            if (buffer[2] == 'T') // HTTP PUT
            {
                stats->method = HTTP_METHOD_PUT;
                return HTTP_REQ_REQ;
            }
            break;
        case 'A':
            if (buffer[2] == 'T' && buffer[3] == 'C' && buffer[4] == 'H') // HTTP PATCH
            {
                stats->method = HTTP_METHOD_PATCH;
                return HTTP_REQ_REQ;
            }
            break;
        default:
            break;
        }
    case 'D':
        if (buffer[1] == 'E' && buffer[2] == 'L' && buffer[3] == 'E' && buffer[4] == 'T' && buffer[5] == 'E') // HTTP DELETE
        {
            stats->method = HTTP_METHOD_DELETE;
            return HTTP_REQ_REQ;
        }
        break;
    case 'H':
        if (buffer[1] == 'T' && buffer[2] == 'T' && buffer[3] == 'P') // response payload
        {
            stats->status_code = HTTP_REQ_RESP;
            goto HTTPRESPONSE;
        }
        else if (buffer[1] == 'E' && buffer[2] == 'A' && buffer[3] == 'D') // HTTP HEAD
        {
            stats->method = HTTP_METHOD_HEAD;
            return HTTP_REQ_REQ;
        }
        break;
    case 'O':
        if (buffer[1] == 'P' && buffer[2] == 'T' && buffer[3] == 'I' && buffer[4] == 'O' && buffer[5] == 'N' && buffer[6] == 'S') // HTTP OPTIONS
        {
            stats->method = HTTP_METHOD_OPTIONS;
            return HTTP_REQ_REQ;
        }
        break;
    // case 'C':
    // if (buffer[1] == 'O' && buffer[2] == 'N' && buffer[3] == 'N' &&
    // buffer[4] == 'E' && buffer[5] == 'C' && buffer[6] == 'T') // HTTP CONNECTION
    // {
    // l7http->method = HTTP_METHOD_CONNECT;
    // return HTTP_REQ_REQ;
    // }
    // break;
    // case 'T':
    // if (buffer[1] == 'R' && buffer[2] == 'A' && buffer[3] == 'C' && buffer[4] == 'E')
    // { // HTTP TRACE
    // l7http->method = HTTP_METHOD_TRACE;
    // return HTTP_REQ_REQ;
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
    stats->status_code = (buffer[9] - '0') * 100 + (buffer[10] - '0') * 10 + (buffer[11] - '0');
    return HTTP_REQ_RESP;
}

static __always_inline void init_ssl_sockfd(void *ssl_ctx, __u32 fd)
{
    bpf_map_update_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx, &fd, BPF_ANY);
}

static __always_inline int record_http_req(void *ctx, struct connection_info *conn, struct layer7_http *stats,
                                           struct l7_buffer *l7buffer, enum MSG_RW rw)
{
    // set payload id
    __builtin_memcpy(&l7buffer->id, &stats->req_payload_id, sizeof(struct payload_id));

    // payload used
    l7buffer->req_ts = stats->req_ts;

    switch (rw)
    {
    case MSG_READ:
        stats->direction = CONN_DIRECTION_INCOMING;
        break;
    case MSG_WRITE:
        stats->direction = CONN_DIRECTION_OUTGOING;
        break;
    }

    // send data to datakit-ebpf agent
    bpf_perf_event_output(ctx, &bpfmap_l7_buffer_out, bpf_get_smp_processor_id(), l7buffer, sizeof(struct l7_buffer));

    bpf_map_update_elem(&bpfmap_http_stats, conn, stats, BPF_NOEXIST);

    return 0;
}

static __always_inline int record_http_resp(void *ctx, struct connection_info *conn,
                                            struct layer7_http *stats, enum MSG_RW rw)
{

    struct layer7_http *stats_cached = bpf_map_lookup_elem(&bpfmap_http_stats, conn);
    if (stats_cached == NULL)
    {
        return 0;
    }

    struct http_req_finished http_finished = {0};

    __builtin_memcpy(&http_finished.conn_info, conn, sizeof(struct connection_info));
    __builtin_memcpy(&http_finished.http, stats_cached, sizeof(struct layer7_http));

    // delete cached stats
    bpf_map_delete_elem(&bpfmap_http_stats, conn);

    http_finished.http.status_code = stats->status_code;
    http_finished.http.http_version = stats->http_version;

    http_finished.http.resp_ts = stats->resp_ts;

    bpf_perf_event_output(ctx, &bpfmap_httpreq_fin_event, bpf_get_smp_processor_id(),
                          &http_finished, sizeof(struct http_req_finished));

    return 0;
}

static __always_inline req_resp_t checkHTTP(struct socket *skt, __u8 *buf, __u64 k_time, struct connection_info *conn, struct layer7_http *stats)
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

    stats->req_payload_id.pid_tid = pid_tgid;
    stats->req_payload_id.ktime = k_time;
    stats->req_payload_id.cpuid = bpf_get_smp_processor_id();
    stats->req_payload_id.prandom = bpf_get_prandom_u32();

    __u8 tmp_buffer[32] = {0};
    bpf_probe_read(&tmp_buffer, sizeof(tmp_buffer), buf);

    // 判断请求/响应以及是否为服务端
    return parse_layer7_http1(tmp_buffer, stats);
}

static __always_inline int parse_http1x(void *ctx, struct l7_buffer *l7buffer, __u64 k_time, struct connection_info *conn, struct layer7_http *stats, req_resp_t req_type, enum MSG_RW rw)
{
    // 判断请求/响应以及是否为服务端
    switch (req_type)
    {
    case HTTP_REQ_REQ:
        stats->req_ts = k_time;
        record_http_req(ctx, conn, stats, l7buffer, rw);
        break;
    case HTTP_REQ_RESP:
        stats->resp_ts = k_time;
        record_http_resp(ctx, conn, stats, rw);
        break;
    default:
        return -1;
    }
    return 0;
}

#endif // !__L7_UTILS_