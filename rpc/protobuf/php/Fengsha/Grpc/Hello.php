<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: hello.proto

namespace Fengsha\Grpc;

class Hello
{
    public static $is_initialized = false;

    public static function initOnce() {
        $pool = \Google\Protobuf\Internal\DescriptorPool::getGeneratedPool();

        if (static::$is_initialized == true) {
          return;
        }
        $pool->internalAddGeneratedFile(
            "\x0A\xE5\x01\x0A\x0Bhello.proto\x12\x0Ahelloworld\"\x1C\x0A\x0CHelloRequest\x12\x0C\x0A\x04name\x18\x01 \x01(\x09\"\x1D\x0A\x0AHelloReply\x12\x0F\x0A\x07message\x18\x01 \x01(\x092I\x0A\x07Greeter\x12>\x0A\x08SayHello\x12\x18.helloworld.HelloRequest\x1A\x16.helloworld.HelloReply\"\x00B:Z\x1Ago_utils/rpc/service/hello\xCA\x02\x0CFengsha\\Grpc\xE2\x02\x0CFengsha\\Grpcb\x06proto3"
        , true);

        static::$is_initialized = true;
    }
}
