// 字符串处理函数
//   变更历史
//     2017-03-30  lixiaoya  新建
package utils

import (
	"strings"
	"regexp"
)

// DelRepeat 删除重复的字符
// 比如 a//b//c////c => a/b/c
//   参数
//     s: 要处理的字符串
//   返回
//     返回处理完的字符串
func DelRepeat(s string, c byte) string {
	b := []byte{}
	j := 0
	flag := false
	for i := 0; i < len(s); i++ {
		if flag && s[i] == c {
			continue
		}

		if s[i] == c {
			flag = true
		} else {
			flag = false
		}

		b = append(b, s[i])
		j++
	}

	return string(b)
}

// DelRepeat 删除重复的字符
// 如果olds和news数量不一样，则以数据少的为准
// 比如 1m12m33 => 1+2=3
//   参数
//     s:    要处理的字符串
//     olds: 要替换的数据
//     news: 替换的新数据
//     n:    替换的个数，0不替换， 小于0 全部替换
//   返回
//     成功返回处理完的字符串，失败返回错误信息
func Replace(src string, olds []string, news []string, n int) string {
	oldCnt := len(olds)
	newCnt := len(news)
	if src == "" || oldCnt == 0 || newCnt == 0 || n == 0 {
		return src
	}

	cnt := oldCnt
	if cnt > newCnt {
		cnt = newCnt
	}

	dst := src
	for i := 0; i < cnt; i++ {
		dst = strings.Replace(dst, olds[i], news[i], n)
	}

	return dst
}

// Substr 字符串截取
//   参数
//     str:    要处理的字符串
//     start:  截取开始下标
//     length: 截取长度
//   返回
//     返回截取的字符串
func Substr(str string, start, length int) string {
	l := len(str)
	if start >= l {
		return ""
	}

	if start < 0{
		start = l + start
		if start < 0 {
			start = 0
		}
	}

	end := start + length
	if end > l {
		end = l
	}

	return str[start:end]
}

// Empty 判断字符串是否为空
//   参数
//     str: 要判断的字符串
//   返回
//     为空时返回true，否则返回false
func Empty(str string) bool {
	if len(str) == 0 {
		return true
	}

	return false
}

// GetSafeSql 防sql注入，将sql特殊字符进行删除
//   参数
//     str: 要处理的字符串
//   返回
//     删除后的数据
func GetSafeSql(str string) string {
	if Empty(str) {
		return ""
	}
	pattern := `\b(?i:sleep|delay|waitfor|and|exec|execute|insert|select|delete|update|count|master|char|declare|net user|xp_cmdshell|or|create|drop|table|from|grant|use|group_concat|column_name|information_schema.columns|table_schema|union|where|orderhaving|having|by|truncate|like)\b`
	reg := regexp.MustCompile(pattern)

	return reg.ReplaceAllString(str, "")
}

// Html2Str Html字符替换
//   参数
//     html: 要处理的字符串
//   返回
//     替换后的数据
func Html2Str(html string) string {
	src := string(html)

	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//remove STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//remove SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

// StrSplit 字符串按指定长度分割到切片中
//   参数
//     src: 源字符串
//     length: 分割的长度，如果长度小于等于0则返回空串
//   返回
//     分割后的切片
func StrSplit(src string, length ...int) []string {
	length = append(length, 1)
	res := []string{}
	if length[0] <= 0 {
		return res
	}
	srcLen := len(src)
	if srcLen <= length[0] {
		return append(res, src)
	}

	pos := 0
	tmp := make([]byte, length[0])
	for i := 0; i < srcLen; i++ {
		tmp[pos] = src[i]
		pos++
		if pos == length[0] {
			res = append(res, string(tmp))
			tmp = make([]byte, length[0])
			pos = 0
		}
	}

	if pos > 0 {
		res = append(res, string(tmp))
	}

	return res
}
