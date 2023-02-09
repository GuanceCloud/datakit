# BPF_PROG_TYPE_KPROBE type eBPF program uses kprobes technology to dynamically insert kernel functions

On the x86_64 platform, trigger the software interrupt instruction int3, which is optimized as a jump instruction

If the function is optimized as an inline function by the compiler, the kprobes detection point registration will fail. You can check `/proc/kallsyms` or `System.map`:

```sh
sudo cat /proc/kallsyms
```

Get the kprobes blacklist, that is, the breakpoint function cannot be set, there are kprobe_int3_handler, etc.:

```sh
sudo cat /sys/kernel/debug/kprobes/blacklist
```