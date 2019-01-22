// 语言包
//   变更历史
//     2017-03-20  lixiaoya  新建
package lang

import (
	"github.com/lixy529/gotools/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

var LANG_EXT = ".json"

// Lang 语言包结构体
type Lang struct {
	langPath string                            // 语言包文件的根目录
	data     map[string]map[string]interface{} // 语言包文件内容
}

// NewLang 实例化语言包对象
//   参数
//     root: 语言包文件的根目录
//   返回
//     成功返回语言包对象，失败返回错误信息
func NewLang(root string) (*Lang, error) {
	l := &Lang{
		langPath: root,
		data:     make(map[string]map[string]interface{}),
	}

	err := l.loadLang()
	if err != nil {
		return nil, err
	}

	return l, err
}

// loadLang 加载语言包文件，以子目录为key
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func (l *Lang) loadLang() error {
	if l.langPath == "" {
		return errors.New("Lang: Root path is empty")
	}

	ok, err := utils.IsDir(l.langPath)
	if err != nil {
		return fmt.Errorf("Lang: err [%s]", err.Error())
	} else if !ok {
		return errors.New("Lang: Root path is't directory")
	}

	// 遍历根目录
	fis, err := ioutil.ReadDir(l.langPath)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		fileName := fi.Name()
		if !strings.HasSuffix(fileName, LANG_EXT) {
			continue
		}

		err = l.parseFile(fileName)
		if err != nil {
			return fmt.Errorf("Lang: ParseFile failed, [%s]", err.Error())
		}
	}

	return err
}

// parseFile 解析语言包文件
//   参数
//     fileName: 语言名文件名
//   返回
//     成功返回nil，失败返回错误信息
func (l *Lang) parseFile(fileName string) error {
	n := len(fileName)
	lang := fileName[0:n-5]
	file := path.Join(l.langPath, fileName)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	var res map[string]interface{}
	dataJson := []byte(data)
	err = json.Unmarshal(dataJson, &res)
	if err != nil {
		return err
	}

	l.data[lang] = res

	return nil
}

// ReadLang 读取语言包数据
//   参数
//     lang: 语言，如zh-cn、en-us
//     key:  语言包key值
//   返回
//     对应语言包数据、key是否存在
func (l *Lang) ReadLang(lang, key string) (interface{}, bool) {
	r, ok := l.data[lang]
	if !ok {
		return "", false
	}

	val, ok := r[key]
	if !ok {
		return "", false
	}

	return val, true
}

// String 读取语言包数据，value为string型使用此函数
//   参数
//     lang: 语言，如zh-cn、en-us
//     key:  语言包key值，如果是取多层的数据，则可以用:拼接每次的key值，如aa:bb:cc,取的是aa下的bb下的cc对应的值
//   返回
//     对应语言包数据
func (l *Lang) String(lang, key string) string {
	if key == "" {
		return ""
	}
	val, exist := l.ReadLang(lang, key)
	if exist {
		if str, ok := val.(string); ok {
			return str
		}
	}

	keys := strings.Split(key, ":")
	kLen := len(keys)
	if kLen < 2 {
		return ""
	}
	val, exist = l.ReadLang(lang, keys[0])
	if !exist {
		return ""
	}

	var ok bool
	var t map[string]interface{}
	for i := 1; i < kLen; i++ {
		if t, ok = val.(map[string]interface{}); !ok {
			return ""
		}
		k := keys[i]

		if i == kLen-1 {
			v, ok := t[k]
			if !ok {
				return ""
			}

			if str, ok := v.(string); ok {
				return str
			}
		} else {
			val, ok = t[k]
			if !ok {
				return ""
			}
		}
	}

	return ""
}

// Map 读取语言包数据，value为map型使用此函数
//   参数
//     lang: 语言，如zh-cn、en-us
//     key:  语言包key值
//   返回
//     对应语言包数据
func (l *Lang) Map(lang, key string) map[string]string {
	val, exist := l.ReadLang(lang, key)
	if !exist || val == "" || val == nil {
		return nil
	}

	m := make(map[string]string)
	if mInter, ok := val.(map[string]interface{}); ok {
		if len(mInter) == 0 {
			return nil
		}

		for k, v := range mInter {
			m[k] = v.(string)
		}
		return m
	}

	return nil
}
