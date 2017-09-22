// 模板相关
//   变更历史
//     2017-03-20  lixiaoya  新建
package bingo

import (
	"github.com/bingo/utils"
	"html/template"
	"time"
	"html"
)

// ShowTime 显示时间
//   参数
//     t: 时间
//     format: 格式
//   返回
//     格式化后的时间串
func ShowTime(t time.Time, format string) string {
	return t.Format(format)
}

// Html2str Html字符替换
//   参数
//     html: 要处理的字符串
//   返回
//     替换后的数据
func Html2Str(html string) string {
	return utils.Html2Str(html)
}

// Str2html 将字符串转换为template.HTML类型
//   参数
//     raw: 普通字符串
//   返回
//     返回template.HTML类型字符串
func Str2Html(raw string) template.HTML {
	return template.HTML(raw)
}

// HtmlQuote 将html替换成对应的符号
//   参数
//     src: html字符串
//   返回
//     替换后的字符串
func HtmlQuote(src string) string {
	return html.EscapeString(src)
}

// HtmlUnquote 将html替换成对应的符号
//   参数
//     src: html字符串
//   返回
//     替换后的字符串
func HtmlUnquote(src string) string {
	return html.UnescapeString(src)
}

// Substr 字符串截取
//   参数
//     str:    要处理的字符串
//     start:  截取开始下标
//     length: 截取长度
//   返回
//     返回截取的字符串
func Substr(str string, start, length int) string {
	return utils.Substr(str, start, length)
}

// Sum 求和
//   参数
//     n: 待相加的数字
//   返回
//     相加和
func Sum(n ...int) int {
	sum := 0
	for _, i := range n {
		sum += i
	}

	return sum
}

// Lang 读取语言包数据
//   参数
//     lang: 语言，如zh-cn、en-us
//     key:  语言包key值
//   返回
//     对应语言包数据
func Lang(lang, key string) string {
	if GLang == nil {
		return ""
	}

	return GLang.String(lang, key)
}
