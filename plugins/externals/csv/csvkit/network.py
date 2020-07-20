# -*- encoding: utf8 -*-

import threading
import requests
import os
import gzip
import datetime
import hashlib
import hmac
import base64
from csvkit.const import *

test_enable = False

def test_init(is_test):
    global test_enable
    test_enable = is_test


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

class Sender:

    def __init__(self, rpc_server=''):
        self.rpc_server = rpc_server
        chan = grpc.insecure_channel(self.rpc_server)
        self.stub = rpc.dk_pb2_grpc.DataKitStub(chan)

    def send(self, data):
        req = rpc.dk_pb2.Request(Lines=data, Name='csvkit')
        try:
            resp = self.stub.Send(req, None)
        except Exception as ex:
            pass # TODO


class Uploder:
    content_type = "text/plain"
    content_coding = "gzip"
    def __init__(self, yaml):
        self.yaml = yaml
        self.batch_size = self.yaml.get(BATCH)
        self.buff = ""
        self.count = 0

    def upload(self, data):
        if test_enable:
            return

        self.buff += data
        self.count += 1
        if self.count >= self.batch_size:
            self.send()
            self.buff = ""
            self.count = 0

    def flush(self):
        if self.buff != "":
            self.send()
            self.buff = ""
            self.count = 0

    def send(self):
        compress_data = gzip.compress(self.buff.encode())
        header = self._build_http_heraer(compress_data)
        pesponse = requests.post(url=self.yaml.get(URL), headers=header,
                                 data=compress_data)
        return pesponse.status_code

    def _build_http_heraer(self, data):
        header = {}
        date = self._http_date()
        header["Content-Encoding"] = self.content_coding
        header["Content-Type"]     = self.content_type
        header["X-Datakit-UUID"]   = self.yaml.get(UUID)
        header["Date"]             = date

        if self.yaml.get(AUTH):
            header["Authorization"] = self._make_auth(data, date)
        return header

    def _http_date(self):
        dt = datetime.datetime.utcnow()
        weekday = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"][dt.weekday()]
        month = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep",
                 "Oct", "Nov", "Dec"][dt.month - 1]
        return "{}, {:02d} {} {:4d} {:02d}:{:02d}:{:02d} GMT".format(weekday, dt.day, month,
                                                        dt.year, dt.hour, dt.minute, dt.second)

    def _make_auth(self, data, date):
        signature = "DWAY " + self.yaml[PK] + ":"
        cont_md5 = hashlib.md5(data).digest()
        cont_md5 = base64.standard_b64encode(cont_md5).decode()
        s = "POST" + "\n" + cont_md5 + "\n" + self.content_type + "\n" + date
        return signature + self._hash_hmac(self.yaml[SK], s)

    def _hash_hmac(self, key, code):
        hmac_code = hmac.new(key.encode(), digestmod=hashlib.sha1)
        hmac_code.update(code.encode())
        bs = hmac_code.digest()
        return base64.standard_b64encode(bs).decode()
