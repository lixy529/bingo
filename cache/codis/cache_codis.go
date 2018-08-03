// codis adapter
// 用于分布式的redis操作
//   变更历史
//     2017-02-21  lixiaoya  新建
package codis

import (
	"github.com/lixy529/bingo/utils"
	"github.com/lixy529/bingo/cache"
	"github.com/gomodule/redigo/redis"
	json "github.com/json-iterator/go"
	"errors"
	"fmt"
	"strconv"
	"time"
	"strings"
)

// CodisCache Codis缓存
type CodisCache struct {
	connPool    []*redis.Pool // 每个主机有一个连接池
	conns       string        // 连接主机和端口，多个主机用逗号分割，如127.0.0.1:1900,127.0.0.2:1900
	auth        string        // 授权密码
	dbNum       int           // db编号，默认为0
	maxIdle     int           // 最大空闲数，默认为3
	maxActive   int           // 最大的激活连接数，默认为0，0不限制
	idleTimeOut time.Duration // 空闲超时时间，默认为180秒，0关闭

	encodeKey []byte // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewCodisCache 新建一个CodisCache适配器.
func NewCodisCache() cache.Cache {
	return &CodisCache{}
}

// Init 初始化
//   参数
//     config: 配置josn串
//       {
//         "conns":"127.0.0.1:19100,127.0.0.2:19100,127.0.0.3:19100",
//         "auth":"xxxx",
//         "dbNum":"1",
//         "maxIdle":"3",
//         "maxActive":"10",
//         "idleTimeOut":"180"
//         "encodeKey":"abcdefghij123456",
//       }
//       conns:       连接主机和端口，多个主机用逗号分割，如127.0.0.1:1900,127.0.0.2:1900
//       auth:        授权密码
//       dbNum:       db编号，默认为0
//       maxIdle:     最大空闲连接数，默认为3
//       maxActive:   最大连接数，0不限制，默认为0
//       idelTimeOut: 最大空闲时间，单位秒，0不限制，默认为0
//       encodeKey:   数据如果要加密，传的加密密钥
//   返回
//     成功时返回nil，失败返回错误信息
func (c *CodisCache) Init(config string) error {
	var mapCfg map[string]string
	var err error

	err = json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("CodisCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	// 连接串
	c.conns = mapCfg["conns"]

	// 授权
	c.auth = mapCfg["auth"]

	// 最大空闲连接数
	maxIdle, err := strconv.Atoi(mapCfg["maxIdle"])
	if err != nil || maxIdle < 0 {
		c.maxIdle = 3
	} else {
		c.maxIdle = maxIdle
	}

	// 最大激活连接数
	maxActive, err := strconv.Atoi(mapCfg["maxActive"])
	if err != nil || maxActive < 0 {
		c.maxActive = 0
	} else {
		c.maxActive = maxActive
	}

	// 空闲连接的超时时间
	idelTimeOut, err := strconv.Atoi(mapCfg["idleTimeOut"])
	if err != nil || idelTimeOut < 0 {
		c.idleTimeOut = 0
	} else {
		c.idleTimeOut = time.Duration(idelTimeOut)
	}

	// dbNum
	mDbNum, err := strconv.Atoi(mapCfg["dbNum"])
	if err != nil {
		c.dbNum = 0
	} else {
		c.dbNum = mDbNum
	}

	// 加密密钥
	if tmp, ok := mapCfg["encodeKey"]; ok && tmp != "" {
		c.encodeKey = []byte(tmp)
	}

	// 设置连接池
	for _, v := range strings.Split(c.conns, ",") {
		rp, err := c.connect(v)
		if err != nil {
			continue
		}
		c.connPool = append(c.connPool, rp)
	}

	return nil
}

// connect 建立连接
//   参数
//     host: 其中一台主机信息
//   返回
//     这台主机的连接池、错误信息
func (c *CodisCache) connect(host string) (*redis.Pool, error) {
	// 连接函数
	dialFunc := func() (conn redis.Conn, err error) {
		conn, err = redis.Dial("tcp", host)
		if err != nil {
			return nil, err
		}

		if c.auth != "" {
			if _, err := conn.Do("AUTH", c.auth); err != nil {
				conn.Close()
				return nil, err
			}
		}

		if c.dbNum > 0 {
			_, selErr := conn.Do("SELECT", c.dbNum)
			if selErr != nil {
				conn.Close()
				return nil, selErr
			}
		}


		return
	}

	connPool := &redis.Pool{
		MaxIdle:     c.maxIdle,
		MaxActive:   c.maxActive,
		IdleTimeout: c.idleTimeOut * time.Second,
		Dial:        dialFunc,
	}

	return connPool, nil
}

// GetConn 获取一个连接
// 外面调用完需要通过c.Close()放回连接池
//   参数
//     host: 其中一台主机信息
//   返回
//     这台主机的连接池、错误信息
func (c *CodisCache) getConn() redis.Conn {
	connCnt := len(c.connPool)
	if connCnt == 0 {
		return nil
	} else if connCnt == 1 {
		return c.connPool[0].Get()
	}

	k := utils.Irand(0, connCnt-1)
	return c.connPool[k].Get()
}

// Do 从连接池里取一个连接，调用redis命令
//   参数
//     commandName:  命令串
//     args:         参数
//   返回
//     成功时返回结果信息，失败时返回错误信息
func (c *CodisCache) do(commandName string, args ...interface{}) (interface{}, error) {
	conn := c.getConn()
	if conn == nil {
		return "", errors.New("connIsNil")
	}

	defer conn.Close()

	return conn.Do(commandName, args...)
}

// Set 向缓存设置一个值，访问主库
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (c *CodisCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
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

	if expire > 0 {
		_, err = c.do("SETEX", key, expire, data)
	} else {
		_, err = c.do("SET", key, data)
	}

	return err
}

// Get 从缓存取一个值，访问从库
//   参数
//     key: key值
//     val: 保存结果地址
//   返回
//     错误信息，是否存在
func (c *CodisCache) Get(key string, val interface{}) (error, bool) {
	v, err := c.do("GET", key)
	if err != nil {
		return err, false
	}

	if v == nil {
		return nil, false
	}

	// 解密判断
	data, err := cache.Decode(v.([]byte), c.encodeKey)
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

// Del 从缓存删除一个值，访问主库
//   参数
//     key: key值
//   返回
//     成功时返回nil，失败返回错误信息
func (c *CodisCache) Del(key string) error {
	_, err := c.do("DEL", key)

	return err
}

// MSet 同时设置一个或多个key-value对，访问主库
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功返回查询结果，失败返回错误信息，key不存在时对应的val为nil
func (c *CodisCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
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

		v = append(v, key, data)
	}

	_, err := c.do("MSET", v...)
	if err != nil {
		return err
	}

	// 设置失效时间
	if expire > 0 {
		for key := range mList {
			_, err = c.do("EXPIRE", key, expire)
			if err != nil {
				c.do("DEL", key)
			}
		}
	}

	return err
}

// MGet 同时获取一个或多个key的value，访问从库
//   参数
//     keys:  要查询的key值
//   返回
//     成功返回查询结果，失败返回错误信息
func (c *CodisCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})
	var args []interface{}
	for _, k := range keys {
		args = append(args, k)
	}

	v, err := c.do("MGET", args...)
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
			data, err := cache.Decode(val.([]byte), c.encodeKey)
			if err != nil {
				return mList, err
			}

			mList[keys[i]] = data
		}

		i++
	}

	return mList, nil
}

// MDel 同时删除一个或多个key，访问主库
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (c *CodisCache) MDel(keys ...string) error {
	for _, key := range keys {
		err := c.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Incr 缓存里的值自增，访问主库
// key不存在时会新建一个，再返回1
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (c *CodisCache) Incr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	v, err := c.do("INCRBY", key, delta[0])
	if err != nil {
		return 0, err
	} else {
		return v.(int64), nil
	}
}

// Decr 缓存里的值自减，访问主库
// key不存在时会新建一个，再返回-1
//   参数
//     key:   递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (c *CodisCache) Decr(key string, delta ...uint64) (int64, error) {
	delta = append(delta, 1)
	v, err := c.do("INCRBY", key, 0-int64(delta[0]))
	if err != nil {
		return 0, err
	} else {
		return v.(int64), nil
	}
}

// IsExist 判断key值是否存在，访问从库
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false
func (c *CodisCache) IsExist(key string) (bool, error) {
	b, err := redis.Bool(c.do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return b, nil
}

// ClearAll 清空所有数据，访问主库
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (c *CodisCache) ClearAll() error {
	keys, err := redis.Strings(c.do("KEYS", "*"))
	if err != nil {
		return err
	}

	for _, str := range keys {
		if _, err = c.do("DEL", str); err != nil {
			return err
		}
	}

	return nil
}

// Hset 添加哈希表，访问主库
//   参数
//     key:    哈希表key值
//     field:  哈希表field值
//     val:    哈希表value值
//     expire: 缓存过期时间，以秒为单位：从现在开始的相对时间，“0”表示项目没有到期时间
//   返回
//     成功时返回添加的个数，失败返回错误信息
func (c *CodisCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return -1, err
	}

	v, err := c.do("HSET", key, field, data)
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		_, err = c.do("EXPIRE", key, expire)
		if err != nil {
			c.do("DEL", key)
		}
	}

	return v.(int64), err
}

// HGet 查询哈希表数据，访问从库
//   参数
//     key:   哈希表key值
//     field: 哈希表field值
//     val:   保存结果地址
//   返回
//     错误信息，是否存在
func (c *CodisCache) HGet(key string, field string, val interface{}) (error, bool) {
	v, err := c.do("HGET", key, field)
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

// HDel 删除哈希表数据，访问主库
//   参数
//     key:   哈希表key值
//     fields: 哈希表field值
//   返回
//     成功返回nil，失败返回错误信息
func (c *CodisCache) HDel(key string, fields ...string) error {
	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 1; i < len(fields)+1; i++ {
		args[i] = fields[i-1]
	}
	_, err := c.do("HDEL", args...)

	return err
}

// HGetAll 返回哈希表 key 中，所有的域和值，struct、map类型需要业务层调用json.Unmarshal
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (c *CodisCache) HGetAll(key string) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	v, err := c.do("HGETALL", key)
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
func (c *CodisCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := 1; i < len(fields)+1; i++ {
		args[i] = fields[i-1]
	}

	v, err := c.do("HMGET", args...)
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

// HVals 返回哈希表 key 中，所有的域和值
//   参数
//     key: 有序集合key值
//   返回
//     查询的结果数据和错误码
func (c *CodisCache) HVals(key string) ([]interface{}, error) {
	v, err := c.do("HVALS", key)
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
//     成功添加的数据和错误码
func (c *CodisCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	var args []interface{}
	args = append(args, key)
	args = append(args, val...)

	n, err := c.do("ZADD", args...)
	if err != nil {
		return -1, err
	}

	if expire > 0 {
		_, err = c.do("EXPIRE", key, expire)
		if err != nil {
			c.do("DEL", key)
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
func (c *CodisCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	var err error
	var v interface{}
	var res []string

	if isRev {
		if withScores {
			v, err = c.do("ZREVRANGE", key, start, stop, "WITHSCORES")
		} else {
			v, err = c.do("ZREVRANGE", key, start, stop)
		}
	} else {
		if withScores {
			v, err = c.do("ZRANGE", key, start, stop, "WITHSCORES")
		} else {
			v, err = c.do("ZRANGE", key, start, stop)
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
func (c *CodisCache) ZDel(key string, field ...string) (int64, error) {
	var args []interface{}
	args = append(args, key)
	for _, f := range field {
		args = append(args, f)
	}

	n, err := c.do("ZREM", args...)
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
func (c *CodisCache) ZCard(key string) (int64, error) {
	n, err := c.do("ZCARD", key)
	if err != nil {
		return -1, err
	}

	return n.(int64), err
}

// Pipeline 执行pipeline命令
//   参数
//     cmds: 要执行的命令
//   返回
//     成功刊返回命令执行的结果，失败返回错误信息
func (c *CodisCache) Pipeline(cmds ...cache.Cmd) ([]cache.PipeRes) {
	conn := c.getConn()
	defer conn.Close()

	n := 0
	for _, cmd := range cmds {
		err := conn.Send(cmd.Name, cmd.Args...)
		if err != nil {
			break
		}
		n++
	}

	ret := []cache.PipeRes{}
	err := conn.Flush()
	if err != nil {
		return ret
	}

	for i := 0; i < n; i++ {
		reply, err := conn.Receive()
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
//     成功刊返回命令执行的结果，失败返回错误信息
func (c *CodisCache) Exec(cmds ...cache.Cmd) (interface{}, error) {
	conn := c.getConn()
	defer conn.Close()

	err := conn.Send("MULTI")
	if err != nil {
		return nil, err
	}

	for _, cmd := range cmds {
		err = conn.Send(cmd.Name, cmd.Args...)
		if err != nil {
			return nil, err
		}
	}

	return conn.Do("EXEC")
}
