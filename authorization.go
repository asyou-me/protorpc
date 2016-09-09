package protorpc

// AuthorizationFunc 权限检查方法定义
type AuthorizationFunc func(p *AuthorizationHeader) error

// AuthorizationHeader 权限检查头文件
type AuthorizationHeader struct {
	Token  string
	Method string
}

// DefaultAuthorizationFunc 默认权限检查方法
func DefaultAuthorizationFunc(p *AuthorizationHeader) error {
	return nil
}
