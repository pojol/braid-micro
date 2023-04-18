package bconsul

import (
	consul "github.com/hashicorp/consul/api"
)

// ConsulRegistReq regist req dat
type ConsulRegistReq struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Name"`
	Tags    []string `json:"Tags"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`
}

// ServiceRegister 服务注册
//  https://developer.hashicorp.com/consul/api-docs/agent/service#register-service
//
func (c *Client) ServiceRegister(parm consul.AgentServiceRegistration) error {
	// 服务注册
	//check := new(consul.AgentServiceCheck)

	return c.Client().Agent().ServiceRegister(&parm)
}

func (c *Client) ServiceDeregister(id string) error {
	// 服务注销
	return c.Client().Agent().ServiceDeregister(id)
}
