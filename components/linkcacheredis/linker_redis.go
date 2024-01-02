// 实现文件 linkerredis 基于 redis 实现的链路缓存
package linkcacheredis

import (
	"context"
	"encoding/json"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
)

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, ps module.IPubsub, cli *redis.Client, opts ...Option) module.ILinkCache {

	p := Parm{
		Mode:             LinkerRedisModeRedis,
		SyncTick:         1000 * 10, // 10 second
		SyncOfflineTick:  60,
		SyncRelationTick: 5,
	}

	for _, opt := range opts {
		opt(&p)
	}

	rl := &redisLinker{
		info:          info,
		parm:          p,
		client:        cli,
		ps:            ps,
		log:           log,
		electorState:  meta.EWait,
		activeNodeMap: make(map[string]meta.Node),
		local: &localLinker{
			serviceName: info.Name,
			tokenMap:    make(map[string]linkInfo),
			relationSet: make(map[string]int),
		},
	}

	return rl
}

func (rl *redisLinker) findToken(token string, serviceName string) (target *linkInfo, err error) {

	info := linkInfo{}

	byt, err := rl.client.HGet(context.TODO(), RoutePrefix+splitFlag+rl.info.Name+splitFlag+serviceName, token).Bytes()
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

	info, err := rl.findToken(token, serviceName)
	if err != nil {
		return "", err
	}

	return info.TargetAddr, err
}

func (rl *redisLinker) redisLink(token string, target meta.Node) error {

	var cnt uint64
	var err error

	info := linkInfo{
		TargetAddr: target.Address,
		TargetID:   target.ID,
		TargetName: target.Name,
	}

	byt, _ := json.Marshal(&info)

	cnt, err = rl.client.HSet(context.TODO(), RoutePrefix+splitFlag+rl.info.Name+splitFlag+target.Name, token, byt).Uint64()
	if err == nil && cnt != 0 {
		relationKey := rl.getLinkNumKey(target.Name, target.ID)
		rl.client.SAdd(context.TODO(), RelationPrefix, relationKey)
		rl.client.Incr(context.TODO(), relationKey)
	}

	//l.logger.Debugf("linked parent %s, target %s, token %s", l.serviceName, cia, token)
	return err
}

func (rl *redisLinker) redisUnlink(token string, target string) error {

	var cnt uint64
	var err error
	var info *linkInfo

	info, err = rl.findToken(token, target)
	if err != nil {
		return nil
	}

	cnt, err = rl.client.HDel(context.TODO(), RoutePrefix+splitFlag+rl.info.Name+splitFlag+target, token).Uint64()
	if err == nil && cnt == 1 {
		rl.client.Decr(context.TODO(), rl.getLinkNumKey(info.TargetName, info.TargetID))
	}

	return nil
}

// todo
func (rl *redisLinker) redisDown(target meta.Node) error {

	var info *linkInfo
	var cnt uint64

	routekey := RoutePrefix + splitFlag + rl.info.Name + splitFlag + target.Name
	ctx := context.TODO()

	tokenMap, err := rl.client.HGetAll(ctx, routekey).Result()
	if err != nil {
		return err
	}

	for key := range tokenMap {

		info, err = rl.findToken(key, target.Name)
		if err != nil {
			rl.log.Debugf("redis down find token err %v", err.Error())
			continue
		}

		if info.TargetID == target.ID {
			rmcnt, _ := rl.client.HDel(ctx, routekey, key).Uint64()
			cnt += rmcnt
		}

	}

	rl.log.Debugf("redis down route del cnt:%v, total:%v, key:%v", cnt, len(tokenMap), routekey)

	relationKey := rl.getLinkNumKey(target.Name, target.ID)
	rl.client.SRem(ctx, RelationPrefix, relationKey)
	rl.client.Del(ctx, relationKey)

	return nil
}
