# DataKit Security Statement

As a system-level observability agent deployed on your nodes, DataKit's core mission is to provide you with the deepest and most comprehensive insights into your infrastructure. To achieve this goal—such as collecting host process metadata, tracing network flows, utilizing eBPF to capture kernel events, and accessing protected system files—DataKit inevitably requires higher privileges than ordinary stateless applications.

We fully acknowledge that powerful privileges come with significant security responsibilities. This document aims to transparently explain the purpose of each privilege, clarify its direct link to specific collection capabilities, and provide suggestions for adjustments based on your security baseline.

---

## Privilege Details {#permission-details}

The following are common alert items triggered by security scanning tools, along with explanations of their necessity for DataKit's core functionality.

### Pod-Level Privileges {#pod-permissions}

These privileges breach the Pod isolation boundary and are crucial for DataKit to collect host information as a node agent.

1. **Shared Host PID Namespace (`hostPID: true`)**
    * **Purpose**: Allows DataKit to view all processes on the node, not just its own.
    * **Dependent Features**:
        * **eBPF Network Tracing**: Needs to associate network traffic (TCP/UDP connections) with specific processes (PIDs), thereby solving critical questions like "Which Pod of which service initiated this network connection."

1. **Shared Host Network Namespace (`hostNetwork: true`)**
    * **Purpose**: Allows DataKit to directly access and listen on the node's network interfaces.
    * **Dependent Features**:
        * **eBPF Network Tracing (`netflow`, `bpf`)**: Must operate under the host network to capture all incoming and outgoing traffic on the node.
        * **Convenient Data Reception**: Allows services like StatsD and OpenTelemetry to send data directly to `NodeIP:Port`, simplifying service discovery and network configuration.

1. **Privileged Mode (`securityContext.privileged: true`)**
    * **Purpose**: This is the most powerful privilege, granting the container kernel access capabilities almost equivalent to the host root user.
    * **Dependent Features**:
        * **eBPF Monitoring**: Loading and running eBPF programs is a privileged operation requiring direct interaction with the kernel. This is a mandatory requirement for enabling eBPF functionality.
        * **Accessing Debug Filesystem**: Accessing paths like `/sys/kernel/debug` to obtain deep kernel and hardware metrics.

### Container-Level Privileges {#container-permissions}

These privileges define the behavior and capabilities within the DataKit container itself.

1. **Start Container as Root User**
    * **Purpose**: Run with `root` privileges to read protected system files and directories.
    * **Dependent Features**: Accessing process information under `/proc`, kernel parameters under `/sys`, and various log files under `/var/log`. This is the foundation for most node-level metric and log collection.

1. **Listen on Node Host Port (`ports.hostPort`)**
    * **Purpose**: Expose the container's port directly on the node's IP address.
    * **Dependent Features**: Consistent with the goal of `hostNetwork: true`, providing a stable and easily accessible entry point for external services to report data.

1. **Writable Filesystem (`readOnlyRootFilesystem: false`)**
    * **Purpose**: Allows DataKit to write data within its container filesystem.
    * **Dependent Features**:
        * **Data Caching**: Write permissions are required for cache paths specified in `volumeMounts` (e.g., `/usr/local/datakit/cache`).
        * **Dynamic Content**: Plugins downloaded at runtime (such as Python collection scripts) or generated temporary configuration files.
        * **Internal Logs**: Recording logs regarding DataKit's own operational status.

---

## Configuration Recommendations {#recommends}

We fully understand and respect your organization's security policies. DataKit is designed to be modular; you can disable specific high-privilege settings according to your needs, though this will directly impact the corresponding data collection capabilities.

To help you make informed decisions, the table below summarizes the correspondence between major features and required privileges:

| If you do not need this feature | You can safely modify the DaemonSet YAML as follows |
| :--- | :--- |
| **Host Process Metrics (`host_processes`)** | Set `hostPID: false` |
| **eBPF Monitoring (Network, Security)** | Set `securityContext.privileged: false` and `hostPID: false`, and remove `ebpf`-related collectors from `ENV_DEFAULT_ENABLED_INPUTS`. |
| **Host Network Metrics (`net`)** | Set `hostNetwork: false` and remove the `ports.hostPort` definition. |
| **Receive data directly on the Node (StatsD, OpenTelemetry)** | Set `hostNetwork: false` and remove `ports.hostPort`. You can still expose ports via Kubernetes `Service`. |

**Recommendations**:

* **Configure on Demand**: Please refer to the table above and trim DataKit's privileges based on your actual monitoring requirements before deployment.
* **Environment Isolation**: We recommend deploying DataKit in a trusted environment. Using Kubernetes `Tolerations` and `NodeSelector`, you can restrict its deployment to specific node pools rather than the entire cluster.
* **Stay Updated**: Please keep an eye on our official releases for the latest security updates and best practices.

We are committed to maximizing respect for your security needs while providing powerful functionality. If you have any further questions, please feel free to contact us.
