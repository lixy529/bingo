// redis cluster adapter
// 用于redis cluster操作
//   变更历史
//     2018-10-10  lixiaoya  新建
package redisc

import (
	"github.com/lixy529/bingo/cache"
	"github.com/go-redis/redis"
	json "github.com/json-iterator/go"
	"fmt"
	"strconv"
	"time"
	"strings"
	"errors"
)

const NOT_EXIST = "redis: nil"

// RediscCache Redis cluster 缓存
type RediscCache struct {
	client *redis.ClusterClient // 连接客户端

	addr         string        // 连接主机和端口，多个主机用逗号分割，如127.0.0.1:1900,127.0.0.2:1900
	auth         string        // 授权密码
	dialTimeout  time.Duration // 连接超时时间，单位秒，默认5秒
	readTimeout  time.Duration // 读超时时间，单位秒，-1-不超时，0-使用默认3秒
	writeTimeout time.Duration // 写超时时间，单位秒，默认为readTimeout
	poolSize     int           // 每个节点连接池的连接数，默认为cpu个数的5倍
	minIdleConns int           // 最少空闲连接数，默认为0
	maxConnAge   time.Duration // 最大连接时间，单位秒，超时时间自动关闭，默认为0
	poolTimeout  time.Duration // 如果所有连接都忙时的等待时间，默认为readTimeout+1秒
	idleTimeout  time.Duration // 最大空闲时间，单位秒，默认为5分钟

	prefix    string // key前缀，如果配置里有，则所有key前自动添加此前缀
	encodeKey []byte // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewRediscCache 新建一个RediscCache适配器.
func NewRediscCache() cache.Cache {
	return &RediscCache{}
}

// Init 初始化
//   参数
//     config: 配置josn串
//       {
//         "addr":"127.0.0.1:19100,127.0.0.2:19100,127.0.0.3:19100",
//         "auth":"xxxx",
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
//       dialTimeout:  连接超时时间，单位秒，默认5秒
//       readTimeout:  读超时时间，单位秒，-1-不超时，0-使用默认3秒
//       writeTimeout: 写超时时间，单位秒，默认为readTimeout
//       poolSize:     每个节点连接池的连接数，默认为cpu个数的5倍
//       minIdleConns: 最少空闲连接数，默认为0
//       maxConnAge:   最大连接时间，单位秒，超时时间自动关闭，默认为0
//       poolTimeout:  如果所有连接都忙时的等待时间，默认为readTimeout+1秒
//       idleTimeout:  最大空闲时间，单位秒，默认为5分钟
//       prefix:       key前缀，如果配置里有，则所有key前自动添加此前缀
//       encodeKey:    数据如果要加密，传的加密密钥
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RediscCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RediscCache: Unmarshal json[%s] error, %s", config, err.Error())
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

	// 连接
	c.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        strings.Split(c.addr, ","),
		Password:     c.auth,
		DialTimeout:  c.dialTimeout * time.Second,
		ReadTimeout:  c.readTimeout * time.Second,
		WriteTimeout: c.writeTimeout * time.Second,
		PoolSize:     c.poolSize,
		MinIdleConns: c.minIdleConns,
		MaxConnAge:   c.maxConnAge * time.Second,
		PoolTimeout:  c.poolTimeout * time.Second,
		IdleTimeout:  c.idleTimeout * time.Second,
	})

	return nil
}

// Set 向缓存设置一个值
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RediscCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
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

	return c.client.Set(key, data, time.Duration(expire)*time.Second).Err()
}

// Get 从缓存取一个值
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (c *RediscCache) Get(key string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.client.Get(key).Result()
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
func (c *RediscCache) Del(key string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.Del(key).Err()
}

// MSet 同时设置一个或多个key-value对
// CROSSSLOT Keys in request don't hash to the same slot.
// 存在上面的错误，所以每个每单独写入，没用MSet函数
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (c *RediscCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	for k, v := range mList {
		err := c.Set(k, v, expire, encode...)
		if err != nil {
			return err
		}
	}

	return nil
}

// MGet 同时获取一个或多个key的value
// CROSSSLOT Keys in request don't hash to the same slot.
// 存在上面的错误，所以每个每单独写入，没用MGet函数
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息
func (c *RediscCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	for _, k := range keys {
		fmt.Println("k = ", k)
		v := ""
		err, b := c.Get(k, &v)
		if err != nil {
			return mList, err
		} else if !b {
			v = ""
		}
		mList[k] = v
	}

	return mList, nil
}

// MDel 同时删除一个或多个key
// CROSSSLOT Keys in request don't hash to the same slot.
// 存在上面的错误，所以每个每单独写入，没用Del函数
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (c *RediscCache) MDel(keys ...string) error {
	for _, key := range keys {
		err := c.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Incr 缓存里的值自增
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (c *RediscCache) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.client.IncrBy(key, int64(delta[0])).Result()
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
func (c *RediscCache) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}
	v, err := c.client.DecrBy(key, int64(delta[0])).Result()
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
func (c *RediscCache) IsExist(key string) (bool, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	n, err := c.client.Exists(key).Result()
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
func (c *RediscCache) ClearAll() error {
	keys, err := c.client.Keys("*").Result()
	if err != nil {
		return err
	}

	return c.client.Del(keys...).Err()
}

// Hset 添加哈希表
//   参数
//     key:    哈希表key值
//     field:  哈希表field值
//     val:    哈希表value值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     成功时返回添加的个数，失败返回错误信息
func (c *RediscCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	if c.prefix != "" {
		key = c.prefix + key
	}

	err = c.client.HSet(key, field, data).Err()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
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
func (c *RediscCache) HGet(key string, field string, val interface{}) (error, bool) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	v, err := c.client.HGet(key, field).Result()
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
func (c *RediscCache) HDel(key string, fields ...string) error {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HDel(key, fields...).Err()
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (c *RediscCache) HGetAll(key string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	res := make(map[string]interface{})
	val, err := c.client.HGetAll(key).Result()
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
func (c *RediscCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	res := make(map[string]interface{})

	v, err := c.client.HMGet(key, fields...).Result()
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
func (c *RediscCache) HVals(key string) ([]interface{}, error) {
	vals, err := c.client.HVals(key).Result()
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
func (c *RediscCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HIncrBy(key, fields, int64(delta[0])).Result()
}

// HDecr 哈希表的值自减
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递减后的结果、失败返回错误信息
func (c *RediscCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.HIncrBy(key, fields, 0-int64(delta[0])).Result()
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(必须为float64), 后面为变量
//   返回
//     成功添加的数据和错误码
func (c *RediscCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
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

	n, err := c.client.ZAdd(key, vals...).Result()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		c.client.Expire(key, time.Duration(expire)*time.Second)
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
func (c *RediscCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	vals := []redis.Z{}
	res := []string{}

	if c.prefix != "" {
		key = c.prefix + key
	}

	if isRev {

		if withScores {
			vals, err = c.client.ZRevRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.client.ZRevRange(key, int64(start), int64(stop)).Result()
		}
	} else {
		if withScores {
			vals, err = c.client.ZRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return c.client.ZRange(key, int64(start), int64(stop)).Result()
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
func (c *RediscCache) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	for _, f := range field {
		args = append(args, f)
	}

	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.client.ZRem(key, args...).Result()
}

// ZRemRangeByRank 删除指定排名区间内的有序集合数据
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RediscCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByRank(key, start, end).Result()
}

// ZRemRangeByScore 删除指定分值区间内的有序集合数据
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RediscCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByScore(key, start, end).Result()
}

// ZRemRangeByLex 删除指定变量区间内的有序集合数据
// 对于一个所有成员的分值都相同的有序集合键 key 来说， 这个命令会移除该集合中， 成员介于 min 和 max 范围内的所有元素。
//   参数
//     key:   有序集合key值
//     start: 开始值
//     end:   结束值
//   返回
//     成功删除的数据个数和错误码
func (c *RediscCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}

	return c.client.ZRemRangeByLex(key, start, end).Result()
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (c *RediscCache) ZCard(key string) (int64, error) {
	if c.prefix != "" {
		key = c.prefix + key
	}
	return c.client.ZCard(key).Result()
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
func (c *RediscCache) Pipeline(isTx bool) cache.Pipeliner {
	p := cache.Pipeliner{}
	if isTx {
		p.Pipe = c.client.TxPipeline()
	} else {
		p.Pipe = c.client.Pipeline()
	}

	return p
}
