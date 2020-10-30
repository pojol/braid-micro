package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/plugin/grpcclient/bproto"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/zaplogger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type rpcServer struct {
	bproto.ListenServer
}

func (rs *rpcServer) Routing(ctx context.Context, req *bproto.RouteReq) (*bproto.RouteRes, error) {
	out := new(bproto.RouteRes)
	var err error

	if req.Service == "test" {
		err = nil
	} else {
		err = errors.New("err")
	}

	return out, err
}

func TestMain(m *testing.M) {

	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	sb := server.GetBuilder(grpcserver.Name)
	sb.AddOption(grpcserver.WithListen(":1205"))
	s, _ := sb.Build("test", log)
	bproto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()
	time.Sleep(time.Millisecond * 10)

	m.Run()

	s.Close()
}

func TestGRPCPool(t *testing.T) {

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1205", grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := NewGRPCPool(f, 10, 64, time.Second*120)
	assert.Equal(t, err, nil)

	conn, err := p.Get(context.Background())
	assert.Equal(t, err, nil)

	caCtx, caCancel := context.WithTimeout(context.Background(), time.Second)
	defer caCancel()

	rres := new(bproto.RouteRes)
	err = conn.Invoke(caCtx, "/bproto.listen/routing", &bproto.RouteReq{
		ReqBody: []byte(`{"Val1":1, "Val2":2}`),
		Service: "test",
		Nod:     "normal",
	}, rres)
	assert.Equal(t, err, nil)

	conn.Put()

	p.Available()
	p.Capacity()

	p.Close()
}

func TestUnhealth(t *testing.T) {

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1205", grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := NewGRPCPool(f, 10, 64, time.Second*120)
	assert.Equal(t, err, nil)

	conn, err := p.Get(context.Background())
	assert.Equal(t, err, nil)
	conn.Unhealthy()
	conn.Put()

	p.Close()
}

func TestIdle(t *testing.T) {

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1205", grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := NewGRPCPool(f, 1, 5, time.Millisecond)
	assert.Equal(t, err, nil)

	time.Sleep(time.Millisecond * 10)

	for i := 0; i < 10; i++ {
		ctx, cal := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cal()
		p.Get(ctx)
		time.Sleep(time.Millisecond)
	}

	p.Close()
}

func TestErr(t *testing.T) {
	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1205", grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	var tests = []struct {
		Init int
		Cap  int
	}{
		{0, 0},
		{1, 1},
	}

	for _, v := range tests {
		p, _ := NewGRPCPool(f, v.Init, v.Cap, time.Millisecond)
		if p != nil {
			p.Close()
			p.Close()
			p.Get(context.Background())
		}
	}

}

var syncmap sync.Map
var synctick uint64

func getpool(address string) (p *GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		conn, err = grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		atomic.AddUint64(&synctick, 1)

		return conn, nil
	}

	pi, ok := syncmap.Load(address)
	if !ok {
		p, err = NewGRPCPool(factory, 10, 20, time.Second*10)
		if err != nil {
			goto EXT
		}

		syncmap.Store(address, p)

		pi = p
	}

	p = pi.(*GRPCPool)

EXT:
	return p, err
}

func getConn(address string) (*ClientConn, error) {
	var caConn *ClientConn
	var caPool *GRPCPool

	caPool, err := getpool(address)
	if err != nil {
		return nil, err
	}

	connCtx, connCancel := context.WithTimeout(context.Background(), time.Second)
	defer connCancel()
	caConn, err = caPool.Get(connCtx)
	if err != nil {
		return nil, err
	}

	return caConn, nil
}

func TestFactory(t *testing.T) {

	//assert.Equal(t, synctick, uint64(10))

	for i := 0; i < 1000; i++ {
		c1, _ := getConn("127")
		c1.Put()
		c2, _ := getConn("128")
		c2.Put()
	}

	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, synctick, uint64(40))

}

func BenchmarkGRPCByOriginal(b *testing.B) {
	testEndpoint := ""
	testMethod := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _ := grpc.Dial(testEndpoint, grpc.WithInsecure())
		rres := new(bproto.RouteRes)
		err := conn.Invoke(context.Background(), testMethod, &bproto.RouteReq{
			ReqBody: []byte(`{"Val1":1, "Val2":2}`),
		}, rres)
		if err != nil {
			b.Error(err)
		}
		conn.Close()
	}
}

func BenchmarkGRPCByPool(b *testing.B) {
	testEndpoint := ""
	testMethod := ""

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(testEndpoint, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	p, err := NewGRPCPool(f, 8, 32, time.Second*120)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		conn, err := p.Get(context.Background())
		if err != nil {
			b.Error(err)
		}

		rres := new(bproto.RouteRes)
		err = conn.Invoke(context.Background(), testMethod, &bproto.RouteReq{
			ReqBody: []byte(`{"Val1":1, "Val1":}`),
		}, rres)
		if err != nil {
			conn.Unhealthy()
			b.Error(err)
		}

		conn.Put()
	}

	p.Close()
}
