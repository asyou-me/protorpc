// Copyright 2013 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protorpc

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"

	wire "github.com/asyou-me/protorpc/protobuf"
	"github.com/golang/protobuf/proto"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	// temporary work space
	respHeader wire.ResponseHeader

	// Protobuf-RPC responses include the request id but not the request method.
	// Package rpc expects both.
	// We save the request method in pending when sending a request
	// and then look it up by request ID when filling out the rpc Response.
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // map request id to method name
}

// NewClientCodec returns a new rpc.ClientCodec using Protobuf-RPC on conn.
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		r:       conn,
		w:       conn,
		c:       conn,
		pending: make(map[uint64]string),
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return fmt.Errorf(
				"protorpc.ClientCodec.WriteRequest: %T does not implement proto.Message",
				param,
			)
		}
	}
	err := writeRequest(c.w, r.Seq, r.ServiceMethod, request)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	header := wire.ResponseHeader{}
	err := readResponseHeader(c.r, &header)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	r.Seq = header.Id
	r.Error = header.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	c.respHeader = header
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return fmt.Errorf(
				"protorpc.ClientCodec.ReadResponseBody: %T does not implement proto.Message",
				x,
			)
		}
	}

	err := readResponseBody(c.r, &c.respHeader, response)
	if err != nil {
		return nil
	}

	c.respHeader = wire.ResponseHeader{}
	return nil
}

// Close closes the underlying connection.
func (c *clientCodec) Close() error {
	return c.c.Close()
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

// Dial connects to a Protobuf-RPC server at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}

// DialTimeout connects to a Protobuf-RPC server at the specified network address.
func DialTimeout(network, address string, timeout time.Duration) (*rpc.Client, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}
