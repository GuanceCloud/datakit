#ifndef __NETFLOW_BPFMAP_H
#define __NETFLOW_BPFMAP_H

#include "../conntrack/maps.h"

#include "bpf_helpers.h"
#include "conn_stats.h"

// ------------------------------------------------------
// ---------------------- BPF MAP -----------------------

struct bpf_map_def SEC("maps/bpfmap_conn_stats") bpfmap_conn_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info),
    .value_size = sizeof(struct connection_stats),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_conn_tcp_stats") bpfmap_conn_tcp_stats = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info), // pid shoud be set to 0
    .value_size = sizeof(struct connection_tcp_stats),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_closed_event") bpfmap_closed_event = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),   // smp_processor_id
    .value_size = sizeof(__u32), // perf file fd
    .max_entries = 0,
};

// Temporarily store the pid_tgid(key, u64) and port(value, u16) when inet_bind(v4/v6) is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_inetbind") bpfmap_tmp_inetbind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(__u16),
    .max_entries = 65536,
};

// map key: struct port_bind
// map value: PORT_CLOSED or PORT_LISTENING
struct bpf_map_def SEC("maps/bpfmap_port_bind") bpfmap_port_bind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct port_bind),
    .value_size = sizeof(__u8),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_udp_port_bind") bpfmap_udp_port_bind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct port_bind),
    .value_size = sizeof(__u8),
    .max_entries = 65536,
};
struct udp_revcmsg_tmp
{
    struct sock *sk;
    struct msghdr *msg;
};

struct bpf_map_def SEC("maps/bpf_map_tmp_udprecvmsg") bpf_map_tmp_udprecvmsg = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct udp_revcmsg_tmp),
    .max_entries = 65536,
};

// Temporarily store the pid_tgid(key, u64) and sockfd(value, u32) when sockfd_lookup_light is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_sockfdlookuplight") bpfmap_tmp_sockfdlookuplight = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(__u32),
    .max_entries = 65536,
};

// key: struct pid_fd, value: struct sock pointer
struct bpf_map_def SEC("maps/bpfmap_sockfd") bpfmap_sockfd = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct pid_fd),
    .value_size = sizeof(struct sock *),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_sockfd_inverted") bpfmap_sockfd_inverted = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct sock *),
    .value_size = sizeof(struct pid_fd),
    .max_entries = 65536,
};

// key: pig_tgid, value: sock ptr
struct bpf_map_def SEC("maps/bpfmap_tmp_sendfile") bpfmap_tmp_sendfile = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct sock *),
    .max_entries = 65536,
};

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

static __always_inline void send_conn_closed_event(struct pt_regs *ctx, struct connection_closed_info event, __u64 cpu)
{
    bpf_perf_event_output(ctx, &bpfmap_closed_event, cpu, &event, sizeof(event));
}

// param direction: connetction direction, automatic judgment | incoming | outgoing | unknown
// param count_typpe: packet count type, 1: init, 2:increment
static __always_inline void update_conn_stats(struct connection_info *conn, size_t sent_bytes, size_t recv_bytes, u64 ts, int direction,
                                              __u32 packets_out, __u32 packets_in, int count_type)
{
    struct connection_stats *val = NULL;

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

    do_dnapt(conn, val->nat_daddr, &val->nat_dport);
}

#endif // !__BPFMAP_H
