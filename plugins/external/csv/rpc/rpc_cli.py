# -*- encoding: utf8 -*-

import dk_pb2_grpc
import dk_pb2

import grpc
import time

listen = '[::]:4321'
listen = 'unix:///tmp/x.sock'
listen = 'unix:///usr/local/cloudcare/DataFlux/datakit/datakit.sock'

def simpleRPC():
    chan = grpc.insecure_channel(listen)
    stub = dk_pb2_grpc.DataKitStub(chan)

    lines = b'''MeTest,tag1=t1,tag2=t2 class=1i,score=2.2,super=True,desp="ffff" 1583429123000000000
MeTest,tag1=t1,tag2=t2 class=1i,score=2.2,super=True,desp="ffff" 1583429125000000000
MeTest,tag1=t1,tag2=t2 class=1i,score=2.2,super=True,desp="ffff" 1583429126000000000
MeTest,tag1=t1,tag2=t2 class=1i,score=2.2,super=True,desp="ffff" 1583429128000000000
MeTest,tag1=t1,tag2=t2 class=1i,score=2.2,super=True,desp="ffff" 1583429129000000000'''

    while True:
        req = dk_pb2.Request(Lines=lines, Name='py-rpc-test')

        try:
            resp = stub.Send(req, None)
            print(time.time(), resp)
        except Exception:
            pass

        time.sleep(1)

if __name__ == '__main__':
    simpleRPC()
