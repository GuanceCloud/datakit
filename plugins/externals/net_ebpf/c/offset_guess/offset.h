#ifndef __OFFSET_H
#define __OFFSET_H

#ifndef PROCNAMELEN
#define PROCNAMELEN 16
#endif // !PROCNAMELEN

#include <linux/types.h>

#define ERR_G_NOERROR 0
#define ERR_G_SK_NET 19

enum GUESS
{
    GUESS_SK_NUM = 1,
    GUESS_INET_SPORT,
    GUESS_SK_FAMILY,
    GUESS_SK_RCV_SADDR,
    GUESS_SK_DADDR,
    GUESS_SK_DPORT,
    GUESS_TCP_SK_SRTT_US,
    GUESS_TCP_SK_MDEV_US,
    GUESS_FLOWI4_SADDR,
    GUESS_FLOWI4_DADDR,
    GUESS_FLOWI4_SPORT,
    GUESS_FLOWI4_DPORT,
    GUESS_FLOWI6_SADDR,
    GUESS_FLOWI6_DADDR,
    GUESS_FLOWI6_SPORT,
    GUESS_FLOWI6_DPORT,
    GUESS_SKADDR_SIN_PORT,
    GUESS_SKADRR6_SIN6_PORT,
    GUESS_SK_NET,
    GUESS_NS_COMMON_INUM,
};

enum ConnLayerP
{
    CONN_L3_MASK = 0xFF, // 0xFF
    CONN_L3_IPv4 = 0x00, // 0x00
    CONN_L3_IPv6 = 0x01, // 0x01

    CONN_L4_MASK = 0xFF00, // 0xFF00
    CONN_L4_TCP = 0x0000,  // 0x00 << 8
    CONN_L4_UDP = 0x0100,  // 0x01 << 8
};

struct offset_guess
{
    __u64 offset_sk_num;
    __u64 offset_inet_sport;
    __u64 offset_sk_family;
    __u64 offset_sk_rcv_saddr;
    __u64 offset_sk_daddr;
    __u64 offset_sk_v6_rcv_saddr;
    __u64 offset_sk_v6_daddr;
    __u64 offset_sk_dport;
    __u64 offset_tcp_sk_srtt_us;
    __u64 offset_tcp_sk_mdev_us;
    __u64 offset_flowi4_saddr;
    __u64 offset_flowi4_daddr;
    __u64 offset_flowi4_sport;
    __u64 offset_flowi4_dport;
    __u64 offset_flowi6_saddr;
    __u64 offset_flowi6_daddr;
    __u64 offset_flowi6_sport;
    __u64 offset_flowi6_dport;
    __u64 offset_skaddr_sin_port;
    __u64 offset_skaddr6_sin6_port;
    __u64 offset_sk_net;
    __u64 offset_ns_common_inum;

    __u8 process_name[PROCNAMELEN];
    __s64 err;
    __u64 state; // conn info updated times
    __u64 pid_tgid;
    __u32 conn_type; // (tcp/udp | IPv4/IPv6)

    __u16 sport;
    __u16 dport;

    __u32 saddr[4];
    __u32 daddr[4];

    // __u32 pid;
    __u32 netns;
    __u32 meta;
    __u32 rtt;
    __u32 rtt_var;
};

#endif // !__OFFSET_H
