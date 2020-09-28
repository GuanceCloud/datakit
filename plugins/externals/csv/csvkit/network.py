# -*- encoding: utf8 -*-
import requests
import logging
from urllib.parse import urljoin

METRICS_PATH = "/v1/write/metric"
OBJECTS_PATH = "/v1/write/object"
PATH_FORMAT  = "{}?name={}"

HTTP_ADDR = ""
def set_http_addr(addr):
    global HTTP_ADDR
    HTTP_ADDR = addr


class Downloder:
    def __init__(self, file):
        self.down_files = file

    def down(self):
        r = requests.get(self.down_files[0])
        with open(self.down_files[1], "wb") as f:
            f.write(r.content)


class MetricSender:
    def __init__(self):
        self._metrics_path = urljoin(HTTP_ADDR, PATH_FORMAT.format(METRICS_PATH, "csvmetric"))

    def send(self, data):
        try:
            resp = requests.post(self._metrics_path, data.encode())
        except Exception as ex:
            logging.error("send {} exception {}".format(self._metrics_path, ex))
        else:
            logging.info("send {} {}".format(self._metrics_path, resp))


class ObjectSender:
    def __init__(self):
        self._objects_path = urljoin(HTTP_ADDR, PATH_FORMAT.format(OBJECTS_PATH, "csvobject"))

    def send(self, data):
        try:
            resp = requests.post(self._objects_path, data.encode())
        except Exception as ex:
            logging.error("send {} exception {}".format(self._objects_path, ex))
        else:
            logging.info("send {} {}".format(self._objects_path, resp))