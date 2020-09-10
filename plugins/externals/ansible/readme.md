
## 前置条件

- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）
- python 脚本 在 datakit 安装目录下的 externals/ansible/task_info.py 中
- 需要将 python 脚本放到 ansible.cfg callback plugin 指定目录  callback_plugins = /x/x
- 安装 python脚本所需环境 pip install -r requirement.txt
- ansible 中需配置 bin_ansible_callbacks = True
- ansible 中需启动callback plugin 白名单 callback_whitelist = task_info

## 配置


设置：

    python 脚本中  DOCUMENTATION
    callback: task_info
        type: notification
        short_description: write playbook output to es
        description:
            - this plugins will output task_info to datakit
        requirements:
            - pip install -r requests.txt
            - set Whitelist in ansible.cfg
            - move plugins info  callback dir set in ansible.cfg
        options:
            output_task_stats:
                version_added: '2.9'
                default: ["unreachable","failed","ok"]   # 将列表中三种task状态的数据打入 keyevent 更改此处即可
                description: choose task stats
                env:
                  - name: ANSIBLE_TASK_INFO
                ini:
                  - section: callback_task_info
                    key: output_task_stats
            datakit_host :
                version_added: '2.9'
                default: 'http://0.0.0.0:9529'  # 配置 datakit host 要跟 datakit配置一致
                description: choose task stats
                env:
                  - name: ANSIBLE_TASK_INFO
                ini:
                  - section: callback_task_info
                    key: datakit_host



                            
