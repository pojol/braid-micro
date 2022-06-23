package consul

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// NodServiceDat 服务信息
type NodServiceDat struct {
	ID             string
	Address        string
	ServiceAddress string
	ServiceID      string
	ServiceName    string
	ServicePort    int
}

// ServicesList 获取服务列表
func ServicesList(address string) (map[string][]string, error) {

	var res *http.Response
	client := &http.Client{}
	var services map[string][]string

	req, err := http.NewRequest(http.MethodGet, address+"/v1/catalog/services", nil)
	if err != nil {
		goto EXT
	}

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}

		err = json.Unmarshal(body, &services)
		if err != nil {
			goto EXT
		}

	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	return services, err
}

// GetCatalogService 获取service系列信息
func GetCatalogService(address string, serviceName string) (servicelist []NodServiceDat, err error) {

	var res *http.Response
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, address+"/v1/catalog/service/"+serviceName, nil)
	if err != nil {
		goto EXT
	}

	res, err = client.Do(req)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}

		err = json.Unmarshal(body, &servicelist)
		if err != nil {
			goto EXT
		}

	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	return servicelist, err
}

// GetCatalogServices 获取所有的service
func GetCatalogServices(address string, serviceTag string) (map[string]NodServiceDat, error) {

	nodServices := make(map[string]NodServiceDat)

	services, err := ServicesList(address)
	if err != nil {
		goto EXT
	}

	for k, v := range services {
		for _, tag := range v {
			if tag == serviceTag {
				lst, err := GetCatalogService(address, k)
				if err != nil {
					continue
				}

				for _, v := range lst {
					nodServices[v.ServiceID] = v
				}
			}
		}
	}

EXT:
	return nodServices, err
}
