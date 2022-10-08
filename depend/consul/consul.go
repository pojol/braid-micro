package consul

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type Service struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
	Nodes    []*ServiceNode    `json:"nodes"`
	Tags     []string          `json:"tags"`
}

// ServiceNode 服务信息
type ServiceNode struct {
	ID       string
	Address  string
	Port     int
	Metadata map[string]string `json:"metadata"`
}

type Client struct {
	client *consul.Client
	parm   Parm

	address []string

	sync.Mutex
	register map[string]uint64
	// lastChecked tracks when a node was last checked as existing in Consul
	lastChecked map[string]time.Time

	httpClient *http.Client
}

func BuildWithOption(opts ...Option) *Client {

	p := Parm{
		cfg: consul.DefaultNonPooledConfig(), //use default non pooled config
		queryOpt: &consul.QueryOptions{
			AllowStale: true,
		},
	}

	for _, opt := range opts {
		opt(&p)
	}

	cc := &Client{
		register:    make(map[string]uint64),
		lastChecked: make(map[string]time.Time),
		parm:        p,
		client:      nil,
		httpClient:  new(http.Client),
	}

	if cc.parm.timeout != 0 {
		cc.httpClient.Timeout = cc.parm.timeout
	}

	// check if there are any addrs
	var addrs []string

	for _, address := range p.Addrs {
		addr, port, err := net.SplitHostPort(address)
		if err != nil {
			panic(err.Error())
		}

		addrs = append(addrs, net.JoinHostPort(addr, port))
	}

	if len(addrs) > 0 {
		cc.address = addrs
	}

	if p.timeout != 0 {
		cc.httpClient.Timeout = p.timeout
	}

	if p.tlsConfig != nil {
		cc.httpClient.Transport = newTransport(p.tlsConfig)
	}

	// set the addrs
	//if len(addrs) > 0 {
	//	c.Address = addrs
	//	config.Address = c.Address[0]
	//}

	// init client
	cc.Client()

	return cc
}

func getDeregisterTTL(t time.Duration) time.Duration {
	// splay slightly for the watcher?
	splay := time.Second * 5
	deregTTL := t + splay

	// consul has a minimum timeout on deregistration of 1 minute.
	if t < time.Minute {
		deregTTL = time.Minute + splay
	}

	return deregTTL
}

func newTransport(config *tls.Config) *http.Transport {
	if config == nil {
		config = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     config,
	}
	runtime.SetFinalizer(&t, func(tr **http.Transport) {
		(*tr).CloseIdleConnections()
	})
	return t
}

func (c *Client) Client() *consul.Client {
	if c.client != nil {
		return c.client
	}

	for _, addr := range c.address {
		// set the address
		c.parm.cfg.Address = addr

		// create a new client
		tmpClient, err := consul.NewClient(c.parm.cfg)
		if err != nil {
			fmt.Println("consul.NewClient", err.Error())
			continue
		}

		// test the client
		_, err = tmpClient.Agent().Host()
		if err != nil {
			fmt.Println("[Consul] agent host err", err.Error())
			continue
		}

		// set the client
		c.client = tmpClient
		return c.client
	}

	var err error
	// set the default
	c.client, err = consul.NewClient(c.parm.cfg)
	if err != nil {
		fmt.Println("[Consul] new client err", err.Error())
	}

	// return the client
	return c.client
}
