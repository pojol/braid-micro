package bconsul

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	consul "github.com/hashicorp/consul/api"
	"github.com/pojol/braid-go/module/meta"
)

func (c *Client) CatalogListServices() ([]meta.Service, error) {

	rsp, _, err := c.Client().Catalog().Services(c.parm.queryOpt)
	if err != nil {
		return nil, err
	}

	var services []meta.Service

	for name, val := range rsp {
		services = append(services, meta.Service{Info: meta.ServiceInfo{Name: name}, Tags: val})
	}

	return services, nil
}

func (c *Client) CatalogGetService(name string) (meta.Service, error) {
	var rsp []*consul.ServiceEntry
	service := meta.Service{}
	var err error

	// if we're connect enabled only get connect services
	if c.parm.connect {
		rsp, _, err = c.Client().Health().Connect(name, "", false, c.parm.queryOpt)
	} else {
		rsp, _, err = c.Client().Health().Service(name, "", false, c.parm.queryOpt)
	}
	if err != nil {
		return service, err
	}

	for _, s := range rsp {

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
			// 这里是否需要删掉节点？ 还是留给其他服务处理
			continue
		}

		service.Nodes = append(service.Nodes, meta.Node{
			ID:      s.Service.ID,
			Address: s.Service.Address,
			Port:    s.Service.Port,
		})

	}

	return service, nil
}

// ServicesList 获取服务列表
func CatalogServicesList(address string) (map[string][]string, error) {

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
