# -*- encoding: utf8 -*-

import threading
import requests
import os
import logging
from urllib.parse import urljoin

METRICS_PATH = "/v1/write/metric"
OBJECTS_PATH = "/v1/write/object"
PATH_FORMAT = "{}?name=csvkit"

HTTP_ADDR = ""

def set_http_addr(addr):
    global HTTP_ADDR
    HTTP_ADDR = addr


class Downloder:
    def __init__(self, *args):
        self.down_files = args

    def down(self):
        thread = []
        down_succ = []
        down_fail = []

        for url, file_path in self.down_files:
            t = threading.Thread(target=self.down_task, args=(url, file_path))
            thread.append(t)
            t.start()

        for t in thread:
            t.join()

        for url, file_path in self.down_files:
            if os.path.isfile(file_path):
                down_succ.append((url, file_path))
            else:
                down_fail.append((url, file_path))

        return down_succ, down_fail

    def down_task(self, url, file_path):
        r = requests.get(url)
        with open(file_path, "wb") as f:
            f.write(r.content)

# class Sender:
#     def __init__(self, rpc_server=''):
#         chan = grpc.insecure_channel(rpc_server)
#         self.sender = dk_pb2_grpc.DataKitStub(chan)
#
#     def send_metrics(self, data):
#         req = dk_pb2.Request(Lines=data, Name='csvkit', io = dk_pb2.METRIC)
#         try:
#             resp = self.sender.Send(req, None)
#         except Exception as ex:
#             pass # TODO
#
#     def send_objects(self, data):
#         req = dk_pb2.Request(Objects=data, Name='csvkit', io = dk_pb2.OBJECT)
#         try:
#             resp = self.sender.Send(req, None)
#         except Exception as ex:
#             pass # TODO

class Sender:
    def __init__(self):
        self._metrics_path = urljoin(HTTP_ADDR, PATH_FORMAT.format(METRICS_PATH))
        self._objects_path = urljoin(HTTP_ADDR, PATH_FORMAT.format(OBJECTS_PATH))

    def send_metrics(self, data):
        try:
            resp = requests.post(self._metrics_path, data)
        except Exception as ex:
            logging.error("send {} exception {}".format(self._metrics_path, ex))
        else:
            logging.info("send {} {}".format(self._metrics_path, resp))

    def send_objects(self, data):
        try:
            resp = requests.post(self._objects_path, data)
        except Exception as ex:
            logging.error("send {} exception {}".format(self._objects_path, ex))
        else:
            logging.info("send {} {}".format(self._objects_path, resp))
