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
    name = 'DataKitFramework'
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
        E = ""

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
        if 'E' in data:
            E = data['E']
        if M is None and L is None and R is None and O is None and CO is None and E is None:
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
        if E:
            url = origin_url.replace(self.__magic, "keyevent")
            response = self.http_post_json(url, E)

        return response


    def checkArgEmpty(self, name, checkStr):
        if not checkStr:
            raise ValueError('arguments missing \"' + name + '\"')


    def feed_user_event(self, df_user_id=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs):
        self.checkArgEmpty("df_user_id", df_user_id)
        self.checkArgEmpty("df_date_range", df_date_range)
        self.checkArgEmpty("df_status", df_status)
        self.checkArgEmpty("df_event_id", df_event_id)
        self.checkArgEmpty("df_title", df_title)
        if len(tags) == 0:
            tags = {}
        if not df_message:
            df_message = ""

        data = {
            "measurement": "measurement",
            "tags": tags,
            "fields": {
                "df_date_range": df_date_range,
                "df_source": "user",
                "df_user_id": df_user_id,
                "df_status": df_status,
                "df_event_id": df_event_id,
                "df_title": df_title,
                "df_message": df_message,
            }
        }
        for key, value in kwargs.items():
            data["fields"][key] = value

        dataArr = [data]

        in_data = {
            'E': dataArr,
        }

        return self.report(in_data)


    def feed_monitor_event(self, df_dimension_tags=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs):
        self.checkArgEmpty("df_dimension_tags", df_dimension_tags)
        self.checkArgEmpty("df_date_range", df_date_range)
        self.checkArgEmpty("df_status", df_status)
        self.checkArgEmpty("df_event_id", df_event_id)
        self.checkArgEmpty("df_title", df_title)
        if len(tags) == 0:
            tags = {}
        if not df_message:
            df_message = ""

        data = {
            "measurement": "measurement",
            "tags": tags,
            "fields": {
                "df_date_range": df_date_range,
                "df_source": "monitor",
                "df_dimension_tags": df_dimension_tags,
                "df_status": df_status,
                "df_event_id": df_event_id,
                "df_title": df_title,
                "df_message": df_message,
            }
        }
        for key, value in kwargs.items():
            data["fields"][key] = value

        dataArr = [data]

        in_data = {
            'E': dataArr,
        }

        return self.report(in_data)


    def feed_system_event(self, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs):
        self.checkArgEmpty("df_date_range", df_date_range)
        self.checkArgEmpty("df_status", df_status)
        self.checkArgEmpty("df_event_id", df_event_id)
        self.checkArgEmpty("df_title", df_title)
        if len(tags) == 0:
            tags = {}
        if not df_message:
            df_message = ""

        data = {
            "measurement": "measurement",
            "tags": tags,
            "fields": {
                "df_date_range": df_date_range,
                "df_source": "system",
                "df_status": df_status,
                "df_event_id": df_event_id,
                "df_title": df_title,
                "df_message": df_message,
            }
        }
        for key, value in kwargs.items():
            data["fields"][key] = value

        dataArr = [data]

        in_data = {
            'E': dataArr,
        }

        return self.report(in_data)


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