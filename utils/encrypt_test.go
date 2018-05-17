// 加密解密相关函数测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"testing"
	"fmt"
)

// TestSha1 测试Sha1函数
func TestSha1(t *testing.T) {
	str := "Hello World!"
	sh := Sha1(str)
	fmt.Println(sh)
}

// TestHmacSha1 测试Sha1函数
func TestHmacSha1(t *testing.T) {
	str := "Hello World!"
	key := "123456"
	sh := HmacSha1(str, key)
	fmt.Println(sh)

	sh = HmacSha1(str, key, true)
	fmt.Println(sh)
}

// TestGobEncode 测试GobEncode和GobDecode函数
func TestGobEncode(t *testing.T) {
	str := "Hello World!"
	// 加密
	e, err := GobEncode(str)
	if err != nil {
		t.Errorf("GobEncode failed. err: %s.", err.Error())
		return
	}

	// 解密
	infoNew := new(string)
	err = GobDecode(e, infoNew)
	if err != nil {
		t.Errorf("GobDecode failed. err: %s.", err.Error())
		return
	} else if *infoNew != str {
		t.Errorf("GobDecode failed. Got %s, expected %s.", *infoNew, str)
		return
	}
}

// TestZlibEncode 测试ZlibEncode和ZlibDecode函数
func TestZlibEncode(t *testing.T) {
	s1 := "Hello World!"
	// 压缩
	b, err := ZlibEncode([]byte(s1))
	if err != nil {
		t.Errorf("ZlibEncode failed. err: %s.", err.Error())
		return
	}

	// 解压
	s2, err := ZlibDecode(b)
	if err != nil {
		t.Errorf("ZlibDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("ZlibDecode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestZlibEncode 测试RsaEncode和RsaDecode函数
func TestRsaEncode(t *testing.T) {
	s1 := "Hello World!"
	pubPem := "./data/public.pem"
	priPem := "./data/private.pem"

	// 公钥加密
	b, err := RsaEncode([]byte(s1), pubPem, MODE_RSA_PUB)
	if err != nil {
		t.Errorf("RsaEncode failed. err: %s.", err.Error())
		return
	}

	// 私钥解密
	s2, err := RsaDecode(b, priPem, MODE_RSA_PRI)
	if err != nil {
		t.Errorf("RsaDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("RsaDecode failed. Got %s, expected %s.", s2, s1)
		return
	}

	// 私钥加密
	b, err = RsaEncode([]byte(s1), priPem, MODE_RSA_PRI)
	if err != nil {
		t.Errorf("RsaEncode failed. err: %s.", err.Error())
		return
	}

	// 公钥解密
	s2, err = RsaDecode(b, pubPem, MODE_RSA_PUB)
	if err != nil {
		t.Errorf("RsaDecode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("RsaDecode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestBase64Encode 测试Base64Encode和Base64Decode函数
func TestBase64Encode(t *testing.T) {
	s1 := "Hello World!"

	// 加密
	str := Base64Encode([]byte(s1), "m", "m1", "+", "m2", "/", "m3")

	// 解密
	s2, err := Base64Decode(str, "m3", "/", "m2", "+", "m1", "m")
	if err != nil {
		t.Errorf("Base64Decode failed. err: %s.", err.Error())
		return
	} else if s1 != string(s2) {
		t.Errorf("Base64Decode failed. Got %s, expected %s.", s2, s1)
		return
	}
}

// TestPadding 测试Padding和UnPadding函数
func TestPadding(t *testing.T) {
	src := []byte("12345")
	dst := Padding(src, 16)
	if len(dst) != 16 {
		t.Errorf("Padding failed. Got %d, expected %d.", len(dst), 16)
		return
	}
	src2 := UnPadding(dst)
	if string(src) != string(src2) {
		t.Errorf("UnPadding failed. Got %s, expected %s.", src2, src)
		return
	}
}

// TestDesEncode 测试DesEncode和DesDecode函数
func TestDesEncode(t *testing.T) {
	key := "12345678"
	src := "HelloWorld!"
	dst, err := DesEncode([]byte(src), []byte(key))
	if err != nil {
		t.Errorf("DesEncode error. %s.", err.Error())
		return
	}

	src2, err := DesDecode(dst, []byte(key))
	if err != nil {
		t.Errorf("DesDecode error. %s.", err.Error())
		return
	} else if (src != string(src2)) {
		t.Errorf("DesDecode failed. Got %s, expected %s.", src2, src)
		return
	}
}

// TestAesEncode 测试AesEncode和AesDecode函数
func TestAesEncode(t *testing.T) {
	key := "12345678901234561234567890123456"
	src := "HelloWorld!"
	dst, err := AesEncode([]byte(src), []byte(key))
	if err != nil {
		t.Errorf("AesEncode error. %s.", err.Error())
		return
	}

	src2, err := AesDecode(dst, []byte(key))
	if err != nil {
		t.Errorf("AesDecode error. %s.", err.Error())
		return
	} else if (src != string(src2)) {
		t.Errorf("AesDecode failed. Got %s, expected %s.", src2, src)
		return
	}
}
