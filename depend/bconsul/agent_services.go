package bconsul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ConsulRegistReq regist req dat
type ConsulRegistReq struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Name"`
	Tags    []string `json:"Tags"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`
}

// Regist regist service 2 consul
func ServiceRegist(address string, req ConsulRegistReq) error {
	var err error
	byt, _ := json.Marshal(&req)
	reqBuf := bytes.NewBuffer(byt)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	api := address + "/v1/agent/service/register"

	httpReq, err := http.NewRequest("PUT", api, reqBuf)
	if err != nil {
		return fmt.Errorf("failed to new request api:%v err:%v", api, err.Error())
	}
	httpReq.Header.Set("Content-type", "application/json")

	httpRes, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to request to server api:%v err:%v", api, err.Error())
	}
	defer httpRes.Body.Close()

	_, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body api:%v err:%v", api, err.Error())
	}

	return nil
}

// curl -X PUT localhost:8500/v1/agent/service/deregister/:service_id

// Deregist deregist service 2 consul
func ServiceDeregist(address string, id string) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	api := address + "/v1/agent/service/deregister/" + id
	httpReq, err := http.NewRequest("PUT", api, nil)
	if err != nil {
		return fmt.Errorf("failed to new request api:%v err:%v", api, err.Error())
	}
	httpReq.Header.Set("Content-type", "application/json")

	httpRes, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to request to server api:%v err:%v", api, err.Error())
	}
	defer httpRes.Body.Close()

	_, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body api:%v err:%v", api, err.Error())
	}

	return nil
}
