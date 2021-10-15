package protobuf

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/golang/protobuf/proto"
)

type ProtobufCodec struct {
}

// 编码器的名称
func (self *ProtobufCodec) Name() string {
	return "protobuf"
}

func (self *ProtobufCodec) MimeType() string {
	return "application/x-protobuf"
}

func (self *ProtobufCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {
	return proto.Marshal(msgObj.(proto.Message))
}

func (self *ProtobufCodec) Decode(data interface{}, msgObj interface{}) error {
	return proto.Unmarshal(data.([]byte), msgObj.(proto.Message))
}

func init() {
	// 注册编码器
	codec.RegisterCodec(new(ProtobufCodec))
}