// memcache adapter
//   变更历史
//     2017-02-20  lixiaoya  新建
package memcache

import (
	json "github.com/json-iterator/go"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/lixy529/bingo/cache"
	"github.com/lixy529/bingo/utils"
	"strconv"
	"strings"
	"time"
)

const (
	COMPRESSION_ZLIB        = "zlib"
	FLAGES_INT_UNCOMPRESS   = 1
	FLAGES_FLOAT_UNCOMPRESS = 2
	FLAGES_JSON_UNCOMPRESS  = 6
	FLAGES_JSON_COMPRESS    = 54
	FLAGES_STR_UNCOMPRESS   = 0
	FLAGES_STR_COMPRESS     = 48

	NOT_EXIST = "cache miss"
)

// MemcCache memcache缓存
type MemcCache struct {
	conn      *memcache.Client
	connCfg   []string
	maxIdle   int           // 最大空闲连接数，默认为2，如果配置值小于1则使用默认值
	ioTimeOut time.Duration // io超时时间，默认为100毫秒，传0为默认时间，单位毫秒
	prefix    string        // key前缀，如果配置里有，则所有key前自动添加此前缀

	serializer        string // 序列化，目前只支持json
	compressType      string // 压缩类型，目前只支持zlib
	compressThreshold int    // 超过大小就进行压缩

	encodeKey []byte // 加解密密钥，使用Aes加密，长度为16的倍数
}

// NewMemcCache 新建一个MemcCache适配器.
func NewMemcCache() cache.Cache {
	return &MemcCache{}
}

// connect 连接memcache
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func (mc *MemcCache) connect() error {
	if mc.conn != nil {
		return nil
	}

	mc.conn = memcache.New(mc.connCfg...)
	if mc.conn == nil {
		return fmt.Errorf("MemcCache Connect memcache [%s] failed", strings.Join(mc.connCfg, ","))
	}

	if mc.maxIdle > 0 {
		mc.conn.MaxIdleConns = mc.maxIdle
	}

	if mc.ioTimeOut >= 0 {
		mc.conn.Timeout = mc.ioTimeOut * time.Millisecond
	}

	return nil
}

// Init 初始化
//   参数
//     config: 配置josn串，如:
//       {
//       "addr":"127.0.0.1:11211,127.0.0.2:11211",
//       "maxIdle":"3",
//       "ioTimeOut":"1",
//       "prefix":"le_",
//       "serializer":"json",
//       "compressType":"zlib",
//       "compressThreshold":"256",
//       "encodeKey":"abcdefghij123456",
//       }
//   返回
//     成功返回nil，失败返回错误信息
func (mc *MemcCache) Init(config string) error {
	var mapCfg map[string]string
	var ok bool
	err := json.Unmarshal([]byte(config), &mapCfg)
	if err != nil {
		return fmt.Errorf("MemcCache: Unmarshal json[%s] error, %s", config, err.Error())
	}

	if _, ok = mapCfg["addr"]; !ok {
		return errors.New("MemcCache: Config hasn't address.")
	}

	mc.connCfg = strings.Split(mapCfg["addr"], ",")
	if _, ok = mapCfg["maxIdle"]; ok {
		mc.maxIdle, _ = strconv.Atoi(mapCfg["maxIdle"])
	} else {
		mc.maxIdle = -1
	}
	if _, ok = mapCfg["ioTimeOut"]; ok {
		ioTimeOut, _ := strconv.Atoi(mapCfg["ioTimeOut"])
		mc.ioTimeOut = time.Duration(ioTimeOut)
	} else if _, ok = mapCfg["idelTimeOut"]; ok {
		// 之前理解错误，配置做一下兼容
		ioTimeOut, _ := strconv.Atoi(mapCfg["idelTimeOut"])
		mc.ioTimeOut = time.Duration(ioTimeOut)
	} else {
		mc.ioTimeOut = -1
	}
	if prefix, ok := mapCfg["prefix"]; ok {
		mc.prefix = prefix
	}

	// 序列化，目前只支持json
	mc.serializer = "json"

	// 压缩，压缩类型目前只支持zlib
	mc.compressType, _ = mapCfg["compressType"]
	if mc.compressType != "" {
		if mc.compressType != COMPRESSION_ZLIB {
			return fmt.Errorf("MemcCache: Compress type don't support %s", mc.compressType)
		}

		if _, ok = mapCfg["compressThreshold"]; ok {
			var err error
			mc.compressThreshold, err = strconv.Atoi(mapCfg["compressThreshold"])
			if err != nil {
				return fmt.Errorf("MemcCache: Compress threshold error, %s", err.Error())
			}
		}

	}

	// 加密密钥
	if tmp, ok := mapCfg["encodeKey"]; ok {
		mc.encodeKey = []byte(tmp)
	}

	err = mc.connect()
	if err != nil {
		return err
	}

	return nil
}

// Set 向缓存设置一个值
//   参数
//     key:    key值
//     val:    value值
//     expire: 到期是缓存过期时间，以秒为单位，从现在开始的相对时间，“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (mc *MemcCache) Set(key string, val interface{}, expire int32, encode ...bool) error {
	if err := mc.connect(); err != nil {
		return err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}

	// 超过30天使用Unix纪元时间
	if expire > 86400*30 {
		expire = int32(time.Now().Unix()) + expire
	}
	item := memcache.Item{Key: key, Expiration: expire}

	// 类型转换
	data, err := cache.InterToByte(val)
	if err != nil {
		return err
	}

	// 加密判断
	encode = append(encode, false)
	if encode[0] {
		data, err = cache.Encode(data, mc.encodeKey)
		if err != nil {
			return err
		}
	}

	// flags设置
	flags := FLAGES_JSON_UNCOMPRESS
	valType := ""
	if _, ok := val.(string); ok {
		valType = "string"
		flags = FLAGES_STR_UNCOMPRESS
	} else if _, ok := val.(float64); ok {
		valType = "float"
		flags = FLAGES_FLOAT_UNCOMPRESS
	} else if _, ok := val.(int64); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	} else if _, ok := val.(int32); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	} else if _, ok := val.(int); ok {
		valType = "int"
		flags = FLAGES_INT_UNCOMPRESS
	}

	// 目前只支持zlib压缩
	if mc.compressType == COMPRESSION_ZLIB {
		dataLen := len(data)
		if dataLen > mc.compressThreshold {
			if valType == "string" {
				flags = FLAGES_STR_COMPRESS
			} else {
				flags = FLAGES_JSON_COMPRESS
			}
			data, err = utils.ZlibEncode(data)
			if err != nil {
				return err
			}
			data = []byte(string(utils.Int32ToByte(int32(dataLen), false)) + string(data))
		}
	}

	item.Flags = uint32(flags)
	item.Value = data
	return mc.conn.Set(&item)
}

// Get 从缓存取一个值
//   参数
//     key:    key值
//   返回
//     错误信息，是否存在
func (mc *MemcCache) Get(key string, val interface{}) (error, bool) {
	if err := mc.connect(); err != nil {
		return err, false
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	item, err := mc.conn.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), NOT_EXIST) {
			return nil, false
		}
		return err, false
	}

	// 解压
	data := item.Value
	if item.Flags == FLAGES_JSON_COMPRESS || item.Flags == FLAGES_STR_COMPRESS {
		data, err = utils.ZlibDecode(data[4:])
		if err != nil {
			return err, true
		}
	}

	// 解密判断
	data, err = cache.Decode(data, mc.encodeKey)
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
func (mc *MemcCache) Del(key string) error {
	if err := mc.connect(); err != nil {
		return err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}

	err := mc.conn.Delete(key)
	if err == nil || strings.Contains(err.Error(), NOT_EXIST) {
		return nil
	}

	return err
}

// MSet 同时设置一个或多个key-value对
//   参数
//     mList:  key-value对
//     expire: 到期是缓存过期时间，以秒为单位：从现在开始的相对时间。“0”表示项目没有到期时间。
//     encode: 是否加密标识
//   返回
//     成功时返回nil，失败返回错误信息
func (mc *MemcCache) MSet(mList map[string]interface{}, expire int32, encode ...bool) error {
	for key, val := range mList {
		err := mc.Set(key, val, expire, encode...)
		if err != nil {
			return err
		}
	}
	return nil
}

// MSet 同时获取一个或多个key的value
//   参数
//     keys:  要查询的key值
//   返回
//     查询结果
func (mc *MemcCache) MGet(keys ...string) (map[string]interface{}, error) {
	mList := make(map[string]interface{})

	if err := mc.connect(); err != nil {
		return mList, err
	}

	if mc.prefix != "" {
		for k, v := range keys {
			keys[k] = mc.prefix + v
		}
	}

	mv, err := mc.conn.GetMulti(keys)
	if err != nil {
		return mList, err
	}

	for key, val := range mv {
		// 解压
		data := val.Value
		if val.Flags == FLAGES_JSON_COMPRESS || val.Flags == FLAGES_STR_COMPRESS {
			data, err = utils.ZlibDecode(data[4:])
			if err != nil {
				mList[key] = nil
				continue
			}
		}

		// 解密判断
		data, err = cache.Decode(data, mc.encodeKey)
		if err != nil {
			return mList, err
		}

		if mc.prefix != "" {
			key = key[len(mc.prefix):]
		}
		mList[key] = data
	}

	return mList, nil
}

// MDel 同时删除一个或多个key
//   参数
//     keys:  要查询的key值
//   返回
//     成功时返回nil，失败返回错误信息
func (mc *MemcCache) MDel(keys ...string) error {
	for _, key := range keys {
		err := mc.Del(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Incr 缓存里的值自增
// key不存在时返回NOT_FOUND错误
//   参数
//     key:   递增的key值
//     delta: 递增的量
//   返回
//     递增后的结果，失败返回错误信息
func (mc *MemcCache) Incr(key string, delta ...uint64) (int64, error) {
	if err := mc.connect(); err != nil {
		return 0, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	delta = append(delta, 1)
	v, err := mc.conn.Increment(key, delta[0])
	return int64(v), err
}

// Decr 缓存里的值自减
// key不存在时返回NOT_FOUND错误
// key对应的值为0时，返回还是0
//   参数
//     key:   递减的key值
//     delta: 递减的量
//   返回
//     递减后的结果，失败返回错误信息
func (mc *MemcCache) Decr(key string, delta ...uint64) (int64, error) {
	if err := mc.connect(); err != nil {
		return 0, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	delta = append(delta, 1)
	v, err := mc.conn.Decrement(key, delta[0])
	return int64(v), err
}

// IsExist 判断key值是否存在
//   参数
//     key:  要查询的key值
//   返回
//     存在返回true，不存在返回false
func (mc *MemcCache) IsExist(key string) (bool, error) {
	if err := mc.connect(); err != nil {
		return false, err
	}

	if mc.prefix != "" {
		key = mc.prefix + key
	}
	_, err := mc.conn.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), NOT_EXIST) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ClearAll 清空所有数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (mc *MemcCache) ClearAll() error {
	if err := mc.connect(); err != nil {
		return err
	}

	return mc.conn.FlushAll()
}

// Hset 添加哈希表，memcache没有哈希表
func (mc *MemcCache) HSet(key string, field string, val interface{}, expire int32) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HSet")
}

// HGet 查询哈希表数据，memcache没有哈希表
func (mc *MemcCache) HGet(key string, field string, val interface{}) (error, bool) {
	return errors.New("MemcCache: Memcache don't support HGet"), false
}

// HDel 删除哈希表数据，memcache没有哈希表
func (mc *MemcCache) HDel(key string, fields ...string) error {
	return errors.New("MemcCache: Memcache don't support HDel")
}

// HGetAll 返回哈希表 key 中，所有的域和值，memcache没有哈希表
func (mc *MemcCache) HGetAll(key string) (map[string]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HGetAll")
}

// HMGet 返回哈希表 key 中，一个或多个给定域的值，memcache没有哈希表
func (mc *MemcCache) HMGet(key string, fields ...string) (map[string]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HMGet")
}

// HVals 删除哈希表数据，memcache没有哈希表
func (mc *MemcCache) HVals(key string) ([]interface{}, error) {
	return nil, errors.New("MemcCache: Memcache don't support HVals")
}

// HIncr 哈希表的值自增，memcache没有哈希表
func (mc *MemcCache) HIncr(key, fields string, delta ...uint64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HIncr")
}

// HDecr 哈希表的值自减，memcache没有哈希表
func (mc *MemcCache) HDecr(key, fields string, delta ...uint64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support HDecr")
}

// ZSet 添加有序集合，memcache没有有序集合
func (mc *MemcCache) ZSet(key string, expire int32, val ...interface{}) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZSet")
}

// ZGet 查询有序集合，memcache没有有序集合
func (mc *MemcCache) ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error) {
	return nil, errors.New("MemcCache: Memcache don't support ZGet")
}

// ZDel 删除有序集合数据，memcache没有有序集合
func (rc *MemcCache) ZDel(key string, field ...string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZDel")
}

// ZRemRangeByRank 删除有序集合数据，memcache没有有序集合
func (rc *MemcCache) ZRemRangeByRank(key string, start, end int64) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByRank")
}

// ZRemRangeByScore 删除有序集合数据，memcache没有有序集合
func (rc *MemcCache) ZRemRangeByScore(key string, start, end string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByScore")
}

// ZRemRangeByLex 删除有序集合数据，memcache没有有序集合
func (rc *MemcCache) ZRemRangeByLex(key string, start, end string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZRemRangeByLex")
}

// ZCard 返回有序集 key 的基数，memcache没有有序集合
func (mc *MemcCache) ZCard(key string) (int64, error) {
	return 0, errors.New("MemcCache: Memcache don't support ZCard")
}

// Pipeline 执行pipeline命令，memcache不支持pipeline
func (mc *MemcCache) Pipeline(isTx bool) cache.Pipeliner {
	return cache.Pipeliner{}
}
