#include <net/netfilter/nf_conntrack.h>
#include "bpf_helpers.h"
#include "offset.h"
#include "filter.h"

// ------------ bpf maps --------------

BPF_HASH_MAP(bpfmap_offset_conntrack, __u64, struct offset_conntrack, 1);

// ------------------------------------

static __always_inline void swap_u16(__u16 *v)
{
    *v = ((*v & 0x00FF) << 8) | ((*v & 0xFF00) >> 8);
}

static __always_inline int read_offset_ct(struct offset_conntrack *dst)
{
    __u64 key = 0;

    struct offset_conntrack *ptr =
        (struct offset_conntrack *)bpf_map_lookup_elem(&bpfmap_offset_conntrack, &key);

    if (ptr == NULL)
    {
        return -1;
    }

    bpf_probe_read(dst, sizeof(struct offset_conntrack), ptr);
    return 0;
}

static __always_inline void write_offset_ct(struct offset_conntrack *offset)
{
    __u64 key = 0;
    bpf_map_update_elem(&bpfmap_offset_conntrack, &key, offset, BPF_ANY);
}

static __always_inline void ct_get_netns(struct nf_conn *ct, struct offset_conntrack *offset)
{
    struct net *sknet = NULL;
    bpf_probe_read(&sknet, sizeof(sknet), (__u8 *)ct + offset->offset_ct_net);
    int err = bpf_probe_read(&offset->netns, sizeof(__u32), (__u8 *)sknet + offset->offset_ns_common_inum);
    if (err == -EFAULT || offset->offset_ns_common_inum > 200)
    {
        offset->err = ERR_G_NS_INUM;
    }
    else
    {
        offset->err = 0;
    }
}

static __always_inline int get_nf_conntrack_tuple(struct nf_conn_tuple *conn, struct nf_conntrack_tuple *tuple)
{
    // l3num: AF_INET or AF_INET6
    bpf_probe_read(&conn->l3num, sizeof(__u16), &tuple->src.l3num);
    switch (conn->l3num)
    {
    case AF_INET:
        bpf_probe_read(&conn->src_ip[3], sizeof(__u32), &tuple->src.u3.ip);
        bpf_probe_read(&conn->dst_ip[3], sizeof(__u32), &tuple->dst.u3.ip);
        break;
    case AF_INET6:
        bpf_probe_read(&conn->src_ip, sizeof(__u32[4]), &tuple->src.u3.ip6);
        bpf_probe_read(&conn->dst_ip, sizeof(__u32[4]), &tuple->dst.u3.ip6);
        break;
    default:
        return -1;
    }

    // l4proto: IPPROTO_TCP or IPPROTO_UDP
    bpf_probe_read(&conn->l4proto, sizeof(__u8), &tuple->dst.protonum);
    switch (conn->l4proto)
    {
    case IPPROTO_TCP:
        break;
    case IPPROTO_UDP:
        break;
    default:
        return -1;
    }

    bpf_probe_read(&conn->src_port, sizeof(__u16), &tuple->src.u);
    swap_u16(&conn->src_port);
    bpf_probe_read(&conn->dst_port, sizeof(__u16), &tuple->dst.u);
    swap_u16(&conn->dst_port);

    return 0;
}

SEC("kprobe/__nf_conntrack_hash_insert")
int kprobe___nf_conntrack_hash_insert(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct offset_conntrack offset = {0};
    if (read_offset_ct(&offset) != 0)
    {
        return 0;
    }

    if (skipConn(offset.process_name, offset.pid_tgid) != 0)
    {
        return 0;
    }

    offset.state++;

    struct nf_conn *ct = (struct nf_conn *)PT_REGS_PARM1(ctx);
    if (ct == NULL)
    {
        return 0;
    }

    // 从内核源码的 commits 记录上看 struct nf_conntrack_tuple 这个结构体较稳定，
    // 无须推算该结构体成员的偏移量

    // origin tuple: &ct->tuplehash[0].tuple
    get_nf_conntrack_tuple(&offset.origin,
                           (struct nf_conntrack_tuple *)((__u8 *)(ct) + offset.offset_ct_origin_tuple));

    // reply tuple: &ct->tuplehash[1].tuple
    get_nf_conntrack_tuple(&offset.reply,
                           (struct nf_conntrack_tuple *)((__u8 *)(ct) + offset.offset_ct_reply_tuple));

    ct_get_netns(ct, &offset);
#ifdef __DK_DEBUG__
    if (offset.origin.src_ip[3] == 0x200007f || offset.origin.dst_ip[3] == 0x200007f)
    {
        bpf_printk("origin addr : %x %x", offset.origin.src_ip[3], offset.origin.dst_ip[3]);
        bpf_printk("reply addr  : %x %x", offset.reply.src_ip[3], offset.reply.dst_ip[3]);
        bpf_printk("origin sport: %d %d", offset.origin.src_port, offset.origin.dst_port);
        bpf_printk("reply sport : %d %d", offset.reply.src_port, offset.reply.dst_port);
    }

    bpf_printk("offset %d, %d", (__u64)(&ct->tuplehash[0].tuple) - (__u64)(ct),
               (__u64)(&ct->tuplehash[1].tuple) - (__u64)(ct));
    unsigned long st = 0;
    bpf_probe_read(&st, sizeof(st), &ct->status);
    bpf_printk("status 0x%x", st);
#endif

    write_offset_ct(&offset);

    return 0;
};

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;