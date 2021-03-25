# coding:utf-8
from __future__ import absolute_import, print_function

__metaclass__ = type

import datetime
import json

import requests

from ansible.plugins.callback import CallbackBase
from influxdb.line_protocol import make_line

DOCUMENTATION = '''
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
            default: ["unreachable","failed","ok"]
            description: choose task stats
            env:
              - name: ANSIBLE_TASK_INFO
            ini:
              - section: callback_task_info
                key: output_task_stats
        datakit_host :
            version_added: '2.9'
            default: 'http://0.0.0.0:9529'
            description: choose task stats
            env:
              - name: ANSIBLE_TASK_INFO
            ini:
              - section: callback_task_info
                key: datakit_host
    '''


class CallbackModule(CallbackBase):
    def __init__(self):
        super(CallbackBase).__init__()
        self.play_book_name = None

    def set_options(self, task_keys=None, var_options=None, direct=None):
        """
          set config
        """
        super(CallbackModule, self).set_options(task_keys=task_keys, var_options=var_options, direct=direct)
        self.output_task_stats = self.get_option("output_task_stats")
        self.datakit_host = self.get_option("datakit_host")

    def make_line_protocol(self, measurement, tags, fields, time_stamp=None):
        if not time_stamp:
            time_stamp = datetime.datetime.utcnow()
        return make_line(measurement, tags, fields, time_stamp)



    def v2_playbook_on_stats(self, stats):
        url = "{}/ansible?type=metric".format(self.datakit_host)
        for host in stats.processed:
            tags = {
                "host": host,
                "play_book_name": self.play_book_name
            }
            fields = stats.summarize(host)
            line = self.make_line_protocol("play_book_stats", tags, fields)
            requests.post(url, data=line.encode("utf-8"), timeout=5)
