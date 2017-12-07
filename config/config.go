// 解析配置文件
//   变更历史
//     2017-09-18  lixiaoya  新建
package config

import (
	"strings"
	"strconv"
	"bufio"
	"path"
	"os"
	"io"
)

const (
	Line_Char = '\\'
	Line_Str  = "\\"
	Inc_Str   = "include "
)

// GetConfig 获取配置对象
//   参数
//     cfgFile: 配置文件
//   返回
//     配置对象、错误信息
func NewConfig(cfgFile string) (*StConfig, error) {
	oConfig := &StConfig{
		filePath: cfgFile,
		configList: make(map[string]map[string]string),
	}
	err := oConfig.ParseFile(cfgFile)
	return oConfig, err
}

// ParseFile 解析配置文件
//   参数
//     cfgFile: 配置文件
//   返回
//     错误信息
func (c *StConfig) ParseFile(cfgFile string) error {
	c.filePath = cfgFile
	c.configList = make(map[string]map[string]string)

	err := c.parseOne(c.filePath)
	if err != nil {
		return err
	}

	// 包含配置文件的处理
	for _, file := range c.includeFile {
		if path.IsAbs(file) {
			err = c.parseOne(file)
			if err != nil {
				return err
			}
		}

		p, _ := path.Split(c.filePath)
		f := path.Join(p, file)
		err = c.parseOne(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseOne 读取一个配置文件
//   参数
//     filePath: 配置文件路径
//   返回
//     成功返回nil，失败返回错误信息
func (c *StConfig) parseOne(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	var section string
	buf := bufio.NewReader(file)
	var realLine string

	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		n := len(line)
		if err != nil {
			if err != io.EOF {
				return err
			}

			if n == 0 {
				break
			}
		}

		// 处理换行
		if n > 0 && line[n-1] == Line_Char {
			realLine += strings.TrimSpace(strings.TrimRight(line, Line_Str))
			continue
		} else if len(realLine) > 0 {
			realLine += strings.TrimSpace(line)
		} else {
			realLine = line
		}

		n = len(realLine)
		switch {
		case n == 0:
		case string(realLine[0]) == "#": // 配置文件备注
			realLine = ""
		case realLine[0] == '[' && realLine[len(realLine)-1] == ']':
			section = strings.ToUpper(strings.TrimSpace(realLine[1 : len(realLine)-1]))
			c.configList[section] = make(map[string]string)
			realLine = ""
		case n > 8 && realLine[0:8] == Inc_Str: // 包含文件
			f := realLine[8:]
			c.includeFile = append(c.includeFile, f)
			realLine = ""

		default:
			tmpLine := realLine
			if i := strings.IndexAny(realLine, "#"); i > 0 {
				// 存在备注
				tmpLine = realLine[0:i]
			}
			realLine = ""
			i := strings.IndexAny(tmpLine, "=")
			if i < 1 {
				continue
			}
			key := strings.ToUpper(strings.TrimSpace(tmpLine[0:i]))
			value := strings.TrimSpace(tmpLine[i+1 : len(tmpLine)])
			c.configList[section][key] = value
		}
	}

	return nil
}

// 解析配置文件基类
type StConfig struct {
	filePath    string                       // 配置文件路径
	includeFile []string                     // 包含的文件
	configList  map[string]map[string]string // 配置文件内容
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
