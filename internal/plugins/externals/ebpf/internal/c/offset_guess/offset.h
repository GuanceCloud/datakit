#ifndef __OFFSET_H__
#define __OFFSET_H__

#include <linux/types.h>
// #include "../netflow/conn_stats.h"

#ifndef KERNEL_TASK_COMM_LEN
#define KERNEL_TASK_COMM_LEN 16
#endif

#define ERR_G_NOERROR 0
#define ERR_G_NS_INUM 19

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
    GUESS_SOCKET_SK,
};

struct offset_guess
{
    // netflow
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
    __u64 offset_ns_common_inum; // +conntrack
    __u64 offset_socket_sk;
    // tcp seq
    __u64 offset_copied_seq;
    __u64 offset_write_seq;

    // apiflow
    __u64 offset_task_struct_files;
    __u64 offset_files_struct_fdt;
    __u64 offset_socket_file;
    __u64 offset_file_private_data;

    // conntrack
    __u64 offset_ct_net;
    __u64 offset_origin_tuple;
    __u64 offset_reply_tuple;

    __u8 process_name[KERNEL_TASK_COMM_LEN];
    __s64 err;
    __u64 state; // conn info updated times
    __u64 pid_tgid;
    __u32 conn_type; // (tcp/udp | IPv4/IPv6)

    __u16 sport;
    __u16 dport;
    __u16 sport_skt;
    __u16 dport_skt;

    __u32 saddr[4];
    __u32 daddr[4];

    // __u32 pid;
    __u32 netns;
    __u32 meta;
    __u32 rtt;
    __u32 rtt_var;

    __u32 _pad;
};

struct offset_httpflow
{
    __u8 process_name[KERNEL_TASK_COMM_LEN];

    __u64 pid_tgid;

    __s32 offset_task_struct_files;

    // eBPF prog loop 0 ~ 300 times
    __s32 offset_files_struct_fdt;

    // TODO
    __s32 offset_fdtable_fd;

    // offset_sock - sizeof(void *)
    __s32 offset_socket_file;

    __s32 offset_file_private_data;

    __s32 times;

    __s32 state; // 0b1 | 0b10, ok

    __s32 fd;
};

struct packet_tuple
{
    __u32 src_ip[4];
    __u32 dst_ip[4];

    __u16 src_port;
    __u16 dst_port;

    __u32 _pad0;
};

struct packet_info
{
    __u8 ctrl_syn;
    __u8 ctrl_syn_ack;
    __u8 ctrl_ack;

    __u8 scale;

    __u16 rcv_wnd;

    __u16 _pad0;

    __u32 seq;
    __u32 ack;
};

struct offset_tcp_seq
{
    __u8 process_name[KERNEL_TASK_COMM_LEN];
    __u64 pid_tgid;

    // guessed offset
    __s32 gs_rtt;

    // from packet data, now unused
    // __s32 da_seq;
    // __s32 da_ack;
    // __s32 da_wnd;
    //
    // __s32 _pad0;

    __s32 offset_copied_seq;
    __s32 offset_write_seq;

    __s32 state; // 0b1 | 0b10, ok
};

struct nf_conn_tuple
{
    __u32 src_ip[4];
    __u32 dst_ip[4];

    __u16 src_port;
    __u16 dst_port;

    __u16 l3num;
    __u8 l4proto;

    __u8 _pad;
};

struct offset_conntrack
{
    __u8 process_name[KERNEL_TASK_COMM_LEN];
    __s64 err; // ERR_G_NOERROR or ERR_G_SK_NET

    __u64 state; // conn info updated times
    __u64 pid_tgid;
    // __u32 conn_type; // (tcp/udp | IPv4/IPv6)

    __u64 offset_ct_origin_tuple;
    __u64 offset_ct_reply_tuple;

    __u64 offset_ct_net;
    __u64 offset_ns_common_inum;

    struct nf_conn_tuple origin, reply;

    __u32 netns;
    __u32 _pad;
};

struct comm_getsockopt_arg
{
    void *skt;
    void *file;
};

#endif // !__OFFSET_H__
