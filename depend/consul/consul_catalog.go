package consul

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	consul "github.com/hashicorp/consul/api"
)

func (c *ConsulClient) ListServices() ([]*Service, error) {

	rsp, _, err := c.Client().Catalog().Services(c.parm.queryOpt)
	if err != nil {
		return nil, err
	}

	var services []*Service

	for name, val := range rsp {
		services = append(services, &Service{Name: name, Tags: val})
	}

	return services, nil
}

func (c *ConsulClient) GetService(name string) (*Service, error) {
	var rsp []*consul.ServiceEntry
	service := &Service{}
	var err error

	// if we're connect enabled only get connect services
	if c.parm.connect {
		rsp, _, err = c.Client().Health().Connect(name, "", false, c.parm.queryOpt)
	} else {
		rsp, _, err = c.Client().Health().Service(name, "", false, c.parm.queryOpt)
	}
	if err != nil {
		return nil, err
	}

	for _, s := range rsp {

		service.Name = s.Service.ID

		var del bool

		for _, check := range s.Checks {
			// delete the node if the status is critical
			if check.Status == "critical" {
				del = true
				break
			}
		}

		// if delete then skip the node
		if del {
			continue
		}

		service.Nodes = append(service.Nodes, &ServiceNode{
			ID:       s.Service.ID,
			Address:  s.Service.Address,
			Metadata: decodeMetadata(s.Service.Tags),
		})

	}

	return service, nil
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
