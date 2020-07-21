# -*- encoding: utf8 -*-

import threading
import requests
import os
import grpc

from rpc import dk_pb2_grpc
from rpc import dk_pb2

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
        chan = grpc.insecure_channel(rpc_server)
        self.sender = dk_pb2_grpc.DataKitStub(chan)

    def send(self, data):
        req = dk_pb2.Request(Lines=data, Name='csvkit')
        try:
            resp = self.sender.Send(req, None)
        except Exception as ex:
            pass # TODO
