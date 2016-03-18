package protorpc

import (
	"bytes"
	"errors"
	//"github.com/gogo/protobuf/proto"
	"encoding/binary"
	"io"
	"net/rpc"
	"sync"
)

var errMissingParams = errors.New("jsonrpc: request body missing params")

type serverCodec struct {
	c io.ReadWriteCloser

	methods map[uint64]string

	// temporary work space
	req []byte

	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]*[]byte
}

// NewServerCodec returns a new rpc.ServerCodec using JSON-RPC on conn.
func NewServerCodec(conn io.ReadWriteCloser, methods *map[uint64]string) rpc.ServerCodec {
	return &serverCodec{
		c:       conn,
		pending: make(map[uint64]*[]byte),
		methods: *methods,
	}
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {

	b := make([]byte, 128)
	_, err := c.c.Read(b)

	if err != nil {
		return err
	}

	c.req = b[16:]

	method_buf := bytes.NewBuffer(b[8:16])
	var method uint64
	binary.Read(method_buf, binary.BigEndian, &method)

	r.ServiceMethod = c.methods[method]

	c.mutex.Lock()
	c.seq++

	id := b[:8]

	c.pending[c.seq] = &id
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {

	if x == nil {
		return nil
	}

	p := x.(codec_type)
	p.Unmarshal(c.req)

	return nil
}

var null = []byte("null")

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		// Invalid request so no id.  Use JSON null.
		b = &null
	}

	// 写入id
	b_buf := bytes.NewBuffer(*b)

	if x == nil {
		// 将数据写入tcp数据流
		_, err := c.c.Write(b_buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	}

	// 写入主要数据
	p := x.(codec_type)
	b2, err := p.Marshal()
	if err != nil {
		return err
	}
	b_buf.Write(b2)

	// 将数据写入tcp数据流
	_, err = c.c.Write(b_buf.Bytes())

	return err
}

func (c *serverCodec) Close() error {
	return c.c.Close()
}

// ServeConn runs the JSON-RPC server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
func ServeConn(conn io.ReadWriteCloser, methods *map[uint64]string) {
	rpc.ServeCodec(NewServerCodec(conn, methods))
}
