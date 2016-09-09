package protorpc

import (
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"sync"
	"time"

	wire "github.com/asyou-me/protorpc/protobuf"
	"github.com/golang/protobuf/proto"
)

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	// temporary work space
	reqHeader wire.RequestHeader

	// Package rpc expects uint64 request IDs.
	// We assign uint64 sequence numbers to incoming requests
	// but save the original request ID in the pending map.
	// When rpc responds, we use the sequence number in
	// the response to find the original request ID.
	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]uint64
	auth    AuthorizationFunc
}

// NewServerCodec returns a serverCodec that communicates with the ClientCodec
// on the other end of the given conn.
func NewServerCodec(conn io.ReadWriteCloser, auth AuthorizationFunc) rpc.ServerCodec {
	return &serverCodec{
		r:       conn,
		w:       conn,
		c:       conn,
		pending: make(map[uint64]uint64),
		auth:    auth,
	}
}

// ReadRequestHeader 读取请求头
func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	header := wire.RequestHeader{}
	err := readRequestHeader(c.r, &header)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = header.Id
	r.ServiceMethod = header.Method
	r.Seq = c.seq
	c.mutex.Unlock()

	c.reqHeader = header

	err = c.auth(&AuthorizationHeader{
		Token:  header.Token,
		Method: header.Method,
	})
	if err != nil {
		go func() {
			// 当用户授权 token 无效时直接结束当前连接
			time.Sleep(time.Millisecond * 100)
			c.Close()
		}()
		return err
	}

	return nil
}

// 读取请求内容
func (c *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	request, ok := x.(proto.Message)
	if !ok {
		return fmt.Errorf(
			"protorpc.ServerCodec.ReadRequestBody: %T does not implement proto.Message",
			x,
		)
	}

	err := readRequestBody(c.r, &c.reqHeader, request)
	if err != nil {
		return nil
	}

	c.reqHeader = wire.RequestHeader{}
	return nil
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. It is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

// WriteResponse 写入结果到客户端
func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		if response, ok = x.(proto.Message); !ok {
			if _, ok = x.(struct{}); !ok {
				c.mutex.Lock()
				delete(c.pending, r.Seq)
				c.mutex.Unlock()
				return fmt.Errorf(
					"protorpc.ServerCodec.WriteResponse: %T does not implement proto.Message",
					x,
				)
			}
		}
	}

	c.mutex.Lock()
	id, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("protorpc: invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	err := writeResponse(c.w, id, r.Error, response)
	if err != nil {
		return err
	}

	return nil
}

// Close 关闭一个服务
func (c *serverCodec) Close() error {
	return c.c.Close()
}
