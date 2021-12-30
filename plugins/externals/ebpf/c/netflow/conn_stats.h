#ifndef __CONN_STATS_H
#define __CONN_STATS_H

#include <linux/types.h>
// ------------------------------------------------------
// ---------------- define or enum ----------------------

enum
{
    PORT_CLOSED = 0,
    PORT_LISETINING
};

enum
{
    CONN_DIRECTION_AUTO = 0x0,
    CONN_DIRECTION_INCOMING,
    CONN_DIRECTION_OUTGOING,
    CONN_DIRECTION_UNKNOWN,
};

enum
{
    CONN_SYNC_SENT_MASK = 0b0001,   // 0b0001
    CONN_SYNC_RCVD_MASK = 0b0010,   // 0b0010
    CONN_ESTABLISHED_MASK = 0b0100, // 0b0100

    CONN_SYNC_SENT = 0b0001,   // 0b0001 << 0
    CONN_SYNC_RCVD = 0b0010,   // 0b0001 << 1
    CONN_ESTABLISHED = 0b0100, // 0b0001 << 2
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

// key of bpf map conn_stats
struct connection_info
{
    __u32 saddr[4]; //src ip address； Use the last element to store the IPv4 address
    __u32 daddr[4]; // dst ip address
    __u16 sport;    // src port
    __u16 dport;    // dst port
    __u32 pid;
    __u32 netns; // network namespace
    __u32 meta;  // first byte: 0x0000|IPv4 or 0x0001|IPv6; second byte 0x0000|TCP or 0x0100|UDP; ...
};

struct connection_stats
{
    __u64 sent_bytes;
    __u64 recv_bytes;

    __u64 sent_packets;
    __u64 recv_packets;

    __u32 flags; // tcp three-way handshake
    __u8 direction;

    __u64 timestamp;
};

struct connection_tcp_stats
{
    __u16 state_transitions; // TCP_CLOSE, TCP_ESTABLISHED
    __s32 retransmits;
    __u32 rtt;
    __u32 rtt_var;
};

struct connection_closed_info
{
    struct connection_info conn_info;
    struct connection_stats conn_stats;
    struct connection_tcp_stats conn_tcp_stats;
};

struct port_bind
{
    __u32 netns;
    __u16 port;
};

struct pid_fd
{
    __u32 pid;
    __s32 fd;
};

#endif // !__CONN_STATS_H
