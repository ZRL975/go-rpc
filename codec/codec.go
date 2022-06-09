package codec

import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

//编解码器接口

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType Type = "application/gob"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
