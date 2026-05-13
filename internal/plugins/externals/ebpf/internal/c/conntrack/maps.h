#ifndef __CONNTRACK_MAPS__
#define __CONNTRACK_MAPS__
#include "bpf_helpers.h"
#include "../netflow/conn_stats.h"
#include "utils.h"

BPF_BEST_EFFORT_LRU_HASH_MAP(bpfmap_conntrack_tuple,
                              struct nf_origin_tuple, struct nf_reply_tuple, 65535);

BPF_ARRAY_MAP(bpfmap_conntrack_update_fail, __u32, __u64, 1);

static __always_inline void record_conntrack_update_fail()
{
    __u32 key = 0;
    __u64 *count = bpf_map_lookup_elem(&bpfmap_conntrack_update_fail, &key);
    if (count != NULL)
    {
        __sync_fetch_and_add(count, 1);
    }
}

static __always_inline void do_dnapt(struct connection_info *conn , __u32 *dst_nat_addr, __u16 *dst_nat_port)
{
    // DNAPT
    struct nf_origin_tuple key = {0};

    __builtin_memcpy(&key.src_ip, conn->saddr, sizeof(__u32[4]));
    __builtin_memcpy(&key.dst_ip, conn->daddr, sizeof(__u32[4]));
    key.src_port = conn->sport;
    key.dst_port = conn->dport;
    key.netns = conn->netns;

    struct nf_reply_tuple *reply = (struct nf_reply_tuple *)bpf_map_lookup_elem(&bpfmap_conntrack_tuple, &key);
    if (reply == NULL && key.netns != 0)
    {
        key.netns = 0;
        reply = (struct nf_reply_tuple *)bpf_map_lookup_elem(&bpfmap_conntrack_tuple, &key);
    }
    if (reply != NULL)
    {
        __builtin_memcpy(dst_nat_addr, reply->src_ip, sizeof(__u32[4]));
        *dst_nat_port = reply->src_port;
#ifdef __DK_DEBUG__
        bpf_printk("update daddr %x\n", dst_nat_addr[3]);
        bpf_printk("update dport %d\n", *dst_nat_port);
#endif
    }
};

#endif
