# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import network_pb2 as network__pb2


class NetworkStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetNetworkInfo = channel.unary_unary(
                '/pactus.Network/GetNetworkInfo',
                request_serializer=network__pb2.GetNetworkInfoRequest.SerializeToString,
                response_deserializer=network__pb2.GetNetworkInfoResponse.FromString,
                )
        self.GetNodeInfo = channel.unary_unary(
                '/pactus.Network/GetNodeInfo',
                request_serializer=network__pb2.GetNodeInfoRequest.SerializeToString,
                response_deserializer=network__pb2.GetNodeInfoResponse.FromString,
                )


class NetworkServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetNetworkInfo(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetNodeInfo(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_NetworkServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetNetworkInfo': grpc.unary_unary_rpc_method_handler(
                    servicer.GetNetworkInfo,
                    request_deserializer=network__pb2.GetNetworkInfoRequest.FromString,
                    response_serializer=network__pb2.GetNetworkInfoResponse.SerializeToString,
            ),
            'GetNodeInfo': grpc.unary_unary_rpc_method_handler(
                    servicer.GetNodeInfo,
                    request_deserializer=network__pb2.GetNodeInfoRequest.FromString,
                    response_serializer=network__pb2.GetNodeInfoResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'pactus.Network', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Network(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetNetworkInfo(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/pactus.Network/GetNetworkInfo',
            network__pb2.GetNetworkInfoRequest.SerializeToString,
            network__pb2.GetNetworkInfoResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetNodeInfo(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/pactus.Network/GetNodeInfo',
            network__pb2.GetNodeInfoRequest.SerializeToString,
            network__pb2.GetNodeInfoResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
