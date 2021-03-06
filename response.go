// 响应信息相关
//   变更历史
//     2017-02-07  lixiaoya  新建
package bingo

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
	"strconv"
	"io"
	"net/url"
)

type Response struct {
	w           http.ResponseWriter
	Status      int
	encoding    string // 返回要设置的压缩编码
	contentType string // OutPut时使用的Content-Type，默认为"text/html; charset=utf-8"
}

func (rsp *Response) reSet(w http.ResponseWriter, encoding string) {
	rsp.w = w
	rsp.Status = 0
	rsp.encoding = encoding
}

// GetResponse 返回http.ResponseWriter
//   参数
//     void
//   返回
//     返回ResponseWriter
func (rsp *Response) GetResponse() http.ResponseWriter {
	return rsp.w
}

// Header 设置response的header
//   参数
//     key: Header的key值
//     val: Header的value值
//   返回
//     void
func (rsp *Response) Header(key, val string) {
	rsp.w.Header().Set(key, val)
}

// SetContentType 设置Content-Type
// 这个函数只用于OutPut函数的设置Content-Type
// 有的业务给页面输了文本时需要设置Content-Type，如设置成“application/json;charset=utf-8”
//   参数
//     key: Header的key值
//   返回
//     void
func (rsp *Response) SetContentType(contentType string) {
	rsp.contentType = contentType
}

// OutPut 输出文本数据
//   参数
//     content: 数据
//   返回
//     错误信息
func (rsp *Response) OutPut(content []byte) error {
	if rsp.contentType == "" {
		rsp.contentType = "text/html; charset=utf-8"
	}
	rsp.Header("Content-Type", rsp.contentType)
	var buf = &bytes.Buffer{}
	b, n, err := Compress(rsp.encoding, buf, content)
	if err != nil {
		http.Error(rsp.w, err.Error(), http.StatusInternalServerError) // 500
		return err
	} else if b {
		rsp.Header("Content-Encoding", n)
	} else {
		rsp.Header("Content-Length", strconv.Itoa(len(content)))
	}

	io.Copy(rsp.w, buf)
	return nil
}

// Image 写二进制文件，如图片等
//   参数
//     data:        要输出的数据
//     contentType: Content-Type，如：png图片为 image/png，jpg图片为 image/jpeg
//   返回
//     成功返回nil，失败返回错误信息
func (rsp *Response) Binary(data []byte, contentType string) error {
	rsp.Header("Content-Type", contentType)
	_, err := rsp.w.Write([]byte(data))
	return err
}

// SetCookie 设置cookie
//   参数
//     name:   cookie的key值
//     value:  cookie的value值
//     others: 其它参数，依次为下面几项
//         MxAge    int    设置过期时间，对应浏览器cookie的MaxAge属性
//         Path     string 路径
//         Domain   string 域名
//         Secure   bool   设置Secure属性
//         HttpOnly bool   设置httpOnly属性
//   返回
//     void
func (rsp *Response) SetCookie(name, value string, others ...interface{}) {
	var b bytes.Buffer
	value = url.QueryEscape(replaceValue(value))
	fmt.Fprintf(&b, "%s=%s", replaceName(name), value)

	// 有效时间
	if len(others) > 0 {
		var maxAge int64

		switch v := others[0].(type) {
		case int:
			maxAge = int64(v)
		case int32:
			maxAge = int64(v)
		case int64:
			maxAge = v
		}

		switch {
		case maxAge > 0:
			fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d", time.Now().Add(time.Duration(maxAge) * time.Second).UTC().Format(time.RFC1123), maxAge)
		case maxAge < 0:
			fmt.Fprintf(&b, "; Max-Age=0")
		}
	}

	// 路径，默认为 "/"
	if len(others) > 1 {
		if v, ok := others[1].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Path=%s", replaceValue(v))
		}
	} else {
		fmt.Fprintf(&b, "; Path=%s", "/")
	}

	// 域名，默认为空
	if len(others) > 2 {
		if v, ok := others[2].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Domain=%s", replaceValue(v))
		}
	}

	// Secure属性， 默认为false
	if len(others) > 3 {
		var secure bool
		switch v := others[3].(type) {
		case bool:
			secure = v
		default:
			if others[3] != nil {
				secure = true
			}
		}
		if secure {
			fmt.Fprintf(&b, "; Secure")
		}
	}

	// 设置httpOnly属性，默认为false
	if len(others) > 4 {
		if v, ok := others[4].(bool); ok && v {
			fmt.Fprintf(&b, "; HttpOnly")
		}
	}

	rsp.w.Header().Add("Set-Cookie", b.String())
}

// Error 输出error http code
//   参数
//     errCode: 错误code
//     errMsg:  错误信息
//   返回
//     void
func (rsp *Response) Error(errCode int, errMsg string) {
	http.Error(rsp.w, errMsg, errCode)
}

// replaceName 替换cookie名字里的特殊字符
//   参数
//     cookie名字
//   返回
//     替换后的cookie名字
func replaceName(n string) string {
	var cookieNameReplace = strings.NewReplacer("\n", "-", "\r", "-")
	return cookieNameReplace.Replace(n)
}

// replaceValue 替换cookie值里的特殊字符
//   参数
//     cookie值
//   返回
//     替换后的cookie值
func replaceValue(v string) string {
	var cookieValueReplace = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")
	return cookieValueReplace.Replace(v)
}
