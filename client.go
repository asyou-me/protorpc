package protorpc

import (
	//"github.com/gogo/protobuf/proto"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"net/rpc"
	//"os"
	"sync"
)

type clientCodec struct {
	c       io.ReadWriteCloser
	methods map[string]uint64

	// temporary work space
	resp []byte

	// RPC responses include the request id but not the request method.
	// Package rpc expects both.
	// We save the request method in pending when sending a request
	// and then look it up by request ID when filling out the rpc Response.
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // map request id to method name
}

// NewClientCodec returns a new rpc.ClientCodec using JSON-RPC on conn.
func NewClientCodec(conn io.ReadWriteCloser, methods *map[string]uint64) rpc.ClientCodec {
	return &clientCodec{
		c:       conn,
		pending: make(map[uint64]string),
		methods: *methods,
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {

	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	r.Seq = r.Seq

	b_buf := bytes.NewBuffer([]byte{})
	// 写入8字节Request id
	binary.Write(b_buf, binary.BigEndian, r.Seq)
	// 写入8字节Request Method
	binary.Write(b_buf, binary.BigEndian, c.methods[r.ServiceMethod])

	// 定义数据的格式为 codec_type
	p := param.(codec_type)
	b, err := p.Marshal()
	if err != nil {
		return err
	}
	// 写入包主体
	b_buf.Write(b)

	// 将数据写入tcp数据流
	_, err = c.c.Write(b_buf.Bytes())

	return err
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	b := make([]byte, 1024)
	_, err := c.c.Read(b)

	if err != nil {
		return err
	}

	// 读取8字节Response id
	c.resp = b[8:]

	// []byte to uint64
	b_buf := bytes.NewBuffer(b[:8])
	var id uint64
	binary.Read(b_buf, binary.BigEndian, &id)

	c.mutex.Lock()
	r.ServiceMethod = c.pending[id]
	delete(c.pending, id)
	c.mutex.Unlock()

	// 将得到的Response id给Response
	r.Seq = id

	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}

	// 转换字节到结构体
	p := x.(codec_type)
	p.Unmarshal(c.resp)

	return nil
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser, methods *map[string]uint64) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn, methods))
}

// Dial connects to a JSON-RPC server at the specified network address.
func Dial(network, address string, methods *map[string]uint64) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, methods), err
}
