## Braid
**Braid** plug-ins become service grid

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">


#### Plug-ins
* RPC
  - grpc-client
  - grpc-server
* Linker
  -  linker-redis   `linker based in the nsq and redis`
* Discover
  - discover-consul
* Balancer
  - smooth-weight-round-robin
* Elector
  - elector-consul
  - elector-k8s
* Pub-sub
  - pubsub-proc
  - pubsub-nsq
* Tracing
  - tracing-jeager


#### rpc-client sample
```go
b := New("test")
b.RegistPlugin(
	Discover(
		discoverconsul.Name,
		discoverconsul.WithConsulAddress(mock.ConsulAddr)),
	Balancer(balancerswrr.Name),
	GRPCClient(grpcclient.WithPoolCapacity(128)))

b.Run()
defer b.Close()

Client().Invoke(context.TODO(), "target", "method", "", args, reply)
```


#### Quick start

> get braid `v1.1.x` Preview version

```bash
go get github.com/pojol/braid@latest
```

> hello,braid

* Architecture based on Docker
  https://github.com/pojol/braid/wiki/quick-start-with-docker

* Architecture based on k8s
  https://github.com/pojol/braid/wiki/quick-start-with-k8s

#### Web
> sankey chart

<img src="https://i.postimg.cc/sX0xHZmF/image.png" width="600">

#### Wiki
https://github.com/pojol/braid/wiki
