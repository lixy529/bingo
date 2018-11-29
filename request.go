// 请求信息相关
//   变更历史
//     2017-02-07  lixiaoya  新建
package bingo

import (
	"net/http"
	"strconv"
	"strings"
	"net/url"
)

type Request struct {
	r         *http.Request
	getParam  map[string][]string
	postParam map[string][]string
	pathParam map[string][]string
}

func (req *Request) reSet(r *http.Request) {
	req.r = r
}

// GetRequest 返回*http.Request
//   参数
//     void
//   返回
//     Request
func (req *Request) GetRequest() *http.Request {
	return req.r
}

// Uri 返回带参数的url
//   参数
//     void
//   返回
//     带参数的url，如: /user/index?id=1001&name=lish
func (req *Request) Uri() string {
	if req.r.RequestURI != "" {
		return req.r.RequestURI
	}

	if req.r.URL.RawQuery == "" {
		return req.r.URL.Path
	}

	return req.r.URL.Path + "?" + req.r.URL.RawQuery
}

// UrlPath 返回不带参数的url
//   参数
//
//   返回
//     不带参数的url，如: /user/index
func (req *Request) UrlPath() string {
	return req.r.URL.Path
}

// Query 返回请求参数
//   参数
//     void
//   返回
//     地址参数，如: id=1001&name=lish
func (req *Request) Query() string {
	return req.r.URL.RawQuery
}

// Protocol 返回协议名称
//   参数
//
//   返回
//     协议名称， 如：HTTP/1.1
func (req *Request) Protocol() string {
	return req.r.Proto
}

// Header 返回Header项
//   参数
//     key: Header的key值
//   返回
//     Header项，如果key不存在，返回nil
func (req *Request) Header(key string) string {
	return req.r.Header.Get(key)
}

// Scheme 返回请求的协议
//   参数
//     void
//   返回
//     请求协议，如: "http" or "https"
func (req *Request) Scheme() string {
	if scheme := req.Header("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if req.r.URL.Scheme != "" {
		return req.r.URL.Scheme
	}
	if req.r.TLS == nil {
		return "http"
	}
	return "https"
}

// Domain 返回域名，不带协议和端口号
// 如果host为空返回localhost
//   参数
//     void
//   返回
//     域名，如: www.xxx.com
func (req *Request) Domain() string {
	if req.r.Host != "" {
		hostParts := strings.Split(req.r.Host, ":")
		if len(hostParts) > 0 {
			return hostParts[0]
		}
		return req.r.Host
	}
	return "localhost"
}

// Port 返回端口号
// 如果出错或者为空时，返回80
//   参数
//     void
//   返回
//     端口号
func (req *Request) Port() int {
	parts := strings.Split(req.r.Host, ":")
	if len(parts) == 2 {
		port, _ := strconv.Atoi(parts[1])
		return port
	}
	return 80
}

// Site 返回路径
//   参数
//     void
//   返回
//     路径，如: http://www.xxx.com:9090
func (req *Request) Site() string {
	scheme := req.Scheme()
	port := req.Port()
	site := scheme + "://" + req.Domain()
	if port == 80 || port == 443 {
		return site
	}
	return site + ":" + strconv.Itoa(port)
}

// Method 返回请求的方法
//   参数
//     void
//   返回
//     请求的方法，如：POST、GET
func (req *Request) Method() string {
	return req.r.Method
}

// Is 返回请求是否是指定的方法, 如：Is("POST").
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) Is(method string) bool {
	return req.Method() == strings.ToUpper(method)
}

// IsGet 是否是GET请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsGet() bool {
	return req.Is("GET")
}

// IsPost 返回是否是POST请求
//   参数
//
//   返回
//     是返回true，否则返回false
func (req *Request) IsPost() bool {
	return req.Is("POST")
}

// IsHead 返回是否是HEAD请求
func (req *Request) IsHead() bool {
	return req.Is("HEAD")
}

// IsOptions 是否是OPTIONS请求
//   参数
//
//   返回
//     是返回true，否则返回false
func (res *Request) IsOptions() bool {
	return res.Is("OPTIONS")
}

// IsPut 返回是否是PUT请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsPut() bool {
	return req.Is("PUT")
}

// IsDelete 返回是否是DELETE请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsDelete() bool {
	return req.Is("DELETE")
}

// IsPatch 返回是否是PATCH请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsPatch() bool {
	return req.Is("PATCH")
}

// IsAjax 返回是否是Ajax请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsAjax() bool {
	return req.Header("X-Requested-With") == "XMLHttpRequest"
}

// IsHttps 返回是否是https请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsHttps() bool {
	return req.Scheme() == "https"
}

// IsWebsocket 返回是否是websocket请求
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsWebsocket() bool {
	return req.Header("Upgrade") == "websocket"
}

// IsUpload 返回是否支持文件上传
//   参数
//     void
//   返回
//     是返回true，否则返回false
func (req *Request) IsUpload() bool {
	return strings.Contains(req.Header("Content-Type"), "multipart/form-data")
}

// UserAgent 返回请求的User-Agent数据
//   参数
//     void
//   返回
//     User-Agent数据
func (req *Request) UserAgent() string {
	return req.Header("User-Agent")
}

// Referer 返回请求的Referer数据
//   参数
//     void
//   返回
//     Referer数据
func (req *Request) Referer() string {
	return req.Header("Referer")
}

// ClientIp 返回客户端IP
// 为空时返回127.0.0.1
//   参数
//     void
//   返回
//     客户端IP
func (req *Request) ClientIp() string {
	// 业务自己添加的header
	if AppCfg.ServerCfg.ForwardName != "" {
		if ips := req.Header(AppCfg.ServerCfg.ForwardName); ips != "" {
			t := strings.Split(ips, ",")
			if len(t) > 0 {
				if AppCfg.ServerCfg.ForwardRev {
					return t[len(t)-1]
				} else {
					return t[0]
				}
			}
		}
	}

	// Cdn-Src-Ip
	if ip := req.Header("Cdn-Src-Ip"); ip != "" {
		return ip
	}

	// X-Forwarded-For
	if ips := req.Header("X-Forwarded-For"); ips != "" {
		t := strings.Split(ips, ",")
		for _, ip := range t {
			if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.16.") {
				continue
			}
			return ip
		}
	}

	// Client_Ip
	if ip := req.Header("Client-Ip"); ip != "" {
		return ip
	}

	// RemoteAddr
	addr := strings.Split(req.r.RemoteAddr, ":")
	if len(addr) > 0 {
		if addr[0] != "[" {
			return addr[0]
		}
	}
	return "127.0.0.1"
}

// ClientPort 返回客户端端口号
//   参数
//     void
//   返回
//     客户端端口号
func (req *Request) ClientPort() string {
	addr := req.r.RemoteAddr
	addrs := strings.Split(addr, ":")
	if len(addrs) > 1 {
		return addrs[1]
	}

	return ""
}

// parseGetParam 解析GET提交的数据
//   参数
//     void
//   返回
//     void
func (req *Request) parseGetParam() {
	if req.getParam == nil {
		req.getParam = make(map[string][]string)
	}
	req.getParam = req.r.URL.Query()
}

// parseGetParam 解析GET提交的数据
// 等同req.r.PostFormValue("xxx")
//   参数
//     void
//   返回
//     void
func (req *Request) parsePostParam() {
	if req.r.PostForm == nil {
		req.r.ParseMultipartForm(32 << 20) // 32M
	}

	req.postParam = req.r.PostForm
}

// parsePathParam 解析path提交的数据
//   参数
//     param: path参数
//   返回
//     void
func (req *Request) parsePathParam(param map[string]string) {
	req.pathParam = make(map[string][]string)
	for k, v := range param {
		req.pathParam[k] = []string{v}
	}
}

// GetCookie 获取cookie
//   参数
//     name: cookie的key值
//   返回
//     返回cookie的value值
func (req *Request) GetCookie(name string) string {
	ck, err := req.r.Cookie(name)
	if err != nil {
		return ""
	}
	val, _ := url.QueryUnescape(ck.Value)
	return val
}
