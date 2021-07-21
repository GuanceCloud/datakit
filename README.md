# DataKit

DataKit is collection agent for [DataFlux](https://dataflux.cn/)

# Build

DataKit is designed to be build across Darwin, Linux and Windows. For example, you can build windows/amd64 release under Darwin:

```shell
# build windows exe on Mac
LOCAL=windows/amd64 make

# build linux binary on Mac
LOCAL=linux/amd64 make

# build on different OS/Arch, DataKit build tool will
# detect current OS/Arch and build the default one
make
```

and is's possible to huild all release under Darwin:

```shell
make testing
```

All the output binary under `dist`:

```
dist/
├── datakit-darwin-amd64
│   └── datakit
└── local
    ├── installer-darwin-amd64
    └── version
```

> Note: The Darwin release can not be build under Linux/Darwin, because we applied CGO for Darwin release

# More references

- [Datakit-How-TO](https://www.yuque.com/dataflux/datakit/datakit-how-to)
- Datakit Install
	- [On host](https://www.yuque.com/dataflux/datakit/datakit-install)
	- [DaemonSet](https://www.yuque.com/dataflux/datakit/datakit-daemonset-deploy)
