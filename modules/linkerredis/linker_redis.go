// 实现文件 linkerredis 基于 redis 实现的链路缓存
package linkerredis

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/pojol/braid-go/module/discover"
)

func (rl *redisLinker) getConn() redis.Conn {
	return rl.client.pool.Get()
}

func (rl *redisLinker) findToken(conn redis.Conn, token string, serviceName string) (target *linkInfo, err error) {

	info := linkInfo{}

	byt, err := redis.Bytes(conn.Do("HGET", RoutePrefix+splitFlag+rl.serviceName+splitFlag+serviceName,
		token),
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byt, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (rl *redisLinker) redisTarget(token string, serviceName string) (target string, err error) {

	target, err = rl.local.target(token, serviceName)
	if err == nil {
		return target, err
	}

	conn := rl.getConn()
	defer conn.Close()

	info, err := rl.findToken(conn, token, serviceName)
	if err != nil {
		return "", err
	}

	rl.local.link(token, discover.Node{
		Address: info.TargetAddr,
		ID:      info.TargetID,
		Name:    info.TargetName,
	})

	return info.TargetAddr, err
}

func (rl *redisLinker) redisLink(token string, target discover.Node) error {

	conn := rl.getConn()
	defer conn.Close()

	var cnt int
	var err error

	info := linkInfo{
		TargetAddr: target.Address,
		TargetID:   target.ID,
		TargetName: target.Name,
	}

	byt, _ := json.Marshal(&info)

	cnt, err = redis.Int(conn.Do("HSET", RoutePrefix+splitFlag+rl.serviceName+splitFlag+target.Name,
		token,
		byt,
	))

	if err == nil && cnt != 0 {

		relationKey := rl.getLinkNumKey(target.Name, target.ID)

		rl.local.link(token, target)
		if !rl.local.isRelationMember(relationKey) {
			rl.local.addRelation(relationKey)

			conn.Do("SADD", RelationPrefix, relationKey)
		}

		conn.Do("INCR", relationKey)
	}

	//l.logger.Debugf("linked parent %s, target %s, token %s", l.serviceName, cia, token)
	return err
}

func (rl *redisLinker) redisUnlink(token string, target string) error {

	conn := rl.getConn()
	defer conn.Close()

	var cnt int
	var err error
	var info *linkInfo

	info, err = rl.findToken(conn, token, target)
	if err != nil {
		return nil
	}

	cnt, err = redis.Int(conn.Do("HDEL", RoutePrefix+splitFlag+rl.serviceName+splitFlag+target,
		token))

	if err == nil && cnt == 1 {
		rl.local.unlink(token, target)

		conn.Do("DECR", rl.getLinkNumKey(info.TargetName, info.TargetID))
	}

	return nil
}

// todo
func (rl *redisLinker) redisDown(target discover.Node) error {

	conn := rl.getConn()
	defer conn.Close()

	var info *linkInfo
	var cnt int64

	routekey := RoutePrefix + splitFlag + rl.serviceName + splitFlag + target.Name

	tokenMap, err := redis.StringMap(conn.Do("HGETALL", routekey))
	if err != nil {
		return err
	}

	for key := range tokenMap {

		info, err = rl.findToken(conn, key, target.Name)
		if err != nil {
			rl.logger.Debugf("redis down find token err %v", err.Error())
			continue
		}

		if info.TargetID == target.ID {
			rmcnt, _ := redis.Int64(conn.Do("HDEL", routekey,
				key))

			cnt += rmcnt
		}

	}

	rl.logger.Debugf("redis down route del cnt:%v, total:%v, key:%v", cnt, len(tokenMap), routekey)

	rl.local.down(target)

	relationKey := rl.getLinkNumKey(target.Name, target.ID)
	rl.local.rmvRelation(relationKey)

	conn.Do("SREM", RelationPrefix, relationKey)
	conn.Do("DEL", relationKey)

	return nil
}
