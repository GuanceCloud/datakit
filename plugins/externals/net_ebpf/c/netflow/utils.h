#ifndef __UTILS_H
#define __UTILS_H
#include <linux/types.h>
#include <asm-generic/errno-base.h>
#include "conn_stats.h"
#include "bpfmap.h"
#include "load_const.h"

static __always_inline __u64 load_kernel_version()
{
    __u64 var = 0;
    LOAD_OFFSET("kernel_version", var);
    return var;
}

static __always_inline int pre_kernel_4_1_0()
{
    __u64 k_version = load_kernel_version();
    if (k_version < 0x0004000100000000)
    {
        return 1;
    }
    else
    {
        return 0;
    }
}

static __always_inline int pre_kernel_4_7_0()
{
    __u64 k_version = load_kernel_version();
    if (k_version < 0x0004000700000000)
    {
        return 1;
    }
    else
    {
        return 0;
    }
}

static __always_inline __u64 load_offset_sk_num()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_num", var);
    return var;
}

static __always_inline __u64 load_offset_inet_sport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_inet_sport", var);
    return var;
}

static __always_inline __u64 load_offset_sk_family()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_family", var);
    return var;
}

static __always_inline __u64 load_offset_sk_rcv_saddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_rcv_saddr", var);
    return var;
}

static __always_inline __u64 load_offset_sk_daddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_daddr", var);
    return var;
}

static __always_inline __u64 load_offset_sk_v6_rcv_saddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_v6_rcv_saddr", var);
    return var;
}
static __always_inline __u64 load_offset_sk_v6_daddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_v6_daddr", var);
    return var;
}
static __always_inline __u64 load_offset_sk_dport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_dport", var);
    return var;
}
static __always_inline __u64 load_offset_rtt()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_tcp_sk_srtt_us", var);
    return var;
}

static __always_inline __u64 load_offset_rtt_var()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_tcp_sk_mdev_us", var);
    return var;
}

static __always_inline __u64 load_offset_flowi4_saddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi4_saddr", var);
    return var;
}
static __always_inline __u64 load_offset_flowi4_daddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi4_daddr", var);
    return var;
}
static __always_inline __u64 load_offset_flowi4_sport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi4_sport", var);
    return var;
}

static __always_inline __u64 load_offset_flowi4_dport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi4_dport", var);
    return var;
}
static __always_inline __u64 load_offset_flowi6_saddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi6_saddr", var);
    return var;
}
static __always_inline __u64 load_offset_flowi6_daddr()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi6_daddr", var);
    return var;
}
static __always_inline __u64 load_offset_flowi6_sport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi6_sport", var);
    return var;
}

static __always_inline __u64 load_offset_flowi6_dport()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_flowi6_dport", var);
    return var;
}

static __always_inline __u64 load_offset_sk_net()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_sk_net", var);
    return var;
}

static __always_inline __u64 load_offset_ns_common_inum()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_ns_common_inum", var);
    return var;
}

// value of sknet: &(struct net *) or &(struct possible_net_t)
static __always_inline __u32 read_netns(void *sk)
{
    __u32 inum = 0;
    struct net *netptr = NULL;
    // read the memory address of a net instance from *sknet,
    // possible_net_t has only one field: struct net *
    __u64  offset_sk_net = load_offset_sk_net();
    __u64 offset_ns_common_inum = load_offset_ns_common_inum();
    bpf_probe_read(&netptr, sizeof(netptr), (__u8 *)sk + offset_sk_net);

    bpf_probe_read(&inum, sizeof(inum), (__u8 *)netptr + offset_ns_common_inum);
    return inum;
}


// param direction: connetction direction, automatic judgment | incoming | outgoing | unknown
// param count_typpe: packet count type, 1: init, 2:increment
static __always_inline void update_conn_stats(struct connection_info *conn, size_t sent_bytes, size_t recv_bytes, u64 ts, int direction,
                                              __u32 packets_out, __u32 packets_in, int count_type)
{
    struct connection_stats *val;

    // initialize-if-no-exist the connection stat, and load it
    struct connection_stats empty = {};
    __builtin_memset(&empty, 0, sizeof(struct connection_stats));
    bpf_map_update_elem(&bpfmap_conn_stats, conn, &empty, BPF_NOEXIST);
    val = bpf_map_lookup_elem(&bpfmap_conn_stats, conn);

    if (val == NULL)
    {
        return;
    }

    if (sent_bytes > 0)
    {
        __sync_fetch_and_add(&val->sent_bytes, sent_bytes);
    }
    if (recv_bytes > 0)
    {
        __sync_fetch_and_add(&val->recv_bytes, recv_bytes);
    }
    if ((conn->meta & CONN_L4_MASK) == CONN_L4_TCP)
    { // tcp three-way handshake
        if (recv_bytes == 0 && sent_bytes > 0)
        {
            val->flags = (val->flags & ~CONN_SYNC_SENT_MASK) | CONN_SYNC_SENT;
        }
        else if (sent_bytes == 0 && recv_bytes > 0)
        {
            val->flags = (val->flags & ~CONN_SYNC_RCVD_MASK) | CONN_SYNC_RCVD;
        }
        else if (sent_bytes > 0 && recv_bytes > 0)
        {
            val->flags = (val->flags & ~CONN_ESTABLISHED_MASK) | CONN_ESTABLISHED;
        }
    }

    val->timestamp = ts;

    // direction
    if (direction == CONN_DIRECTION_AUTO)
    {
        struct port_bind bind = {};
        __u8 *port_state = NULL;
        bind.port = conn->sport;
        if ((conn->meta & CONN_L4_MASK) == CONN_L4_TCP)
        {
            bind.netns = conn->netns;
            port_state = bpf_map_lookup_elem(&bpfmap_port_bind, &bind);
        }
        else
        {
            port_state = bpf_map_lookup_elem(&bpfmap_udp_port_bind, &bind);
        }
        val->direction = (port_state != NULL) ? CONN_DIRECTION_INCOMING : CONN_DIRECTION_OUTGOING;
    }
    else
    {
        val->direction = direction;
    }
}

// key conn_info remove pid
static __always_inline void update_tcp_stats(struct connection_info conn_info, struct connection_tcp_stats stats)
{
    // value copy

    // query stats without the PID from the tuple
    conn_info.pid = 0;

    struct connection_tcp_stats empty = {};
    // initialize-if-no-exist the connetion state, and load it
    bpf_map_update_elem(&bpfmap_conn_tcp_stats, &conn_info, &empty, BPF_NOEXIST);
    struct connection_tcp_stats *val = bpf_map_lookup_elem(&bpfmap_conn_tcp_stats, &conn_info);

    if (val == NULL)
    {
        return;
    }

    if (stats.rtt > 0)
    {
        val->rtt = stats.rtt;
        val->rtt_var = stats.rtt_var;
    }

    if (stats.retransmits > 0)
    {
        __sync_fetch_and_add(&val->retransmits, stats.retransmits);
    }

    if (stats.state_transitions > 0)
    {
        val->state_transitions |= stats.state_transitions;
    }
}

static __always_inline int update_tcp_retransmit(struct connection_info conn, int segs)
{
    __u64 pid_tgid = 0;
    conn.pid = 0;
    struct connection_tcp_stats tcpstats = {
        .retransmits = segs,
        .rtt = 0,
        .rtt_var = 0,
    };
    update_tcp_stats(conn, tcpstats);
    return 0;
}

// Remove conn from bpfmap_conn_stats.
// In addition if it is a TCP conn, remove it from bpfmap_conn_tcp_stats.
static __always_inline void remove_from_conn_map(struct connection_info conn_info, struct connection_closed_info *event)
{
    event->conn_info = conn_info;

    __u32 tcp_or_udp = conn_info.meta & CONN_L4_MASK;
    struct connection_tcp_stats *tcp_sts = NULL;

    if (tcp_or_udp == CONN_L4_TCP)
    {
        __u32 pid = conn_info.pid;
        conn_info.pid = 0;
        tcp_sts = bpf_map_lookup_elem(&bpfmap_conn_tcp_stats, &conn_info);
        if (tcp_sts != NULL)
        {
            event->conn_tcp_stats = *tcp_sts;
            event->conn_tcp_stats.state_transitions |= (1 << TCP_CLOSE);
        }
        bpf_map_delete_elem(&bpfmap_conn_tcp_stats, &conn_info);
        conn_info.pid = pid;
    }

    struct connection_stats *conn_sts = bpf_map_lookup_elem(&bpfmap_conn_stats, &conn_info);
    if (conn_sts != NULL)
    {
        event->conn_stats = *conn_sts;
    }
    bpf_map_delete_elem(&bpfmap_conn_stats, &conn_info);
}

static __always_inline void send_conn_closed_event(struct pt_regs *ctx, struct connection_closed_info event, __u64 cpu)
{
    bpf_perf_event_output(ctx, &bpfmap_closed_event, cpu, &event, sizeof(event));
}

static __always_inline void swap_u16(__u16 *v)
{
    __u16 tmpv = *v & 0xFF;
    *v >>= 8;
    *v |= tmpv << 8;
}

// network byte order (big-endian), u8 -> u32.
static __always_inline void in6addr_to_u32arr(u32 ipv6[], __u8 *in6)
{
    // in6->in6_u.u6_addr8
    bpf_probe_read(ipv6, sizeof(__u8) * 16, in6);
}

// IPv4-mapped IPv6 address, ::FFFF:xxx.xxx.xxx.xxx, 4 * 32bits
static __always_inline bool is_ipv4_mapped_ipv6(__u32 saddr[4], __u32 daddr[4])
{
#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
    return (
        (saddr[0] == 0x0 && saddr[1] == 0x0 && saddr[2] == 0xFFFF0000) ||
        (daddr[0] == 0x0 && daddr[1] == 0x0 && daddr[2] == 0xFFFF0000));
#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
    return ((saddr[0] == 0x0 && saddr[1] == 0x0 && saddr[2] == 0x0000FFFF) ||
            (daddr[0] == 0x0 && daddr[1] == 0x0 && daddr[2] == 0x0000FFFF));
#else
#error "The machine's __BYTE_ORDER__ is unknown."
#endif
}

static __always_inline __u16 read_sock_src_port(struct sock *sk)
{
    __u16 sport = 0;
    // sport: sk->sk_num
    __u64 offset_sk_num = load_offset_sk_num();
    bpf_probe_read(&sport, sizeof(sport), (__u8 *)sk + offset_sk_num);
    if (sport == 0)
    {
        // sport: &inet_sk(sk)->inet_sport
        __u64 offset_inet_sport = load_offset_inet_sport();
        bpf_probe_read(&sport, sizeof(sport), (__u8 *)sk + offset_inet_sport);
        swap_u16(&sport); // default in little endian system
    }
    return sport;
}

static __always_inline void read_tcp_segment_counts(struct sock *skp, __u32 *packets_in, __u32 *packets_out)
{
    bpf_probe_read(packets_out, sizeof(*packets_out), &tcp_sk(skp)->segs_out);
    bpf_probe_read(packets_in, sizeof(*packets_in), &tcp_sk(skp)->segs_in);
}

// read pid, meta, saddr, daddr
static __always_inline int read_connection_info(struct sock *sk, struct connection_info *conn_info,
                                                __u64 pid_tgid, enum ConnLayerP l4_p)
{
    // read L4 protocol, pid
    conn_info->meta = (conn_info->meta & ~CONN_L4_MASK) | l4_p;
    conn_info->pid = pid_tgid >> 32;

    conn_info->netns = read_netns(sk);
    // read src addr, dst addr and L3 protocol
    unsigned short family = AF_UNSPEC;

    __u64 offset_sk_family = load_offset_sk_family();

    // family: sk->sk_family
    bpf_probe_read(&family, sizeof(family), (__u8 *)sk + offset_sk_family);
    if (family == AF_INET)
    {
        __u64 offset_sk_rcv_saddr = load_offset_sk_rcv_saddr();
        __u64 offset_sk_daddr = load_offset_sk_daddr();

        // Use the last element to store the IPv4 address
        conn_info->meta = (conn_info->meta & ~CONN_L3_MASK) | CONN_L3_IPv4;
        // saddr: sk->sk_rcv_saddr
        bpf_probe_read(conn_info->saddr + 3, sizeof(__be32), (__u8 *)sk + offset_sk_rcv_saddr);
        // daddr: sk->sk_daddr
        bpf_probe_read(conn_info->daddr + 3, sizeof(__be32), (__u8 *)sk + offset_sk_daddr);
        if ((conn_info->daddr[3] | conn_info->saddr[3]) == 0)
        {
            return -1;
        }
    }
    else if (family == AF_INET6)
    {
        __u64 offset_sk_v6_daddr = load_offset_sk_v6_daddr();
        __u64 offset_sk_v6_rcv_saddr = load_offset_sk_v6_rcv_saddr();

        conn_info->meta = (conn_info->meta & ~CONN_L3_MASK) | CONN_L3_IPv6;
        // saddr: sk->sk_v6_rcv_saddr
        in6addr_to_u32arr(conn_info->saddr, (__u8 *)sk + offset_sk_v6_rcv_saddr);
        // daddr: sk->sk_v6_daddr
        in6addr_to_u32arr(conn_info->daddr, (__u8 *)sk + offset_sk_v6_daddr);
        if ((conn_info->saddr[0] | conn_info->saddr[1] | conn_info->saddr[2] | conn_info->saddr[3] |
             conn_info->daddr[0] | conn_info->daddr[1] | conn_info->daddr[2] | conn_info->daddr[3]) == 0)
        {
            return -1;
        }
    }

    // read sport and dport

    conn_info->sport = read_sock_src_port(sk);

    // dport: sk->sk_dport
    __u64 offset_sk_dport = load_offset_sk_dport();
    bpf_probe_read(&conn_info->dport, sizeof(conn_info->dport), (__u8 *)sk + offset_sk_dport);
    swap_u16(&conn_info->dport);

    if ((conn_info->sport | conn_info->dport) == 0)
    {
        return -1;
    }
    return 0;
}

// read tcp info: rtt, rtt_var
static __always_inline int read_tcp_rtt(struct sock *sk, struct connection_tcp_stats *tcp_stats)
{
    __u32 srtt_us = 0;
    __u32 mdev_us = 0;

    // rtt: &tcp_sk(sk)->srtt_us
    // rtt_var: &tcp_sk(sk)->mdev_us
    __u64 offset_rtt = load_offset_rtt();
    __u64 offset_rtt_var = load_offset_rtt_var();

    bpf_probe_read(&srtt_us, sizeof(srtt_us), (__u8 *)sk + offset_rtt);
    bpf_probe_read(&mdev_us, sizeof(mdev_us), (__u8 *)sk + offset_rtt_var);

    tcp_stats->rtt = srtt_us >> 3;
    tcp_stats->rtt_var = mdev_us >> 2;

    return 0;
}


#endif // !__UTILS_H
