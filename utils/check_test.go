// 验证相关函数测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"testing"
)

// TestCheckIp 测试CheckIp函数
func TestCheckIp(t *testing.T) {
	if CheckIp("127.0.0.1") == false {
		t.Error("CheckIp failed. Got false, expected true.")
		return
	}

	if CheckIp("255.255.255.255") == false {
		t.Error("CheckIp failed. Got false, expected true.")
		return
	}

	if CheckIp("127.0.0.") == true {
		t.Error("CheckIp failed. Got true, expected false.")
		return
	}

	if CheckIp("aa.bb.cc.dd") == true {
		t.Error("CheckIp failed. Got true, expected false.")
		return
	}

	if CheckIp("999.999.999.999") == true {
		t.Error("CheckIp failed. Got true, expected false.")
		return
	}
}

// TestCheckEmail 测试CheckEmail函数
func TestCheckEmail(t *testing.T) {
	if CheckEmail("lixiaoya@le.com") == false {
		t.Error("CheckEmail failed. Got false, expected true.")
		return
	}

	if CheckEmail("LXY@SINA.COM") == false {
		t.Error("CheckEmail failed. Got false, expected true.")
		return
	}

	if CheckEmail("123@SINA.COM") == false {
		t.Error("CheckEmail failed. Got false, expected true.")
		return
	}

	if CheckEmail("lixioaya") == true {
		t.Error("CheckEmail failed. Got true, expected false.")
		return
	}

	if CheckEmail("lixiaoya@") == true {
		t.Error("CheckIp failed. Got true, expected false.")
		return
	}

	if CheckEmail("@sina.com") == true {
		t.Error("CheckIp failed. Got true, expected false.")
		return
	}
}

// TestCheckMobile 测试CheckMobile函数
func TestCheckMobile(t *testing.T) {
	// 中国大陆手机号
	mobile := "15812345678"
	if !CheckMobile(mobile) {
		t.Error("CheckMobile failed. Got false, expected true.")
		return
	}

	// 其它
	mobile = "123123123"
	if CheckMobile(mobile) {
		t.Error("CheckMobile failed. Got true, expected false.")
		return
	}
}