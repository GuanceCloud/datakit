#include <uapi/linux/ptrace.h>
#include <linux/tcp.h>

#include "bpf_helpers.h"
#include "../netflow/conn_stats.h"
#include "../netflow/netflow_utils.h"
#include "offset.h"
#include "../apiflow/packet.h"
#include "filter.h"

BPF_HASH_MAP(bpfmap_packet_tuple, struct packet_tuple,
             struct packet_info, 65536);

BPF_HASH_MAP(bpfmap_offset_tcp_seq,
             __u64, struct offset_tcp_seq, 4096);

static __always_inline int read_offset_tcp_seq(struct offset_tcp_seq *dst)
{
    __u64 key = 0;
    struct offset_tcp_seq *ptr =
        (struct offset_tcp_seq *)bpf_map_lookup_elem(&bpfmap_offset_tcp_seq, &key);
    if (ptr == NULL)
    {
        return -1;
    }
    bpf_probe_read(dst, sizeof(struct offset_tcp_seq), ptr);
    return 0;
}

static __always_inline void write_offset_tcp_seq(struct offset_tcp_seq *offset)
{
    __u64 key = 0;
    bpf_map_update_elem(&bpfmap_offset_tcp_seq, &key, offset, BPF_ANY);
}

SEC("socket/packet_tcp_header")
int socket__packet_tcp_header(struct __sk_buff *skb)
{
    struct conn_skb_l4_info skbinfo = {0};
    struct connection_info conn_info = {0};

    if (read_connection_info_skb(skb, &skbinfo, &conn_info) != 0)
    {
        goto out;
    }

    // ipv4 only
    if ((conn_info.meta & CONN_L3_MASK) != CONN_L3_IPv4)
    {
        return 0;
    }

    // the ip address of the tcp server is 127.0.0.1
    if (conn_info.daddr[3] != 0x0200007F && conn_info.saddr[3] != 0x0200007F)
    {
        return 0;
    }
    struct packet_tuple key = {0};
    key.src_ip[3] = conn_info.saddr[3];
    key.dst_ip[3] = conn_info.daddr[3];
    key.src_port = conn_info.sport;
    key.dst_port = conn_info.dport;

    struct packet_info *pkt_inf = NULL;

    switch (skbinfo.tcp_flags)
    {
    case TCP_SYN:

        pkt_inf = bpf_map_lookup_elem(&bpfmap_packet_tuple, &key);

        if (pkt_inf != NULL)
        {
            return 0;
        }

        struct packet_info pkt = {0};

        // 3whs, client sends SYN
        pkt.ctrl_syn = 1;

        pkt.ack = skbinfo.seg_ack;
        pkt.seq = skbinfo.seg_seq;
        pkt.rcv_wnd = skbinfo.wnd;
        pkt.scale = skbinfo.scale;

        bpf_map_update_elem(&bpfmap_packet_tuple, &key, &pkt, BPF_ANY);
        break;
    case TCP_SYN | TCP_ACK:
        // reversal
        key.src_ip[3] = conn_info.daddr[3];
        key.dst_ip[3] = conn_info.saddr[3];
        key.src_port = conn_info.dport;
        key.dst_port = conn_info.sport;

        pkt_inf = bpf_map_lookup_elem(&bpfmap_packet_tuple, &key);

        if (pkt_inf == NULL || pkt_inf->ctrl_syn_ack == 1)
        {
            return 0;
        }

        pkt_inf->ctrl_syn_ack = 1;

        break;
    case TCP_ACK:
        pkt_inf = bpf_map_lookup_elem(&bpfmap_packet_tuple, &key);

        if (pkt_inf == NULL || pkt_inf->ctrl_ack == 1)
        {
            return 0;
        }

        pkt_inf->ctrl_ack = 1;

        // update
        pkt_inf->ack = skbinfo.seg_ack;
        pkt_inf->seq = skbinfo.seg_seq;
        pkt_inf->rcv_wnd = skbinfo.wnd;

        // bpf_printk("ack seq wnd %u %u %u", pkt_inf->ack, pkt_inf->seq, pkt_inf->rcv_wnd);
        break;
    case TCP_FIN:
    case TCP_FIN | TCP_ACK:
    case TCP_RST:
    case TCP_RST | TCP_ACK:
        bpf_map_delete_elem(&bpfmap_packet_tuple, &key);
        break;
    }

out:
    return 0;
}

SEC("kprobe/tcp_getsockopt")
int kprobe__tcp_getsockopt(struct pt_regs *ctx)
{
    void *sk = (void *)PT_REGS_PARM1(ctx);
    int level = (int)PT_REGS_PARM2(ctx);
    int optname = (int)PT_REGS_PARM3(ctx);
    if (level != SOL_TCP || optname != TCP_INFO)
    {
        return 0;
    }

    __u64 pid_tgid = bpf_get_current_pid_tgid();

    if (sk == NULL)
    {
        return 0;
    }

    struct tcp_sock *tp = tcp_sk(sk);

    struct offset_tcp_seq offset = {0};

    // not ready
    if (read_offset_tcp_seq(&offset) != 0)
    {
        return 0;
    }

    // skip
    if (skipConn(offset.process_name, offset.pid_tgid) != 0)
    {
        return 0;
    }

    struct connection_info conn_info = {0};

    read_connection_info(sk, &conn_info, pid_tgid, CONN_L4_TCP);

    struct packet_tuple key = {0};
    key.src_ip[3] = conn_info.saddr[3];
    key.dst_ip[3] = conn_info.daddr[3];
    key.src_port = conn_info.sport;
    key.dst_port = conn_info.dport;

    struct packet_info *pkt = bpf_map_lookup_elem(&bpfmap_packet_tuple, &key);
    if (pkt == NULL)
    {
        return 0;
    }

    __u64 offset_rtt = load_offset_rtt();

    if (offset.offset_copied_seq == 0)
    {
        offset.offset_copied_seq = offset_rtt;
    }

    if (offset.offset_write_seq == 0)
    {
        offset.offset_write_seq = offset_rtt;
    }


    __u64 half_half = 0;
    // low
    // guess copied seq and rcv next
#pragma unroll
    for (int i = 0; i < 10; i++)
    {
        half_half = 0;
        if ((offset.state & 0b1) == 0b1)
        {
            break;
        }
        offset.offset_copied_seq -= 1;
        bpf_probe_read(&half_half, sizeof(half_half), (__u8 *)sk + offset.offset_copied_seq);
        if (half_half >> 32 == pkt->seq && (half_half << 32) >> 32 == pkt->ack)
        {
            __u32 copied_seq = 0;

            __s32 cp = offset.offset_copied_seq - 4;
            bpf_probe_read(&copied_seq, sizeof(copied_seq), (__u8 *)sk + cp);
            if (copied_seq == pkt->ack)
            {
                offset.offset_copied_seq = cp;
                offset.state |= 0b1;
                break;
            }
        }
    }

    __u32 win = pkt->rcv_wnd << pkt->scale;

    // high
    // guess write_seq (and snd_wnd/pushed_seq)
#pragma unroll
    for (int i = 0; i < 10; i++)
    {
        half_half = 0;
        if ((offset.state & 0b10) == 0b10)
        {
            break;
        }
        offset.offset_write_seq += 1;
        bpf_probe_read(&half_half, sizeof(half_half), (__u8 *)sk + offset.offset_write_seq);
        // we need to infer from the context
        if (half_half << 32 >> 32 == win && half_half >> 32 == pkt->seq)
        {
            if (win == 0)
            { // probably not
                half_half = 0;
                bpf_probe_read(&half_half, sizeof(half_half), (__u8 *)sk + (offset.offset_write_seq + 8));
                if (half_half << 32 >> 32 == pkt->seq || half_half >> 32 == pkt->seq)
                {
                    offset.offset_write_seq = offset.offset_write_seq + 4;
                    offset.state |= 0b10;
                    break;
                }
            }
            else
            {
                offset.offset_write_seq = offset.offset_write_seq + 4;
                offset.state |= 0b10;
                break;
            }
        }
    }

    write_offset_tcp_seq(&offset);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
