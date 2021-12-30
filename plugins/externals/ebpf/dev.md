# net_ebpf

## 注意

* 当前仅适用于使用小端的系统

* 在低版本内核(源码编译), 如 Linux Kernel 4.4.1 中：
  * 需要初始化局部变量, 由于这可能导致 BPF 验证器提示如 invalid indirect read from stack 等;
  * bpfmap 的 value 无法作为另一个 bpfmap 的 key, 需要先拷贝至局部变量;
