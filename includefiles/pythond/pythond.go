// Package pythond contains pythond core scripts files.
package pythond

/*

files' location like this:

├── datakit
└── python.d
    ├── core
        └── datakit_framework.py
        └── demo.py
*/

// PythonDCoreFiles python.d/core.
var PythonDCoreFiles = map[string]string{
	"datakit_framework.py": pyDatakitFramework,
	"demo.py":              pyDemo,
} // var

const (
	pyDatakitFramework = `
#encoding: utf-8

import os
import sys
from string import Template
import logging
from logging.handlers import RotatingFileHandler
import requests

logger = logging.getLogger('pythond_framework')

'''
DataKitFramework
所有plugin的基类
'''

class DataKitFramework(object):
    '''
    框架基类，提供必要功能和接口定义
    '''
    __name = 'DataKitFramework'
    __dk_host = "127.0.0.1"
    __dk_port = 9529
    __magic = "{xxx}"
    log_name = ""
    is_init_log = False

    def __init__(self, **kwargs):
        ip = kwargs.get("ip")
        if ip:
            self.__dk_host = ip
        port = kwargs.get("port")
        if port:
            self.__dk_port = port

    def run(self):
        raise NotImplementedError()

    def test(self):
        print("log_name = ", self.log_name)
        mylog("789")

    def report(self, data):
        M = ""
        L = ""
        R = ""
        O = ""
        CO = ""

        if 'M' in data:
            M = data['M']
        if 'L' in data:
            L = data['L']
        if 'R' in data:
            R = data['R']
        if 'O' in data:
            O = data['O']
        if 'CO' in data:
            CO = data['CO']
        if M is None and L is None and R is None and O is None and CO is None:
            return

        precision = ""
        input = ""
        ignore_global_tags = ""
        version = ""

        if 'precision' in data:
            precision = data['precision']
        if 'input' in data:
            input = data['input']
        if 'ignore_global_tags' in data:
            ignore_global_tags = data['ignore_global_tags']
        if 'version' in data:
            version = data['version']

        s = Template('http://${s1}:${s2}/v1/write/${s3}?')
        origin_url = s.safe_substitute(s1=self.__dk_host, s2=self.__dk_port, s3=self.__magic)
        if precision:
            origin_url += "precision=" + precision + '&'
        if input:
            origin_url += "input=" + input + '&'
        if ignore_global_tags:
            origin_url += "ignore_global_tags=" + ignore_global_tags + '&'
        if version:
            origin_url += "version=" + version + '&'

        if origin_url[-1] in ('&', '?'):
            origin_url = origin_url[:-1]
        url = origin_url[:-1]

        response = ""

        if M:
            url = origin_url.replace(self.__magic, "metric")
            response = self.http_post_json(url, M)
        if L:
            url = origin_url.replace(self.__magic, "logging")
            response = self.http_post_json(url, L)
        if R:
            url = origin_url.replace(self.__magic, "rum")
            response = self.http_post_json(url, R)
        if O:
            url = origin_url.replace(self.__magic, "object")
            response = self.http_post_json(url, O)
        if CO:
            url = origin_url.replace(self.__magic, "custom_object")
            response = self.http_post_json(url, CO)

        return response


    def construct_url(self, path):
        s = Template('http://${s1}:${s2}/${s3}')
        return s.safe_substitute(s1=self.__dk_host, s2=self.__dk_port, s3=path)


    def set_lasterror(self, input_name, err_msg):
        base_url = self.construct_url("v1/lasterror")
        raw_data = {
            "input":input_name,
            "err_content":err_msg
        }
        return self.http_post_json(base_url, raw_data)

    def http_post_json(self, url, raw_data):
        return self.http_post_data_core(url, raw_data, 'POST', True)

    def http_post_raw(self, url, raw_data):
        return self.http_post_data_core(url, raw_data, 'POST', False)

    def http_post_data_core(self, url, raw_data, method, is_json):
        headers = {}
        send_data = bytes()
        if is_json is True:
            headers = {"Content-Type":"application/json"}
        else:
            send_data = bytes(str(raw_data),'utf8')

        html = requests.Response()
        try:
            if method == 'POST':
                if is_json is True:
                    html = requests.post(url=url, headers=headers, json=raw_data)
                else:
                    html = requests.post(url=url, headers=headers, data=send_data)
            elif method == 'GET':
                if is_json is True:
                    html = requests.get(url=url, params=send_data)
                else:
                    html = requests.get(url=url, params=send_data)
        except requests.exceptions.RequestException as e:
            mylog("requests.RequestException:", e.strerror)
        except:
            mylog("Unexpected error:", sys.exc_info()[0])

        return html.text

def init_log():
    log_path = os.path.join(os.path.expanduser('~'), "_datakit_pythond_framework_" + DataKitFramework.log_name + "_.log")
    print(log_path)
    logger.setLevel(logging.DEBUG)
    handler = RotatingFileHandler(log_path, maxBytes=100000, backupCount=10)
    logger.addHandler(handler)

def mylog(msg, *args, **kwargs):
    if DataKitFramework.is_init_log is False:
        init_log()
        DataKitFramework.is_init_log = True
    logger.debug(msg, *args, **kwargs)
`

	pyDemo = `
from datakit_framework import DataKitFramework

class Demo(DataKitFramework):
    __name = 'Demo'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

    def run(self):
        print("Demo")
        data = [
                {
                    "measurement": "abc",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                },

                {
                    "measurement": "def",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                }
            ]

        in_data = {
            'M':data,
            'input': "datakitpy"
        }

        return self.report(in_data) # you must call self.report here
`
) // const
