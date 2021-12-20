// #include <linux/kconfig.h>
#include <uapi/linux/ptrace.h>
#include <uapi/linux/tcp.h>
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
    char actual[PROCNAMELEN] = {};
    bpf_get_current_comm(&actual, PROCNAMELEN);
    for (int i = 0; i < PROCNAMELEN - 1; i++)
    {
        if (actual[i] != status->process_name[i])
        {
            return 1;
        }
    }
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    if ((pid_tgid >> 32) != (status->pid_tgid >> 32))
    {
        return 1;
    }
    return 0;
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

static __always_inline __u32 read_netns(void *sknet)
{
    __u32 inum = 0;
#ifdef CONFIG_NET_NS
    struct net *netptr = NULL;
    // read the memory address of a net instance from *sknet,
    // possible_net_t has only one field: struct net *
    bpf_probe_read(&netptr, sizeof(netptr), sknet);

    bpf_probe_read(&inum, sizeof(inum), &netptr->ns.inum);
#endif
    return inum;
}

static __always_inline void read_netns_inum(struct sock *sk, struct offset_guess *status)
{

    struct net *sknet = NULL;

    bpf_probe_read(&sknet, sizeof(sknet), (__u8 *)sk + status->offset_sk_net);
    int err = bpf_probe_read(&status->netns, sizeof(__u32), (__u8 *)sknet + status->offset_ns_common_inum);
    if (err == -EFAULT || status->offset_ns_common_inum > 200 || status->offset_sk_net > 128)
    {
        status->err = ERR_G_SK_NET;
    }
    else
    {
        status->err = 0;
    }
}

static __always_inline int read_conn_info(struct sock *sk, struct offset_guess *status)
{
    unsigned short family = AF_UNSPEC;
    // read_netns_inum(sk, status);

    bpf_probe_read(&family, sizeof(family), (__u8 *)sk + status->offset_sk_family);
    if (family == AF_INET)
    {
        status->meta = CONN_L3_IPv4;
    }
    else if (family == AF_INET6)
    {
        status->meta = CONN_L3_IPv6;
    }
    else
    {
        status->meta = CONN_L3_MASK; // unknown
    }

    if ((status->conn_type && CONN_L3_MASK) == CONN_L3_IPv4)
    {
        // src ip
        bpf_probe_read(status->daddr + 3, sizeof(__be32), (__u8 *)sk + status->offset_sk_daddr);

        bpf_probe_read(&status->dport, sizeof(__u16), (__u8 *)sk + status->offset_sk_dport);
        swap_u16(&status->dport);
        bpf_probe_read(&status->sport, sizeof(__u16), (__u8 *)sk + status->offset_inet_sport);
        swap_u16(&status->sport);
    }
    else
    {
        bpf_probe_read(status->daddr, sizeof(__be32) * 4, (__u8 *)sk + status->offset_sk_v6_daddr);
    }
    return 0;
}

SEC("kprobe/tcp_getsockopt")
int kprobe__tcp_getsockopt(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    int level = (int)PT_REGS_PARM2(ctx);
    int optname = (int)PT_REGS_PARM3(ctx);
    if (level != SOL_TCP || optname != TCP_INFO)
    {
        return 0;
    }

    struct offset_guess status = {};
    if (read_offset(&status) != 0)
    {
        return 0;
    }

    if (skipConn(&status) != 0)
    {
        return 0;
    }

    read_conn_info(sk, &status);
    read_netns_inum(sk, &status);
    status.state++;

    // rtt
    bpf_probe_read(&status.rtt, sizeof(__u32), (__u8 *)sk + status.offset_tcp_sk_srtt_us);
    status.rtt >>= 3;
    bpf_probe_read(&status.rtt_var, sizeof(__u32), (__u8 *)sk + status.offset_tcp_sk_mdev_us);
    status.rtt_var >>= 2;

    update_offset(&status);
    return 0;
}

SEC("kprobe/tcp_v6_connect")
int kprobe__tcp_v6_connect(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    bpf_map_update_elem(&bpfmap_tcpv6conn, &pid_tgid, &sk, BPF_ANY);
    return 0;
}

SEC("kretprobe/tcp_v6_connect")
int kretprobe__tcp_v6_connect(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct sock **skpp = bpf_map_lookup_elem(&bpfmap_tcpv6conn, &pid_tgid);
    if (skpp == NULL)
    {
        return 0;
    }
    struct sock *sk = *skpp;
    bpf_map_delete_elem(&bpfmap_tcpv6conn, &pid_tgid);

    struct offset_guess status = {};
    if (read_offset(&status) != 0)
    {
        return 0;
    }
    if (skipConn(&status) != 0)
    {
        return 0;
    }
    read_conn_info(sk, &status);
    status.state++;
    update_offset(&status);

    return 0;
}

SEC("kprobe/ip_make_skb")
int kprobe__ip_make_skb(struct pt_regs *ctx)
{
    struct offset_guess status = {};
    if (read_offset(&status) != 0)
    {
        return 0;
    }
    if (skipConn(&status) != 0)
    {
        return 0;
    }
    struct flowi4 *fl4 = (struct flowi4 *)PT_REGS_PARM2(ctx);

    bpf_probe_read(status.daddr + 3, sizeof(__be32), (__u8 *)fl4 + status.offset_flowi4_daddr);
    bpf_probe_read(status.saddr + 3, sizeof(__be32), (__u8 *)fl4 + status.offset_flowi4_saddr);

    bpf_probe_read(&status.dport, sizeof(__u16), (__u8 *)fl4 + status.offset_flowi4_dport);
    swap_u16(&status.dport);

    status.state++;
    update_offset(&status);

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
