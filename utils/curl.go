// 调用curl相关函数
//   变更历史
//     2017-02-06  lixiaoya  新建
package utils

import (
	"strings"
	"net/http"
	"net/url"
	"time"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
)

const (
	OPT_PROXY = iota
	OPT_HTTPHEADER
	OPT_SSLCERT
)

// Curl Get或Post数据
//   参数
//     url:     访问的url地址
//     data:    提交的数据，json或http query格式
//     method:  方法标识，取值： POST | GET，默认为GET
//     timeout: 超时时间，单位秒，默认为5秒
//     params:  其它参数
//       目前支持如下参数：
//         OPT_PROXY：     代理，如:http://10.12.34.53:2443
//         OPT_SSLCERT:    https证书，传map[string]string型，certFile（cert证书）、keyFile（key证书，为空时使用cert证书）、caFile（根ca证书，可为空）
//         OPT_HTTPHEADER: http请求头，传map[string]string型
//   返回
//     结果串、http状态、错误内容
func Curl(urlAddr, data, method string, timeout time.Duration, params ...map[int]interface{}) (string, int, error) {
	if timeout <= 0 {
		timeout = 5
	}

	if strings.ToUpper(method) == "POST" {
		method = "POST"
	} else {
		method = "GET"
	}

	// 设置Content-Type
	headers := make(map[string]string)
	if data != "" && data[0] == '{' { // json: {"a":"1", "b":"2", "c":"3"}
		headers["Content-Type"] = "application/json; charset=utf-8"
	} else { // http query: a=1&b=2&c=3
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	tpFlag := false
	tp := http.Transport{}
	if len(params) > 0 {
		param := params[0]

		// 设置代理
		if v, ok := param[OPT_PROXY]; ok {
			if proxyAddr, ok := v.(string); ok {
				proxy := func(_ *http.Request) (*url.URL, error) {
					return url.Parse(proxyAddr)
				}
				tp.Proxy = proxy
				tpFlag = true
			}
		}

		// 设置证书
		if v, ok := param[OPT_SSLCERT]; ok {
			if t, ok := v.(map[string]string); ok {
				if certFile, ok := t["certFile"]; ok && certFile != "" {
					keyFile := ""
					if keyFile, ok = t["keyFile"]; !ok || keyFile == "" {
						keyFile = certFile
					}
					caFile, _ := t["caFile"]

					tlsCfg, err := parseTLSConfig(certFile, keyFile, caFile)
					if err == nil {
						tp.TLSClientConfig = tlsCfg
						tpFlag = true
					}
				}
			}
		}

		// 设置HEADER
		if v, ok := param[OPT_HTTPHEADER]; ok {
			if t, ok := v.(map[string]string); ok {
				for key,val := range t {
					headers[key] = val
				}
			}
		}
	}

	req, err := http.NewRequest(method, urlAddr, strings.NewReader(data))
	if err != nil {
		return "", -1, err
	}

	// 设置HEADER
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	// 设置Transport
	if tpFlag {
		client.Transport = &tp
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1, err
	}

	return string(body), resp.StatusCode, nil
}

// parseTLSConfig 解析证书文件
//   请求
//     certFile: cert证书
//     keyFile:  key证书
//     caFile:   根ca证书，可为空
//   返回
//     解析结果、错误信息
func parseTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// load cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := tls.Config{
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{cert},
	}

	// load root ca
	if caFile != "" {
		caData, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		tlsCfg.RootCAs = pool
	}

	return &tlsCfg, nil
}