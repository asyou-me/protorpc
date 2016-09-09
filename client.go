package protorpc

import (
	"io"
	"net"
	"net/rpc"
	"time"
)

// Client 自定义 rpc 连接
type Client struct {
	*rpc.Client
}

// Dial 创建一个 rpc 连接
func Dial(network, address string, auth ...string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, auth...), err
}

// DialTimeout 创建一个设定超时时间的 rpc 连接
func DialTimeout(network, address string, timeout time.Duration, auth ...string) (*Client, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, auth...), err
}

// NewClient 创建新的客户端
func NewClient(conn io.ReadWriteCloser, auth ...string) *Client {
	client := &Client{}
	client.Client = rpc.NewClientWithCodec(NewClientCodec(conn, auth...))
	return client
}
