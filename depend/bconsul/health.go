package bconsul

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ServiceHealthCheck 节点健康检查 check
type ServiceHealthCheck struct {
	ID          string `json:"ID"`
	Node        string `json:"Node"`
	Name        string `josn:"Name"`
	Status      string `json:"Status"` // passing
	ServiceID   string `json:"ServiceID"`
	ServiceName string `json:"ServiceName"`
}

// ServiceHealthService 节点健康检查 service
type ServiceHealthService struct {
	ID      string `json:"ID"`
	Service string `json:"Service"`
	Address string `json:"Address"`
}

// ServiceHealthRes 节点健康检查返回
type ServiceHealthRes struct {
	Checks  []ServiceHealthCheck
	Service ServiceHealthService
}

// GetHealthNode 获取service中的健康节点
func GetHealthNode(address string, service string) (nodes []string) {

	healthRes := []ServiceHealthRes{}

	path := address + "/v1/health/service/" + service
	res, err := http.Get(path)
	if err != nil {
		goto EXT
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			goto EXT
		}
		err = json.Unmarshal(body, &healthRes)
		if err != nil {
			goto EXT
		}

		for _, v := range healthRes {
			for _, cv := range v.Checks {
				if cv.Status == "passing" {
					nodes = append(nodes, v.Service.ID)
				}
			}
		}

	} else {
		err = NewHTTPError(res.StatusCode)
		goto EXT
	}

EXT:
	if err != nil {
	}

	return nodes
}
