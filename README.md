# DataKit

DataKit is collection agent for [DataFlux](https://dataflux.cn/)

# Build

## Dependencies

- `apt-get install gcc-multilib`: for building oracle input
- `apt-get install tree`: for Makefile usage
- `packr2`: for packaging manuals
- `goyacc`: for pipeline grammar generation

DataKit was designed to be build across Darwin, Linux and Windows. For example, you can build windows/amd64 release under Darwin:

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
# build all os-arch platform under testing rule
make testing
```

All the output binary under `dist`:

```
dist
├  [4.0K]  datakit-linux-386
│   ├ [ 51M]  datakit
│   └ [4.0K]  externals
│       └ [8.5M]  oracle
├  [4.0K]  datakit-linux-amd64
│   ├ [ 60M]  datakit
│   └ [4.0K]  externals
│       └ [9.9M]  oracle
├  [4.0K]  datakit-linux-arm
│   └ [ 51M]  datakit
├  [4.0K]  datakit-linux-arm64
│   └ [ 56M]  datakit
├  [4.0K]  datakit-windows-386
│   └ [ 52M]  datakit.exe
├  [4.0K]  datakit-windows-amd64
│   └ [ 60M]  datakit.exe
└  [4.0K]  test
    ├  [ 12M]  installer-linux-386
    ├  [ 15M]  installer-linux-amd64
    ├  [ 12M]  installer-linux-arm
    ├  [ 14M]  installer-linux-arm64
    ├  [ 13M]  installer-windows-386.exe
    ├  [ 15M]  installer-windows-amd64.exe
    └  [ 204]  version

9 directories, 15 files
```

> Note: The Darwin release can not be build under Linux/Windows, because we applied CGO for Darwin release. BTW, Windows lack(default) of many build tools(such as `make`), we still recommand to build under Linux and Darwin.

# More references

- [Datakit-How-TO](https://www.yuque.com/dataflux/datakit/datakit-how-to)
- Datakit Install
	- [On host](https://www.yuque.com/dataflux/datakit/datakit-install)
	- [DaemonSet](https://www.yuque.com/dataflux/datakit/datakit-daemonset-deploy)
