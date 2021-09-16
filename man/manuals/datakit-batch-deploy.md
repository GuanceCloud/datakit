{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

我们可以通过 Ansible 等方式来批量安装 DataKit。

## 前置条件

- 管理机安装 Ansible
- 在 Ansible 默认配置路径 `/etc/ansible/` 下配置好 `host` 文件跟 `install.yaml` 文件
- 如果通过 Ansible 管理 Windows 机器，参考 [Ansible 文档](https://ansible-tran.readthedocs.io/en/latest/docs/intro_windows.html#windows-installing) 做相应前置准备

## 配置

Ansible `host` 文件配置示例：

```toml
[linux]
# ansible_become_pass 提权用户密码 默认不指定为 root 可以通过 become_user 指定(具体参照官方文档)
10.200.6.58    ansible_ssh_user=xxx   ansible_ssh_pass=xxx   ansible_become_pass=xxx
10.100.64.117  ansible_ssh_user=xxx   ansible_ssh_pass=xxx   ansible_become_pass=xxx

[windows]
# ansible_connection 连接使用 winrm(具体参照官方文档)
10.100.65.17 ansible_ssh_user="xxx" ansible_ssh_pass="xxx" ansible_ssh_port=5986 ansible_connection="winrm" ansible_winrm_server_cert_validation=ignore
```

Ansible `install.yaml` 文件配置示例

```yaml
- hosts: linux # 此处对应 host配置文件中 linux机器
  become: true
  gather_facts: no
  tasks:
  - name: install
    # 此处的 shell 为批量安装，通过指定 dataway 地址、默认开启的主机采集器(cpu,disk,mem)等，设置了 -global-tags host=__datakit_hostname 等
    shell: DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    async: 120  # 代表了这个任务执行时间的上限值。即任务执行所用时间如果超出这个时间，则认为任务失败。此参数若未设置，则为同步执行 poll: 10 # 代表了任务异步执行时轮询的时间间隔，如果poll为0，就相当于一个不关心结果的任务

- hosts: windows # 此处对应 host 配置文件中 Windows 机器
  gather_facts: no
  tasks:
  - name: install
		win_shell: $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    async: 120
    poll: 10
```

## 部署

在管理机上运行如下命令即可实现批量部署：

```shell
ansible-playbook -i hosts install.yaml
```
