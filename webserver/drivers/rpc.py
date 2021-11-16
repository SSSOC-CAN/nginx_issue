from app import app
import codecs
import dependencies.rpc_pb2 as r
import dependencies.rpc_pb2_grpc as rpc
from drivers.base_rpc import BaseRPC
import grpc
import os

os.environ["GRPC_SSL_CIPHER_SUITES"] = 'HIGH+ECDSA'

class RPC(BaseRPC):
    source = 0
    macaroon = 0
    def __init__(self, source):
        RPC.set_source(source)
        super(RPC, self).__init__()

    @staticmethod
    def set_source(source):
        RPC.source = source

    @staticmethod
    def frames():
        # Monkey patch for gunicorn grpc compatibility
        import grpc.experimental.gevent
        grpc.experimental.gevent.init_gevent()
        channel = grpc.insecure_channel(RPC.source)
        stub = rpc.RpcServiceStub(channel)
        for rtd in stub.SubscribeDataStream(r.SubscribeDataRequest()):
            yield rtd