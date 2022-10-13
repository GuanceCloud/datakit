#ifndef __L7_UTILS_
#define __L7_UTILS_

#define KEEPPACKET -1
#define DROPPACKET 0

#include <uapi/linux/if_ether.h>
#include <uapi/linux/in.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/ipv6.h>
#include <uapi/linux/tcp.h>
#include <uapi/linux/udp.h>

#include "bpf_helpers.h"
#include "netflow_utils.h"
#include "bpfmap_l7.h"
#include "l7_stats.h"

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

enum
{
    HTTP_REQ_UNKNOWN = 0b00,
    HTTP_REQ_REQ = 0b01,
    HTTP_REQ_RESP = 0b10
};

static __always_inline void read_ipv4_from_skb(struct __sk_buff *skb, struct connection_info *conn_info)
{
    // 读取数据示例 load_half -> 0x7f00, load_word -> 0x7f000001
    conn_info->saddr[3] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct iphdr, saddr)));
    conn_info->daddr[3] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct iphdr, daddr)));

    conn_info->meta = (conn_info->meta & ~CONN_L3_MASK) | CONN_L3_IPv4;
}

static __always_inline void read_ipv6_from_skb(struct __sk_buff *skb, struct connection_info *conn_info)
{
    conn_info->saddr[0] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, saddr)));
    conn_info->saddr[1] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, saddr) + 4));
    conn_info->saddr[2] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, saddr) + 8));
    conn_info->saddr[3] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, saddr) + 12));

    conn_info->daddr[0] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, daddr)));
    conn_info->daddr[1] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, daddr) + 4));
    conn_info->daddr[2] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, daddr) + 8));
    conn_info->daddr[3] = __builtin_bswap32(load_word(skb, ETH_HLEN + offsetof(struct ipv6hdr, daddr) + 12));

    conn_info->meta = (conn_info->meta & ~CONN_L3_MASK) | CONN_L3_IPv6;
}

static __always_inline void read_payload_skb(struct conn_skb_l4_info *l4, __u32 *payload, struct __sk_buff *skb)
{
    __u32 offset = l4->hdr_len + 0;

    __u32 pkt_read_size = l4->hdr_len + HTTP_PAYLOAD_MAXSIZE;
    if (pkt_read_size > skb->len)
    {
        pkt_read_size = skb->len;
    }

// 越界访问 skb 数据将有异常
#pragma unroll
    for (int i = 0; i < HTTP_PAYLOAD_LOOP_SIZE; i++) // arr[HTTP_PAYLOAD_MAXSIZE - 1] == EOF
    {
        if (offset + 4 > pkt_read_size)
        {
            break;
        }

        bpf_skb_load_bytes(skb, offset, payload, 4);
        offset += 4;
        payload++;
    }
}


// TODO:
// Looks like the BPF stack limit of 512 bytes is exceeded.
//
// static __always_inline void read_payload_skb_for_oldversion(struct conn_skb_l4_info *l4, __u32 *payload, struct __sk_buff *skb)
// {
//     __u32 offset = l4->hdr_len + 0;
//     __u32 pkt_read_size = l4->hdr_len + HTTP_PAYLOAD_MAXSIZE;
//     if (pkt_read_size > skb->len)
//     {
//         pkt_read_size = skb->len;
//     }
// // 越界访问 skb 时读取的 payload 数据有异常
// #pragma unroll
//     for (int i = 0; i < HTTP_PAYLOAD_LOOP_SIZE - 1; i++) // arr[HTTP_PAYLOAD_MAXSIZE - 1] == EOF
//     {
//         if (offset + 4 > pkt_read_size)
//         {
//             break;
//         }
//         asm volatile("" ::
//                          : "r1");
//         *payload = __builtin_bswap32(load_word(skb, offset));
//         offset += 4;
//         payload++;
//     }
// }

// 从结构体 __skb_buff 读取连接信息 和 eth 帧头、ip 头、tcp/udp 头的总字节数
static __always_inline int read_connection_info_skb(struct __sk_buff *skb, struct conn_skb_l4_info *skbinfo, struct connection_info *conn_info)
{
    skbinfo->hdr_len = ETH_HLEN; // sizeof(struct ethhdr);

    // eth protocol id
    __be16 eth_next_proto = load_half(skb, offsetof(struct ethhdr, h_proto));
    __u8 ip_next_proto = 0;

    switch (eth_next_proto)
    {
    case ETH_P_IPV6:
        ip_next_proto = load_byte(skb, skbinfo->hdr_len + offsetof(struct ipv6hdr, nexthdr));
        read_ipv6_from_skb(skb, conn_info);
        skbinfo->hdr_len += sizeof(struct ipv6hdr);
        break;
    case ETH_P_IP:
        ip_next_proto = load_byte(skb, skbinfo->hdr_len + offsetof(struct iphdr, protocol));
        read_ipv4_from_skb(skb, conn_info);
        skbinfo->hdr_len += sizeof(struct iphdr);
        break;
    default:
        return -1;
    }

    switch (ip_next_proto)
    {
    case IPPROTO_TCP:
        conn_info->meta = (conn_info->meta & ~CONN_L4_MASK) | CONN_L4_TCP;

        conn_info->sport = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, source));
        conn_info->dport = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, dest));

        // skbinfo->seg_seq = load_word(skb, skbinfo->hdr_len + offsetof(struct tcphdr, seq));
        // skbinfo->seg_ack = load_word(skb, skbinfo->hdr_len + offsetof(struct tcphdr, ack_seq));

        // load_half 将交换字节序
        __u16 doff_and_flags = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, ack_seq) + 4);
        // + tcp data offset (doff * 4bytes: tcp_hdr_len + tcp_opt_len)
        skbinfo->hdr_len += (doff_and_flags & 0xF000) >> 10;
        skbinfo->tcp_flags = doff_and_flags & 0x0FFF;

        break;
    case IPPROTO_UDP:
        conn_info->meta = (conn_info->meta & ~CONN_L4_MASK) | CONN_L4_UDP;
        conn_info->sport = load_half(skb, skbinfo->hdr_len + offsetof(struct udphdr, source));
        conn_info->dport = load_half(skb, skbinfo->hdr_len + offsetof(struct udphdr, dest));
        break;
    default:
        return -1;
    }

    if (skb->len < skbinfo->hdr_len)
    {
        return -1;
    }

    return 0;
}

static __always_inline void swap_conn_src_dst(struct connection_info *conn_info)
{
    __u32 addr = 0;
    __u16 port = 0;

    port = conn_info->dport;
    conn_info->dport = conn_info->sport;
    conn_info->sport = port;

#pragma unroll
    for (int i = 0; i < 4; i++)
    {
        addr = conn_info->daddr[i];
        conn_info->daddr[i] = conn_info->saddr[i];
        conn_info->saddr[i] = addr;
    }
}

static __always_inline int parse_layer7_http(__u8 *buffer, struct layer7_http *l7http)
{
    switch (buffer[0])
    {
    case 'G':
        if (buffer[1] == 'E' && buffer[2] == 'T') // HTTP GET
        {
            l7http->method = HTTP_METHOD_GET;
            l7http->req_status = HTTP_REQ_REQ;
            return 0;
        }
        break;
    case 'P':
        switch (buffer[1])
        {
        case 'O':
            if (buffer[2] == 'S' && buffer[3] == 'T') // HTTP POST
            {
                l7http->method = HTTP_METHOD_POST;
                l7http->req_status = HTTP_REQ_REQ;
                return 0;
            }
            break;
        case 'U':
            if (buffer[2] == 'T') // HTTP PUT
            {
                l7http->method = HTTP_METHOD_PUT;
                l7http->req_status = HTTP_REQ_REQ;
                return 0;
            }
            break;
        case 'A':
            if (buffer[2] == 'T' && buffer[3] == 'C' && buffer[4] == 'H') // HTTP PATCH
            {
                l7http->method = HTTP_METHOD_PATCH;
                l7http->req_status = HTTP_REQ_REQ;
                return 0;
            }
            break;
        default:
            break;
        }
    case 'D':
        if (buffer[1] == 'E' && buffer[2] == 'L' && buffer[3] == 'E' && buffer[4] == 'T' && buffer[5] == 'E') // HTTP DELETE
        {
            l7http->method = HTTP_METHOD_DELETE;
            l7http->req_status = HTTP_REQ_REQ;
            return 0;
        }
        break;
    case 'H':
        if (buffer[1] == 'T' && buffer[2] == 'T' && buffer[3] == 'P') // response payload
        {
            l7http->req_status = HTTP_REQ_RESP;
            goto HTTPRESPONSE;
        }
        else if (buffer[1] == 'E' && buffer[2] == 'A' && buffer[3] == 'D') // HTTP HEAD
        {
            l7http->method = HTTP_METHOD_HEAD;
            l7http->req_status = HTTP_REQ_REQ;
            return 0;
        }
        break;
    case 'O':
        if (buffer[1] == 'P' && buffer[2] == 'T' && buffer[3] == 'I' && buffer[4] == 'O' && buffer[5] == 'N' && buffer[6] == 'S') // HTTP OPTIONS
        {
            l7http->method = HTTP_METHOD_OPTIONS;
            l7http->req_status = HTTP_REQ_REQ;
            return 0;
        }
        break;
    // case 'C':
    //     if (buffer[1] == 'O' && buffer[2] == 'N' && buffer[3] == 'N' &&
    //         buffer[4] == 'E' && buffer[5] == 'C' && buffer[6] == 'T') // HTTP CONNECTION
    //     {
    //         l7http->method = HTTP_METHOD_CONNECT;
    //         l7http->req_status = HTTP_REQ_REQ;
    //         return 0;
    //     }
    //     break;
    // case 'T':
    //     if (buffer[1] == 'R' && buffer[2] == 'A' && buffer[3] == 'C' && buffer[4] == 'E')
    //     { // HTTP TRACE
    //         l7http->method = HTTP_METHOD_TRACE;
    //         l7http->req_status = HTTP_REQ_REQ;
    //         return 0;
    //     }
    default:
        break;
    }

    return -1;

HTTPRESPONSE:
    if (buffer[4] != '/' || buffer[6] != '.' || buffer[8] != ' ')
    {
        return -1;
    }
    l7http->http_version = ((buffer[5] - '0') << 16) + (buffer[7] - '0');
    l7http->status_code = (buffer[9] - '0') * 100 + (buffer[10] - '0') * 10 + (buffer[11] - '0');
    return 0;
}

static __always_inline void map_cache_finished_http_req(struct http_req_finished_info *http_req)
{
    __u16 cpuid = bpf_get_smp_processor_id();
    __u16 map_index = cpuid;
#pragma unroll
    for (int i = 0; i < MAPCANSAVEREQNUM; i++)
    {
        if (bpf_map_lookup_elem(&bpfmap_httpreq_finished, &map_index) == NULL)
        {
            bpf_map_update_elem(&bpfmap_httpreq_finished, &map_index, http_req, BPF_NOEXIST);
            return;
        }
        map_index += MAPCANSAVEREQNUM;
    }
}

static __always_inline void send_httpreq_fin_event(struct pt_regs *ctx)
{
    __u32 cpuid = bpf_get_smp_processor_id();
    __u32 map_index = cpuid;

    struct http_req_finished_info *http_fin = NULL;

#pragma unroll
    for (int i = 0; i < MAPCANSAVEREQNUM; i++)
    {
        http_fin = bpf_map_lookup_elem(&bpfmap_httpreq_finished, &map_index);
        if (http_fin != NULL)
        {
            struct http_req_finished_info fin = {0};
            bpf_probe_read(&fin, sizeof(struct http_req_finished_info), http_fin);
            bpf_perf_event_output(ctx, &bpfmap_httpreq_fin_event, cpuid, &fin, sizeof(struct http_req_finished_info));
            bpf_map_delete_elem(&bpfmap_httpreq_finished, &map_index);
        }
        map_index += MAPCANSAVEREQNUM;
    }
}

static __always_inline void init_ssl_sockfd(void *ssl_ctx, __u32 fd)
{
    struct ssl_sockfd sockfd = {0};
    sockfd.fd = fd;
    bpf_map_update_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx, &sockfd, BPF_ANY);
}

static __always_inline int read_conn_ssl(void *ssl_ctx, __u64 pid_tgid, struct connection_info *conn)
{
    struct ssl_sockfd *sockfd = (struct ssl_sockfd *)bpf_map_lookup_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    if (sockfd == NULL)
    {
        return -1;
    }
    struct pid_fd pidfd = {
        .pid = pid_tgid >> 32,
        .fd = sockfd->fd,
    };
    struct sock **skp = (struct sock **)bpf_map_lookup_elem(&bpfmap_sockfd, &pidfd);
    if (skp == NULL)
    {
        return -1;
    }
    if (read_connection_info(*skp, conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return -1;
    }
    conn->pid = 0;
    conn->netns = 0;
    __builtin_memcpy(&sockfd->conn, conn, sizeof(struct connection_info));
    return 0;
}

static __always_inline int record_http_req(struct connection_info *conn,
                                           struct http_stats *stats, __u8 method)
{
    stats->req_ts = bpf_ktime_get_ns();
    stats->req_method = method;
    bpf_map_update_elem(&bpfmap_http_stats, conn, stats, BPF_NOEXIST);
    return 0;
}

static __always_inline int record_http_resp(struct connection_info *conn,
                                            struct http_stats *stats, struct layer7_http *l7http)
{

    struct http_stats *stats_cached = bpf_map_lookup_elem(&bpfmap_http_stats, conn);
    if (stats_cached == NULL)
    {
        return 0;
    }

    struct http_req_finished_info http_finished = {0};

    __builtin_memcpy(&http_finished.conn_info, conn, sizeof(struct connection_info));
    __builtin_memcpy(&http_finished.http_stats, stats_cached, sizeof(struct http_stats));

    bpf_map_delete_elem(&bpfmap_http_stats, conn);

    http_finished.http_stats.resp_code = l7http->status_code;
    http_finished.http_stats.http_version = l7http->http_version;
    http_finished.http_stats.resp_ts = bpf_ktime_get_ns();

    map_cache_finished_http_req(&http_finished);
    return 0;
}

#endif // !__L7_UTILS_