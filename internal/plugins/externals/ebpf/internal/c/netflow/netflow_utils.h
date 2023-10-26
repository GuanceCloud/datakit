#ifndef __NETFLOW_UTILS_H
#define __NETFLOW_UTILS_H

#include <linux/types.h>
#include <asm-generic/errno-base.h>
#include <linux/tcp.h>

#include "bpf_helpers.h"
#include "conn_stats.h"
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

static __always_inline __u64 load_offset_socket_sk()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_socket_sk", var);
    return var;
}

static __always_inline __u64 load_offset_socket_file()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_socket_file", var);
    return var;
}

static __always_inline __u64 load_offset_task_struct_files()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_task_struct_files", var);
    return var;
}

static __always_inline __u64 load_offset_files_struct_fdt()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_files_struct_fdt", var);
    return var;
}

static __always_inline __u64 load_offset_file_private_data()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_file_private_data", var);
    return var;
}

static __always_inline __u64 load_offset_copied_seq()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_copied_seq", var);
    return var;
}

static __always_inline __u64 load_offset_write_seq()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_write_seq", var);
    return var;
}

// value of sknet: &(struct net *) or &(struct possible_net_t)
static __always_inline __u32 read_netns(void *sk)
{
    __u32 inum = 0;
    struct net *netptr = NULL;
    // read the memory address of a net instance from *sknet,
    // possible_net_t has only one field: struct net *
    __u64 offset_sk_net = load_offset_sk_net();
    __u64 offset_ns_common_inum = load_offset_ns_common_inum();
    bpf_probe_read(&netptr, sizeof(netptr), (__u8 *)sk + offset_sk_net);

    bpf_probe_read(&inum, sizeof(inum), (__u8 *)netptr + offset_ns_common_inum);
    return inum;
}

static __always_inline __u32 read_write_seq(void *sk)
{
    __u32 write_seq = 0;

    __u64 offset_write_seq = load_offset_write_seq();
    bpf_probe_read(&write_seq, sizeof(write_seq),
                   (__u8 *)sk + offset_write_seq);
    return write_seq;
}

static __always_inline __u32 read_copied_seq(void *sk)
{
    __u32 copied_seq = 0;

    __u64 offset_copied_seq = load_offset_copied_seq();
    bpf_probe_read(&copied_seq, sizeof(copied_seq),
                   (__u8 *)sk + offset_copied_seq);
    return copied_seq;
}

static __always_inline void swap_u16(__u16 *v)
{
    *v = __builtin_bswap16(*v);
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
    // (LLVM v10.0.0)
    // EulerOS 2.0 (SP8)/(SP9):
    //
    // --- before ---
    //
    // 12: r1 = 0x10000000000
    // 14: *(u64 *)(r10 -64) = r1
    // ...
    // 57: r2 = *(u32 *)(r10 -60)
    // ERROR: invalid size of register fill.
    //
    // --- after ---
    //
    // 13: r9 = 0
    // 14: *(u64 *)(r10 - 64) = r9
    // 15: r6 = 1
    // 16: *(u8 *)(r10 - 59) = r6
    // ...
    // 58: r2 = *(u32 *)(r10 - 60)
    // PASS.
    //
    conn_info->netns = read_netns(sk);

    // read L4 protocol, pid
    conn_info->meta = (conn_info->meta & ~CONN_L4_MASK) | l4_p;
    conn_info->pid = pid_tgid >> 32;

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
    else
    {
        return -1;
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

#endif // !__NETFLOW_UTILS_H
