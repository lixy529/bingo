// redism adapter
// 用于主从模式的redis操作
//   变更历史
//     2017-02-21  lixiaoya  新建
package redism

import (
	json "github.com/json-iterator/go"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/lixy529/bingo/cache"
	"strconv"
	"time"
)

const NOT_EXIST = "redis: nil"

// RedismCache Redis缓存
type RedismCache struct {
	master *RedisPool // 主库
	slave  *RedisPool // 从库

	mAddr  string // 主库连接串
	mDbNum int    // 主库DbNum
	mAuth  string // 主库授权码

	sAddr  string // 从库连接串
	sDbNum int    // 从库DbNum
	sAuth  string // 从库授权码

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

// NewRedismCache 新建一个RedismCache适配器.
func NewRedismCache() cache.Cache {
	return &RedismCache{}
}

// Init 初始化
//   参数
//     config: 配置josn串
//       {
//         "master":{"addr":"127.0.0.1:6379","dbNum":"0","auth":"xxxxx"},
//         "slave":{"addr":"127.0.0.2:6379","dbNum":"1","auth":"xxxxx"},
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
//       addr:         连接主机和端口，如127.0.0.1:1900
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
func (rc *RedismCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RedismCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// 连接超时时间
	dialTimeout, err := strconv.Atoi(mapCfg["dialTimeout"])
	if err != nil || dialTimeout < 0 {
		rc.dialTimeout = 5
	} else {
		rc.dialTimeout = time.Duration(dialTimeout)
	}

	// 读超时时间
	readTimeout, err := strconv.Atoi(mapCfg["readTimeout"])
	if err != nil {
		rc.readTimeout = 3
	} else if readTimeout < 0 {
		rc.readTimeout = -1
	} else {
		rc.readTimeout = time.Duration(readTimeout)
	}

	// 写超时时间
	writeTimeout, err := strconv.Atoi(mapCfg["writeTimeout"])
	if err != nil {
		rc.writeTimeout = rc.readTimeout
	} else if writeTimeout < 0 {
		rc.writeTimeout = -1
	} else {
		rc.writeTimeout = time.Duration(writeTimeout)
	}

	// 每个节点连接池的连接数
	poolSize, err := strconv.Atoi(mapCfg["poolSize"])
	if err != nil || poolSize < 0 {
		rc.poolSize = 0
	} else {
		rc.poolSize = poolSize
	}

	// 最少空闲连接数
	minIdleConns, err := strconv.Atoi(mapCfg["minIdleConns"])
	if err != nil || minIdleConns < 0 {
		rc.minIdleConns = 0
	} else {
		rc.minIdleConns = minIdleConns
	}

	// 最大连接时间
	maxConnAge, err := strconv.Atoi(mapCfg["maxConnAge"])
	if err != nil || maxConnAge < 0 {
		rc.maxConnAge = 0
	} else {
		rc.maxConnAge = time.Duration(maxConnAge)
	}

	// 如果所有连接都忙时的等待时间
	poolTimeout, err := strconv.Atoi(mapCfg["poolTimeout"])
	if err != nil || poolTimeout < 0 {
		rc.poolTimeout = rc.readTimeout + 1
	} else {
		rc.poolTimeout = time.Duration(poolTimeout)
	}

	// 最大空闲时间
	idleTimeout, err := strconv.Atoi(mapCfg["idleTimeout"])
	if err != nil || idleTimeout < 0 {
		rc.idleTimeout = 300
	} else {
		rc.idleTimeout = time.Duration(idleTimeout)
	}

	// 前缀
	if prefix, ok := mapCfg["prefix"]; ok {
		rc.prefix = prefix
	}

	// 加密密钥
	if tmp, ok := mapCfg["encodeKey"]; ok && tmp != "" {
		rc.encodeKey = []byte(tmp)
	}

	// 主库配置
	rc.mAddr = mapCfg["mAddr"]
	if rc.mAddr == "" {
		return errors.New("RedismCache: Master addr is empty")
	}

	dbNum, err := strconv.Atoi(mapCfg["mDbNum"])
	if err != nil {
		rc.mDbNum = 0
	} else {
		rc.mDbNum = dbNum
	}

	rc.mAuth = mapCfg["mAuth"]

	// 实例化主库
	rc.master = NewRedisPool(rc.mAddr, rc.mAuth, rc.mDbNum, rc.dialTimeout, rc.readTimeout, rc.writeTimeout, rc.poolSize, rc.minIdleConns, rc.maxConnAge, rc.poolTimeout, rc.idleTimeout, rc.prefix, rc.encodeKey)

	// 从库配置
	rc.sAddr = mapCfg["sAddr"]
	if rc.sAddr == "" {
		rc.slave = rc.master
	} else {
		dbNum, err := strconv.Atoi(mapCfg["sDbNum"])
		if err != nil {
			rc.sDbNum = 0
		} else {
			rc.sDbNum = dbNum
		}

		rc.sAuth = mapCfg["sAuth"]
		rc.slave = NewRedisPool(rc.sAddr, rc.sAuth, rc.sDbNum, rc.dialTimeout, rc.readTimeout, rc.writeTimeout, rc.poolSize, rc.minIdleConns, rc.maxConnAge, rc.poolTimeout, rc.idleTimeout, rc.prefix, rc.encodeKey)
	}

	return nil
}

// Set 向缓存设置一个值，访问主库
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedismCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	return rc.master.Set(key, val, expire, encode...)
}

// Get 从缓存取一个值，访问从库
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (rc *RedismCache) Get(key string, val interface{}) (error, bool) {
	return rc.slave.Get(key, val)
}

// Del 从缓存删除一个值，访问主库
//   参数
//     key:    key值
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedismCache) Del(key string) error {
	return rc.master.Del(key)
}

// MSet 同时设置一个或多个key-value对，访问主库
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (rc *RedismCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	return rc.master.MSet(mList, expire, encode...)
}

// MGet 同时获取一个或多个key的value，访问从库
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息
func (rc *RedismCache) MGet(keys ...string) (map[string]interface{}, error) {
	return rc.slave.MGet(keys...)
}

// MDel 同时删除一个或多个key，访问主库
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedismCache) MDel(keys ...string) error {
	return rc.master.MDel(keys...)
}

// Incr 缓存里的值自增，访问主库
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (rc *RedismCache) Incr(key string, delta ...uint64) (int64, error) {
	return rc.master.Incr(key, delta...)
}

// Decr 缓存里的值自减，访问主库
// key不存在时会新建一个，再返回-1
//   参数
//     key:   递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (rc *RedismCache) Decr(key string, delta ...uint64) (int64, error) {
	return rc.master.Decr(key, delta...)
}

// IsExist 判断key值是否存在，访问从库
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false
func (rc *RedismCache) IsExist(key string) (bool, error) {
	return rc.slave.IsExist(key)
}

// ClearAll 清空所有数据，访问主库
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedismCache) ClearAll() error {
	return rc.master.ClearAll()
}

// Hset 添加哈希表，访问主库
//   参数
//     key:    哈希表key值
//     field:  哈希表field值
//     val:    哈希表value值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     成功时返回添加的个数，失败返回错误信息
func (rc *RedismCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	return rc.master.HSet(key, field, val, expire)
}

// HGet 查询哈希表数据，访问从库
//   参数
//     key:   哈希表key值
//     field: 哈希表field值
//     val:   保存结果地址
//   返回
//     错误信息，是否存在
func (rc *RedismCache) HGet(key string, field string, val interface{}) (error, bool) {
	return rc.slave.HGet(key, field, val)
}

// HDel 删除哈希表数据，访问主库
//   参数
//     key:   哈希表key值
//     fields: 哈希表field值
//   返回
//     成功返回nil，失败返回错误信息
func (rc *RedismCache) HDel(key string, fields ...string) error {
	return rc.master.HDel(key, fields...)
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (rc *RedismCache) HGetAll(key string) (map[string]interface{}, error) {
	return rc.slave.HGetAll(key)
}

// HMGet 返回哈希表 key 中，一个或多个给定域的值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//   返回
//     查询的结果数据和错误码
func (rc *RedismCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	return rc.slave.HMGet(key, fields...)
}

// HVals 返回哈希表 key 中，所有的域和值
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (rc *RedismCache) HVals(key string) ([]interface{}, error) {
	return rc.slave.HVals(key)
}

// HIncr 哈希表的值自增
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递增后的结果、失败返回错误信息
func (rc *RedismCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	return rc.master.HIncr(key, fields, delta...)
}

// HDecr 哈希表的值自减
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递减后的结果、失败返回错误信息
func (rc *RedismCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	return rc.master.HDecr(key, fields, delta...)
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(整数值或双精度浮点数), 后面为变量
//   返回
//     成功添加的数据和错误码
func (rc *RedismCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	return rc.master.ZSet(key, expire, val...)
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
func (rc *RedismCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	return rc.slave.ZGet(key, start, stop, withScores, isRev)
}

// ZDel 删除有序集合数据
//   参数
//     key:   有序集合key值
//     field: 要删除的数据
//   返回
//     成功删除的数据个数和错误码
func (rc *RedismCache) ZDel(key string, field ...string) (int64, error) {
	return rc.master.ZDel(key, field...)
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (rc *RedismCache) ZCard(key string) (int64, error) {
	return rc.slave.ZCard(key)
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
//     成功刊返回命令执行的结果
func (rc *RedismCache) Pipeline(isTx bool) cache.Pipeliner {
	return rc.master.Pipeline(isTx)
}

// RedisPool Redis缓存
type RedisPool struct {
	client *redis.Client // 连接池

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

// NewRedisPool 实例化RedisPool对象
//   参数
//     addr:         连接主机和端口
//     auth:         授权码
//     dbNum:        db号
//     dialTimeout:  连接超时时间，单位秒，默认5秒
//     readTimeout:  读超时时间，单位秒，-1-不超时，0-使用默认3秒
//     writeTimeout: 写超时时间，单位秒，默认为readTimeout
//     poolSize:     每个节点连接池的连接数，默认为cpu个数的10倍
//     minIdleConns: 最少空闲连接数，默认为0
//     maxConnAge:   最大连接时间，单位秒，超时时间自动关闭，默认为0
//     poolTimeout:  如果所有连接都忙时的等待时间，默认为readTimeout+1秒
//     idleTimeout:  最大空闲时间，单位秒，默认为5分钟
//     prefix:       key前缀，如果配置里有，则所有key前自动添加此前缀
//     encodeKey:    加密key
//   返回
//     成功时Redis连接池
func NewRedisPool(addr, auth string, dbNum int, dialTimeout, readTimeout, writeTimeout time.Duration, poolSize, minIdleConns int, maxConnAge, poolTimeout, idleTimeout time.Duration, prefix string, encodeKey []byte) *RedisPool {
	rp := &RedisPool{
		addr:         addr,
		auth:         auth,
		dbNum:        dbNum,
		dialTimeout:  dialTimeout,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		poolSize:     poolSize,
		minIdleConns: minIdleConns,
		maxConnAge:   maxConnAge,
		poolTimeout:  poolTimeout,
		idleTimeout:  idleTimeout,
		prefix:       prefix,
		encodeKey:    encodeKey,
	}

	rp.connect()

	return rp
}

// connect 连接redis
//   参数
//
//   返回
//     成功时返回nil，失败时返回错误信息
func (rp *RedisPool) connect() {
	rp.client = redis.NewClient(&redis.Options{
		Addr:         rp.addr,
		Password:     rp.auth,
		DB:           rp.dbNum,
		DialTimeout:  rp.dialTimeout * time.Second,
		ReadTimeout:  rp.readTimeout * time.Second,
		WriteTimeout: rp.writeTimeout * time.Second,
		PoolSize:     rp.poolSize,
		MinIdleConns: rp.minIdleConns,
		MaxConnAge:   rp.maxConnAge * time.Second,
		PoolTimeout:  rp.poolTimeout * time.Second,
		IdleTimeout:  rp.idleTimeout * time.Second,
	})

	return
}

// Set 向缓存设置一个值
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (rp *RedisPool) Set(key string, val interface{}, expire int32, encode ...bool) error {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// 加密判断
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, rp.encodeKey)
		if err != nil {
			return err
		}
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.Set(key, data, time.Duration(expire)*time.Second).Err()
}

// Get 从缓存取一个值
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (rp *RedisPool) Get(key string, val interface{}) (error, bool) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	v, err := rp.client.Get(key).Result()
	if err != nil {
		if err.Error() == NOT_EXIST {
			return nil, false
		}
		return err, false
	}

	// 解密判断
	data, err := cache.Decode([]byte(v), rp.encodeKey)
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
//     key:    key值
//   返回
//     成功时返回nil，失败返回错误信息
func (rp *RedisPool) Del(key string) error {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.Del(key).Err()
}

// MSet 同时设置一个或多个key-value对
//   参数
//     mList:  key-value对
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (rp *RedisPool) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
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
			data, err = cache.Encode(data, rp.encodeKey)
			if err != nil {
				return err
			}
		}

		if rp.prefix != "" {
			key = rp.prefix + key
		}
		v = append(v, key, data)
	}

	err := rp.client.MSet(v...).Err()
	if err != nil {
		return err
	}

	// 设置失效时间
	if expire > 0 {
		for key := range mList {
			if rp.prefix != "" {
				key = rp.prefix + key
			}
			rp.client.Expire(key, time.Duration(expire)*time.Second)
		}
	}

	return err
}

// MGet 同时获取一个或多个key的value，
// struct、map类型返回的结果需要调用方做一下json.Unmarshal处理
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (rp *RedisPool) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	args := []string{}
	for _, k := range keys {
		if rp.prefix != "" {
			k = rp.prefix + k
		}
		args = append(args, k)
	}

	v, err := rp.client.MGet(args...).Result()
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
			data, err := cache.Decode([]byte(val.(string)), rp.encodeKey)
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
func (rp *RedisPool) MDel(keys ...string) error {
	args := make([]string, len(keys))
	for k, v := range keys {
		if rp.prefix != "" {
			v = rp.prefix + v
		}
		args[k] = v
	}

	rp.client.Del(args...)
	return nil
}

// Incr 缓存里的值自增
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (rp *RedisPool) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	v, err := rp.client.IncrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// Decr 缓存里的值自减
// key不存在时会新建一个，再返回-1
//   参数
//     key:   要递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (rp *RedisPool) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	v, err := rp.client.DecrBy(key, int64(delta[0])).Result()
	if err != nil {
		return 0, err
	}

	return v, nil
}

// IsExist 判断key值是否存在
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false，出错时返回错误信息
func (rp *RedisPool) IsExist(key string) (bool, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	n, err := rp.client.Exists(key).Result()
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
func (rp *RedisPool) ClearAll() error {
	keys, err := rp.client.Keys("*").Result()
	if err != nil {
		return err
	}

	return rp.client.Del(keys...).Err()
}

// Hset 添加哈希表
//   参数
//     key:    哈希表key值
//     field:  哈希表field值
//     val:    哈希表value值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     成功时返回添加的个数，失败返回错误信息
func (rp *RedisPool) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	err = rp.client.HSet(key, field, data).Err()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
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
func (rp *RedisPool) HGet(key string, field string, val interface{}) (error, bool) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	v, err := rp.client.HGet(key, field).Result()
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
//     key:    哈希表key值
//     fields: 哈希表field值
//   返回
//     成功返回nil，失败返回错误信息
func (rp *RedisPool) HDel(key string, fields ...string) error {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HDel(key, fields...).Err()
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
//   示例
//     key := "addr"
//     val, err := adapter.HGetAll(key)
//     if err != nil {
//         fmt.Printf("Redis HGetAll failed. err: %s.", err.Error())
//         return
//     }
//     for k, v := range val {
//         var data StUser
//         if v != nil {
//             json.Unmarshal(v.([]byte), &data)
//         }
//         fmt.Println(k, data)
//     }
func (rp *RedisPool) HGetAll(key string) (map[string]interface{}, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	res := make(map[string]interface{})
	val, err := rp.client.HGetAll(key).Result()
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
func (rp *RedisPool) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	res := make(map[string]interface{})

	v, err := rp.client.HMGet(key, fields...).Result()
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

// HVals 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key:        有序集合key值
//   返回
//     查询的结果数据和错误码
//   示例
//     key := "addr"
//     val, err := adapter.HVals(key)
//     if err != nil {
//         fmt.Printf("Redis HVals failed. err: %s.", err.Error())
//         return
//     }
//     for _, v := range val {
//         var data StUser
//         if v != nil {
//             json.Unmarshal(v.([]byte), &data)
//         }
//         fmt.Println(data)
//     }
func (rp *RedisPool) HVals(key string) ([]interface{}, error) {
	vals, err := rp.client.HVals(key).Result()
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
func (rp *RedisPool) HIncr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HIncrBy(key, fields, int64(delta[0])).Result()
}

// HDecr 哈希表的值自减
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//     delta:  递增的量，默认为1
//   返回
//     递减后的结果、失败返回错误信息
func (rp *RedisPool) HDecr(key, fields string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	if rp.prefix != "" {
		key = rp.prefix + key
	}

	return rp.client.HIncrBy(key, fields, 0-int64(delta[0])).Result()
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(整数值或双精度浮点数), 后面为变量
//   返回
//     成功添加的数据个数和错误码
func (rp *RedisPool) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
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

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	n, err := rp.client.ZAdd(key, vals...).Result()
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		rp.client.Expire(key, time.Duration(expire)*time.Second)
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
func (rp *RedisPool) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	vals := []redis.Z{}
	res := []string{}

	if rp.prefix != "" {
		key = rp.prefix + key
	}

	if isRev {

		if withScores {
			vals, err = rp.client.ZRevRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return rp.client.ZRevRange(key, int64(start), int64(stop)).Result()
		}
	} else {
		if withScores {
			vals, err = rp.client.ZRangeWithScores(key, int64(start), int64(stop)).Result()
			if err != nil {
				return res, err
			}
		} else {
			return rp.client.ZRange(key, int64(start), int64(stop)).Result()
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
func (rp *RedisPool) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	for _, f := range field {
		args = append(args, f)
	}

	if rp.prefix != "" {
		key = rp.prefix + key
	}
	return rp.client.ZRem(key, args...).Result()
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (rp *RedisPool) ZCard(key string) (int64, error) {
	if rp.prefix != "" {
		key = rp.prefix + key
	}
	return rp.client.ZCard(key).Result()
}

// Pipeline 执行pipeline命令
//   实例：
//     pipe := rc.Pipeline(false)
//     incr := pipe.Incr("pipeline_counter")
//     pipe.Expire("pipeline_counter", time.Hour)
//     _, err := pipe.Exec()
//     fmt.Println(incr.Val(), err)
//   参数
//     isTx: 是否事务模式
//   返回
//     成功返回命令执行的结果，失败返回错误信息
func (rp *RedisPool) Pipeline(isTx bool) cache.Pipeliner {
	p := cache.Pipeliner{}
	if isTx {
		p.Pipe = rp.client.TxPipeline()
	} else {
		p.Pipe = rp.client.Pipeline()
	}

	return p
}
