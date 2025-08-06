---
title     : 'SCheck'
summary   : 'Collect Securiry data from SCheck'
tags:
  - 'SECURITY'
__int_icon: 'icon/scheck'
---

Operating system support: :fontawesome-brands-linux: :fontawesome-brands-windows:

---

DataKit has direct access to Security Checker's data. For specific use of Security Checker, see [here](../scheck/scheck-install.md).
<!-- markdownlint-disable MD013 -->
## To Install the Security Checker Installation Through the DataKit {#install}
<!-- markdownlint-enable -->
```shell
sudo datakit install --scheck
```

After installation, security Checker sends data to the DataKit `:9529/v1/write/security` interface by default.
