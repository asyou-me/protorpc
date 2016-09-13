package protorpc

import (
	"errors"
	"sync"
	"time"
)

// Pool rpc 连接池
type Pool struct {
	Dial          func() (*Client, error)
	TestFunc      func(c *Client, t time.Time) error
	TestOnBorrow  bool
	TestOnReturn  bool
	TestWhileIdle bool
	Max           int
	// 正在使用
	Connected int
	// 空闲连接
	list  *List
	mutex *sync.RWMutex
}

// NewPool 新建一个连接池
func NewPool(
	Dial func() (*Client, error),
	TestFunc func(c *Client, t time.Time) error,
	Max int,
) *Pool {
	pool := &Pool{
		Dial:     Dial,
		TestFunc: TestFunc,
		Max:      Max,
	}
	pool.Init()
	return pool
}

// Init 初始化一个连接池
func (pool *Pool) Init() {
	pool.list = NewList()
	pool.mutex = new(sync.RWMutex)
	if pool.Max == 0 {
		pool.Max = 10
	}
}

// Call 调用远程的方法
func (pool *Pool) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	element, err := pool.get()
	if err != nil {
		return err
	}
	// 使用完成之后将连接送回连接池
	defer func() {
		if err == nil {
			pool.put(element)
		} else {
			pool.test(element)
		}
	}()
	err = element.Value.Call(serviceMethod, args, reply)
	return
}

// 从连接池获取一个连接
func (pool *Pool) get() (*Element, error) {
	pool.mutex.Lock()
	item := pool.list.Front()
	connected := pool.Connected
	if item == nil {
		if connected >= pool.Max*2 {
			pool.mutex.Unlock()
			return nil, errors.New("连接数到达上限")
		}
		cli, err := pool.Dial()
		if err != nil {
			pool.mutex.Unlock()
			return nil, err
		}
		pool.Connected = pool.Connected + 1
		pool.mutex.Unlock()
		return &Element{
			Value: cli,
		}, nil
	}
	pool.mutex.Unlock()
	return item, nil
}

// 将一个连接放到连接池
func (pool *Pool) put(c *Element) {
	pool.mutex.Lock()
	connected := pool.Connected
	pool.mutex.Unlock()
	if connected >= pool.Max {
		pool.mutex.Lock()
		pool.Connected = pool.Connected - 1
		pool.mutex.Unlock()
		c.Value.Close()
	} else {
		pool.mutex.Lock()
		pool.list.PushBack(c.Value)
		pool.mutex.Unlock()
	}
}

// 当一个连接错误时会用用此方法测试这个连接
func (pool *Pool) test(c *Element) {
	pool.mutex.Lock()
	c.Value.Close()
	pool.Connected = pool.Connected - 1
	pool.list.Remove(c)
	pool.mutex.Unlock()
}
