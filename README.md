## braid
A lightweight microservice framework that provides commonly used components. Using braid will enable us to focus on implementation without having to care about master-slave, adding and deleting services, scheduling, load balancing and other microservice logic. In addition, users can also implement their own plug-ins and customize the internal logic of the module.

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">

> `v1.1.x` Preview version

#### Feature
* RPC
* Linker
* Discover
* Balancer
* Elector
* Tracing

#### rpc-client sample
```go
b := New("test")
b.RegistPlugin(DiscoverByConsul(mock.ConsulAddr),
  BalancerBySwrr(),
  GRPCClient(grpcclient.WithPoolCapacity(128)))

b.Run()
defer b.Close()

Client().Invoke(context.TODO(), "target", "method", "", args, reply)
```


#### Quick start

> get braid

```bash
go get github.com/pojol/braid@latest
```

> hello,braid

* Architecture based on Docker
  https://github.com/pojol/braid/wiki/quick-start-with-docker

* Architecture based on k8s
  https://github.com/pojol/braid/wiki/quick-start-with-k8s



#### Wiki
https://github.com/pojol/braid/wiki