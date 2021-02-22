package linkerredis

import (
	"fmt"

	"github.com/pojol/braid-go/module/discover"
)

type localLinker struct {
	serviceName string

	//tokenMap map["base_mail_token"] : "127.0.0.1:8001"
	tokenMap map[string]linkInfo

	relationSet map[string]int
}

func (ll *localLinker) link(token string, target discover.Node) {
	key := ll.serviceName + splitFlag + target.Name + splitFlag + token
	if _, ok := ll.tokenMap[key]; !ok {
		ll.tokenMap[key] = linkInfo{
			TargetAddr: target.Address,
			TargetID:   target.ID,
			TargetName: target.Name,
		}
	}
}

func (ll *localLinker) target(token string, serviceName string) (string, error) {
	var targetAddr string
	var err error

	// hash ll.serviceName + "_" + serviceName + "_" + token : target addr
	key := ll.serviceName + splitFlag + serviceName + splitFlag + token

	if _, ok := ll.tokenMap[key]; ok {
		targetAddr = ll.tokenMap[key].TargetAddr
	} else {
		err = fmt.Errorf("can't find token by service %v", serviceName)
	}

	return targetAddr, err
}

func (ll *localLinker) unlink(token string, target string) linkInfo {
	key := ll.serviceName + splitFlag + target + splitFlag + token

	var info linkInfo

	if _, ok := ll.tokenMap[key]; ok {
		info = ll.tokenMap[key]
		delete(ll.tokenMap, key)
	}

	return info
}

func (ll *localLinker) down(target discover.Node) int {

	var cnt int

	for key := range ll.tokenMap {
		if ll.tokenMap[key].TargetID == target.ID {
			delete(ll.tokenMap, key)
			cnt++
		}
	}

	return cnt
}

func (ll *localLinker) addRelation(relation string) {

	if _, ok := ll.relationSet[relation]; !ok {
		ll.relationSet[relation] = 1
	}

}

func (ll *localLinker) isRelationMember(relation string) bool {

	_, ok := ll.relationSet[relation]

	return ok
}

// 这里将来或许还要添加周期检查的逻辑
func (ll *localLinker) rmvRelation(relation string) {
	if _, ok := ll.relationSet[relation]; ok {
		delete(ll.relationSet, relation)
	}
}

func (rl *redisLinker) localTarget(token string, serviceName string) (string, error) {
	return rl.local.target(token, serviceName)
}

func (rl *redisLinker) localLink(token string, target discover.Node) error {

	conn := rl.getConn()
	defer conn.Close()

	rl.local.link(token, target)

	relationKey := rl.getRelationKey(target.Name, target.ID)
	if !rl.local.isRelationMember(relationKey) {
		rl.local.addRelation(relationKey)

		conn.Do("SADD", RelationPrefix, relationKey)
	}

	conn.Do("INCR", rl.getRelationKey(target.Name, target.ID))

	return nil
}

func (rl *redisLinker) localUnlink(token string, target string) error {

	conn := rl.getConn()
	defer conn.Close()

	info := rl.local.unlink(token, target)

	if info.TargetID != "" {
		conn.Do("DECR", rl.getRelationKey(info.TargetName, info.TargetID))
	}

	return nil
}

func (rl *redisLinker) localDown(target discover.Node) error {

	conn := rl.getConn()
	defer conn.Close()

	cnt := rl.local.down(target)

	relationKey := rl.getRelationKey(target.Name, target.ID)
	rl.local.rmvRelation(relationKey)

	conn.Do("SREM", RelationPrefix, relationKey)

	if cnt != 0 {
		conn.Do("DECRBY", rl.getRelationKey(target.Name, target.ID), cnt)
	}

	return nil
}
