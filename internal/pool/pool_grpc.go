package pool

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// 实现来自 https://github.com/processout/grpc-go-pool

// GRPCConnFactory 用于提供创建grpc.ClientConn
type GRPCConnFactory func() (*grpc.ClientConn, error)

var (
	// ErrPoolFull 池满
	ErrPoolFull = errors.New("grpc pool closing a ClientConn into a full pool")
	// ErrPoolTimeout 超时
	ErrPoolTimeout = errors.New("grpc pool get client timeout")
	// ErrPoolClosed 已关闭
	ErrPoolClosed = errors.New("grpc pool is closed")
	// ErrPoolCapacity 错误的容量设置
	ErrPoolCapacity = errors.New("grpc pool wrong capacity")
)

// GRPCPool grpc client pool
type GRPCPool struct {
	clients chan ClientConn
	/*
		idleTimeout
		如果客户端长期没有使用这个创建的Conn，那么这个连接会成为空闲连接，
		在一段时间之后，这个连接可能会被服务端的超时策略关闭 (serverOptions.connectionTimeout)
		这个时候再次使用这个连接，则通信会失败，
		因此最好将maxLifeDuration设置为比服务器超时时间短些的时间。
	*/
	idleTimeout time.Duration
	factory     GRPCConnFactory
	mu          sync.Mutex
}

// ClientConn grpc.ClientConn 的包装
type ClientConn struct {
	*grpc.ClientConn
	pool          *GRPCPool
	timeInitiated time.Time
	timeUsed      time.Time

	/*
		unhealthy
		当服务器发生重启，本地的连接池连接都会失效，
		这个时候需要客户端在连接失败时设置连接unhealthy，
		通过unhealthy标记，pool会在使用这个连接的时候将其回收，并重新获取一个新的连接。
	*/
	unhealthy bool
}

// NewGRPCPool 新建 grpc 连接池
func NewGRPCPool(factory GRPCConnFactory, initNum, capacity int, idleTimeout time.Duration) (*GRPCPool, error) {

	if capacity <= 0 {
		return nil, ErrPoolCapacity
	}

	if initNum > capacity {
		initNum = capacity
	}

	p := &GRPCPool{
		clients:     make(chan ClientConn, capacity),
		factory:     factory,
		idleTimeout: idleTimeout,
	}

	for i := 0; i < initNum; i++ {
		c, err := factory()
		if err != nil {
			return nil, err
		}

		p.clients <- ClientConn{
			ClientConn:    c,
			pool:          p,
			timeInitiated: time.Now(),
			timeUsed:      time.Now(),
		}
	}

	for i := 0; i < capacity-initNum; i++ {
		// fill empty client
		p.clients <- ClientConn{
			pool: p,
		}
	}

	return p, nil
}

func (p *GRPCPool) getClient() chan ClientConn {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.clients
}

// Get 从池中获取一个client conn，传入WithTimeout的context 可以在 #池满# 的时候获得超时信号。
func (p *GRPCPool) Get(ctx context.Context) (*ClientConn, error) {
	var err error
	clients := p.getClient()
	if clients == nil {
		return nil, ErrPoolClosed
	}

	wrapper := ClientConn{
		pool: p,
	}

	select {
	case wrapper = <-clients:
		// All good
	case <-ctx.Done():
		return nil, ErrPoolTimeout
	}

	idleTimeout := p.idleTimeout
	if wrapper.ClientConn != nil && idleTimeout > 0 &&
		wrapper.timeUsed.Add(idleTimeout).Before(time.Now()) {

		wrapper.ClientConn.Close()
		wrapper.ClientConn = nil
	}

	if wrapper.ClientConn == nil {
		wrapper.ClientConn, err = p.factory()
		if err != nil {
			// If there was an error, we want to put back a placeholder
			// client in the channel
			clients <- ClientConn{
				pool: p,
			}
		}
		// This is a new connection, reset its initiated time
		wrapper.timeInitiated = time.Now()

	}

	return &wrapper, err
}

// Put 放回池中
func (c *ClientConn) Put() error {

	wrapper := ClientConn{
		pool:       c.pool,
		ClientConn: c.ClientConn,
		timeUsed:   time.Now(),
	}

	if c.unhealthy {
		wrapper.ClientConn.Close()
		wrapper.ClientConn = nil
	} else {
		wrapper.timeInitiated = c.timeInitiated
	}

	select {
	case c.pool.clients <- wrapper:
		// All good
	default:
		return ErrPoolFull
	}

	c.ClientConn = nil // Mark as closed
	return nil
}

// Close 关闭并清理pool
func (p *GRPCPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	clients := p.clients
	p.clients = nil

	if clients == nil {
		return
	}

	close(clients)
	for client := range clients {
		if client.ClientConn == nil {
			continue
		}

		client.ClientConn.Close()
	}
}

// Unhealthy 将连接设置为不健康状态
func (c *ClientConn) Unhealthy() {
	c.unhealthy = true
}

// Capacity 返回池的总容量
func (p *GRPCPool) Capacity() int {

	return cap(p.clients)
}

// Available 返回可用的client数量
func (p *GRPCPool) Available() int {

	return len(p.clients)
}
