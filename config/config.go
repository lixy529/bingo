// 解析配置文件
//   变更历史
//     2017-09-18  lixiaoya  新建
package config

import (
	"strings"
	"strconv"
)

// 解析配置文件接口
type ConfigInter interface {
	ParseFile(config string) error
	GetString(section, key string, def ...string) string
	GetBool(section, key string, def ...bool) bool
	GetInt(section, key string, def ...int) int
	GetInt32(section, key string, def ...int32) int32
	GetInt64(section, key string, def ...int64) int64
	GetFloat64(section, key string, def ...float64) float64
	SetValue(section, key, value string)
	GetSec(section string) (map[string]string, bool)
	GetSecs() []string
}

// GetConfig 获取配置对象
//   请求
//     cfgFile: 配置文件
//     cfgType: 配置文件类型，取值范围: json-json文件 ini-ini文件，默认为ini
//   返回
//     配置对象、错误信息
func GetConfig(cfgFile, cfgType string) (ConfigInter, error) {
	var oConfig ConfigInter
	if cfgType == "json" {
		oConfig = NewStIni() // todo 添加json
	} else {
		oConfig = NewStIni()
	}

	err := oConfig.ParseFile(cfgFile)
	return oConfig, err
}


// 解析配置文件基类
type StConfig struct {
	filePath    string                       // 配置文件路径
	includeFile []string                     // 包含的文件
	configList  map[string]map[string]string // 配置文件内容
}

// ParseFile 解析配置文件
//   请求
//     cfgFile: 配置文件
//   返回
//     错误信息
func (c *StConfig) ParseFile(cfgFile string) error {
	return nil
}

// getValue 根据key值获取对应的value值
//   参数
//     section: 段名
//     key:     key值
//   返回
//     key存在时返回Value值，不存在返回空串
func (c *StConfig) getValue(section, key string) (string, bool) {
	if mapSec, ok := c.configList[strings.ToUpper(section)]; ok {
		if val, ok := mapSec[strings.ToUpper(key)]; ok {
			return val, true
		}
	}

	return "", false
}

// GetString 根据key值获取对应的value值，返回结果为string型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetString(section, key string, def ...string) string {
	if val, ok := c.getValue(section, key); ok {
		return val
	} else {
		def = append(def, "")
		return def[0]
	}

	return ""
}

// GetBool 根据key值获取对应的value值，返回结果为bool型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetBool(section, key string, def ...bool) bool {
	if val, ok := c.getValue(section, key); ok {
		switch strings.ToUpper(val) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	} else {
		def = append(def, false)
		return def[0]
	}
}

// GetInt 根据key值获取对应的value值，返回结果为int型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetInt(section, key string, def ...int) int {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.Atoi(val); err == nil {
			return val
		} else {
			def = append(def, 0)
			return def[0]
		}
	} else {
		def = append(def, 0)
		return def[0]
	}
}

// GetInt32 根据key值获取对应的value值，返回结果为int32型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetInt32(section, key string, def ...int32) int32 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseInt(val, 10, 64); err == nil {
			return int32(val)
		} else {
			def = append(def, 0)
			return def[0]
		}
	} else {
		def = append(def, 0)
		return def[0]
	}
}

// GetInt64 根据key值获取对应的value值，返回结果为int64型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetInt64(section, key string, def ...int64) int64 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseInt(val, 10, 64); err == nil {
			return val
		} else {
			def = append(def, 0)
			return def[0]
		}
	} else {
		def = append(def, 0)
		return def[0]
	}
}

// GetFloat64 根据key值获取对应的value值，返回结果为float64型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func (c *StConfig) GetFloat64(section, key string, def ...float64) float64 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseFloat(val, 64); err == nil {
			return val
		} else {
			def = append(def, 0.00)
			return def[0]
		}
	} else {
		def = append(def, 0.00)
		return def[0]
	}
}

// SetValue 设置一个key值
// 如果key值不存在，则新建一个
// 如果key值存在，则更新
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     void
func (c *StConfig) SetValue(section, key, value string) {
	section = strings.ToUpper(section)
	key = strings.ToUpper(key)
	_, ok := c.configList[section]
	if !ok {
		c.configList[section] = make(map[string]string)
		c.configList[section][key] = value
		return
	}

	c.configList[section][key] = value
}

// GetSec 根据section值获取段下所有的配置
//   参数
//     section: 段名
//   返回
//     段下对应的所有Value值
func (c *StConfig) GetSec(section string) (map[string]string, bool) {
	mapSec, ok := c.configList[strings.ToUpper(section)]
	if ok {
		return mapSec, true
	}

	return mapSec, false
}

// GetSecs 获取所有段名
//   参数
//     void
//   返回
//     所有段列表
func (c *StConfig) GetSecs() []string {
	var secs []string
	for sec, _ := range c.configList {
		secs = append(secs, sec)
	}

	return secs
}
