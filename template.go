// 模板相关
//   变更历史
//     2017-02-08  lixiaoya  新建
package bingo

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"github.com/bingo/utils"
	"os"
	"path/filepath"
	"strings"
	"net/url"
)

var (
	tplFuncMap = make(template.FuncMap)
)

// Template
type Template struct {
	viewsDir  string
	viewsExt  string
	viewFiles []string
	ViewTemp  *template.Template
}

// NewTemplate 实例化Template
//   参数
//     viewsDir: 模板文件的目录
//     viewsExt: 模板文件的扩展名
//   返回
//     Template对象
func NewTemplate(viewsDir, viewsExt string) *Template {
	return &Template{
		viewsDir: viewsDir,
		viewsExt: viewsExt,
	}
}

// buildViews 编译viewsDir下的所有模板文件
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func (t *Template) buildViews() error {
	if t.viewsDir == "" || t.viewsExt == "" {
		return fmt.Errorf("Template: Views directory or extension is empty")
	}

	err := filepath.Walk(t.viewsDir, func(path string, info os.FileInfo, err error) error {
		t.find(path, info, err)
		return nil
	})

	if err != nil {
		return err
	}

	return t.parseFiles()
}

// find 查找模板文件
//   参数
//     path: 模板文件路径
//     f:    模板文件信息
//     err:  错误信息
//   返回
//     成功返回nil，失败返回错误信息
func (t *Template) find(path string, f os.FileInfo, err error) error {
	isFile, err := utils.IsFile(path)
	if err != nil || !isFile {
		return fmt.Errorf("Template: path [%s] is not file", path)
	}

	if filepath.Ext(path) != t.viewsExt {
		return fmt.Errorf("Template: path [%s] extension is error", path)
	}

	t.viewFiles = append(t.viewFiles, path)

	return nil
}

// ParseFiles 解析所有模板文件
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func (t *Template) parseFiles() error {
	for _, filename := range t.viewFiles {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		s := string(b)
		var tmpl *template.Template
		name := strings.Replace(filename, t.viewsDir, "", 1)[1:]
		if t.ViewTemp == nil {
			t.ViewTemp = template.New(name)

			// 添加自定义函数
			t.ViewTemp.Funcs(tplFuncMap)
		}
		if name == t.ViewTemp.Name() {
			tmpl = t.ViewTemp
		} else {
			tmpl = t.ViewTemp.New(name)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// init 初始化函数
//   参数
//     void
//   返回
//     void
func init() {
	tplFuncMap["showtime"] = ShowTime
	tplFuncMap["html2str"] = Html2Str
	tplFuncMap["str2html"] = Str2Html
	tplFuncMap["urlencode"] = url.QueryEscape
	tplFuncMap["urldecode"] = url.QueryUnescape
	tplFuncMap["htmlquote"] = HtmlQuote
	tplFuncMap["htmlunquote"] = HtmlUnquote
	tplFuncMap["trim"] = strings.Trim
	tplFuncMap["trimleft"] = strings.TrimLeft
	tplFuncMap["trimright"] = strings.TrimRight
	tplFuncMap["substr"] = Substr
	tplFuncMap["sum"] = Sum
	tplFuncMap["lang"] = Lang
}
