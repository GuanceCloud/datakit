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

    def send_task_event(self, host, res):
        tags = {
            "host": host,
            "__status": res.get('__status'),
            "ansible_status": res.get('ansible_status'),
            "__source": "ansible"
        }
        fields = {
            "__content": json.dumps(res),
            "__title": "Ansible task {} on {}".format(res.get("ansible_status"), host)
        }
        format_string = "%Y-%m-%d %H:%M:%S.%f"
        start = res.get("start")
        if not start:
            return
        time_stamp = datetime.datetime.strptime(start, format_string)
        line = self.make_line_protocol("__keyevent", tags, fields, time_stamp)
        url = "{}/ansible?type=keyevent".format(self.datakit_host)
        requests.post(url, data=line.encode("utf-8"),timeout=5)

    def runner_on_failed(self, host, res, ignore_errors=False):
        if ignore_errors:
            return
        if 'failed' not in self.output_task_stats:
            return
        res['__status'] = "error"
        res['ansible_status'] = "failed"
        self.send_task_event(host, res)

    def runner_on_ok(self, host, res):
        if 'ok' not in self.output_task_stats:
            return
        res['__status'] = "info"
        res['ansible_status'] = "ok"
        self.send_task_event(host, res)

    def runner_on_unreachable(self, host, res):
        if 'unreachable' not in self.output_task_stats:
            return
        res['__status'] = "error"
        res['ansible_status'] = "unreachable"
        self.send_task_event(host, res)

    def v2_playbook_on_play_start(self, play):
        self.play_book_name = play.name
        tags = {
            "__status": "info",
            "ansible_status": "play_start",
            "__source": "ansible"
        }
        fields = {
            "__title": "Ansible playbook {} start".format(self.play_book_name)
        }
        line = self.make_line_protocol("__keyevent", tags, fields)
        url = "{}/ansible?type=keyevent".format(self.datakit_host)
        requests.post(url, data=line.encode("utf-8"),timeout=5)

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
