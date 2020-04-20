package pool

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/caller/brpc"
	"google.golang.org/grpc"
)

var testEndpoint = "192.168.50.138:10101"
var testMethod = "/brpc.calculate/addition"

func TestGRPCPool(t *testing.T) {
	/*
		f := func() (*grpc.ClientConn, error) {
			conn, err := grpc.Dial(testEndpoint, grpc.WithInsecure())
			if err != nil {
				return nil, err
			}

			return conn, nil
		}

		p, err := NewGRPCPool(f, 10, 64, time.Second*120)
		if err != nil {
			t.Error(err)
		}

		conn, err := p.Get(context.Background())
		if err != nil {
			t.Error(err)
		}

		caCtx, caCancel := context.WithTimeout(context.Background(), time.Second)
		defer caCancel()

		rres := new(brpc.RouteRes)
		err = conn.Invoke(caCtx, testMethod, &brpc.RouteReq{
			ReqBody: []byte(`{"Val1":1, "Val2":2}`),
		}, rres)
		if err != nil {
			conn.Unhealthy()
			t.Error(err)
		}

		conn.Put()
		p.Close()
	*/
}

func BenchmarkGRPCByOriginal(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _ := grpc.Dial(testEndpoint, grpc.WithInsecure())
		rres := new(brpc.RouteRes)
		err := conn.Invoke(context.Background(), testMethod, &brpc.RouteReq{
			ReqBody: []byte(`{"Val1":1, "Val2":2}`),
		}, rres)
		if err != nil {
			b.Error(err)
		}
		conn.Close()
	}
}

func BenchmarkGRPCByPool(b *testing.B) {
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

		rres := new(brpc.RouteRes)
		err = conn.Invoke(context.Background(), testMethod, &brpc.RouteReq{
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
