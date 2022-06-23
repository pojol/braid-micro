package consul

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// StatusLeaderRes 获取consul master节点信息
type StatusLeaderRes struct {
	Dc string
}

// GetConsulLeader 获取consul的master节点
func GetConsulLeader(address string) (string, error) {

	path := address + "/v1/status/leader"
	res, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("get consul status read err %v", err.Error())
		}

		leaderinfo := string(body)

		if strings.Compare(leaderinfo, "") == 0 {
			return "", errors.New("get consul status info err")
		}

		return leaderinfo, nil
	}

	return "", fmt.Errorf("get consul status http err code:%v", res.StatusCode)
}
