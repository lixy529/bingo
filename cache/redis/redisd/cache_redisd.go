// redisd adapter
// 用于分布式的redis操作
//   变更历史
//     2017-02-21  lixiaoya  新建
package redisd

import (
	"github.com/lixy529/bingo/cache"
	"github.com/go-redis/redis"
	json "github.com/json-iterator/go"
	"fmt"
	"strconv"
	"time"
	"strings"
	"errors"
	"math/rand"
)

const NOT_EXIST = "redis: nil"

// RedisdCache缓存
type RedisdCache struct {
	connClient []*redis.Client // 每个主机有一个连接池

	addr         string        // 连接主机和端口，多个主机用逗号分割，如127.0.0.1:1900,127.0.0.2:1900
	auth         string        // 授权密码
	dbNum        int           // db编号，默认为0
	dialTimeout  time.Duration // 连接超时时间，单位秒，默认5秒
	readTimeout  time.Duration // 读超时时间，单位秒，-1-不超时，0-使用默认3秒
	writeTimeout time.Duration // 写超时时间，单位秒，默认为readTimeout
	poolSize     int           // 每个节点连接池的连接数，默认为cpu个数的10倍
	minIdleConns int           // 最少空闲连接数，默认为0
	maxConnAge   time.Duration // 最大连接时间，单位秒，超时时间自动关闭，默认为0
	poolTimeout  time.Duration // 如果所有连接都忙时的等待时间，默认为readTimeout+1秒
	idleTimeout  time.Duration // 最大空闲时间，单位秒，默认为5分钟

	prefix    string // key前缀，如果配置里有，则所有key前自动添加此前缀
	encodeKey []byte // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewRedisdCache 新建一个RedisdCache适配器.
func NewRedisdCache() cache.Cache {
	return &RedisdCache{}
}

// Init 初始化
//   参数
//     config: 配置josn串
//       {
//         "addr":"127.0.0.1:19100,127.0.0.2:19100,127.0.0.3:19100",
//         "auth":"xxxx",
//         "dbNum":"1",
//         "dialTimeout":"5",
//         "readTimeout":"5",
//         "writeTimeout":"5",
//         "poolSize":"5",
//         "minIdleConns":"5",
//         "maxConnAge":"5",
//         "poolTimeout":"5",
//         "idleTimeout":"5",
//         "prefix":"le_",
//         "encodeKey":"abcdefghij123456",
//       }
//       addr:         连接主机和端口，多个主机用逗号分割，如127.0.0.1:1900,127.0.0.2:1900
//       auth:         授权密码
//       dbNum:        db编号，默认为0
//       dialTimeout:  连接超时时间，单位秒，默认5秒
//       readTimeout:  读超时时间，单位秒，-1-不超时，0-使用默认3秒
//       writeTimeout: 写超时时间，单位秒，默认为readTimeout
//       poolSize:     每个节点连接池的连接数，默认为cpu个数的10倍
//       minIdleConns: 最少空闲连接数，默认为0
//       maxConnAge:   最大连接时间，单位秒，超时时间自动关闭，默认为0
//       poolTimeout:  如果所有连接都忙时的等待时间，默认为readTimeout+1秒
//       idleTimeout:  最大空闲时间，单位秒，默认为5分钟
//       prefix:       key前缀，如果配置里有，则所有key前自动添加此前缀
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RedisdCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RedisdCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// 连接串
	c.addr = mapCfg["addr"]

	// 授权
	c.auth = mapCfg["auth"]

	// 连接超时时间
	dialTimeout, err := strconv.Atoi(mapCfg["dialTimeout"])
	if err != nil || dialTimeout < 0 {
		c.dialTimeout = 5
	} else {
		c.dialTimeout = time.Duration(dialTimeout)
	}

	// 读超时时间
	readTimeout, err := strconv.Atoi(mapCfg["readTimeout"])
	if err != nil {
		c.readTimeout = 3
	} else if readTimeout < 0 {
		c.readTimeout = -1
	} else {
		c.readTimeout = time.Duration(readTimeout)
	}

	// 写超时时间
	writeTimeout, err := strconv.Atoi(mapCfg["writeTimeout"])
	if err != nil {
		c.writeTimeout = c.readTimeout
	} else if writeTimeout < 0 {
		c.writeTimeout = -1
	} else {
		c.writeTimeout = time.Duration(writeTimeout)
	}

	// 每个节点连接池的连接数
	poolSize, err := strconv.Atoi(mapCfg["poolSize"])
	if err != nil || poolSize < 0 {
		c.poolSize = 0
	} else {
		c.poolSize = poolSize
	}

	// 最少空闲连接数
	minIdleConns, err := strconv.Atoi(mapCfg["minIdleConns"])
	if err != nil || minIdleConns < 0 {
		c.minIdleConns = 0
	} else {
		c.minIdleConns = minIdleConns
	}

	// 最大连接时间
	maxConnAge, err := strconv.Atoi(mapCfg["maxConnAge"])
	if err != nil || maxConnAge < 0 {
		c.maxConnAge = 0
	} else {
		c.maxConnAge = time.Duration(maxConnAge)
	}

	// 如果所有连接都忙时的等待时间
	poolTimeout, err := strconv.Atoi(mapCfg["poolTimeout"])
	if err != nil || poolTimeout < 0 {
		c.poolTimeout = c.readTimeout + 1
	} else {
		c.poolTimeout = time.Duration(poolTimeout)
	}

	// 最大空闲时间
	idleTimeout, err := strconv.Atoi(mapCfg["idleTimeout"])
	if err != nil || idleTimeout < 0 {
		c.idleTimeout = 300
	} else {
		c.idleTimeout = time.Duration(idleTimeout)
	}

	// 前缀
	if prefix, ok := mapCfg["prefix"]; ok {
		c.prefix = prefix
	}

	// 加密密钥
	if tmp, ok := mapCfg["encodeKey"]; ok && tmp != "" {
		c.encodeKey = []byte(tmp)
	}

	// dbNum
	dbNum, err := strconv.Atoi(mapCfg["dbNum"])
	if err != nil {
		c.dbNum = 0
	} else {
		c.dbNum = dbNum
	}

	// 设置连接池
	for _, v := range strings.Split(c.addr, ",") {
		rp, err := c.connect(v)
		if err != nil {
			continue
		}
		c.connClient = append(c.connClient, rp)
	}

	return nil
}

// connect 建立连接
//   参数
//     host: 其中一台主机信息
//   返回
//     这台主机的连接池、错误信息
func (c *RedisdCache) connect(host string) (*redis.Client, error) {
	// 连接客户端
	client := redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     c.auth,
		DB:           c.dbNum,
		DialTimeout:  c.dialTimeout * time.Second,
		ReadTimeout:  c.readTimeout * time.Second,
		WriteTimeout: c.writeTimeout * time.Second,
		PoolSize:     c.poolSize,
		MinIdleConns: c.minIdleConns,
		MaxConnAge:   c.maxConnAge * time.Second,
		PoolTimeout:  c.poolTimeout * time.Second,
		IdleTimeout:  c.idleTimeout * time.Second,
	})

	return client, nil
}

// getClient 获取一个连接
// 外面调用完需要通过c.Close()放回连接池
//   参数
//     host: 其中一台主机信息
//   返回
//     这台主机的连接池
func (c *RedisdCache) getClient() *redis.Client {
	connCnt := len(c.connClient)
	if connCnt == 0 {
		return nil
	} else if connCnt == 1 {
		return c.connClient[0]
	}

	k := rand.Intn(connCnt)
	return c.connClient[k]
}

// Set 向缓存设置一个值
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RedisdCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// 加密判断
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, c.encodeKey)
		if err != nil {
			return err
		}
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().Set(key, data, time.Duration(expire)*time.Second).Err()
}

// Get 从缓存取一个值
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (c *RedisdCache) Get(key string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.getClient().Get(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// 解密判断
	data, err := cache.Decode([]byte(v), c.encodeKey)
	if err != nil {
		return err, true
	}

	// 类型转换
	err = cache.ByteToInter(data, val)
	if err != nil {
		return err, true
	}

	return nil, true
}

// Del 从缓存删除一个值
//   参数
//     key: key值
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RedisdCache) Del(key string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().Del(key).Err()
}

// MSet 同时设置一个或多个key-value对
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (c *RedisdCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	var v []interface{}
	for key, val := range mList {
		// 类型转换
		data, err := cache.InterToByte(val)
		if err != nil {
			return err
		}

		// 加密判断
		encode = append(encode, false)
		if encode[0] {
			data, err = cache.Encode(data, c.encodeKey)
			if err != nil {
				return err
			}
		}

		if c.prefix != "" {
			key = c.prefix + key
		}
		v = append(v, key, data)
	}

	err := c.getClient().MSet(v...).Err()
	if err != nil {
		return err
	}

	// 设置失效时间
	if expire > 0 {
		for key := range mList {
			if c.prefix != "" {
				key = c.prefix + key
			}
			c.getClient().Expire(key, time.Duration(expire)*time.Second)
		}
	}

	return err
}

// MGet 同时获取一个或多个key的value
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息
func (c *RedisdCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	args := []string{}
	for _, k := range keys {
		if c.prefix != "" {
			k = c.prefix + k
		}
		args = append(args, k)
	}

	v, err := c.getClient().MGet(args...).Result()
	if err != nil {
		return mList, err
	}

	i := 0
	for _, val := range v {
		if val == nil {
			// 不存在
			mList[keys[i]] = nil
		} else {
			// 解密判断
			data, err := cache.Decode([]byte(val.(string)), c.encodeKey)
			if err != nil {
				return mList, err
			}

			mList[keys[i]] = string(data)
		}

		i++
	}

	return mList, nil
}

// MDel 同时删除一个或多个key
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RedisdCache) MDel(keys ...string) error {
	args := make([]string, len(keys))
	for k, v := range keys {
		if c.prefix != "" {
			v = c.prefix + v
		}
		args[k] = v
	}

	c.getClient().Del(args...)
	return nil
}

// Incr 缓存里的值自增
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (c *RedisdCache) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.getClient().IncrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// Decr 缓存里的值自减
// key不存在时会新建一个，再返回-1
//   参数
//     key:   递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (c *RedisdCache) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.getClient().DecrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// IsExist 判断key值是否存在
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false
func (c *RedisdCache) IsExist(key string) (bool, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	n, err := c.getClient().Exists(key).Result()
	if err != nil {
		return false, err
	}

	if n == 1 {
		return true, nil
	}

	return false, nil
}

// ClearAll 清空所有数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RedisdCache) ClearAll() error {
	keys, err := c.getClient().Keys("*").Result()
	if err != nil {
		return err
	}

	return c.getClient().Del(keys...).Err()
}

// Hset 添加哈希表
//   参数
//     key:    哈希表key值
//     field:  哈希表field值
//     val:    哈希表value值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     成功时返回添加的个数，失败返回错误信息
func (c *RedisdCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	err = c.getClient().HSet(key, field, data).Err()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.getClient().Expire(key, time.Duration(expire)*time.Second)
	}

	return 1, err
}

// HGet 查询哈希表数据
//   参数
//     key:   哈希表key值
//     field: 哈希表field值
//     val:   保存结果地址
//   返回
//     错误信息，是否存在
func (c *RedisdCache) HGet(key string, field string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.getClient().HGet(key, field).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// 类型转换
	err = cache.ByteToInter([]byte(v), val)
	if err != nil {
		return err, true
	}

	return nil, true
}

// HDel 删除哈希表数据
//   参数
//     key:   哈希表key值
//     fields: 哈希表field值
//   返回
//     成功返回nil，失败返回错误信息
func (c *RedisdCache) HDel(key string, fields ...string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().HDel(key, fields...).Err()
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (c *RedisdCache) HGetAll(key string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	res := make(map[string]interface{})
	val, err := c.getClient().HGetAll(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return res, nil
		}

		return nil, err
	}

	for k, v := range val {
		res[k] = v
	}

	return res, err
}

// HMGet 返回哈希表 key 中，一个或多个给定域的值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//   返回
//     查询的结果数据和错误码
func (c *RedisdCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	res := make(map[string]interface{})

	v, err := c.getClient().HMGet(key, fields...).Result()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return res, nil
	}

	i := 0
	for _, field := range fields {
		res[field] = v[i]
		i++
	}

	return res, err
}

// HVals 返回哈希表 key 中，所有的域和值
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (c *RedisdCache) HVals(key string) ([]interface{}, error) {
	vals, err := c.getClient().HVals(key).Result()
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(vals))
	for k, v := range vals {
		res[k] = v
	}

	return res, nil
}

// HIncr 哈希表的值自增
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递增后的结果、失败返回错误信息
func (c *RedisdCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().HIncrBy(key, fields, int64(delta[0])).Result()
}

// HDecr 哈希表的值自减
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递减后的结果、失败返回错误信息
func (c *RedisdCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().HIncrBy(key, fields, 0-int64(delta[0])).Result()
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(必须为float64), 后面为变量
//   返回
//     成功添加的数据和错误码
func (c *RedisdCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	valLen := len(val)
	if valLen < 2 || valLen%2 != 0 {
		return -1, errors.New("val param error")
	}
	vals := []redis.Z{}
	for i := 0; i < valLen-1; i += 2 {
		stZ := redis.Z{
			Score:  val[i].(float64),
			Member: val[i+1],
		}
		vals = append(vals, stZ)
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	n, err := c.getClient().ZAdd(key, vals...).Result()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.getClient().Expire(key, time.Duration(expire)*time.Second)
	}

	return n, err
}

// ZGet 查询有序集合
//   参数
//     key:        有序集合key值
//     start:      要查询有序集开始下标，0表示第一个，-1表示最后一个，-2表示倒数第二个
//     stop:       要查询有序集结束下标，0表示第一个，-1表示最后一个，-2表示倒数第二个
//     withScores: 是否带上score
//     isRev:      true-递减排列，使用ZREVRANGE命令 false-递增排列，使用ZRANGE命令
//   返回
//     查询的结果数据和错误码
func (c *RedisdCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	vals := []redis.Z{}
	res := []string{}

	if c.prefix != "" {
		key = c.prefix + key
	}

	if isRev {

		if withScores {
			vals, err = c.getClient().ZRevRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.getClient().ZRevRange(key, int64(start), int64(stop)).Result()
		}
	} else {
		if withScores {
			vals, err = c.getClient().ZRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.getClient().ZRange(key, int64(start), int64(stop)).Result()
		}
	}

	for _, val := range vals {
		res = append(res, fmt.Sprintf("%f", val.Score), fmt.Sprintf("%v", val.Member))
	}

	return res, err
}

// ZDel 删除有序集合数据
//   参数
//     key:   有序集合key值
//     field: 要删除的数据
//   返回
//     成功删除的数据个数和错误码
func (c *RedisdCache) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	for _, f := range field {
		args = append(args, f)
	}

	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.getClient().ZRem(key, args...).Result()
}

// ZRemRangeByRank 删除指定排名区间内的有序集合数据
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RedisdCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().ZRemRangeByRank(key, start, end).Result()
}

// ZRemRangeByScore 删除指定分值区间内的有序集合数据
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RedisdCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().ZRemRangeByScore(key, start, end).Result()
}

// ZRemRangeByLex 删除指定变量区间内的有序集合数据
// 对于一个所有成员的分值都相同的有序集合键 key 来说， 这个命令会移除该集合中， 成员介于 min 和 max 范围内的所有元素。
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RedisdCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.getClient().ZRemRangeByLex(key, start, end).Result()
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (c *RedisdCache) ZCard(key string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.getClient().ZCard(key).Result()
}

// Pipeline 执行pipeline命令
//   实例：
//     pipe := rc.Pipeline(false).Pipe
//     incr := pipe.Incr("pipeline_counter")
//     pipe.Expire("pipeline_counter", time.Hour)
//     _, err := pipe.Exec()
//     fmt.Println(incr.Val(), err)
//   参数
//     isTx: 是否事务模式
//   返回
//     成功刊返回命令执行的结果，失败返回错误信息
func (c *RedisdCache) Pipeline(isTx bool) cache.Pipeliner {
	p := cache.Pipeliner{}
	if isTx {
		p.Pipe = c.getClient().TxPipeline()
	} else {
		p.Pipe = c.getClient().Pipeline()
	}

	return p
}
