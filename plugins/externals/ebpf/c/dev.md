# BPF_PROG_TYPE_KPROBE 类型 eBPF 程序使用 kprobes 技术对内核函数动态插桩

在 x86_64 平台，触发软件中断指令 int3，优化后为 jump 指令

若函数被编译器优化为内联函数将导致 kprobes 探测点注册失败，可检查 `/proc/kallsyms` 或 `System.map`:

```sh
sudo cat /proc/kallsyms
```

获取 kprobes 黑名单，即无法设置断点函数, 有 kprobe_int3_handler 等:

```sh
sudo cat /sys/kernel/debug/kprobes/blacklist
```
