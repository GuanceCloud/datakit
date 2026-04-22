#include "bash_history.h"
#include "bpf_helpers.h"
#include <uapi/linux/ptrace.h>

struct bpf_map_def SEC("maps/bpfmap_bash_readline") bpfmap_bash_readline = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .key_size = sizeof(__u32),   // cpu id
    .value_size = sizeof(__u32), // fd
    .max_entries = 0,
};

SEC("uretprobe/readline")
int uretprobe__readline(struct pt_regs *ctx) {
  void *ret = (void *)PT_REGS_RC(ctx);

  if (ret == NULL) {
    return 0;
  }

  struct bash_event event = {};

  /*
   * Keep Ubuntu 18.04 (4.15 kernel) compatibility:
   * bpf_probe_read_user() is not available on older kernels/loaders.
   * For uprobes, bpf_probe_read() can safely read userspace memory here.
   */
  if (bpf_probe_read(&event.line, sizeof(event.line), ret) != 0) {
    return 0;
  }

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
