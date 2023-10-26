#include <linux/fdtable.h>
#include <uapi/linux/ptrace.h>
#include <uapi/linux/tcp.h>
#include <net/sock.h>

#include "../apiflow/tp_arg.h"
#include "bpf_helpers.h"
#include "offset.h"
#include "bpfmap.h"
#include "load_const.h"
#include "filter.h"

static __always_inline __s32 read_offset(struct offset_httpflow *dst)
{
    __u64 key = 0;

    struct offset_httpflow *ptr =
        (struct offset_httpflow *)bpf_map_lookup_elem(&bpf_map_offset_httpflow, &key);

    if (ptr == NULL)
    {
        return -1;
    }

    bpf_probe_read(dst, sizeof(struct offset_httpflow), ptr);
    return 0;
}

static __always_inline __s32 update_offset(struct offset_httpflow *src)
{
    __u64 key = 0;
    bpf_map_update_elem(&bpf_map_offset_httpflow, &key, src, BPF_ANY);
    return 0;
}

// Used to calculate offsetof(struct file, private_data)
SEC("kprobe/sock_common_getsockopt")
int kprobe__sock_common_getsockopt(struct pt_regs *ctx)
{
    // Before calculating the offset, you need to lock the thread,
    // otherwise the tgid may not match and the address of the cached file cannot be found.
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct offset_httpflow offset = {};
    if (read_offset(&offset) != 0)
    {
        return 0;
    }

    if (skipConn(offset.process_name, offset.pid_tgid) != 0)
    {
        return 0;
    }

    // socket addr
    void *skt = (struct socket *)PT_REGS_PARM1(ctx);
    if (skt == NULL)
    {
        return 0;
    }

    // prog task_struct
    struct task_struct *task = bpf_get_current_task();

    struct file *file;
    // socket.file addr
    bpf_probe_read(&file, sizeof(__u8 *),
                   (__u8 *)skt + offset.offset_socket_file);

    if (file == NULL)
    {
        goto update;
    }

    struct comm_getsockopt_arg arg = {
        .file = file,
        .skt = skt,
    };

    // save file, for task_struct guess
    bpf_map_update_elem(&bpf_map_sock_common_getsockopt_arg, &pid_tgid, &arg, BPF_ANY);

    void *private_data = NULL;

    if (offset.offset_file_private_data != 0)
    {
        return 0;
    }

#pragma unroll
    for (__u32 i = 0; i < 300; i++)
    {

        // file.private_data (== socket addr)
        bpf_probe_read(&private_data, sizeof(__u8 *), (__u8 *)file + i);

        if (private_data != NULL && private_data == skt)
        {
            offset.offset_file_private_data = i;
            offset.state |= 0b10;
            goto update;
        }
    }

update:
    offset.times++;
    update_offset(&offset);

    return 0;
}

// Used to calculate offset(struct task_struct, files)
// and offset(struct files_struct)
SEC("kretprobe/sock_common_getsockopt")
int kpretrobe__sock_common_getsockopt(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct offset_httpflow offset = {};
    if (read_offset(&offset) != 0)
    {
        return 0;
    }

    if (skipConn(offset.process_name, offset.pid_tgid) != 0)
    {
        return 0;
    }

    int fd = offset.fd;
    if (fd < 3)
    {
        return 0;
    }

    struct comm_getsockopt_arg *arg =
        bpf_map_lookup_elem(&bpf_map_sock_common_getsockopt_arg, &pid_tgid);

    if (arg == NULL)
    {
        return 0;
    }

    void *file = NULL;

    struct files_struct *files = NULL;

    struct task_struct *task = bpf_get_current_task();

    bpf_probe_read(&files, sizeof(files),
                   (__u8 *)task + offset.offset_task_struct_files);

    if (files == NULL)
    {
        goto offset_plusplus;
    }

    struct fdtable *fdt = NULL;
    struct file **farry = NULL;
    void *skfile = NULL;

    void *skt = (void *)bpf_map_lookup_elem(&bpf_map_sock_common_getsockopt_arg, &pid_tgid);
    if (skt == NULL)
    {
        return 0;
    }

#pragma unroll
    for (int i = 0; i < 125; i++)
    {
        bpf_probe_read(&fdt, sizeof(fdt),
                       (__u8 *)files + i);

        // bpf_printk("fdt %u", fdt);
        if (fdt == NULL)
        {
            continue;
        }

        bpf_probe_read(&farry, sizeof(farry), &fdt->fd);
        // bpf_printk("farry %u", farry);

        if (farry == NULL)
        {
            continue;
        }

        bpf_probe_read(&skfile, sizeof(skfile), (void **)farry + fd);
        // bpf_printk("skfile %u", skfile);

        if (skfile == NULL)
        {
            continue;
        }

        if (skfile == arg->file)
        {
            offset.offset_files_struct_fdt = i;
            offset.state |= 0b1;
            goto tail;
        }
    }

offset_plusplus:
    offset.offset_task_struct_files += 1;

tail:
    offset.times++;
    update_offset(&offset);

    bpf_map_delete_elem(&bpf_map_sock_common_getsockopt_arg, &pid_tgid);
    bpf_map_delete_elem(&bpf_map_file_ptr, &pid_tgid);

    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
