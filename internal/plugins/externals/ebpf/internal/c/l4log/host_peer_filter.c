#include <linux/bpf.h>
#include <linux/types.h>

#include "../common/bpf_helpers.h"

#define DK_ETH_P_IP 0x0800
#define DK_ETH_P_IPV6 0x86DD
#define DK_IPPROTO_TCP 6
#define DK_IPPROTO_UDP 17
#define DK_IPPROTO_FRAGMENT 44
#define DK_VXLAN_PORT_8472 8472
#define DK_VXLAN_PORT_4789 4789
#define DK_HOST_PEER_IFINDEX_MAX 1024

BPF_HASH_MAP(bpfmap_host_peer_ifindex, __u32, __u8, DK_HOST_PEER_IFINDEX_MAX);

static __always_inline int allow_ifindex(struct __sk_buff *skb) {
  __u32 ifindex = skb->ifindex;
  __u8 *enabled = bpf_map_lookup_elem(&bpfmap_host_peer_ifindex, &ifindex);
  if (enabled == NULL || *enabled == 0) {
    return 0;
  }
  return 1;
}

static __always_inline int allow_udp_port(__u16 sport, __u16 dport) {
  return sport == DK_VXLAN_PORT_8472 || sport == DK_VXLAN_PORT_4789 ||
         dport == DK_VXLAN_PORT_8472 || dport == DK_VXLAN_PORT_4789;
}

static __always_inline int filter_ipv4(struct __sk_buff *skb) {
  __u8 proto = (__u8)load_byte(skb, 23);
  if (proto == DK_IPPROTO_TCP) {
    return -1;
  }
  if (proto != DK_IPPROTO_UDP) {
    return 0;
  }

  __u16 frag_off = (__u16)load_half(skb, 20);
  if ((frag_off & 0x1FFF) != 0) {
    return 0;
  }

  __u8 ihl = (__u8)load_byte(skb, 14);
  __u32 l4_off = 14 + ((__u32)(ihl & 0x0F) << 2);
  __u16 sport = (__u16)load_half(skb, l4_off);
  __u16 dport = (__u16)load_half(skb, l4_off + 2);

  if (allow_udp_port(sport, dport)) {
    return -1;
  }
  return 0;
}

static __always_inline int filter_ipv6(struct __sk_buff *skb) {
  __u8 next_header = (__u8)load_byte(skb, 20);
  __u32 l4_off = 54;

  if (next_header == DK_IPPROTO_TCP) {
    return -1;
  }

  if (next_header == DK_IPPROTO_FRAGMENT) {
    next_header = (__u8)load_byte(skb, 54);
    l4_off = 62;
    if (next_header == DK_IPPROTO_TCP) {
      return -1;
    }
  }

  if (next_header != DK_IPPROTO_UDP) {
    return 0;
  }

  __u16 sport = (__u16)load_half(skb, l4_off);
  __u16 dport = (__u16)load_half(skb, l4_off + 2);
  if (allow_udp_port(sport, dport)) {
    return -1;
  }
  return 0;
}

SEC("socket/host_peer_filter")
int socket__host_peer_filter(struct __sk_buff *skb) {
  if (allow_ifindex(skb) == 0) {
    return 0;
  }

  __u16 eth_proto = (__u16)load_half(skb, 12);
  if (eth_proto == DK_ETH_P_IP) {
    return filter_ipv4(skb);
  }
  if (eth_proto == DK_ETH_P_IPV6) {
    return filter_ipv6(skb);
  }
  return 0;
}

char __license[] SEC("license") = "Dual BSD/GPL";
