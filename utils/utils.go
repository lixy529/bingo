// 通用函数
//   变更历史
//     2017-02-06  lixiaoya  新建
package utils

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"path"
	"regexp"
	"net/url"
)

const (
	RAND_KIND_NUM    = 0 // 纯数字
	RAND_KIND_LOWER  = 1 // 小写字母
	RAND_KIND_UPPER  = 2 // 大写字母
	RAND_KIND_LETTER = 3 // 大小写字母
	RAND_KIND_ALL    = 4 // 数字、大小写字母
)

// Uniqid 生成唯一串
//   参数
//     prefix: 前缀
//     r:      带上随机串
//   返回
//     唯一串
func Uniqid(prefix string, r ...bool) string {
	t := time.Now()
	str := fmt.Sprintf("%s%x%x", prefix, t.Unix(), t.UnixNano() - t.Unix() * 1000000000)
	if len(r) > 0 && r[0] {
		str += "." + Krand(8, RAND_KIND_NUM)
	}
	return str
}

// Guid 生成Guid字符串
//   参数
//     void
//   返回
//     生成的Guid串
func Guid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		return ""
	}
	return Md5(base64.URLEncoding.EncodeToString(b) + Uniqid(""))
}

// Krand 随机字符串
//   参数
//     size: 字符串个数
//     kind: 字符串类型，取值为：RAND_KIND_NUM、RAND_KIND_LOWER、RAND_KIND_UPPER、RAND_KIND_LETTER、RAND_KIND_ALL
//   返回
//     生成的字符串
func Krand(size int, kind int) string {
	ikind, bases, scopes, result := kind, []int{48, 97, 65}, []int{10, 26, 26}, make([]byte, size)
	is_all := kind > 3 || kind < 0
	mrand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if is_all {
			ikind = mrand.Intn(3)
		} else if kind == RAND_KIND_LETTER {
			ikind = RAND_KIND_LOWER + mrand.Intn(2)
		}

		base, scope := bases[ikind], scopes[ikind]
		result[i] = uint8(base + mrand.Intn(scope))
	}
	return string(result)
}

// Irand 随机生成一个指定范围的数字，范围[start, end]
//   参数
//     start: 开始值
//     end:   结束值
//   返回
//     生成的数字
func Irand(start, end int) int {
	if start >= end {
		return end
	}
	mrand.Seed(time.Now().UnixNano())
	ikind := mrand.Intn(end - start + 1) + start
	return ikind
}

// RangeInt 生成值为start到end的切片，范围[start, end]
//   参数
//     start: 开始值
//     end:   结束值
//   返回
//     生成的数字切片
func RangeInt(start, end int) []int {
	res := make([]int, end - start + 1)
	for i := 0; i <= end - start; i++ {
		res[i] = start + i
	}

	return res
}

// getTopDomain 获取一级域名，sessionId的cookie记录到一级域名下
// 比如: www.baidu.com 返回 baidu.com
//   参数
//     domain: 域名
//   返回
//     一级域名
func GetTopDomain(domain string) string {
	if domain == "" {
		return ""
	}

	// 解析url
	domain = strings.ToLower(domain)
	urlAddr := domain
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		urlAddr = "http://" + domain
	}
	urlObj, err := url.Parse(urlAddr)
	if err != nil {
		return domain
	}
	urlHost := urlObj.Host
	if strings.Contains(urlHost, ":") {
		urlList := strings.Split(urlHost, ":")
		urlHost = urlList[0]
	}
	if urlHost == "" {
		return domain
	}

	// ip直接返回
	if CheckIp(urlHost) {
		return urlHost
	}

	// 获取一级域名
	domainParts := strings.Split(urlHost, ".")
	l := len(domainParts)
	if l > 1 {
		urlHost = domainParts[l-2] + "." + domainParts[l-1]
	}

	return urlHost
}

// GetLocalIp 获取本机IP
//   参数
//     domain: 域名
//   返回
//     本机IP
func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// Stack 获取异常代码的函数名、文件名、行数
//   参数
//     depth: 堆栈的开始深度和结束深度，默认取1到10
//   返回
//     调用代码信息，格式：函数名:文件名:行数 函数名:文件名:行数 函数名:文件名:行数
func Stack(depth ...int) string {
	var stack string
	var start int = 1
	var end int = 20
	if len(depth) > 0 {
		start = depth[0]
	}
	if len(depth) > 1 {
		end = depth[1]
	}

	for i := start; i < end; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
			funcName := runtime.FuncForPC(pc).Name()
			names := strings.Split(funcName, "/")
			if len(names) > 0 {
				funcName = names[len(names)-1]
			}

			if stack == "" {
				stack = fmt.Sprintf("\n%v:%v:%v", funcName, file, line)
			} else {
				stack = stack + "\n" + fmt.Sprintf("%v:%v:%v", funcName, file, line)
			}
		}
	}

	return stack
}

// getCall 获取调用代码文件名和行数
//   参数
//     depth: 函数调用深度
//   返回
//     文件路径和代码所在行
func GetCall(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "unknown"
		line = 0
	}

	list := strings.Split(file, "/")
	n := len(list)
	if n > 1 {
		file = path.Join(list[n-2], list[n-1])
	}

	return file, line
}

// handleSignals 捕获信号
//   参数
//     void
//   返回
//     信号编号和名称
func HandleSignals() (os.Signal, string) {
	var sig os.Signal
	signalChan := make(chan os.Signal)

	signal.Notify(
		signalChan,
		syscall.SIGTERM,
		syscall.SIGUSR2,
		syscall.SIGHUP,
	)

	for {
		sig = <-signalChan

		switch sig {
		case syscall.SIGTERM:
			return sig, "SIGTERM"

		case syscall.SIGHUP:
			return sig, "SIGHUP"

		case syscall.SIGUSR2:
			return sig, "SIGUSR2"

		default:
			return sig, "unknown"
		}
	}
}

// GetTerminal 获取客户终端信息
//   参数
//     userAgent: 客户的USER_AGENT
//   返回
//     终端类型pc、phone、pad）和终端操作系统（win、unix、linux、mac、ios、android）
//     终端类型（
//     终端操作类型
func GetTerminal(userAgent string) (string, string) {
	userAgent = strings.ToLower(userAgent)

	if m, _ := regexp.MatchString("ipad", userAgent); m {
		return "pad", "ios"
	} else if m, _ := regexp.MatchString("jakarta|iphone|ipod", userAgent); m {
		return "phone", "ios"
	} else if m, _ := regexp.MatchString("windows phone", userAgent); m {
		return "phone", "win"
	} else if m, _ := regexp.MatchString("resty|android", userAgent); m {
		return "phone", "android"
	} else if m, _ := regexp.MatchString("mac", userAgent); m {
		return "pc", "mac"
	} else {
		return "pc", "win"
	}

	return "pc", "win"
}

// SelStrVal 根据条件返回相应选项
//   参数
//     con:  条件
//     opt1: 选项1
//     opt2: 选项2
//   返回
//     如果条件为true，返回选项1，否则返回选项2
func SelStrVal(con bool, opt1, opt2 string) string {
	if con {
		return opt1
	}

	return opt2
}

// SelIntVal 根据条件返回相应选项
//   参数
//     con:  条件
//     opt1: 选项1
//     opt2: 选项2
//   返回
//     如果条件为true，返回选项1，否则返回选项2
func SelIntVal(con bool, opt1, opt2 int) int {
	if con {
		return opt1
	}

	return opt2
}
