#include "bpf_helpers.h"
#include "bash_history.h"
#include <uapi/linux/ptrace.h>

struct bpf_map_def SEC("maps/bpfmap_bash_readline") bpfmap_bash_readline = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),   // cpu id
    .value_size = sizeof(__u32), // fd
    .max_entries = 0,
};


SEC("uretprobe/readline")
int uretprobe_readline(struct pt_regs *ctx)
{
        void *ret = (void *)PT_REGS_RC(ctx);

        struct bash_event event = {};

        bpf_probe_read(&event.line, sizeof(event.line), ret);

        event.pid_tgid = bpf_get_current_pid_tgid();
        event.uid_gid = bpf_get_current_uid_gid();

        __u64 cpu = bpf_get_smp_processor_id();

        bpf_perf_event_output(ctx, &bpfmap_bash_readline, cpu, &event, sizeof(event));
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
