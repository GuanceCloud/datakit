#if !defined(__PRINT_APUFLOW_H__)
#define __PRINT_APUFLOW_H__

#include "../netflow/netflow_utils.h"

static __always_inline void print_seq(struct tcp_sock *t)
{
#ifdef __DK_DEBUG__
    __u32 rcv_nxt = 0;
    bpf_probe_read(&rcv_nxt, sizeof(rcv_nxt), &t->rcv_nxt);

    __u32 copied_seq = 0;
    bpf_probe_read(&copied_seq, sizeof(copied_seq), &t->copied_seq);

    __u32 rcv_wup = 0;
    bpf_probe_read(&rcv_wup, sizeof(rcv_wup), &t->rcv_wup);

    __u32 snd_nxt = 0;
    bpf_probe_read(&snd_nxt, sizeof(snd_nxt), &t->snd_nxt);

    __u64 bytes_sent = 0;
    bpf_probe_read(&bytes_sent, sizeof(bytes_sent), &t->bytes_sent);

    __u64 bytes_acked = 0;
    bpf_probe_read(&bytes_acked, sizeof(bytes_acked), &t->bytes_acked);

    __u32 snd_una = 0;
    bpf_probe_read(&snd_una, sizeof(snd_una), &t->snd_una);

    bpf_printk("================================");
    bpf_printk("rcv_nxt %u", rcv_nxt);
    bpf_printk("copied_seq %u", copied_seq);
    bpf_printk("rcv_wup %u", rcv_wup);
    bpf_printk("snd_nxt %u", snd_nxt);

#endif
}

static __always_inline void print_conn(__u64 pid_tgid, struct connection_info *info, int rw)
{
#ifdef __DK_DEBUG__
    __u8 saddr[4] = {0};
    __builtin_memcpy(&saddr, info->saddr, sizeof(saddr));
    __u8 daddr[4] = {0};
    __builtin_memcpy(&daddr, info->daddr, sizeof(daddr));
    __u32 s = saddr[3] + saddr[2] * 1000 + saddr[1] * 1000000 + saddr[0] * 1000000000;
    __u32 d = daddr[3] + daddr[2] * 1000 + daddr[1] * 1000000 + daddr[0] * 1000000000;

    if (rw == 0)
    {
        bpf_printk("r pid:%d, tid:%d", pid_tgid >> 32, pid_tgid << 32 >> 32);
        bpf_printk("r daddr: %d:%d", s, info->sport);
        bpf_printk("r daddr: %d:%d", d, info->dport);
    }
    else
    {
        bpf_printk("w pid:%d, tid:%d", pid_tgid >> 32, pid_tgid << 32 >> 32);
        bpf_printk("w daddr: %d:%d", s, info->sport);
        bpf_printk("w daddr: %d:%d", d, info->dport);
    }
#endif
}

#endif // __PRINT_APUFLOW_H__
