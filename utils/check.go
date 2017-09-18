// 验证相关函数
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"regexp"
)

// CheckIp 检验字符串是否是ip
//   参数
//     ip: IP地址
//   返回
//     是否IP地址
func CheckIp(ip string) bool {
	pattern := `^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
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
