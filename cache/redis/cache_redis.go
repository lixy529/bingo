// redis adapter
//   变更历史
//     2017-02-21  lixiaoya  新建
package redis

import (
	json "github.com/json-iterator/go"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/lixy529/bingo/cache"
	"strconv"
	"time"
)

// RedisCache Redis缓存
type RedisCache struct {
	master *RedisPool // 主库
	slave  *RedisPool // 从库

	mConn  string // 主库连接串
	mDbNum int    // 主库DbNum
	mAuth  string // 主库授权码

	sConn  string // 从库连接串
	sDbNum int    // 从库DbNum
	sAuth  string // 从库授权码

	maxIdle     int           // 最大空闲连接数，默认为3
	maxActive   int           // 最大的激活连接数，默认为0，0不限制
	idleTimeOut time.Duration // 空闲超时时间，默认为180秒，0关闭

	encodeKey []byte // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewRedisCache 新建一个RedisCache适配器.
func NewRedisCache() cache.Cache {
	return &RedisCache{}
}

// Init 初始化
//   参数
//     config: 配置josn串
//       {
//       "master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"xxxxx"},
//       "slave":{"conn":"127.0.0.2:6379","dbNum":"1","auth":"xxxxx"},
//       "maxIdle":"3",
//       "maxActive":"10",
//       "idleTimeOut":"180"
//       "encodeKey":"abcdefghij123456",
//       }
//       auth: 授权密码
//       maxIdle: 最大空闲连接数，默认为3
//       maxActive: 最大连接数，0不限制，默认为0
//       idelTimeOut: 最大空闲时间，单位秒，0不限制，默认为0
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedisCache) Init(config string) error {
	var mapCfg map[string]interface{}
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("RedisCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// 最大空闲连接数
	rc.maxIdle, err = strconv.Atoi(mapCfg["maxIdle"].(string))
	if err != nil {
		rc.maxIdle = 3
	}

	// 最大激活连接数
	rc.maxActive, err = strconv.Atoi(mapCfg["maxActive"].(string))
	if err != nil {
		rc.maxActive = 0
	}

	// 空闲连接的超时时间
	idelTimeOut, err := strconv.Atoi(mapCfg["idleTimeOut"].(string))
	if err != nil {
		rc.idleTimeOut = 0
	} else {
		rc.idleTimeOut = time.Duration(idelTimeOut)
	}

	// 主库配置
	var ok bool
	var temp interface{}
	temp, ok = mapCfg["master"]
	if !ok {
		return errors.New("RedisCache: Master config is empty")
	}

	masterCfg := temp.(map[string]interface{})
	temp, ok = masterCfg["conn"]
	if !ok {
		return errors.New("RedisCache: Master conn is empty")
	}
	rc.mConn = temp.(string)

	temp, ok = masterCfg["dbNum"]
	if ok {
		rc.mDbNum, _ = strconv.Atoi(temp.(string))
	} else {
		rc.mDbNum = 0
	}

	rc.mAuth = masterCfg["auth"].(string)

	// 实例化主库
	rc.master, err = NewRedisPool(rc.mConn, rc.mDbNum, rc.mAuth, rc.maxIdle, rc.maxActive, rc.idleTimeOut, rc.encodeKey)
	if err != nil {
		return fmt.Errorf("RedisCache: New master pool err: %s", err.Error())
	}

	// 从库配置
	temp, ok = mapCfg["slave"]
	if !ok {
		rc.slave = rc.master
	} else {
		slaveCfg := temp.(map[string]interface{})
		temp, ok = slaveCfg["conn"]
		if !ok {
			return errors.New("RedisCache: Slave conn is empty")
		}
		rc.sConn = temp.(string)

		temp, ok = slaveCfg["dbNum"]
		if ok {
			rc.sDbNum, _ = strconv.Atoi(temp.(string))
		} else {
			rc.sDbNum = 0
		}

		rc.sAuth = slaveCfg["auth"].(string)

		// 实例化从库
		rc.slave, err = NewRedisPool(rc.sConn, rc.sDbNum, rc.sAuth, rc.maxIdle, rc.maxActive, rc.idleTimeOut, rc.encodeKey)
		if err != nil {
			return fmt.Errorf("RedisCache: New slave pool err: %s", err.Error())
		}
	}

	// 加密密钥
	if tmp, ok := mapCfg["encodeKey"]; ok {
		rc.encodeKey = []byte(tmp.(string))
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
func (rc *RedisCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	return rc.master.Set(key, val, expire, encode...)
}

// Get 从缓存取一个值，访问从库
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (rc *RedisCache) Get(key string, val interface{}) (error, bool) {
	return rc.slave.Get(key, val)
}

// Del 从缓存删除一个值，访问主库
//   参数
//     key:    key值
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedisCache) Del(key string) error {
	return rc.master.Del(key)
}

// MSet 同时设置一个或多个key-value对，访问主库
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (rc *RedisCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	return rc.master.MSet(mList, expire, encode...)
}

// MGet 同时获取一个或多个key的value，访问从库
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息
func (rc *RedisCache) MGet(keys ...string) (map[string]interface{}, error) {
	return rc.slave.MGet(keys...)
}

// MDel 同时删除一个或多个key，访问主库
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedisCache) MDel(keys ...string) error {
	return rc.master.MDel(keys...)
}

// Incr 缓存里的值自增，访问主库
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (rc *RedisCache) Incr(key string, delta ...uint64) (int64, error) {
	return rc.master.Incr(key, delta...)
}

// Decr 缓存里的值自减，访问主库
// key不存在时会新建一个，再返回-1
//   参数
//     key:   递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (rc *RedisCache) Decr(key string, delta ...uint64) (int64, error) {
	return rc.master.Decr(key, delta...)
}

// IsExist 判断key值是否存在，访问从库
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false
func (rc *RedisCache) IsExist(key string) (bool, error) {
	return rc.slave.IsExist(key)
}

// ClearAll 清空所有数据，访问主库
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (rc *RedisCache) ClearAll() error {
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
func (rc *RedisCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	return rc.master.HSet(key, field, val, expire)
}

// HGet 查询哈希表数据，访问从库
//   参数
//     key:   哈希表key值
//     field: 哈希表field值
//     val:   保存结果地址
//   返回
//     错误信息，是否存在
func (rc *RedisCache) HGet(key string, field string, val interface{}) (error, bool) {
	return rc.slave.HGet(key, field, val)
}

// HDel 删除哈希表数据，访问主库
//   参数
//     key:   哈希表key值
//     fields: 哈希表field值
//   返回
//     成功返回nil，失败返回错误信息
func (rc *RedisCache) HDel(key string, fields ...string) error {
	return rc.master.HDel(key, fields...)
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (rc *RedisCache) HGetAll(key string) (map[string]interface{}, error) {
	return rc.slave.HGetAll(key)
}

// HMGet 返回哈希表 key 中，一个或多个给定域的值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key:    有序集合key值
//     fields: 给定域的集合
//   返回
//     查询的结果数据和错误码
func (rc *RedisCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	return rc.slave.HMGet(key, fields...)
}

// HVals 返回哈希表 key 中，所有的域和值
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (rc *RedisCache) HVals(key string) ([]interface{}, error) {
	return rc.slave.HVals(key)
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(整数值或双精度浮点数), 后面为变量
//   返回
//     成功添加的数据和错误码
func (rc *RedisCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
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
func (rc *RedisCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	return rc.slave.ZGet(key, start, stop, withScores, isRev)
}

// ZDel 删除有序集合数据
//   参数
//     key:   有序集合key值
//     field: 要删除的数据
//   返回
//     成功删除的数据个数和错误码
func (rc *RedisCache) ZDel(key string, field ...string) (int64, error) {
	return rc.master.ZDel(key, field...)
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (rc *RedisCache) ZCard(key string) (int64, error) {
	return rc.slave.ZCard(key)
}

// Pipeline 执行pipeline命令
//   参数
//     cmds: 要执行的命令
//   返回
//     成功刊返回命令执行的结果，失败返回错误信息
func (rc *RedisCache) Pipeline(cmds ...cache.Cmd) ([]cache.PipeRes) {
	return rc.master.Pipeline(cmds...)
}

// Exec 执行pipeline事务命令
//   参数
//     cmds: 要执行的命令
//   返回
//     成功刊返回命令执行的结果，失败返回错误信息
func (rc *RedisCache) Exec(cmds ...cache.Cmd) (interface{}, error) {
	return rc.master.Exec(cmds...)
}

// RedisCache Redis缓存
type RedisPool struct {
	connPool    *redis.Pool   // 连接池
	connCfg     string        // 连接配置
	dbNum       int           // 默认数据库
	auth        string        // 认证
	maxIdle     int           // 最大空闲连接数，默认为3
	maxActive   int           // 最大的激活连接数，默认为0，0不限制
	idleTimeOut time.Duration // 空闲超时时间，默认为180秒，0关闭
	encodeKey   []byte        // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewRedisPool 实例化RedisPool对象
//   参数
//     conn:  连接配置
//     dbNum: db号
//     auth:  授权码
//     maxIdle: 最大空闲连接数，默认为3
//     maxActive: 最大的激活连接数，默认为0，0不限制
//     idleTimeOut: 空闲超时时间，默认为180秒，0关闭
//   返回
//     成功时Redis连接池，失败时返回错误信息
func NewRedisPool(conn string, dbNum int, auth string, maxIdle int, maxActive int, idleTimeOut time.Duration, encodeKey []byte) (*RedisPool, error) {
	rp := &RedisPool{
		connCfg:     conn,
		dbNum:       dbNum,
		auth:        auth,
		maxIdle:     maxIdle,
		maxActive:   maxActive,
		idleTimeOut: idleTimeOut,
		encodeKey:   encodeKey,
	}

	err := rp.connect()
	if err != nil {
		return nil, err
	}

	c := rp.connPool.Get()
	defer c.Close()

	return rp, c.Err()
}

// connect 连接redis
//   参数
//
//   返回
//     成功时返回nil，失败时返回错误信息
func (rp *RedisPool) connect() error {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rp.connCfg)
		if err != nil {
			return nil, err
		}

		if rp.auth != "" {
			if _, err := c.Do("AUTH", rp.auth); err != nil {
				c.Close()
				return nil, err
			}
		}

		if rp.dbNum > 0 {
			_, selErr := c.Do("SELECT", rp.dbNum)
			if selErr != nil {
				c.Close()
				return nil, selErr
			}
		}

		return
	}
	// initialize a new pool
	rp.connPool = &redis.Pool{
		MaxIdle:     rp.maxIdle,
		MaxActive:   rp.maxActive,
		IdleTimeout: rp.idleTimeOut * time.Second,
		Dial:        dialFunc,
	}

	return nil
}

// do 从连接池里取一个连接，调用redis命令
//   参数
//     commandName:  命令串
//     args:         参数
//   返回
//     成功时返回结果信息，失败时返回错误信息
func (rp *RedisPool) do(commandName string, args ...interface{}) (interface{}, error) {
	c := rp.connPool.Get()
	defer c.Close()

	return c.Do(commandName, args...)
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

	if expire > 0 {
		_, err = rp.do("SETEX", key, expire, data)
	} else {
		_, err = rp.do("SET", key, data)
	}

	return err
}

// Get 从缓存取一个值
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (rp *RedisPool) Get(key string, val interface{}) (error, bool) {
	v, err := rp.do("GET", key)
	if err != nil {
		return err, false
	}

	if v == nil {
		return nil, false
	}

	// 解密判断
	data, err := cache.Decode(v.([]byte), rp.encodeKey)
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
	_, err := rp.do("DEL", key)

	return err
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

		v = append(v, key, data)
	}

	_, err := rp.do("MSET", v...)
	if err != nil {
		return err
	}

	// 设置失效时间
	if expire > 0 {
		for key := range mList {
			_, err = rp.do("EXPIRE", key, expire)
			if err != nil {
				rp.do("DEL", key)
			}
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
	var args []interface{}
	for _, k := range keys {
		args = append(args, k)
	}

	v, err := rp.do("MGET", args...)
	if err != nil {
		return mList, err
	}

	i := 0
	for _, val := range v.([]interface{}) {
		if val == nil {
			// 不存在
			mList[keys[i]] = nil
		} else {
			// 解密判断
			data, err := cache.Decode(val.([]byte), rp.encodeKey)
			if err != nil {
				return mList, err
			}

			mList[keys[i]] = data
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
	for _, key := range keys {
		err := rp.Del(key)
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
func (rp *RedisPool) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	v, err := rp.do("INCRBY", key, delta[0])
	if err != nil {
		return 0, err
	} else {
		return v.(int64), nil
	}
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
	v, err := rp.do("INCRBY", key, 0-int64(delta[0]))
	if err != nil {
		return 0, err
	} else {
		return v.(int64), nil
	}
}

// IsExist 判断key值是否存在
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false，出错时返回错误信息
func (rp *RedisPool) IsExist(key string) (bool, error) {
	b, err := redis.Bool(rp.do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return b, nil
}

// ClearAll 清空所有数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (rp *RedisPool) ClearAll() error {
	keys, err := redis.Strings(rp.do("KEYS", "*"))
	if err != nil {
		return err
	}

	for _, str := range keys {
		if _, err = rp.do("DEL", str); err != nil {
			return err
		}
	}

	return nil
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

	v, err := rp.do("HSET", key, field, data)
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		_, err = rp.do("EXPIRE", key, expire)
		if err != nil {
			rp.do("DEL", key)
		}
	}

	return v.(int64), err
}

// HGet 查询哈希表数据
//   参数
//     key:   哈希表key值
//     field: 哈希表field值
//     val:   保存结果地址
//   返回
//     错误信息，是否存在
func (rp *RedisPool) HGet(key string, field string, val interface{}) (error, bool) {
	v, err := rp.do("HGET", key, field)
	if err != nil {
		return err, false
	}

	if v == nil {
		return nil, false
	}

	// 类型转换
	err = cache.ByteToInter(v.([]byte), val)
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
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 1; i < len(fields)+1; i++ {
		args[i] = fields[i-1]
	}
	_, err := rp.do("HDEL", args...)

	return err
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
	res := make(map[string]interface{})

	v, err := rp.do("HGETALL", key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return res, nil
	}

	vInter := v.([]interface{})
	n := len(vInter)
	for i := 0; i < n; i += 2 {
		key := string(vInter[i].([]byte))
		res[key] = vInter[i+1]
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
	res := make(map[string]interface{})

	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 1; i < len(fields)+1; i++ {
		args[i] = fields[i-1]
	}

	v, err := rp.do("HMGET", args...)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return res, nil
	}

	vInter := v.([]interface{})
	i := 0
	for _, field := range fields {
		res[field] = vInter[i]
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
	v, err := rp.do("HVALS", key)
	if err != nil {
		return nil, err
	}
	return v.([]interface{}), nil
}

// ZSet 添加有序集合
//   参数
//     key:    有序集合key值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//     val:    有序集合值，数据为成对出来，前面为score(整数值或双精度浮点数), 后面为变量
//   返回
//     成功添加的数据个数和错误码
func (rp *RedisPool) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	var args []interface{}
	args = append(args, key)
	args = append(args, val...)

	n, err := rp.do("ZADD", args...)
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		_, err = rp.do("EXPIRE", key, expire)
		if err != nil {
			rp.do("DEL", key)
		}
	}

	return n.(int64), err
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
	var v interface{}
	var res []string

	if isRev {
		if withScores {
			v, err = rp.do("ZREVRANGE", key, start, stop, "WITHSCORES")
		} else {
			v, err = rp.do("ZREVRANGE", key, start, stop)
		}
	} else {
		if withScores {
			v, err = rp.do("ZRANGE", key, start, stop, "WITHSCORES")
		} else {
			v, err = rp.do("ZRANGE", key, start, stop)
		}
	}

	if err != nil {
		return nil, err
	}

	for _, val := range v.([]interface{}) {
		res = append(res, string(val.([]byte)))
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
	args = append(args, key)
	for _, f := range field {
		args = append(args, f)
	}

	n, err := rp.do("ZREM", args...)
	if err != nil {
		return -1, err
	}

	return n.(int64), err
}

// ZCard 返回有序集 key 的基数
//   参数
//     key: 有序集合key值
//   返回
//     有序集 key 的基数和错误码
func (rp *RedisPool) ZCard(key string) (int64, error) {
	n, err := rp.do("ZCARD", key)
	if err != nil {
		return -1, err
	}

	return n.(int64), err
}

// Pipeline 执行pipeline命令
//   参数
//     cmds: 要执行的命令
//   返回
//     成功返回命令执行的结果，失败返回错误信息
func (rp *RedisPool) Pipeline(cmds ...cache.Cmd) ([]cache.PipeRes) {
	c := rp.connPool.Get()
	defer c.Close()

	n := 0
	for _, cmd := range cmds {
		err := c.Send(cmd.Name, cmd.Args...)
		if err != nil {
			break
		}
		n++
	}

	ret := []cache.PipeRes{}
	err := c.Flush()
	if err != nil {
		return ret
	}

	for i := 0; i < n; i++ {
		reply, err := c.Receive()
		cmdRes := cache.PipeRes{}
		if err != nil {
			cmdRes.CmdErr = err
		} else {
			cmdRes.CmdRes = reply
		}
		ret = append(ret, cmdRes)
	}
	return ret
}

// Exec 执行pipeline事务命令
//   参数
//     cmds: 要执行的命令
//   返回
//     成功返回命令执行的结果，失败返回错误信息
func (rp *RedisPool) Exec(cmds ...cache.Cmd) (interface{}, error) {
	c := rp.connPool.Get()
	defer c.Close()

	err := c.Send("MULTI")
	if err != nil {
		return nil, err
	}

	for _, cmd := range cmds {
		err = c.Send(cmd.Name, cmd.Args...)
		if err != nil {
			return nil, err
		}
	}

	return c.Do("EXEC")
}
