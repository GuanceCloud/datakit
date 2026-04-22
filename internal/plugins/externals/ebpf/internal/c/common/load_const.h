#ifndef __LOAD_CONST_H
#define __LOAD_CONST_H

#ifdef DK_LEGACY_CONST_IMM

#define DK_LEGACY_CONST_MAGIC 0x54A17EAD00000000ULL

#define DK_LEGACY_CONST_kernel_version           (DK_LEGACY_CONST_MAGIC | 0x001ULL)
#define DK_LEGACY_CONST_offset_sk_num            (DK_LEGACY_CONST_MAGIC | 0x002ULL)
#define DK_LEGACY_CONST_offset_inet_sport        (DK_LEGACY_CONST_MAGIC | 0x003ULL)
#define DK_LEGACY_CONST_offset_sk_family         (DK_LEGACY_CONST_MAGIC | 0x004ULL)
#define DK_LEGACY_CONST_offset_sk_rcv_saddr      (DK_LEGACY_CONST_MAGIC | 0x005ULL)
#define DK_LEGACY_CONST_offset_sk_daddr          (DK_LEGACY_CONST_MAGIC | 0x006ULL)
#define DK_LEGACY_CONST_offset_sk_v6_rcv_saddr   (DK_LEGACY_CONST_MAGIC | 0x007ULL)
#define DK_LEGACY_CONST_offset_sk_v6_daddr       (DK_LEGACY_CONST_MAGIC | 0x008ULL)
#define DK_LEGACY_CONST_offset_sk_dport          (DK_LEGACY_CONST_MAGIC | 0x009ULL)
#define DK_LEGACY_CONST_offset_tcp_sk_srtt_us    (DK_LEGACY_CONST_MAGIC | 0x00AULL)
#define DK_LEGACY_CONST_offset_tcp_sk_mdev_us    (DK_LEGACY_CONST_MAGIC | 0x00BULL)
#define DK_LEGACY_CONST_offset_flowi4_saddr      (DK_LEGACY_CONST_MAGIC | 0x00CULL)
#define DK_LEGACY_CONST_offset_flowi4_daddr      (DK_LEGACY_CONST_MAGIC | 0x00DULL)
#define DK_LEGACY_CONST_offset_flowi4_sport      (DK_LEGACY_CONST_MAGIC | 0x00EULL)
#define DK_LEGACY_CONST_offset_flowi4_dport      (DK_LEGACY_CONST_MAGIC | 0x00FULL)
#define DK_LEGACY_CONST_offset_flowi6_saddr      (DK_LEGACY_CONST_MAGIC | 0x010ULL)
#define DK_LEGACY_CONST_offset_flowi6_daddr      (DK_LEGACY_CONST_MAGIC | 0x011ULL)
#define DK_LEGACY_CONST_offset_flowi6_sport      (DK_LEGACY_CONST_MAGIC | 0x012ULL)
#define DK_LEGACY_CONST_offset_flowi6_dport      (DK_LEGACY_CONST_MAGIC | 0x013ULL)
#define DK_LEGACY_CONST_offset_sk_net            (DK_LEGACY_CONST_MAGIC | 0x014ULL)
#define DK_LEGACY_CONST_offset_ns_common_inum    (DK_LEGACY_CONST_MAGIC | 0x015ULL)
#define DK_LEGACY_CONST_offset_socket_sk         (DK_LEGACY_CONST_MAGIC | 0x016ULL)
#define DK_LEGACY_CONST_offset_socket_file       (DK_LEGACY_CONST_MAGIC | 0x017ULL)
#define DK_LEGACY_CONST_offset_task_struct_files (DK_LEGACY_CONST_MAGIC | 0x018ULL)
#define DK_LEGACY_CONST_offset_files_struct_fdt  (DK_LEGACY_CONST_MAGIC | 0x019ULL)
#define DK_LEGACY_CONST_offset_file_private_data (DK_LEGACY_CONST_MAGIC | 0x01AULL)
#define DK_LEGACY_CONST_offset_copied_seq        (DK_LEGACY_CONST_MAGIC | 0x01BULL)
#define DK_LEGACY_CONST_offset_write_seq         (DK_LEGACY_CONST_MAGIC | 0x01CULL)
#define DK_LEGACY_CONST_offset_ct_net            (DK_LEGACY_CONST_MAGIC | 0x01DULL)
#define DK_LEGACY_CONST_offset_ct_ns_common_inum (DK_LEGACY_CONST_MAGIC | 0x01EULL)
#define DK_LEGACY_CONST_offset_ct_origin_tuple   (DK_LEGACY_CONST_MAGIC | 0x01FULL)
#define DK_LEGACY_CONST_offset_ct_reply_tuple    (DK_LEGACY_CONST_MAGIC | 0x020ULL)
#define DK_LEGACY_CONST_apiflow_min_capture_size (DK_LEGACY_CONST_MAGIC | 0x021ULL)

#define LOAD_OFFSET(param, var)             \
    do                                      \
    {                                       \
        var = (__u64)DK_LEGACY_CONST_##param; \
    } while (0)

#else

#include <linux/types.h>

const volatile __u64 kernel_version = 0;

const volatile __u64 offset_sk_num = 0;
const volatile __u64 offset_inet_sport = 0;
const volatile __u64 offset_sk_family = 0;
const volatile __u64 offset_sk_rcv_saddr = 0;
const volatile __u64 offset_sk_daddr = 0;
const volatile __u64 offset_sk_v6_rcv_saddr = 0;
const volatile __u64 offset_sk_v6_daddr = 0;
const volatile __u64 offset_sk_dport = 0;
const volatile __u64 offset_tcp_sk_srtt_us = 0;
const volatile __u64 offset_tcp_sk_mdev_us = 0;
const volatile __u64 offset_flowi4_saddr = 0;
const volatile __u64 offset_flowi4_daddr = 0;
const volatile __u64 offset_flowi4_sport = 0;
const volatile __u64 offset_flowi4_dport = 0;
const volatile __u64 offset_flowi6_saddr = 0;
const volatile __u64 offset_flowi6_daddr = 0;
const volatile __u64 offset_flowi6_sport = 0;
const volatile __u64 offset_flowi6_dport = 0;
const volatile __u64 offset_sk_net = 0;
const volatile __u64 offset_ns_common_inum = 0;
const volatile __u64 offset_socket_sk = 0;
const volatile __u64 offset_socket_file = 0;
const volatile __u64 offset_task_struct_files = 0;
const volatile __u64 offset_files_struct_fdt = 0;
const volatile __u64 offset_file_private_data = 0;
const volatile __u64 offset_copied_seq = 0;
const volatile __u64 offset_write_seq = 0;
const volatile __u64 offset_ct_net = 0;
const volatile __u64 offset_ct_ns_common_inum = 0;
const volatile __u64 offset_ct_origin_tuple = 0;
const volatile __u64 offset_ct_reply_tuple = 0;
const volatile __u64 apiflow_min_capture_size = 0;

#define LOAD_OFFSET(param, var)  \
    do                           \
    {                            \
        var = (param);           \
    } while (0)

#endif

#endif // !__LOAD_CONST_H
