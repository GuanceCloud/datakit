#ifndef __PACKET_H_
#define __PACKET_H_

#include <uapi/linux/if_ether.h>
#include <uapi/linux/in.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/ipv6.h>
#include <uapi/linux/udp.h>
#include <uapi/linux/tcp.h>

#include <net/tcp.h>

#include "../netflow/conn_stats.h"
#include "bpf_helpers.h"

#define TCP_FIN 0b000001
#define TCP_SYN 0b000010
#define TCP_RST 0b000100
#define TCP_PSH 0b001000
#define TCP_ACK 0b010000
#define TCP_URG 0b100000

struct conn_skb_l4_info
{
    __u16 hdr_len;
    __u16 offset;

    __u8 tcp_flags; // 6bits
    __u8 scale;

    __u16 wnd;

    __u32 seg_seq; // seq
    __u32 seg_ack; // ack
};

static __always_inline __u8 tcp_control_flag(__u8 tcp_ctrl, __u8 flag)
{
    return (tcp_ctrl & flag) == flag;
}

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
        // https://datatracker.ietf.org/doc/html/rfc793#section-3.1
        conn_info->meta = (conn_info->meta & ~CONN_L4_MASK) | CONN_L4_TCP;

        conn_info->sport = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, source));
        conn_info->dport = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, dest));

        // seq and ack
        skbinfo->seg_seq = load_word(skb, skbinfo->hdr_len + offsetof(struct tcphdr, seq));
        skbinfo->seg_ack = load_word(skb, skbinfo->hdr_len + offsetof(struct tcphdr, ack_seq));

        // load_half 将交换字节序，如 be16_to_cpu
        __u16 doff_and_flags = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, ack_seq) + 4);

        // + tcp data offset (doff * 4bytes: tcp_hdr_len + tcp_opt_len)
        // tcp_hdr_len = doff << 2
        skbinfo->tcp_flags = doff_and_flags & 0x3F;

        int tcp_hdr_len = (doff_and_flags & 0xF000) >> (12 - 2);

        if (tcp_hdr_len <= 20)
        {
            goto skip_option;
        }

        int opt_len = skbinfo->hdr_len + 20;
        int hdrLen = skbinfo->hdr_len + tcp_hdr_len;

        // tcp option
#pragma unroll
        for (int i = 0; i < 40; i++)
        {
            if (opt_len >= hdrLen)
            {
            }
            else
            {
                __u8 opt = load_byte(skb, opt_len);

                if (opt == TCPOPT_NOP || opt == TCPOPT_EOL)
                {
                    // https://www.rfc-editor.org/rfc/rfc9293.html#appendix-B-1
                    opt_len++;
                }
                else if (opt == TCPOPT_WINDOW)
                {
                    skbinfo->scale = (__u8)load_byte(skb, opt_len + 2);
                    // https://www.iana.org/assignments/tcp-parameters/tcp-parameters.xhtml
                    opt_len += 3;
                }
                else
                {
                    __u8 l = 0;
                    l = load_byte(skb, opt_len + 1);
                    opt_len += l;
                }
            }
        }

    skip_option:
        // rcv window
        skbinfo->wnd = load_half(skb, skbinfo->hdr_len + offsetof(struct tcphdr, window));

        // do that last
        skbinfo->hdr_len += tcp_hdr_len;
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

#ifdef __DK_DEBUG__
    if (tcp_control_flag(skbinfo->tcp_flags, TCP_FIN | TCP_ACK) == 1)
    {
        bpf_printk("FIN|ACK saddr %x %d", conn_info->saddr[3], conn_info->sport);
        bpf_printk("FIN|ACK daddr %x %d", conn_info->daddr[3], conn_info->dport);
        bpf_printk("FIN|ACK seq %u ack %u", skbinfo->seg_seq, skbinfo->seg_ack);
        bpf_printk("FIN|ACK flag %x, len %d", skbinfo->tcp_flags, skbinfo->hdr_len);
    }
    else if (tcp_control_flag(skbinfo->tcp_flags, TCP_FIN) == 1)
    {
        bpf_printk("FIN saddr %x %d", conn_info->saddr[3], conn_info->sport);
        bpf_printk("FIN daddr %x %d", conn_info->daddr[3], conn_info->dport);
        bpf_printk("FIN seq %u ack %u", skbinfo->seg_seq, skbinfo->seg_ack);
        bpf_printk("FIN flag %x, len %d", skbinfo->tcp_flags, skbinfo->hdr_len);
    }
    else if (tcp_control_flag(skbinfo->tcp_flags, TCP_RST | TCP_ACK) == 1)
    {
        bpf_printk("RST|ACK saddr %x %d", conn_info->saddr[3], conn_info->sport);
        bpf_printk("RST|ACK daddr %x %d", conn_info->daddr[3], conn_info->dport);
        bpf_printk("RST|ACK seq %u ack %u", skbinfo->seg_seq, skbinfo->seg_ack);
        bpf_printk("RST|ACK flag %x, len %d", skbinfo->tcp_flags, skbinfo->hdr_len);
    }
    else if (tcp_control_flag(skbinfo->tcp_flags, TCP_RST) == 1)
    {
        bpf_printk("RST saddr %x %d", conn_info->saddr[3], conn_info->sport);
        bpf_printk("RST daddr %x %d", conn_info->daddr[3], conn_info->dport);
        bpf_printk("RST seq %u ack %u", skbinfo->seg_seq, skbinfo->seg_ack);
        bpf_printk("RST flag %x, len %d", skbinfo->tcp_flags, skbinfo->hdr_len);
    }
#endif

    return 0;
}
#endif