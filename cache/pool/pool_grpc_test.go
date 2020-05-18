package pool

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/rpc/dispatcher/bproto"
	"github.com/pojol/braid/rpc/register"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGRPCPool(t *testing.T) {

	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSysLog(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	s := register.New("test", register.WithListen(":1205"))
	s.Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		fmt.Println("pong")
		return nil, nil
	})
	s.Run()

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
	s.Close()
}

func TestUnhealth(t *testing.T) {
	s := register.New("test", register.WithListen(":1206"))
	s.Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		fmt.Println("pong")
		return nil, nil
	})
	s.Run()
	defer s.Close()

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1206", grpc.WithInsecure())
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
	s := register.New("test", register.WithListen(":1306"))
	s.Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		fmt.Println("pong")
		return nil, nil
	})
	s.Run()
	defer s.Close()

	f := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial("localhost:1306", grpc.WithInsecure())
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
		conn, err := grpc.Dial("localhost:1207", grpc.WithInsecure())
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
