// Cache 入口
//   变更历史
//     2017-02-20  lixiaoya  新建
package cache

import (
	json "github.com/json-iterator/go"
	"errors"
	"fmt"
	"legitlab.letv.cn/uc_tp/goweb/utils"
)

const (
	ENCODE_FLAG = "ENC1_"
	ENCODE_LEN  = 5
)

// Cache 所有缓存的接口
type Cache interface {
	Init(config string) error
	Set(key string, val interface{}, expire int32, encode ...bool) error
	Get(key string, val interface{}) (error, bool)
	Del(key string) error

	MSet(mList map[string]interface{}, expire int32, encode ...bool) error
	MGet(keys ...string) (map[string]interface{}, error)
	MDel(keys ...string) error

	Incr(key string, delta ...uint64) (int64, error)
	Decr(key string, delta ...uint64) (int64, error)
	IsExist(key string) (bool, error)
	ClearAll() error

	// 哈希表操作(支持redis)
	HSet(key string, field string, val interface{}, expire int32) (int64, error)
	HGet(key string, field string, val interface{}) (error, bool)
	HDel(key string, fields ...string) error
	HGetAll(key string) (map[string]interface{}, error)
	HMGet(key string, fields ...string) (map[string]interface{}, error)
	HVals(key string) ([]interface{}, error)

	// 有序集合操作(支持redis)
	ZSet(key string, expire int32, val ...interface{}) (int64, error)
	ZGet(key string, start, stop int, withScores bool, isRev bool) ([]string, error)
	ZDel(key string, field ...string) (int64, error)
	ZCard(key string) (int64, error)

	// pipeline支持
	Pipeline(cmds ...Cmd) ([]interface{}, error)
	Exec(cmds ...Cmd) (interface{}, error)
}

// IJson 生成与解析json串接口，如果参数实现了此接口，则生成与解析json串就使用参数的函数
type IJson interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// Cmd 保存命令，用于pipeline传递命令
type Cmd struct {
	Name string
	Args []interface{}
}

// NewCmd 实例化Cmd对象
func NewCmd(name string, args ...interface{}) Cmd {
	return Cmd{
		Name: name,
		Args: args,
	}
}

var Adapters = make(map[string]Cache)

// GetCache 获取一个Cache适配器
// 如果adapterName返回第一个找到的适配器，这个顺序是不确定的，一般只有一个适配器时不需要传值
//   参数
//     adapterName: 适配器名称
//   返回
//     成功时返回Cache对象，失败返回错误信息
func GetCache(adapterName ...string) (Cache, error) {
	if len(adapterName) == 0 {
		for _, adapter := range Adapters {
			return adapter, nil
		}
		return nil, errors.New("Cache: Adapter is empty")
	}

	name := adapterName[0]
	adapter, ok := Adapters[name]
	if !ok {
		return nil, fmt.Errorf("Cache: unknown adapter name %q", name)
	}

	return adapter, nil
}

// Encode 加密数据
//   参数
//     data: 要加密的数据
//     key:  加密密钥
//   返回
//     成功返回加密串，失败返回错误信息
func Encode(data, key []byte) ([]byte, error) {
	encodeText, err := utils.AesEncode(data, key)
	if err != nil {
		return nil, fmt.Errorf("Cache: Encode error, %s", err.Error())
	}
	encode := []byte(ENCODE_FLAG)
	encode = append(encode, encodeText...)
	return encode, nil
}

// Decode 解密数据
//   参数
//     data: 要解密的数据
//     key:  解密密钥
//   返回
//     成功返回加密串，失败返回错误信息
func Decode(data, key []byte) ([]byte, error) {
	if len(data) < ENCODE_LEN {
		return data, nil
	}

	encodeFlag := data[:ENCODE_LEN]
	if string(encodeFlag) == ENCODE_FLAG {
		decode, err := utils.AesDecode(data[ENCODE_LEN:], []byte(key))
		if err != nil {
			return nil, fmt.Errorf("Cache: Decode error, %s", err.Error())
		}
		return decode, nil
	}
	return data, nil
}

// InterToByte 将interface{}类型转成[]byte
//   参数
//     src: 要转换的数据
//   返回
//     转换后的数据，错误信息
func InterToByte(src interface{}) ([]byte, error) {
	var data []byte
	var err error
	if str, ok := src.(string); ok {
		data = []byte(str)
	} else if str, ok := src.(*string); ok {
		data = []byte(*str)
	} else if inter, ok := src.(IJson); ok {
		data, err = inter.MarshalJSON()
		if err != nil {
			return nil, err
		}
	} else {
		data, err = json.Marshal(src)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// ByteToInter 将[]byte类型转成interface{}
//   参数
//     src: 要转换的数据
//     dst: 转换后的数据
//   返回
//     转换后的数据，错误信息
func ByteToInter(src []byte, dst interface{}) error {
	if str, ok := dst.(*string); ok {
		*str = string(src)
		return nil
	} else if inter, ok := dst.(IJson); ok {
		return inter.UnmarshalJSON(src)
	}

	return json.Unmarshal(src, dst)
}
