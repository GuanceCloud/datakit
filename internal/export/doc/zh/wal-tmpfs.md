# 在 WAL 中使用 tmpfs

:fontawesome-brands-linux : :material-kubernetes :

---

某些情况下，如果磁盘性能不够，我们可以用一点内存的空间来满足 WAL 对磁盘读写的要求。

<!-- markdownlint-disable MD046 -->
=== "Linux 主机"

    1. 在主机上创建 tmpfs 挂载目录：
    
        ```shell
        sudo mkdir -p /mnt/wal-ramdisk
        ```
    
    1. 用 1 GiB 的内存空间来构建一个 tmpfs 磁盘：
    
        ```shell
        sudo mount -t tmpfs -o size=1G tmpfs /mnt/wal-ramdisk
        ```
    
    1. 查看挂载情况：
    
        ```shell
        $ df -h /mnt/wal-ramdisk/
        Filesystem      Size  Used Avail Use% Mounted on
        tmpfs           1.0G     0  1.0G   0% /mnt/wal-ramdisk
        ```
    
    1. 修改 *datakit.conf* 中 WAL 目录：
    
        ```toml hl_lines="2"
        [dataway.wal]
          path = "/mnt/wal-ramdisk"
        ```

=== "DaemonSet"

    1. 在 Kubernetes Node 机器上，创建 *ramdisk* 目录：
    
        ```shell
        # 此处我们换了目录，默认情况下，datakit.yaml 中，已经将 Node 主机的 /root/datakit_cache
        # 挂载到了 Datakit 容器的 /usr/local/datakit/cache 目录下
        mkdir -p /root/datakit_cache/ramdisk
        ```
    
    1. 创建 1 GiB 的 tmpfs：
    
        ```shell
        mount -t tmpfs -o size=1G tmpfs /root/datakit_cache/ramdisk
        ```
    
    1. 在 *datakit.yaml* 中增加如下环境变量，然后重启 Datakit 容器：
    
        ```yaml
        - name: ENV_DATAWAY_WAL_PATH
          value: /usr/local/datakit/cache/ramdisk
        ```
<!-- markdownlint-enable -->

---

<!-- markdownlint-disable MD046 -->
???+ danger "酌情调整 tmpfs 大小"

    默认情况下，WAL 每个分类的磁盘空间设置为 2 GiB，一般情况下，这个量基本是足够的。在 tmpfs 场景下，如果每个分类都设置这么大的内存不太现实，此处只用 1 GiB（即所有数据分类共用 1 GiB tmpfs 空间）的内存来满足 WAL 对磁盘的需求，数据量不大且网络（Datakit 和 Dataway 之间的网络）条件正常的情况下，可能也够用了。

    如果主机（或 Kubernetes Node）重启，这些 WAL 中的数据将丢失，但 Datakit 自身重启不会影响。
<!-- markdownlint-enable -->

设置完之后，我们在 cache 目录下即可看到有一个 *ramdisk* 目录，Datakit 启动后，如果有 WAL 产生，即可在 *ramdisk* 目录下看到各种数据分类对应的磁盘文件：

```shell
# 进入 /usr/local/datakit/cache 或 /mnt/wal-ramdisk/
$ du -sh ramdisk/*
8.0K    ramdisk/custom_object
0       ramdisk/dialtesting
4.0K    ramdisk/dynamic_dw
4.0K    ramdisk/fc
4.0K    ramdisk/keyevent
4.3M    ramdisk/logging
1000K   ramdisk/metric
4.0K    ramdisk/network
4.0K    ramdisk/object
4.0K    ramdisk/profiling
4.0K    ramdisk/rum
4.0K    ramdisk/security
4.0K    ramdisk/tracing
```
