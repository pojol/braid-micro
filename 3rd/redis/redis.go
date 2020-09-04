package redis

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

type (

	// Client redis client
	Client struct {
		pool    *redis.Pool
		Address string
	}

	// Config 配置项
	Config struct {
		Address string //connection string, like "redis:// :password@10.0.1.11:6379/0"

		ReadTimeOut    time.Duration // 连接的读取超时时间
		WriteTimeOut   time.Duration // 连接的写入超时时间
		ConnectTimeOut time.Duration // 连接超时时间
		MaxIdle        int           // 最大空闲连接数
		MaxActive      int           // 最大连接数，当为0时没有连接数限制
		IdleTimeout    time.Duration // 闲置连接的超时时间, 设置小于服务器的超时时间 redis.conf : timeout
	}
)

var (
	defaultConnPoolConfig = Config{
		Address:        "redis://127.0.0.1:6379/0",
		ReadTimeOut:    5 * time.Second,
		WriteTimeOut:   5 * time.Second,
		ConnectTimeOut: 2 * time.Second,
		MaxIdle:        16,
		MaxActive:      128,
		IdleTimeout:    0,
	}

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	client *Client
)

// New 构建 Redis client
func New() *Client {
	client = &Client{}
	return client
}

// Get get redis client
func Get() *Client {
	return client
}

// Init 创建redis连接池基于传入配置
func (rc *Client) Init(cfg interface{}) error {
	redisCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}
	rc.Address = redisCfg.Address
	rc.pool = newPool(redisCfg)

	_, err := client.Ping()
	return err
}

// InitWithDefault 基于默认配置进行初始化
func (rc *Client) InitWithDefault() error {
	return rc.Init(defaultConnPoolConfig)
}

// Close 释放redis
func (rc *Client) Close() {
	rc.pool.Close()
}

func newPool(cfg Config) *redis.Pool {

	return &redis.Pool{
		MaxIdle:   cfg.MaxIdle,
		MaxActive: cfg.MaxActive, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(
				cfg.Address,
				redis.DialReadTimeout(cfg.ReadTimeOut),
				redis.DialWriteTimeout(cfg.WriteTimeOut),
				redis.DialConnectTimeout(cfg.ConnectTimeOut),
			)
			return c, err
		},
	}
}

// Run nil
func (rc *Client) Run() {

}

// Conn 获取一个redis连接，使用这个接口的时候一定要记得手动`回收`连接
func (rc *Client) Conn() redis.Conn {
	return rc.pool.Get()
}

// ActiveConnCount 获取 redis 当前连接数量
func (rc *Client) ActiveConnCount() int {
	return rc.pool.ActiveCount()
}

// ---------------------- key -------------------------------

// Keys O(N) 查找所有符合给定模式 pattern 的 key 。注：生产环境注意使用，如果有大量的key会阻塞redis操作。
func (rc *Client) Keys(pattern string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("KEYS", pattern))
	return val, err
}

// Del 删除给定的一个或多个 key
func (rc *Client) Del(key string) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("DEL", key)
	if errDo == nil && reply == nil {
		return 0, nil
	}
	val, err := redis.Int64(reply, errDo)
	return val, err
}

// Expire 设置指定key的过期时间
func (rc *Client) Expire(key string, timeOutSeconds int64) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int64(conn.Do("EXPIRE", key, timeOutSeconds))
	return val, err
}

// ---------------------- string -----------------------------

// SetWithExpire Set带ex
func (rc *Client) SetWithExpire(key string, val interface{}, timeOutSeconds int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("SET", key, val, "EX", timeOutSeconds))
	return val, err
}

// Set 将字符串值 value 关联到 key 。
func (rc *Client) Set(key string, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, val)
	return err
}

// SetEx 将值 value 关联到 key ，并将 key 的生存时间设为 seconds (以秒为单位)。
func (rc *Client) SetEx(key string, timeOutSeconds int64, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := conn.Do("SETEX", key, timeOutSeconds, val)
	return err
}

// Get 返回 key 所关联的字符串值。
func (rc *Client) Get(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, errDo := conn.Do("GET", key)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

// ------- set ----------

// ConnSMembers 返回集合 key 中的所有成员。
func ConnSMembers(conn redis.Conn, key string) ([]string, error) {
	val, err := redis.Strings(conn.Do("SMEMBERS", key))
	return val, err
}

// ---------------------- hash -------------------------------

// HGet 返回哈希表 key 中给定域 field 的值
func (rc *Client) HGet(hashID string, field string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	return ConnHGet(conn, hashID, field)
}

// ConnHGet hget
func ConnHGet(conn redis.Conn, hashID string, field string) (string, error) {
	reply, errDo := conn.Do("HGET", hashID, field)
	if errDo == nil && reply == nil {
		return "", nil
	}
	val, err := redis.String(reply, errDo)
	return val, err
}

// HGetAll 获取指定hash的所有内容
func (rc *Client) HGetAll(hashID string) (map[string]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	reply, err := redis.StringMap(conn.Do("HGetAll", hashID))
	return reply, err
}

// HKeys 返回哈希表 key 中的所有域
func (rc *Client) HKeys(hashID string) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Strings(conn.Do("HKEYS", hashID))
	return val, err
}

// HSet 将哈希表 key 中的域 field 的值设为 value
func (rc *Client) HSet(hashID string, field string, val string) error {
	conn := rc.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", hashID, field, val)
	return err
}

// HExist 返回hash里面field是否存在
func (rc *Client) HExist(hashID string, field string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("HEXISTS", hashID, field))
	return val, err
}

// HLen 返回哈希表 key 中域的数量。
func (rc *Client) HLen(field string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.Int(conn.Do("HLEN", field))
	return val, err
}

// HDel 设置指定hashset的内容, 如果field不存在, 该操作无效, 返回0
func (rc *Client) HDel(args ...interface{}) (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	return ConnHDel(conn, args...)
}

// ConnHDel HDel
func ConnHDel(conn redis.Conn, args ...interface{}) (int64, error) {
	val, err := redis.Int64(conn.Do("HDEL", args...))
	return val, err
}

//----------------- list -----------------------

// LPush 将一个或多个值 value 插入到列表 key 的表头
func (rc *Client) LPush(key string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	ret, err := redis.Int(conn.Do("LPUSH", key, value))
	if err != nil {
		return -1, err
	}

	return ret, nil
}

// RPop 移除并返回列表 key 的尾元素。
func (rc *Client) RPop(key string) (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.String(conn.Do("RPOP", key))
	return resp, err
}

// LRange 返回列表 key 中指定区间内的元素，区间以偏移量 start 和 stop 指定。
func (rc *Client) LRange(key string, start int, stop int) ([]string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Strings(conn.Do("LRANGE", key, start, stop))
	return resp, err
}

// LRem 根据参数 count 的值，移除列表中与参数 value 相等的元素。
func (rc *Client) LRem(key string, count int, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(conn.Do("LREM", key, count, value))
	return resp, err
}

// LLen 返回列表 key 的长度。
func (rc *Client) LLen(key string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	resp, err := redis.Int(conn.Do("LLen", key))
	return resp, err
}

// RPush 将一个或多个值 value 插入到列表 key 的表尾(最右边)。
func (rc *Client) RPush(key string, value string) (int, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	ret, err := redis.Int(conn.Do("RPush", key, value))
	return ret, err
}

//----------------- server ----------------------

// Ping 测试一个连接是否可用
func (rc *Client) Ping() (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("PING"))
	return val, err
}

// DBSize O(1) 返回当前数据库的 key 的数量
func (rc *Client) DBSize() (int64, error) {
	conn := rc.pool.Get()
	defer conn.Close()

	val, err := redis.Int64(conn.Do("DBSIZE"))
	return val, err
}

// FlushDB 删除当前数据库里面的所有数据
// 这个命令永远不会出现失败
func (rc *Client) flushDB() {
	conn := rc.pool.Get()
	defer conn.Close()
	conn.Do("FLUSHALL")
}
