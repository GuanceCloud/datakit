#include <net/netfilter/nf_conntrack_tuple.h>
#include <net/netfilter/nf_conntrack.h>

#include "../netflow/conn_stats.h"

#include "load_const.h"
#include "bpf_helpers.h"

#include "maps.h"
#include "utils.h"

static __always_inline __u64 load_offset_origin_tuple()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_ct_origin_tuple", var);
    return var;
}

static __always_inline __u64 load_offset_reply_tuple()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_ct_reply_tuple", var);
    return var;
}

static __always_inline __u64 load_offset_ct_net()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_ct_net", var);
    return var;
}

static __always_inline __u64 load_offset_ns_common_inum()
{
    __u64 var = 0;
    LOAD_OFFSET("offset_ns_common_inum", var);
    return var;
}

static __always_inline void swap_u16(__u16 *v)
{
    *v = __builtin_bswap16(*v);
}

static __always_inline void ct_get_netns(struct nf_conn *ct, struct nf_origin_tuple *tuple)
{
    struct net *sknet = NULL;
    __u64 offset = load_offset_ct_net();
    bpf_probe_read(&sknet, sizeof(sknet), (__u8 *)ct + offset);
    offset = load_offset_ns_common_inum();
    bpf_probe_read(&tuple->netns, sizeof(__u32), (__u8 *)sknet + offset);
}

static __always_inline int get_nf_tuple_info(struct nf_conn *ct, struct nf_origin_tuple *conn, struct nf_reply_tuple *value)
{
    if (ct == NULL || conn == NULL || value == NULL)
    {
        return -1;
    }

    __u64 offset_origin = load_offset_origin_tuple();
    __u64 offset_reply = load_offset_reply_tuple();

    struct nf_conntrack_tuple *origin = (struct nf_conntrack_tuple *)((__u8 *)(ct) + offset_origin);
    struct nf_conntrack_tuple *reply = (struct nf_conntrack_tuple *)((__u8 *)(ct) + offset_reply);

    // l3num: AF_INET or AF_INET6
    __u16 l3num = 0;
    bpf_probe_read(&l3num, sizeof(__u16), &origin->src.l3num);
    switch (l3num)
    {
    case AF_INET:
        bpf_probe_read(conn->src_ip + 3, sizeof(__u32), &origin->src.u3.ip);
        bpf_probe_read(conn->dst_ip + 3, sizeof(__u32), &origin->dst.u3.ip);

        bpf_probe_read(value->src_ip + 3, sizeof(__u32), &reply->src.u3.ip);
        bpf_probe_read(value->dst_ip + 3, sizeof(__u32), &reply->dst.u3.ip);
        break;
    case AF_INET6:
        bpf_probe_read(&conn->src_ip, sizeof(__u32[4]), &origin->src.u3.ip6);
        bpf_probe_read(&conn->dst_ip, sizeof(__u32[4]), &origin->dst.u3.ip6);

        bpf_probe_read(&value->src_ip, sizeof(__u32[4]), &reply->src.u3.ip6);
        bpf_probe_read(&value->dst_ip, sizeof(__u32[4]), &reply->dst.u3.ip6);
        break;
    default:
        return -1;
    }

    // l4proto: IPPROTO_TCP or IPPROTO_UDP
    __u8 l4proto = 0;
    bpf_probe_read(&l4proto, sizeof(__u8), &origin->dst.protonum);
    switch (l4proto)
    {
    case IPPROTO_TCP:
        break;
    case IPPROTO_UDP:
        break;
    default:
        return -1;
    }

    bpf_probe_read(&conn->src_port, sizeof(__u16), &origin->src.u);
    swap_u16(&conn->src_port);
    bpf_probe_read(&conn->dst_port, sizeof(__u16), &origin->dst.u);
    swap_u16(&conn->dst_port);

    bpf_probe_read(&value->src_port, sizeof(__u16), &reply->src.u);
    swap_u16(&value->src_port);
    bpf_probe_read(&value->dst_port, sizeof(__u16), &reply->dst.u);
    swap_u16(&value->dst_port);

    // filter without DNAPT
    if (((__u64 *)(conn->dst_ip))[0] == ((__u64 *)(value->src_ip))[0] &&
        ((__u64 *)(conn->dst_ip) + 1)[0] == ((__u64 *)(value->src_ip) + 1)[0] &&
        conn->dst_port == value->src_port 

        // ignore SNAPT
        // ((__u64 *)(conn->src_ip))[0] == ((__u64 *)(value->dst_ip))[0] &&
        // ((__u64 *)(conn->src_ip) + 1)[0] == ((__u64 *)(value->dst_ip) + 1)[0] &&
        // conn->src_port == value->dst_port
    )
    {
        return 1;
    }

    return 0;
}

SEC("kprobe/__nf_conntrack_hash_insert")
int kprobe___nf_conntrack_hash_insert(struct pt_regs *ctx)
{
    struct nf_conn *ct = (struct nf_conn *)PT_REGS_PARM1(ctx);

    struct nf_origin_tuple origin = {};
    struct nf_reply_tuple reply = {};
    if (get_nf_tuple_info(ct, &origin, &reply) != 0)
    {
        return 0;
    }
    ct_get_netns(ct, &origin);

#ifdef __DK_DEBUG__
    bpf_printk("netns %u\n", origin.netns);
    bpf_printk("origin saddr %x daddr %x\n", origin.src_ip[3], origin.dst_ip[3]);
    bpf_printk("origin sport %d dport %d\n", origin.src_port, origin.dst_port);
    bpf_printk("reply saddr %x daddr %x\n", reply.src_ip[3], reply.dst_ip[3]);
    bpf_printk("reply sport %d dport %d\n", reply.src_port, reply.dst_port);
#endif

    bpf_map_update_elem(&bpfmap_conntrack_tuple, &origin, &reply, BPF_ANY);
    return 0;
}

SEC("kprobe/nf_ct_delete")
int kprobe__nf_ct_delete(struct pt_regs *ctx)
{
    struct nf_conn *ct = (struct nf_conn *)PT_REGS_PARM1(ctx);

    struct nf_origin_tuple origin = {};
    struct nf_reply_tuple reply = {};
    if (get_nf_tuple_info(ct, &origin, &reply) != 0)
    {
        return 0;
    }
    ct_get_netns(ct, &origin);

#ifdef __DK_DEBUG__
    bpf_printk("DESTROY origin saddr %x daddr %x\n", origin.src_ip[3], origin.dst_ip[3]);
    bpf_printk("DESTROY origin sport %d dport %d\n", origin.src_port, origin.dst_port);
    bpf_printk("DESTROY reply saddr %x daddr %x\n", reply.src_ip[3], reply.dst_ip[3]);
    bpf_printk("DESTROY reply sport %d dport %d\n", reply.src_port, reply.dst_port);
#endif

    bpf_map_delete_elem(&bpfmap_conntrack_tuple, &origin);

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;