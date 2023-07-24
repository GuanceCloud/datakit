
# Scheck Access
---

Operating system support: :fontawesome-brands-linux: :fontawesome-brands-windows:

---

Datakit has direct access to Security Checker's data. For specific use of Security Checker, see [here](../scheck/scheck-install.md).

## To Install the Security Checker Installation Through the DataKit {#install}

```shell
$ sudo datakit install --scheck
```

After installation, security Checker sends data to the DataKit `:9529/v1/write/security` interface by default.
