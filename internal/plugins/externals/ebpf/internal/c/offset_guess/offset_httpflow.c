#include <linux/fdtable.h>
#include <uapi/linux/ptrace.h>
#include <uapi/linux/tcp.h>
#include <net/sock.h>

#include "../netflow/conn_stats.h"
#include "../netflow/netflow_utils.h"
#include "../apiflow/tp_arg.h"
#include "bpf_helpers.h"
#include "offset.h"
#include "bpfmap.h"
#include "load_const.h"
#include "offset_read.h"
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
    if (offset.state == 0b111)
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

    struct file *file = NULL;
    if (offset.offset_socket_file >= 0 && offset.offset_socket_file <= 256)
    {
        DK_READ_VALUE(file, skt, offset.offset_socket_file);
    }

    if (file != NULL && offset.offset_file_private_data == 0)
    {
#pragma unroll
        for (__u32 j = 0; j < 48; j++)
        {
            __u32 private_data_offset = j * sizeof(void *);
            void *private_data = NULL;
            bpf_probe_read(&private_data, sizeof(private_data),
                           (__u8 *)file + private_data_offset);
            if (private_data != NULL && private_data == skt)
            {
                offset.offset_file_private_data = private_data_offset;
                offset.state |= 0b10;
                goto save_file_arg;
            }
        }

        file = NULL;
    }

    if ((offset.state & 0b100) == 0)
    {
#pragma unroll
        for (__u32 j = 0; j < 16; j++)
        {
            __u32 socket_sk_offset = j * sizeof(void *);
            struct sock *sk = NULL;
            conn_inf_t conn = {};

            bpf_probe_read(&sk, sizeof(sk), (__u8 *)skt + socket_sk_offset);
            if (sk == NULL)
            {
                continue;
            }

            if (read_connection_info(sk, &conn, pid_tgid, CONN_L4_TCP) != 0)
            {
                continue;
            }

            if (conn.saddr[3] != offset.saddr[3] || conn.daddr[3] != offset.daddr[3] ||
                conn.sport != offset.sport || conn.dport != offset.dport)
            {
                continue;
            }

            offset.offset_socket_sk = socket_sk_offset;
            offset.state |= 0b100;
            break;
        }
    }

save_file_arg:
    if (file != NULL)
    {
        struct comm_getsockopt_arg arg = {
            .file = file,
            .skt = skt,
        };

        // save file, for task_struct guess
        bpf_map_update_elem(&bpf_map_sock_common_getsockopt_arg, &pid_tgid, &arg, BPF_ANY);
    }
    else if ((offset.state & 0b10) == 0 && offset.offset_socket_file <= 256)
    {
        offset.offset_socket_file += sizeof(void *);
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
    if (offset.state == 0b111)
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

    struct fdtable *fdt = NULL;
    struct file **farry = NULL;
    void *skfile = NULL;

    DK_READ_VALUE(files, task, offset.offset_task_struct_files);

    if (files == NULL)
    {
        goto offset_plusplus;
    }

#pragma unroll
    for (int i = 0; i < 32; i++)
    {
        __u32 files_fdt_offset = i * sizeof(void *);
        bpf_probe_read(&fdt, sizeof(fdt),
                       (__u8 *)files + files_fdt_offset);

        if (fdt == NULL)
        {
            continue;
        }

        bpf_probe_read(&farry, sizeof(farry), &fdt->fd);

        if (farry == NULL)
        {
            continue;
        }

        bpf_probe_read(&skfile, sizeof(skfile), (void **)farry + fd);

        if (skfile == NULL)
        {
            continue;
        }

        if (skfile == arg->file)
        {
            offset.offset_files_struct_fdt = files_fdt_offset;
            offset.state |= 0b1;
            goto tail;
        }
    }

offset_plusplus:
    offset.offset_task_struct_files += sizeof(void *);

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
