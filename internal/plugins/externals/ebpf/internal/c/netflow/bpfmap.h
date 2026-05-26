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

struct tcp_segment_counter
{
    __u32 segs_in;
    __u32 segs_out;
};

struct bpf_map_def SEC("maps/bpfmap_conn_tcp_segments") bpfmap_conn_tcp_segments = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct connection_info), // pid should be set to 0
    .value_size = sizeof(struct tcp_segment_counter),
    .max_entries = 65536,
};

enum netflow_update_fail_reason
{
    NETFLOW_UPDATE_FAIL_CONN_STATS = 0,
    NETFLOW_UPDATE_FAIL_TCP_STATS = 1,
    NETFLOW_UPDATE_FAIL_TCP_SEGMENTS = 2,
    NETFLOW_UPDATE_FAIL_MAX = 3,
};

struct bpf_map_def SEC("maps/bpfmap_netflow_update_fail") bpfmap_netflow_update_fail = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32),
    .value_size = sizeof(__u64),
    .max_entries = NETFLOW_UPDATE_FAIL_MAX,
};

static __always_inline void record_netflow_update_fail(__u32 reason)
{
    __u64 *count = bpf_map_lookup_elem(&bpfmap_netflow_update_fail, &reason);
    if (count != NULL)
    {
        __sync_fetch_and_add(count, 1);
    }
}

struct bpf_map_def SEC("maps/bpfmap_closed_event") bpfmap_closed_event = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),   // smp_processor_id
    .value_size = sizeof(__u32), // perf file fd
    .max_entries = 0,
};

struct inet_bind_tmp
{
    __u32 netns;
    __u16 port;
};

// Temporarily store the pid_tgid(key, u64) and bind info(value) when inet_bind(v4/v6) is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_inetbind") bpfmap_tmp_inetbind = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct inet_bind_tmp),
    .max_entries = 1024,
};

// map key: struct port_bind
// map value: PORT_CLOSED or PORT_LISTENING
struct bpf_map_def SEC("maps/bpfmap_port_bind") bpfmap_port_bind = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct port_bind),
    .value_size = sizeof(__u8),
    .max_entries = 65536,
};

// User-space procfs listener seed map. Keep it separate from bpfmap_port_bind
// so kernel probes and procfs scans never race on the same key/value.
struct bpf_map_def SEC("maps/bpfmap_port_bind_proc") bpfmap_port_bind_proc = {
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
    struct connection_info conn_info;
};

struct bpf_map_def SEC("maps/bpf_map_tmp_udprecvmsg") bpf_map_tmp_udprecvmsg = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct udp_revcmsg_tmp),
    .max_entries = 1024,
};

// Temporarily store the pid_tgid(key, u64) and sockfd(value, u32) when sockfd_lookup_light is called.
struct bpf_map_def SEC("maps/bpfmap_tmp_sockfdlookuplight") bpfmap_tmp_sockfdlookuplight = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(__u32),
    .max_entries = 1024,
};

// key: struct pid_fd, value: struct sock pointer
struct bpf_map_def SEC("maps/bpfmap_sockfd") bpfmap_sockfd = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(struct pid_fd),
    .value_size = sizeof(struct sock *),
    .max_entries = 65536,
};

struct bpf_map_def SEC("maps/bpfmap_sockfd_inverted") bpfmap_sockfd_inverted = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(struct sock *),
    .value_size = sizeof(struct pid_fd),
    .max_entries = 65536,
};

// key: pid_tgid, value: connection_info snapshot
struct bpf_map_def SEC("maps/bpfmap_tmp_sendfile") bpfmap_tmp_sendfile = {
    .type = BPF_BEST_EFFORT_LRU_HASH,
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct connection_info),
    .max_entries = 1024,
};

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
        record_netflow_update_fail(NETFLOW_UPDATE_FAIL_TCP_STATS);
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
        __u16 prev = val->state_transitions;
        val->state_transitions |= stats.state_transitions;

        if ((stats.state_transitions & (1 << TCP_CLOSE)) != 0)
        {
            __u16 handshake = (1 << TCP_SYN_SENT) | (1 << TCP_SYN_RECV) | (1 << TCP_NEW_SYN_RECV);
            if ((prev & handshake) != 0 && (prev & (1 << TCP_ESTABLISHED)) == 0)
            {
                __sync_fetch_and_add(&val->connect_failures, 1);
            }
        }
    }

    if (stats.connect_attempts > 0)
    {
        __sync_fetch_and_add(&val->connect_attempts, stats.connect_attempts);
    }

    if (stats.close_wait > 0)
    {
        __sync_fetch_and_add(&val->close_wait, stats.close_wait);
    }

    if (stats.last_ack > 0)
    {
        __sync_fetch_and_add(&val->last_ack, stats.last_ack);
    }

    if (stats.time_wait > 0)
    {
        __sync_fetch_and_add(&val->time_wait, stats.time_wait);
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

static __always_inline void read_tcp_segment_delta(struct connection_info conn, __u32 cur_in, __u32 cur_out,
                                                   __u32 *delta_in, __u32 *delta_out)
{
    if (delta_in == NULL || delta_out == NULL)
    {
        return;
    }

    *delta_in = 0;
    *delta_out = 0;

    conn.pid = 0;

    struct tcp_segment_counter zero = {};
    bpf_map_update_elem(&bpfmap_conn_tcp_segments, &conn, &zero, BPF_NOEXIST);

    struct tcp_segment_counter *prev = bpf_map_lookup_elem(&bpfmap_conn_tcp_segments, &conn);
    if (prev == NULL)
    {
        record_netflow_update_fail(NETFLOW_UPDATE_FAIL_TCP_SEGMENTS);
        return;
    }

    if (cur_in >= prev->segs_in)
    {
        *delta_in = cur_in - prev->segs_in;
    }
    if (cur_out >= prev->segs_out)
    {
        *delta_out = cur_out - prev->segs_out;
    }

    prev->segs_in = cur_in;
    prev->segs_out = cur_out;
}

static __always_inline void send_conn_closed_event(struct pt_regs *ctx, struct connection_closed_info event, __u64 cpu)
{
    bpf_perf_event_output(ctx, &bpfmap_closed_event, cpu, &event, sizeof(event));
}

static __always_inline int fill_conn_stats(struct connection_stats *dst, struct connection_info *conn, size_t sent_bytes, size_t recv_bytes,
                                           u64 ts, int direction, __u32 packets_out, __u32 packets_in)
{
    if (dst == NULL)
    {
        return -1;
    }

    if (sent_bytes > 0)
    {
        __sync_fetch_and_add(&dst->sent_bytes, sent_bytes);
    }
    if (recv_bytes > 0)
    {
        __sync_fetch_and_add(&dst->recv_bytes, recv_bytes);
    }
    if (packets_out > 0)
    {
        __sync_fetch_and_add(&dst->sent_packets, packets_out);
    }
    if (packets_in > 0)
    {
        __sync_fetch_and_add(&dst->recv_packets, packets_in);
    }
    if ((conn->meta & CONN_L4_MASK) == CONN_L4_TCP)
    { // tcp three-way handshake
        if (recv_bytes == 0 && sent_bytes > 0)
        {
            dst->flags = (dst->flags & ~CONN_SYNC_SENT_MASK) | CONN_SYNC_SENT;
        }
        else if (sent_bytes == 0 && recv_bytes > 0)
        {
            dst->flags = (dst->flags & ~CONN_SYNC_RCVD_MASK) | CONN_SYNC_RCVD;
        }
        else if (sent_bytes > 0 && recv_bytes > 0)
        {
            dst->flags = (dst->flags & ~CONN_ESTABLISHED_MASK) | CONN_ESTABLISHED;
        }
    }

    if (ts > 0)
    {
        dst->timestamp = ts;
    }

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
            if (port_state == NULL || *port_state == PORT_CLOSED)
            {
                port_state = bpf_map_lookup_elem(&bpfmap_port_bind_proc, &bind);
            }
        }
        else
        {
            bind.netns = conn->netns;
            port_state = bpf_map_lookup_elem(&bpfmap_udp_port_bind, &bind);
        }
        dst->direction = (port_state != NULL && *port_state != PORT_CLOSED) ? CONN_DIRECTION_INCOMING : CONN_DIRECTION_OUTGOING;
    }
    else
    {
        dst->direction = direction;
    }

    return 0;
}

// param direction: connetction direction, automatic judgment | incoming | outgoing | unknown
// param count_typpe: packet count type, 1: init, 2:increment
static __always_inline int update_conn_stats(struct connection_info *conn, size_t sent_bytes, size_t recv_bytes, u64 ts, int direction,
                                             __u32 packets_out, __u32 packets_in)
{
    struct connection_stats *val = NULL;
    __u32 nat_daddr[4] = {};
    __u16 nat_dport = 0;

    do_dnapt(conn, nat_daddr, &nat_dport);

    // initialize-if-no-exist the connection stat, and load it
    struct connection_stats empty = {0};
    bpf_map_update_elem(&bpfmap_conn_stats, conn, &empty, BPF_NOEXIST);
    val = bpf_map_lookup_elem(&bpfmap_conn_stats, conn);
    if (fill_conn_stats(val, conn, sent_bytes, recv_bytes, ts, direction, packets_out, packets_in) != 0)
    {
        record_netflow_update_fail(NETFLOW_UPDATE_FAIL_CONN_STATS);
        return -1;
    }
    if (val != NULL)
    {
        __builtin_memcpy(val->nat_daddr, nat_daddr, sizeof(nat_daddr));
        val->nat_dport = nat_dport;
    }
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
        }
        event->conn_tcp_stats.state_transitions |= (1 << TCP_CLOSE);
        bpf_map_delete_elem(&bpfmap_conn_tcp_stats, &conn_info);
        bpf_map_delete_elem(&bpfmap_conn_tcp_segments, &conn_info);
        conn_info.pid = pid;
    }

    update_conn_stats(&conn_info, 0, 0, 0, CONN_DIRECTION_AUTO, 0, 0);
    struct connection_stats *conn_sts = bpf_map_lookup_elem(&bpfmap_conn_stats, &conn_info);
    if (conn_sts != NULL)
    {
        event->conn_stats = *conn_sts;
    }
    else
    {
        __u64 ts = bpf_ktime_get_ns();
        fill_conn_stats(&event->conn_stats, &conn_info, 0, 0, ts, CONN_DIRECTION_AUTO, 0, 0);
    }

    bpf_map_delete_elem(&bpfmap_conn_stats, &conn_info);
}

#endif // !__BPFMAP_H
