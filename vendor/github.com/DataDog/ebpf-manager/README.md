## eBPF Manager
[![](https://godoc.org/github.com/DataDog/ebpf-manager?status.svg)](https://godoc.org/github.com/DataDog/ebpf-manager)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/mit)

This repository implements a manager on top of [Cilium's eBPF library](https://github.com/cilium/ebpf). This declarative manager simplifies attaching and detaching eBPF programs by controlling their entire life cycle. It was built with the intention of unifying how eBPF is used in large scale projects such as the [Datadog Agent](https://github.com/DataDog/datadog-agent). By using the same declarative conventions, multiple teams can quickly collaborate on complex eBPF programs by sharing maps, programs or even hook points without having to worry about the setup of complex program types.

### Requirements

* A version of Go that is [supported by upstream](https://golang.org/doc/devel/release.html#policy)
* Linux 4.4+ (some eBPF features are only available on newer kernel versions, see [eBPF features by Linux version](https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md))

### Getting started

You can find many examples using the manager in [examples/](https://github.com/DataDog/ebpf-manager/tree/main/examples). For a real world use case, check out the [Datadog Agent](https://github.com/DataDog/datadog-agent).

### Useful resources

* [Cilium eBPF library](https://github.com/cilium/ebpf)
* [Cilium eBPF documentation](https://cilium.readthedocs.io/en/latest/bpf/#bpf-guide)
* [Linux documentation on BPF](http://elixir.free-electrons.com/linux/latest/source/Documentation/networking/filter.txt)
* [eBPF features by Linux version](https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md)

## License

- Unless explicitly specified otherwise, the golang code in this repository is under the MIT License.
- The eBPF programs are under the GPL v2 License.