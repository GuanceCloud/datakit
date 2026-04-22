#ifndef __OFFSET_READ_H__
#define __OFFSET_READ_H__

#include <linux/types.h>

static __always_inline const void *dk_offset_ptr(const void *base, __u64 offset)
{
    return (const void *)((const __u8 *)base + offset);
}

#define DK_PTR_AT(type, base, offset) \
    ((type *)dk_offset_ptr((base), (offset)))

#define DK_READ_VALUE(dst, base, offset) \
    bpf_probe_read(&(dst), sizeof(dst), dk_offset_ptr((base), (offset)))

#define DK_READ_INTO(dst, base, offset) \
    bpf_probe_read((dst), sizeof(*(dst)), dk_offset_ptr((base), (offset)))

#define DK_READ_BUF(dst, size, base, offset) \
    bpf_probe_read((dst), (size), dk_offset_ptr((base), (offset)))

#endif
