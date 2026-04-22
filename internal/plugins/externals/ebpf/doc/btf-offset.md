# eBPF Offset BTF 获取说明

现在 offset 解析默认走 `BTF -> 最小化 fallback`，优先不再启动主动探测 runtime。

运行时的 BTF 查找顺序：

1. 环境变量 `DK_EBPF_BTF_PATH`
2. `/sys/kernel/btf/vmlinux`
3. `/boot/vmlinux-$(uname -r)`
4. `/usr/lib/modules/$(uname -r)/vmlinux`
5. `/lib/modules/$(uname -r)/vmlinux`
6. `/usr/lib/debug/boot/vmlinux-$(uname -r)`
7. `/usr/lib/debug/lib/modules/$(uname -r)/vmlinux`

推荐做法：

```bash
ls -lh /sys/kernel/btf/vmlinux
```

如果存在，这就是最优路径，启动时会直接使用内核导出的 BTF。

如果内核没有导出 `/sys/kernel/btf/vmlinux`，安装当前内核匹配的 `vmlinux`/debug/BTF 包，然后显式指定：

```bash
export DK_EBPF_BTF_PATH=/boot/vmlinux-$(uname -r)
```

如果需要生成 `vmlinux.h` 给 C eBPF 程序使用，可以直接从同一个 BTF 文件导出：

```bash
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
```

当前已经接入 BTF 直接解析的偏移包括：

- netflow kernel offsets
- tcp seq offsets
- httpflow file/socket offsets
- conntrack tuple/netns offsets

只有 BTF 缺失或结构字段无法解析时，才会回退到旧的主动探测逻辑。
