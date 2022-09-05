package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type (
	// CreateSessionReq new session dat
	CreateSessionReq struct {
		Name string `json:"Name"`
		TTL  string `json:"TTL"`
	}

	// RenewSessionRes renew
	RenewSessionRes struct {
		ID          string `json:"ID"`
		Name        string `json:"Name"`
		TTL         string `json:"TTL"`
		CreateIndex string `json:"CreateIndex"`
		ModifyIndex string `json:"ModifyIndex"`
	}

	// AcquireReq 获取锁请求 kv update
	AcquireReq struct {
		Acquire string `json:"acquire"`
	}
)

const (
	// SessionTTL session 的超时时间，主要用于在进程非正常退出后锁能够被动释放，
	// 同时在使用TTL后，进程需要间断renew session保活（推荐时间是ttl/2
	SessionTTL = "10s"
)

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int) error {
	return fmt.Errorf("%v status is %v", "http response err", code)
}

func (c *Client) CreateSession(name string) (string, error) {

	id, _, err := c.Client().Session().Create(
		&api.SessionEntry{
			Name: name,
			TTL:  SessionTTL},
		&api.WriteOptions{},
	)

	return id, err
}

func (c *Client) DeleteSession(id string) error {
	_, err := c.Client().Session().Destroy(id, &api.WriteOptions{})
	return err
}

func (c *Client) RefreshSession(id string) error {
	_, _, err := c.Client().Session().Renew(id, &api.WriteOptions{})
	return err
}
