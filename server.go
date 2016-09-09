package protorpc

import (
	"io"
	"net/rpc"
)

// 默认服务
var defalutServer *Server

func init() {
	defalutServer = NewServer()
}

// Server 服务结构
type Server struct {
	auth AuthorizationFunc
	*rpc.Server
}

// NewServer 建立一个服务
func NewServer() *Server {
	server := &Server{
		Server: rpc.NewServer(),
	}
	server.Auth(DefaultAuthorizationFunc)
	return server
}

//Auth 设定所有服务的权限验证函数
func (s *Server) Auth(fn AuthorizationFunc) error {
	s.auth = fn
	return nil
}

// ServeConn 建立服务器连接
func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	codec := NewServerCodec(conn, s.auth)
	s.ServeCodec(codec)
}

// ServeConn 建立一个链接
func ServeConn(conn io.ReadWriteCloser) {
	defalutServer.ServeConn(conn)
}
