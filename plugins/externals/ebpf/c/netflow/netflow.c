#include <linux/kconfig.h>
#include <linux/tcp.h>
#include <net/flow.h>
#include <net/inet_sock.h>
#include <net/sock.h>
#include <net/net_namespace.h>
#include <uapi/linux/udp.h>

#include "bpf_helpers.h"

#include "conn_stats.h"
#include "bpfmap.h"
#include "utils.h"

// ------------------------------------------------------
// --------------- kprobe / kretprobe -------------------

SEC("kprobe/sockfd_lookup_light")
int kprobe__sockfd_lookup_light(struct pt_regs *ctx)
{
    int sockfd = (int)PT_REGS_PARM1(ctx);
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct pid_fd pidfd =
        {
            .pid = pid_tgid >> 32,
            .fd = sockfd,
        };
    struct sock **sk = bpf_map_lookup_elem(&bpfmap_sockfd, &pidfd);
    if (sk != NULL)
    {
        return 0;
    }
    bpf_map_update_elem(&bpfmap_tmp_sockfdlookuplight, &pid_tgid, &sockfd, BPF_ANY);
    return 0;
}

SEC("kretprobe/sockfd_lookup_light")
int kretprobe__sockfd_lookup_light(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    int *sockfd = bpf_map_lookup_elem(&bpfmap_tmp_sockfdlookuplight, &pid_tgid);
    if (sockfd == NULL)
    {
        return 0;
    }
    struct socket *skt = (struct socket *)PT_REGS_RC(ctx);
    enum sock_type sktype = 0;
    if (sktype == SOCK_STREAM)
    { // TCP socket
        struct sock *sk = NULL;
        bpf_probe_read(&sk, sizeof(sk), &skt->sk);
        struct pid_fd pidfd = {
            .pid = pid_tgid >> 32,
            .fd = *sockfd,
        };
        bpf_map_update_elem(&bpfmap_sockfd, &pidfd, sk, BPF_ANY);
        bpf_map_update_elem(&bpfmap_sockfd_inverted, sk, &pidfd, BPF_ANY);
    }
    bpf_map_delete_elem(&bpfmap_tmp_sockfdlookuplight, &pid_tgid);
    return 0;
}

SEC("kprobe/do_sendfile")
int kprobe__do_sendfile(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 fdout = (int)PT_REGS_PARM1(ctx);
    struct pid_fd pidfd = {
        .pid = pid_tgid >> 32,
        .fd = fdout,
    };
    struct sock **skpp = bpf_map_lookup_elem(&bpfmap_sockfd, &pidfd);
    if (skpp == NULL)
    {
        return 0;
    }
    struct sock *sk = *skpp;
    bpf_map_update_elem(&bpfmap_tmp_sendfile, &pid_tgid, &sk, BPF_ANY);
    return 0;
}

SEC("kretprobe/do_sendfile")
int kretprobe__do_sendfile(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 ktime_ns = bpf_ktime_get_ns();
    struct sock **skpp = (struct sock **)bpf_map_lookup_elem(&bpfmap_tmp_sendfile, &pid_tgid);
    if (skpp == NULL)
    {
        return 0;
    }
    struct connection_info conninf = {};
    if (read_connection_info(*skpp, &conninf, pid_tgid, CONN_L4_TCP) == 0)
    {
        size_t sent = (size_t)PT_REGS_RC(ctx);
        update_conn_stats(&conninf, sent, 0, ktime_ns, CONN_DIRECTION_AUTO, 0, 0, -1);
    }
    bpf_map_delete_elem(&bpfmap_tmp_sendfile, &pid_tgid);
    return 0;
}

// ===============================================

// TCP_ESTABLISHED；
// 记录 TCP_CLOSE 后将导致 kprobe__tcp_close 清除的 tcp 连接信息被重新写入 bpfmap,
// https://elixir.bootlin.com/linux/latest/source/net/ipv4/tcp.c#L2707 .
SEC("kprobe/tcp_set_state")
int kprobe__tcp_set_state(struct pt_regs *ctx)
{
    __u8 state = (__u8)PT_REGS_PARM2(ctx);
    if (state == TCP_ESTABLISHED)
    {
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        __u64 pid_tgid = bpf_get_current_pid_tgid();
        struct connection_info conninf = {};
        if (read_connection_info(sk, &conninf, pid_tgid, CONN_L4_TCP) != 0)
        {
            return 0;
        }
        struct connection_tcp_stats tcpstats = {};
        __builtin_memset(&tcpstats, 0, sizeof(struct connection_tcp_stats));
        tcpstats.state_transitions = (1 << state);
        read_tcp_rtt(sk, &tcpstats);
        update_tcp_stats(conninf, tcpstats);
    }
    return 0;
}

//  port state listening (tcp)
// func inet_csk_accept return null or sock pointer
SEC("kretprobe/inet_csk_accept")
int kretprobe__inet_csk_accept(struct pt_regs *ctx)
{
    __u64 ktime_ns = bpf_ktime_get_ns();

    struct sock *sk = (struct sock *)PT_REGS_RC(ctx);
    if (sk == NULL)
    {
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct connection_info conninf = {};
    if (read_connection_info(sk, &conninf, pid_tgid, CONN_L4_TCP != 0))
    {
        return 0;
    }

    struct connection_tcp_stats tcpstat = {};
    __builtin_memset(&tcpstat, 0, sizeof(struct connection_tcp_stats));

    read_tcp_rtt(sk, &tcpstat);
    update_tcp_stats(conninf, tcpstat);

    update_conn_stats(&conninf, 0, 0, ktime_ns, CONN_DIRECTION_INCOMING, 0, 0, 1);

    struct port_bind pb = {
        .netns = conninf.netns,
        .port = conninf.sport,
    };

    __u8 state = PORT_LISETINING;
    bpf_map_update_elem(&bpfmap_port_bind, &pb, &state, BPF_NOEXIST);

    return 0;
}

SEC("kprobe/inet_csk_listen_stop")
int kprobe__inet_csk_listen_stop(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    __u16 port = read_sock_src_port(sk);
    if (port == 0)
    {
        return 0;
    }
    struct port_bind pb = {};
    pb.netns = read_netns(&sk->sk_net);
    pb.port = port;

    bpf_map_delete_elem(&bpfmap_port_bind, &pb);
    return 0;
}

// tcp_close
SEC("kprobe/tcp_close")
int kprobe__tcp_close(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);

    if (sk == NULL)
    {
        return 0;
    }

    // clear bpfmap_sockfd
    struct pid_fd *pidfd = (struct pid_fd *)bpf_map_lookup_elem(&bpfmap_sockfd_inverted, &sk);
    if (pidfd != NULL)
    {
        struct pid_fd pf = {};                             // for linux4.4
        bpf_probe_read(&pf, sizeof(struct pid_fd), pidfd); // for linux4.4
        bpf_map_delete_elem(&bpfmap_sockfd, &pf);
        bpf_map_delete_elem(&bpfmap_sockfd_inverted, &sk);
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct connection_info conn;
    __builtin_memset(&conn, 0, sizeof(conn));

    if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_TCP) != 0)
    {
        return 0;
    }
    struct connection_closed_info event = {};
    remove_from_conn_map(conn, &event);
    __u64 cpu = bpf_get_smp_processor_id();
    send_conn_closed_event(ctx, event, cpu);
    return 0;
}

SEC("kprobe/tcp_retransmit_skb")
int kprobe__tcp_retransmit_skb(struct pt_regs *ctx)
{
    // https://elixir.bootlin.com/linux/v4.6.7/source/include/net/tcp.h#L537
    int pre_4_7_0 = pre_kernel_4_7_0();
    if (pre_4_7_0 == 0)
    {

        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        int segs = (int)PT_REGS_PARM3(ctx);
        struct connection_info conn = {};
        read_connection_info(sk, &conn, 0, CONN_L4_TCP);
        update_tcp_retransmit(conn, segs);
    }
    else
    {
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        struct connection_info conn = {};
        read_connection_info(sk, &conn, 0, CONN_L4_TCP);
        update_tcp_retransmit(conn, 1);
    }
    return 0;
}

SEC("kprobe/tcp_sendmsg")
int kprobe__tcp_sendmsg(struct pt_regs *ctx)
{

    // https://elixir.bootlin.com/linux/v4.0/source/include/net/tcp.h#L352
    int pre_4_1_0 = pre_kernel_4_1_0();
    if (pre_4_1_0 == 0)
    {
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        size_t size = (size_t)PT_REGS_PARM3(ctx);

        __u64 pid_tgid = bpf_get_current_pid_tgid();

        // init connection info struct
        struct connection_info conn_info = {};
        if (read_connection_info(sk, &conn_info, pid_tgid, CONN_L4_TCP) != 0)
        {
            return 0;
        }

        // packets in & out
        __u32 packets_in = 0;
        __u32 packets_out = 0;
        read_tcp_segment_counts(sk, &packets_in, &packets_out);

        __u64 ts = bpf_ktime_get_ns();

        struct connection_tcp_stats tcp_stats = {};
        __builtin_memset(&tcp_stats, 0, sizeof(struct connection_tcp_stats));

        read_tcp_rtt(sk, &tcp_stats);
        update_tcp_stats(conn_info, tcp_stats);

        update_conn_stats(&conn_info, size, 0, ts, CONN_DIRECTION_AUTO, packets_out, packets_in, 1);
    }
    else
    {
        struct sock *sk = (struct sock *)PT_REGS_PARM2(ctx);
        size_t size = (size_t)PT_REGS_PARM4(ctx);

        __u64 pid_tgid = bpf_get_current_pid_tgid();

        // init connection info struct
        struct connection_info conn_info = {};
        if (read_connection_info(sk, &conn_info, pid_tgid, CONN_L4_TCP) != 0)
        {
            return 0;
        }

        // packets in & out
        __u32 packets_in = 0;
        __u32 packets_out = 0;
        read_tcp_segment_counts(sk, &packets_in, &packets_out);

        __u64 ts = bpf_ktime_get_ns();

        struct connection_tcp_stats tcp_stats = {};
        __builtin_memset(&tcp_stats, 0, sizeof(struct connection_tcp_stats));

        read_tcp_rtt(sk, &tcp_stats);
        update_tcp_stats(conn_info, tcp_stats);

        update_conn_stats(&conn_info, size, 0, ts, CONN_DIRECTION_AUTO, packets_out, packets_in, 1);
    }

    return 0;
}

// The function tcp_cleanup_rbuf is called by functions such as
// tcp_read_sock and tcp_recvmsg
SEC("kprobe/tcp_cleanup_rbuf")
int kprobe__tcp_cleanup_buf(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    int copied = (int)PT_REGS_PARM2(ctx);
    if (copied <= 0)
    {
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct connection_info conn_info = {};
    if (read_connection_info(sk, &conn_info, pid_tgid, CONN_L4_TCP) != 0)
    {
        return 0;
    }

    __u32 packets_in = 0;
    __u32 packets_out = 0;
    read_tcp_segment_counts(sk, &packets_in, &packets_out);

    struct connection_tcp_stats tcp_stats = {};
    __builtin_memset(&tcp_stats, 0, sizeof(struct connection_tcp_stats));
    read_tcp_rtt(sk, &tcp_stats);
    update_tcp_stats(conn_info, tcp_stats);

    __u64 ts = bpf_ktime_get_ns();
    update_conn_stats(&conn_info, 0, copied, ts, CONN_DIRECTION_AUTO, packets_out, packets_in, 1);

    return 0;
}

// ===============================================
SEC("kprobe/ip_make_skb")
int kprobe__ip_make_skb(struct pt_regs *ctx)
{
    __u64 ts = bpf_ktime_get_ns();
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    size_t size = (size_t)PT_REGS_PARM5(ctx);
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    size -= sizeof(struct udphdr);
    struct connection_info conninf = {};
    if (read_connection_info(sk, &conninf, pid_tgid, CONN_L4_UDP) != 0)
    {
        __u64 offset_flowi4_daddr = load_offset_flowi4_daddr();
        __u64 offset_flowi4_saddr = load_offset_flowi4_saddr();

        struct flowi4 *fl4 = (struct flowi4 *)PT_REGS_PARM2(ctx);
        // saddr: fl4->saddr, daddr: fl4->daddr
        bpf_probe_read(conninf.saddr + 3, sizeof(__be32), (__u8 *)fl4 + offset_flowi4_saddr);
        bpf_probe_read(conninf.daddr + 3, sizeof(__be32), (__u8 *)fl4 + offset_flowi4_daddr);
        if ((conninf.saddr[3] | conninf.daddr[3]) == 0)
        {
            return 0;
        }

        __u64 offset_flowi4_dport = load_offset_flowi4_dport();
        __u64 offset_flowi4_sport = load_offset_flowi4_sport();
        // sport: fl4->fl4_sport, dport: fl4->fl4_dport
        bpf_probe_read(&conninf.sport, sizeof(__be16), (__u8 *)fl4 + offset_flowi4_sport);
        bpf_probe_read(&conninf.dport, sizeof(__be16), (__u8 *)fl4 + offset_flowi4_dport);
        if ((conninf.sport | conninf.dport) == 0)
        {
            return 0;
        }
        swap_u16(&conninf.sport);
        swap_u16(&conninf.dport);
    }
    update_conn_stats(&conninf, size, 0, ts, CONN_DIRECTION_AUTO, 1, 0, 2);
    return 0;
}

SEC("kprobe/ip6_make_skb")
int kprobe__ip6_make_skb(struct pt_regs *ctx)
{
    // https://elixir.bootlin.com/linux/v4.6.7/source/net/ipv6/ip6_output.c#L1743
    int pre_4_7_0 = pre_kernel_4_7_0();
    if (pre_4_7_0 == 0)
    {
        __u64 ts = bpf_ktime_get_ns();
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        size_t size = (size_t)PT_REGS_PARM4(ctx);
        __u64 pid_tgid = bpf_get_current_pid_tgid();
        size -= sizeof(struct udphdr);
        struct connection_info conn;
        __builtin_memset(&conn, 0, sizeof(conn));
        if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_UDP) != 0)
        {
            __u64 offset_flowi6_daddr = load_offset_flowi6_daddr();
            __u64 offset_flowi6_saddr = load_offset_flowi6_saddr();
            struct flowi6 *fl6 = (struct flowi6 *)PT_REGS_PARM7(ctx);
            bpf_probe_read(&conn.daddr, sizeof(__u32) * 4, (__u8 *)fl6 + offset_flowi6_daddr);
            bpf_probe_read(&conn.saddr, sizeof(__u32) * 4, (__u8 *)fl6 + offset_flowi6_saddr);
            if (((conn.saddr[0] | conn.saddr[1] | conn.saddr[2] | conn.saddr[3]) |
                 (conn.daddr[0] | conn.daddr[1] | conn.daddr[2] | conn.daddr[3])) == 0)
            {
                return 0;
            }
            __u64 offset_flowi6_dport = load_offset_flowi6_dport();
            __u64 offset_flowi6_sport = load_offset_flowi6_sport();
            bpf_probe_read(&conn.dport, sizeof(__u32), (__u8 *)fl6 + offset_flowi6_dport);
            bpf_probe_read(&conn.sport, sizeof(__u32), (__u8 *)fl6 + offset_flowi6_sport);
            swap_u16(&conn.sport);
            swap_u16(&conn.dport);
            update_conn_stats(&conn, size, 0, ts, CONN_DIRECTION_AUTO, 1, 0, 2);
        }
    }
    else
    {
        __u64 ts = bpf_ktime_get_ns();
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        size_t size = (size_t)PT_REGS_PARM4(ctx);
        __u64 pid_tgid = bpf_get_current_pid_tgid();
        size -= sizeof(struct udphdr);
        struct connection_info conn;
        __builtin_memset(&conn, 0, sizeof(conn));
        if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_UDP) != 0)
        {
            __u64 offset_flowi6_daddr = load_offset_flowi6_daddr();
            __u64 offset_flowi6_saddr = load_offset_flowi6_saddr();
            struct flowi6 *fl6 = (struct flowi6 *)PT_REGS_PARM9(ctx);
            bpf_probe_read(&conn.daddr, sizeof(__u32) * 4, (__u8 *)fl6 + offset_flowi6_daddr);
            bpf_probe_read(&conn.saddr, sizeof(__u32) * 4, (__u8 *)fl6 + offset_flowi6_saddr);
            if (((conn.saddr[0] | conn.saddr[1] | conn.saddr[2] | conn.saddr[3]) |
                 (conn.daddr[0] | conn.daddr[1] | conn.daddr[2] | conn.daddr[3])) == 0)
            {
                return 0;
            }
            __u64 offset_flowi6_dport = load_offset_flowi6_dport();
            __u64 offset_flowi6_sport = load_offset_flowi6_sport();
            bpf_probe_read(&conn.dport, sizeof(__u32), (__u8 *)fl6 + offset_flowi6_dport);
            bpf_probe_read(&conn.sport, sizeof(__u32), (__u8 *)fl6 + offset_flowi6_sport);
            swap_u16(&conn.sport);
            swap_u16(&conn.dport);
            update_conn_stats(&conn, size, 0, ts, CONN_DIRECTION_AUTO, 1, 0, 2);
        }
    }

    return 0;
}

SEC("kprobe/udp_recvmsg")
int kprobe__udp_recvmsg(struct pt_regs *ctx)
{
    // https://elixir.bootlin.com/linux/v4.0/source/net/ipv4/udp.c#L1257
    int pre_4_1_0 = pre_kernel_4_1_0();
    if (pre_4_1_0 == 0)
    {
        __u64 pid_tgid = bpf_get_current_pid_tgid();
        struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
        struct msghdr *msg = (struct msghdr *)PT_REGS_PARM2(ctx);
        int flag = (int)PT_REGS_PARM5(ctx);
        if (flag & MSG_PEEK)
        {
            return 0;
        }
        struct udp_revcmsg_tmp rcvd = {
            .sk = NULL,
            .msg = NULL,
        };
        if (sk != NULL)
        {
            bpf_probe_read(&rcvd.sk, sizeof(struct sock *), &sk);
        }
        if (msg != NULL)
        {
            bpf_probe_read(&rcvd.msg, sizeof(struct msghdr *), &msg);
        }
        bpf_map_update_elem(&bpf_map_tmp_udprecvmsg, &pid_tgid, &rcvd, BPF_ANY);
    }
    else
    {
        __u64 pid_tgid = bpf_get_current_pid_tgid();
        struct sock *sk = (struct sock *)PT_REGS_PARM2(ctx);
        struct msghdr *msg = (struct msghdr *)PT_REGS_PARM3(ctx);
        int flag = (int)PT_REGS_PARM6(ctx);
        if (flag & MSG_PEEK)
        {
            return 0;
        }
        struct udp_revcmsg_tmp rcvd = {
            .sk = NULL,
            .msg = NULL,
        };
        if (sk != NULL)
        {
            bpf_probe_read(&rcvd.sk, sizeof(struct sock *), &sk);
        }
        if (msg != NULL)
        {
            bpf_probe_read(&rcvd.msg, sizeof(struct msghdr *), &msg);
        }
        bpf_map_update_elem(&bpf_map_tmp_udprecvmsg, &pid_tgid, &rcvd, BPF_ANY);
    }
    return 0;
}

// https://elixir.bootlin.com/linux/latest/source/net/ipv4/udp.c#L1834
SEC("kretprobe/udp_recvmsg")
int kretprobe__udp_recvmsg(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    int copied = (int)PT_REGS_RC(ctx);
    if (copied < 0)
    {
        return 0;
    }
    struct udp_revcmsg_tmp *rcvd = NULL;
    rcvd = (struct udp_revcmsg_tmp *)bpf_map_lookup_elem(&bpf_map_tmp_udprecvmsg, &pid_tgid);
    if (rcvd == NULL)
    {
        return 0;
    }
    bpf_map_delete_elem(&bpf_map_tmp_udprecvmsg, &pid_tgid);

    struct connection_info conn = {};

    if (read_connection_info(rcvd->sk, &conn, pid_tgid, CONN_L4_UDP) != 0)
    {
        return 0;
    }

    struct sockaddr *ska = NULL;
    if (rcvd->msg != NULL)
    {
        bpf_probe_read(&ska, sizeof(struct sockaddr *), &(rcvd->msg->msg_name));
        if (ska != NULL)
        {
            if ((conn.meta & CONN_L3_MASK) == CONN_L3_IPv4)
            {
                bpf_probe_read(conn.daddr + 3, sizeof(__be32), &(((struct sockaddr_in *)ska)->sin_addr.s_addr));
                bpf_probe_read(&conn.dport, sizeof(__be16), &(((struct sockaddr_in *)ska)->sin_port));
                swap_u16(&conn.dport);
            }
            else
            {
                bpf_probe_read(conn.daddr, sizeof(__u32) * 4, &(((struct sockaddr_in6 *)ska)->sin6_addr));
                bpf_probe_read(&conn.dport, sizeof(__be16), &(((struct sockaddr_in6 *)ska)->sin6_port));
                swap_u16(&conn.dport);
            }
        }
    }

    update_conn_stats(&conn, 0, copied, ts, CONN_DIRECTION_AUTO, 0, 1, 2);
    return 0;
}

SEC("kprobe/inet_bind")
int kprobe__inet_bind(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct socket *sckt = (struct socket *)PT_REGS_PARM1(ctx);

    __u16 sk_type = 0;
    if (sckt == NULL)
    {
        return 0;
    }

    bpf_probe_read(&sk_type, sizeof(sk_type), &sckt->type);

    if ((sk_type & SOCK_DGRAM) == 0)
    { // not connectionless(UDP socket)
        return 0;
    }

    __u16 bindport = 0;
    struct sockaddr_in *sckaddr = (struct sockaddr_in *)PT_REGS_PARM2(ctx);
    bpf_probe_read(&bindport, sizeof(__be16), &sckaddr->sin_port);

    swap_u16(&bindport); // port type __be16
    if (bindport == 0)
    {
        return 0;
    }

    bpf_map_update_elem(&bpfmap_tmp_inetbind, &pid_tgid, &bindport, BPF_ANY);

    return 0;
}

SEC("kretprobe/inet_bind")
int kretprobe__inet_bind(struct pt_regs *ctx)
{
    __s64 ret = (__s64)PT_REGS_RC(ctx); // inet_bind return value
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u16 *port = bpf_map_lookup_elem(&bpfmap_tmp_inetbind, &pid_tgid);
    if (port == NULL)
    {
        return 0;
    }
    bpf_map_delete_elem(&bpfmap_tmp_inetbind, &pid_tgid);
    if (ret != 0)
    {
        return 0;
    }
    struct port_bind pbind = {};
    pbind.netns = 0;
    pbind.port = *port;
    __u8 state = PORT_LISETINING;
    bpf_map_update_elem(&bpfmap_udp_port_bind, &pbind, &state, BPF_ANY);
    return 0;
}

SEC("kprobe/inet6_bind")
int kprobe__inet6_bind(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct socket *sckt = (struct socket *)PT_REGS_PARM1(ctx);

    __u16 sk_type = 0;
    if (sckt == NULL)
    {
        return 0;
    }

    bpf_probe_read(&sk_type, sizeof(sk_type), &sckt->type);

    if ((sk_type & SOCK_DGRAM) == 0)
    { // not connectionless(UDP socket)
        return 0;
    }

    __u16 bindport = 0;
    struct sockaddr_in6 *sckaddr = (struct sockaddr_in6 *)PT_REGS_PARM2(ctx);
    bpf_probe_read(&bindport, sizeof(__be16), &sckaddr->sin6_port);

    swap_u16(&bindport); // port type __be16
    if (bindport == 0)
    {
        return 0;
    }

    bpf_map_update_elem(&bpfmap_tmp_inetbind, &pid_tgid, &bindport, BPF_ANY);

    return 0;
}

SEC("kretprobe/inet6_bind")
int kretprobe__inet6_bind(struct pt_regs *ctx)
{
    __s64 ret = PT_REGS_RC(ctx); // inet_bind return value
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u16 *port = bpf_map_lookup_elem(&bpfmap_tmp_inetbind, &pid_tgid);
    if (port == NULL)
    {
        return 0;
    }
    bpf_map_delete_elem(&bpfmap_tmp_inetbind, &pid_tgid);
    if (ret != 0)
    {
        return 0;
    }
    struct port_bind pbind = {};
    pbind.netns = 0;
    pbind.port = *port;
    __u8 state = PORT_LISETINING;
    bpf_map_update_elem(&bpfmap_udp_port_bind, &pbind, &state, BPF_ANY);
    return 0;
}

SEC("kprobe/udp_destroy_sock")
int kprobe__udp_destroy_sock(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    struct connection_info conn = {};
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u16 port = 0;
    if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_UDP) != 0)
    {
        port = read_sock_src_port(sk);
        goto clear;
    }
    struct connection_closed_info event = {};
    remove_from_conn_map(conn, &event);
    __u64 cpu = bpf_get_smp_processor_id();
    send_conn_closed_event(ctx, event, cpu);
    port = event.conn_info.sport;
clear:
    if (port == 0)
    {
        return 0;
    }
    struct port_bind pb = {};
    pb.netns = 0;
    pb.port = port;
    bpf_map_delete_elem(&bpfmap_udp_port_bind, &pb);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
