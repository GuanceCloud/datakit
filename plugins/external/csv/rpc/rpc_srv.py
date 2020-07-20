# -*- encoding: utf8 -*-

from concurrent import futures

import grpc

import dk_pb2_grpc
import dk_pb2

class ServerImpl(dk_pb2_grpc.DataKitServicer):
    def __init__(self):
        pass

    def Send(self, request, ctx):
        print(request)
        return dk_pb2.Response(err="", points=1, objects=0)


listen = '[::]:4321'
listen = '/tmp/x.sock'
listen = 'unix:///usr/local/cloudcare/DataFlux/datakit/datakit.sock'

def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  dk_pb2_grpc.add_DataKitServicer_to_server(ServerImpl(), server)
  server.add_insecure_port(listen)
  server.start()
  server.wait_for_termination()

if __name__ == '__main__':
    serve()
