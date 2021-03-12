package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
func Regist(address string, req ConsulRegistReq) (err error) {
	byt, _ := json.Marshal(&req)
	reqBuf := bytes.NewBuffer(byt)
	client := &http.Client{}

	httpReq, err := http.NewRequest("PUT", address+"/v1/agent/service/register", reqBuf)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-type", "application/json")

	httpRes, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode == http.StatusOK {
		sidDat, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return err
		}
		fmt.Println("dat", sidDat)
	} else {
		return errors.New("http status err code=" + httpRes.Status)
	}

	return nil
}

// curl -X PUT localhost:8500/v1/agent/service/deregister/:service_id

// Deregist deregist service 2 consul
func Deregist(address string, id string) error {
	client := &http.Client{}
	httpReq, err := http.NewRequest("PUT", address+"/v1/agent/service/deregister/"+id, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-type", "application/json")

	httpRes, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode == http.StatusOK {
		sidDat, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return err
		}
		fmt.Println("dat", sidDat)
	} else {
		return errors.New("http status err code=" + httpRes.Status)
	}

	return nil
}
