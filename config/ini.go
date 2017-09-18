// 解析ini型配置文件
//   变更历史
//     2017-09-18  lixiaoya  新建
package config

import (
	"strings"
	"os"
	"io"
	"bufio"
	"path"
)

const (
	Line_Char = '\\'
	Line_Str  = "\\"
	Inc_Str   = "include "
)

type StIni struct {
	StConfig
}

func NewStIni() *StIni {
	obj := StIni{}

	return &obj
}

// ParseFile 解析配置文件
//   请求
//     cfgFile: 配置文件
//   返回
//     错误信息
func (c *StIni) ParseFile(cfgFile string) error {
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
func (c *StIni) parseOne(filePath string) error {
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
