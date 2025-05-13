# Using tmpfs in WAL

:fontawesome-brands-linux: :material-kubernetes:

---

In some cases, if the disk performance is insufficient, we can use a portion of memory to meet the disk read/write requirements of WAL.

<!-- markdownlint-disable MD046 -->
=== "Linux Host"

    1. Create a tmpfs mount directory on the host:
    
        ```shell
        sudo mkdir -p /mnt/wal-ramdisk
        ```
    
    1. Create a tmpfs disk using 1 GiB of memory space:
    
        ```shell
        sudo mount -t tmpfs -o size=1G tmpfs /mnt/wal-ramdisk
        ```
    
    1. Check the mount status:
    
        ```shell
        $ df -h /mnt/wal-ramdisk/
        Filesystem      Size  Used Avail Use% Mounted on
        tmpfs           1.0G     0  1.0G   0% /mnt/wal-ramdisk
        ```
    
    1. Modify the WAL directory in *datakit.conf*:
    
        ```toml hl_lines="2"
        [dataway.wal]
          path = "/mnt/wal-ramdisk"
        ```

=== "DaemonSet"

    1. On the Kubernetes Node, create a *ramdisk* directory:

        ```shell
        # Here we change the directory; by default, in datakit.yaml, the Node's /root/datakit_cache
        # is already mounted to the /usr/local/datakit/cache directory in the DataKit container
        mkdir -p /root/datakit_cache/ramdisk
        ```
    
    1. Create a 1 GiB tmpfs:

        ```shell
        mount -t tmpfs -o size=1G tmpfs /root/datakit_cache/ramdisk
        ```
    
    1. Add the following environment variable in *datakit.yaml*, then restart the DataKit container:

        ```yaml
        - name: ENV_DATAWAY_WAL_PATH
          value: /usr/local/datakit/cache/ramdisk
        ```
<!-- markdownlint-enable -->

---

<!-- markdownlint-disable MD046 -->
???+ danger "Adjust tmpfs size as needed"

    By default, the disk space for each category in WAL is set to 2 GiB, which is generally sufficient. In a tmpfs scenario, it may not be practical to allocate such a large amount of memory for each category. Here, only 1 GiB (i.e., all data categories share 1 GiB of tmpfs space) of memory is used to meet the disk requirements of WAL. This may be enough under conditions where the data volume is not large and the network (between DataKit and Dataway) is ok.

    If the host (or Kubernetes Node) restarts, the data in WAL will be lost, but a DataKit restart will not affect this.
<!-- markdownlint-enable -->

After setting this up, you will see a *ramdisk* directory in the cache directory. Once DataKit starts, if WAL is generated, you will see various data category disk files in the *ramdisk* directory:

```shell
# Enter /usr/local/datakit/cache or /mnt/wal-ramdisk/
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
