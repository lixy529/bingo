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

// Lang language package struct.
type Lang struct {
	langPath string                            // Root directory of language package files.
	data     map[string]map[string]interface{} // Language package content.
}

// NewLang return Lang object.
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

// loadLang load language package files.
// The subdirectory is key, eg:zh-cn,en-us.
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

// parseFile parse language package files.
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

// ReadLang read language package content.
// Eg: ReadLang("zh-cn", "title")
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

// String read language package content, return a string value.
// If multi-level data is taken, you can use ":" to merge key values.
// eg: String("en-us", "aa"), String("en-us", "aa:bb:cc")
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

// Map read language package content, return a map value.
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
