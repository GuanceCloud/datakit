
# Bulk Deployment
---

We can install DataKit in batches through Ansible and so on.

## Preconditions {#requirements}

- Installing Ansible on the management machine.
- Configure the `host` file and the `install.yaml` `/etc/ansible/` under the default configuration path of Ansible.
- If you manage Windows machines through Ansible, refer to [Ansible doc](https://ansible-tran.readthedocs.io/en/latest/docs/intro_windows.html#windows-installing){:target="_blank"} for corresponding pre-preparation.

## Configuration {#config}

Ansible `host` file configuration examples:

```toml
[linux]
# ansible_become_pass user password is not specified as root by default, but can be specified by become_user (refer to official documents for details)
10.200.6.58    ansible_ssh_user=xxx   ansible_ssh_pass=xxx   ansible_become_pass=xxx
10.100.64.117  ansible_ssh_user=xxx   ansible_ssh_pass=xxx   ansible_become_pass=xxx

[windows]
# ansible_connection using winrm (refer to official documentation for details)
10.100.65.17 ansible_ssh_user="xxx" ansible_ssh_pass="xxx" ansible_ssh_port=5986 ansible_connection="winrm" ansible_winrm_server_cert_validation=ignore
```

Ansible `install.yaml` file configuration example:

```yaml
- hosts: linux # here corresponds to the linux machine in the host configuration file
  become: true
  gather_facts: no
  tasks:
  - name: install
    # here the linux machine in the host configuration file corresponds to the batch installation of the shell here. By specifying the dataway address, the default open host collector (cpu, disk, mem) and so on, the -global-tags host=__datakit_hostname and so on are set
    shell: DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    async: 120  # representing the upper limit of the execution time of this task. That is, if the time taken by the task to execute exceeds this time, the task is considered to have failed. If this parameter is not set, the poll is executed synchronously: 10 # representing the polling time interval when the task is executed asynchronously, and if the poll is 0, it is equivalent to a task that does not care about the result

- hosts: windows # here corresponds to the Windows machine in the host configuration file
  gather_facts: no
  tasks:
  - name: install
    win_shell: $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    async: 120
    poll: 10
```

## Deployment {#deploy}

Batch deployment can be realized by running the following command on the management machine:

```shell
ansible-playbook -i hosts install.yaml
```
