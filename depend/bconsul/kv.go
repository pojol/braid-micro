package bconsul

import (
	"github.com/hashicorp/consul/api"
)

func (c *Client) AcquireLock(name string, id string) (bool, error) {
	ok, _, err := c.Client().KV().Acquire(&api.KVPair{
		Key:     name + "_lead",
		Session: id,
	}, &api.WriteOptions{})
	return ok, err
}

func (c *Client) ReleaseLock(name string, id string) (bool, error) {
	ok, _, err := c.Client().KV().Release(&api.KVPair{
		Key:     name + "_lead",
		Session: id,
	}, &api.WriteOptions{})
	return ok, err
}
