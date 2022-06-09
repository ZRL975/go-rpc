package go_rpc

import (
	"encoding/json"
	"fmt"
	"go-rpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x4b3898bfc

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) serveConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	//使用json对opt进行解码
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: option error:", err)
		return
	}
	//检查MagicNumber是否正确
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	//获取编解码器并且判断是否存在
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handlerRequest(cc, req, sending, wg)
	}
	wg.Wait()
	cc.Close()
}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}
func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.replyv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}
func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}
func (s *Server) handlerRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	//todo 根据对应的method进行处理
	defer wg.Done()
	req.replyv = reflect.ValueOf(fmt.Sprintf("go-rpc resp %d", req.h.Seq)) //对replyv进行赋值
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.serveConn(conn)
	}

}

//对外暴露的调用函数

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
