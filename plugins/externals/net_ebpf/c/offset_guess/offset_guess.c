// #include <linux/kconfig.h>
#include <uapi/linux/ptrace.h>
#include <net/sock.h>

#include "bpf_helpers.h"
#include "offset.h"
#include "bpfmap.h"

static __always_inline void swap_u16(__u16 *v)
{
    *v = ((*v & 0x00FF) << 8) | ((*v & 0xFF00) >> 8);
}

static __always_inline int skipConn(struct offset_guess *status)
{
    char actual[PROCNAMELEN];
    bpf_get_current_comm(actual, PROCNAMELEN);
    // process name only 
    for (int i = 0; i < PROCNAMELEN - 1; i++)
    {
        if (actual[i] != status->process_name[i])
        {
            return 0;
        }
    }
    return 1;
}

static __always_inline int read_offset(struct offset_guess *dst)
{
    __u64 key = 0;
    struct offset_guess *ptr = (struct offset_guess *)bpf_map_lookup_elem(&bpfmap_offset_guess, &key);
    if (ptr == NULL)
    {
        return -1;
    }
    bpf_probe_read(dst, sizeof(struct offset_guess), ptr);

    return 0;
}

static __always_inline int update_offset(struct offset_guess *src)
{
    __u64 key = 0;
    bpf_map_update_elem(&bpfmap_offset_guess, &key, src, BPF_ANY);
    return 0;
}

static __always_inline int read_conn_info(__u8 *sk, struct offset_guess *status)
{
    if ((status->conn_type && CONN_L3_MASK) == CONN_L3_IPv4)
    {
        // src ip
        bpf_probe_read(&status->daddr + 3, sizeof(__be32), sk + status->offset_sk_daddr);
        bpf_probe_read(&status->saddr + 4, sizeof(__be32), sk + status->offset_sk_rcv_saddr);

        bpf_probe_read(&status->dport, sizeof(__u16), sk + status->offset_sk_dport);
        swap_u16(&status->dport);
        bpf_probe_read(&status->sport, sizeof(__u16), sk + status->offset_inet_sport);
        swap_u16(&status->sport);
    }
    else
    {
        bpf_probe_read(&status->daddr, sizeof(__be32) * 4, sk + status->offset_sk_v6_daddr);
        bpf_probe_read(&status->saddr, sizeof(__be32) * 4, sk + status->offset_sk_v6_rcv_saddr);
    }
    // rtt
    bpf_probe_read(&status->rtt, sizeof(__u32), sk + status->offset_tcp_sk_srtt_us);
    status->rtt >>= 3;
    bpf_probe_read(&status->rtt_var, sizeof(__u32), sk + status->offset_tcp_sk_mdev_us);
    status->rtt_var >>= 2;
    return 0;
}

SEC("kprobe/tcp_rcv_established")
int kprobe__tcp_rcv_established(struct pt_regs *ctx)
{
    __u8 *sk = (__u8 *)PT_REGS_PARM1(ctx);

    struct offset_guess status;
    if (read_offset(&status) != 0)
    {
        return 0;
    }
    if (skipConn(&status) != 1)
    {
        return 0;
    }
    read_conn_info(sk, &status);
    status.state++;
    update_offset(&status);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
