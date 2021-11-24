from app import app
import codecs
import dependencies.rpc_pb2 as r
import dependencies.rpc_pb2_grpc as rpc
from drivers.base_rpc import BaseRPC
import grpc
import json
import os

os.environ["GRPC_SSL_CIPHER_SUITES"] = 'HIGH+ECDSA'

import json

class Object:
    def toJson(self):
        return json.dumps(self, default=lambda o: o.__dict__, 
            sort_keys=True, indent=4)

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

    def MessageToJson(self, rtd):
        rtdJson = Object()
        rtdJson.source = rtd.source
        rtdJson.is_scanning = rtd.is_scanning
        rtdJson.timestamp = rtd.timestamp
        rtdJson.data = {}
        # data = [a for a in dir(rtd.data) if not a.startswith('__')]
        # print(data)
        # print(rtd.data.items)
        # print(rtd.data.keys)
        # print(rtd.data.values)
        # for d in data:
        #     rtdJson.data[int(d)] = Object()
        #     rtdJson.data[int(d)].name = rtd.data[int(d)].name
        #     rtdJson.data[int(d)].value = rtd.data[int(d)].value
        # print(type(rtd.data), rtd.data[1])
        for key, value in rtd.data.items():
            rtdJson.data[int(key)] = Object()
            rtdJson.data[int(key)].name = value.name
            rtdJson.data[int(key)].value = value.value
        return rtdJson.toJson()

    def MessageToJsonV2(self, rtd):
        rtdJsonDict = {}
        rtdJsonDict["source"] = rtd.source
        rtdJsonDict["is_scanning"] = rtd.is_scanning
        rtdJsonDict["timestamp"] = rtd.timestamp
        rtdJsonDict["data"] = {}
        for key, v in rtd.data.items():
            rtdJsonDict["data"][int(key)] = {}
            rtdJsonDict["data"][int(key)]["name"] = v.name
            rtdJsonDict["data"][int(key)]["value"] = v.value
        return rtdJsonDict
