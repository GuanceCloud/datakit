#ifndef __CONNTRACK_UTILS__
#define __CONNTRACK_UTILS__
#include <linux/types.h>

struct nf_origin_tuple
{
    // 与 netflow 的 src 含义不同，
    // src 表示报文来源 ip:port,
    // 不一定是本机 ip:port
    __u32 src_ip[4];
    __u32 dst_ip[4];

    __u16 src_port;
    __u16 dst_port;

    __u32 netns;
};

struct nf_reply_tuple
{
    __u64 ts;

    __u32 src_ip[4];
    __u32 dst_ip[4];

    __u16 src_port;
    __u16 dst_port;

    __u32 netns;
};

#endif
