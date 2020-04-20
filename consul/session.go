package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pojol/braid/log"
)

type (
	// CreateSessionReq new session dat
	CreateSessionReq struct {
		Name string `json:"Name"`
		TTL  string `json:"TTL"`
	}

	// CreateSessionRes res
	CreateSessionRes struct {
		ID string `json:"ID"`
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

// CreateSession 创建session
func CreateSession(address string, name string) (ID string, err error) {
	var resBuf *bytes.Buffer
	var req *http.Request
	var res *http.Response
	client := &http.Client{}
	var sidDat []byte
	var cRes CreateSessionRes

	byt, _ := json.Marshal(&CreateSessionReq{Name: name, TTL: SessionTTL})
	resBuf = bytes.NewBuffer(byt)

	req, err = http.NewRequest("PUT", address+"/v1/session/create", resBuf)
	if err != nil {
		goto EXT
	}
	req.Header.Set("Content-type", "application/json")

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		sidDat, err = ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}
		json.Unmarshal(sidDat, &cRes)
	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("consul", "createSession", err.Error())
	}
	return cRes.ID, err
}

// DeleteSession 删除session
func DeleteSession(address string, id string) (succ bool, err error) {
	var req *http.Request
	var res *http.Response
	client := &http.Client{}
	var sidDat []byte

	req, err = http.NewRequest("PUT", address+"/v1/session/destroy/"+id, nil)
	if err != nil {
		goto EXT
	}
	req.Header.Set("Content-type", "application/json")

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		sidDat, err = ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}
		if string(sidDat) == "true" {
			succ = true
		}
	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("consul", "createSession", err.Error())
	}
	return succ, err
}

// RefushSession 刷新session
func RefushSession(address string, id string) (err error) {

	var req *http.Request
	var res *http.Response
	client := &http.Client{}
	var sidDat []byte
	var cRes []RenewSessionRes

	req, err = http.NewRequest("PUT", address+"/v1/session/renew/"+id, nil)
	if err != nil {
		goto EXT
	}
	req.Header.Set("Content-type", "application/json")

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		sidDat, err = ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}
		json.Unmarshal(sidDat, &cRes)
	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("consul", "createSession", err.Error())
	}
	return err
}

// AcquireLock 获取分布式锁
func AcquireLock(address string, name string, id string) (succ bool, err error) {
	var resBuf *bytes.Buffer
	var req *http.Request
	var res *http.Response
	client := &http.Client{}
	var sidDat []byte
	succ = false

	byt, _ := json.Marshal(&AcquireReq{Acquire: id})
	resBuf = bytes.NewBuffer(byt)

	req, err = http.NewRequest("PUT", address+"/v1/kv/"+name+"_lead?acquire="+id, resBuf)
	if err != nil {
		goto EXT
	}
	req.Header.Set("Content-type", "application/json")

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		sidDat, err = ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}

		if string(sidDat) == "true" {
			succ = true
		}
	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("consul", "acquireLock", err.Error())
	}
	return succ, err
}

// ReleaseLock 释放锁 (需要考虑异常退出情况
func ReleaseLock(address string, name string, id string) (err error) {
	var resBuf *bytes.Buffer
	var req *http.Request
	var res *http.Response
	client := &http.Client{}

	byt, _ := json.Marshal(&AcquireReq{Acquire: id})
	resBuf = bytes.NewBuffer(byt)

	req, err = http.NewRequest("PUT", address+"/v1/kv/"+name+"_lead?release="+id, resBuf)
	if err != nil {
		goto EXT
	}
	req.Header.Set("Content-type", "application/json")

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}
	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("consul", "releaseLock", err.Error())
	}
	return err
}
