// 验证相关函数
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"regexp"
	"strings"
)

const (
	IPV4 = 1
	IPV6 = 2
	IPVX = 3
)

// CheckIp 检验字符串是否是ip
//   参数
//     ip:  IP地址
//     ipv: ip类型，可选值：IPV4、IPV6、IPVX，默认为IPV4
//   返回
//     是否IPv4地址
func CheckIp(ip string, ipv ...int) bool {
	ipv = append(ipv, IPV4)

	if ipv[0] & IPV4 == IPV4 && IsIpv4(ip) {
		return true
	}

	if ipv[0] & IPV6 == IPV6 && IsIpv6(ip) {
		return true
	}

	return false
}

// IsIpv4 检验字符串是否是ipv4
//   参数
//     ip: IP地址
//   返回
//     是否IPv6地址
func IsIpv4(ip string) bool {
	pattern := `^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	return false
}

// IsIpv6 检验字符串是否是ipv6
//   参数
//     ip: IP地址
//   返回
//     是否IP地址
func IsIpv6(ip string) bool {
	// CDCD:910A:2222:5498:8475:1111:3900:2020 格式
	pattern := `^([0-9a-fA-Z]{1,4}:){7}[0-9a-fA-Z]{1,4}$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F:F:F::1:1 F:F:F:F:F::1 F::F:F:F:F:1 格式
	pattern = `^(([0-9a-fA-Z]{1,4}:){0,6})((:[0-9a-fA-Z]{1,4}){0,6})$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		t := strings.Split(ip, ":")
		if len(t) > 0 && len(t) <= 8 {
			return true
		}
	}

	// F:F:10F:: 格式
	pattern = `^([0-9a-fA-F]{1,4}:){1,7}:$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// ::F:F:10F 格式
	pattern = `^:(:[0-9a-fA-F]{1,4}){1,7}$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F:0:0:0:0:0:10.0.0.1 格式
	pattern = `^([0-9a-fA-F]{1,4}:){6}((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// F::10.0.0.1 格式
	pattern = `^([0-9a-fA-F]{1,4}:){1,5}:((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	// ::10.0.0.1 格式
	pattern = `^::((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	if m, _ := regexp.MatchString(pattern, ip); m {
		return true
	}

	return false
}

// CheckEmail 检验邮箱格式
//   参数
//     email: 邮箱地址
//   返回
//     是否邮箱地址
func CheckEmail(email string) bool {
	//pattern := "[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?"
	pattern := `^[A-Za-z\d]+([-_.][A-Za-z\d]+)*@([A-Za-z\d]+[-.])+[A-Za-z\d]{2,4}$`
	if m, _ := regexp.MatchString(pattern, email); m {
		return true
	}

	return false
}

// CheckMobile 检验手机格式
//   参数
//     mobile: 手机号
//   返回
//     是否手机号
func CheckMobile(mobile string) bool {
	pattern := `^(1[3|4|5|7|8])\d{9}$`
	if m, _ := regexp.MatchString(pattern, mobile); m {
		return true
	}

	return false
}
