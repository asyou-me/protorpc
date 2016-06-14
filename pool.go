package protorpc

import (
	"errors"
	"net/rpc"
	"sync"
	"time"
)

type Pool struct {
	Dial          func() (*rpc.Client, error)
	TestFunc      func(c *rpc.Client, t time.Time) error
	TestOnBorrow  bool
	TestOnReturn  bool
	TestWhileIdle bool
	Max           int
	Connected     int
	using         int
	list          *List
	mutex         *sync.RWMutex
}

func NewPool(Dial func() (*rpc.Client, error),
	TestFunc func(c *rpc.Client, t time.Time) error, Max int) *Pool {
	pool := &Pool{
		Dial:     Dial,
		TestFunc: TestFunc,
		Max:      Max,
	}
	pool.Init()
	return pool
}

func (this *Pool) Init() {
	this.list = NewList()
	this.mutex = new(sync.RWMutex)
	if this.Max == 0 {
		this.Max = 10
	}
}

func (this *Pool) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	element, err := this.get()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			this.put(element)
		} else {
			this.test(element)
		}
	}()
	err = element.Value.Call(serviceMethod, args, reply)
	return
}

func (this *Pool) get() (*Element, error) {
	this.mutex.Lock()
	item := this.list.Front()
	connected := this.Connected
	if item == nil {
		if connected >= this.Max*2 {
			this.mutex.Unlock()
			return nil, errors.New("连接数到达上限")
		}
		cli, err := this.Dial()
		if err != nil {
			this.mutex.Unlock()
			return nil, err
		}
		this.Connected = this.Connected + 1
		this.mutex.Unlock()
		return &Element{
			Value: cli,
		}, nil
	}
	this.mutex.Unlock()
	return item, nil
}

func (this *Pool) put(c *Element) {
	this.mutex.Lock()
	connected := this.Connected
	this.mutex.Unlock()
	if connected >= this.Max {
		this.mutex.Lock()
		this.Connected = this.Connected - 1
		this.mutex.Unlock()
		c.Value.Close()
	} else {
		this.mutex.Lock()
		this.list.PushBack(c.Value)
		this.mutex.Unlock()
	}
}

func (this *Pool) test(c *Element) {
	this.mutex.Lock()
	c.Value.Close()
	this.list.Remove(c)
	this.mutex.Unlock()
}
