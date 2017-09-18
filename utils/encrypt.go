// 加密解密相关函数
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"io/ioutil"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"encoding/base64"
	"strings"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha1"
	"crypto/hmac"
)

const (
	MODE_RSA_PUB  = 1 // 公钥加密或公钥解密
	MODE_RSA_PRI  = 2 // 私钥加密或私钥解密
)

//Md5 生成32位md5串
//   参数
//     s: 要加密的串
//   返回
//     md5后的结果
func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//Sha1 生成sha1串
//   参数
//     s: 要加密的串
//   返回
//     sha1后的结果
func Sha1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//HmacSha1 生成hash_hmac串
//   参数
//     s: 要加密的串
//     k: 加密密钥
//     isBase64: 是否返回base64加密的串，默认false
//   返回
//     hash_hmac后的结果
func HmacSha1(s, k string, isBase64 ...bool) string {
	h := hmac.New(sha1.New, []byte(k))
	h.Write([]byte(s))

	if len(isBase64) > 0 && isBase64[0] {
		return Base64Encode(h.Sum(nil))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// GobEncode 用gob进行数据编码
//   调用
//     e, err := GobEncode("Hello World!")
//   参数
//     data: 要加密的串
//   返回
//     加密后的结果
func GobEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode 用gob进行数据解码
//   调用
//     infoNew := new(string)
//     err := GobDecode(e, infoNew)
//   参数
//     data: 加密串
//   返回
//     解密后的结果
func GobDecode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

// ZlibEncode zlib压缩
//   参数
//     data: 要压缩的数据
//   返回
//     成功时返回压缩后的数据，失败返回错误信息
func ZlibEncode(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		w.Close()
		return nil, err
	}
	w.Close()

	return b.Bytes(), nil
}

// ZlibDecode zlib解压
//   参数
//     data: 要解压的数据
//   返回
//     成功时返回解压后的数据，失败返回错误信息
func ZlibDecode(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.Bytes(), err
}

// RsaEncode Rsa加密
//   参数
//     data:    要加密的数据
//     pemFile: pem证书文件
//     mode:    加密模式，取值: MODE_RSA_PUB、MODE_RSA_PRI
//   返回
//     成功返回加密结果，失败返回错误信息
func RsaEncode(data []byte, pemFile string, mode int) ([]byte, error) {
	if mode == MODE_RSA_PRI {
		// 私钥加密
		privateKey, err := ioutil.ReadFile(pemFile)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(privateKey)
		if block == nil {
			return nil, errors.New("private key error!")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		return EncPKCS1v15(rand.Reader, priv, data)
	}

	// 公钥加密
	publicKey, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// RsaDecode Rsa解密
// 参数 https://github.com/dgkang/rsa
//   参数
//     data:    要解密的数据
//     pemFile: pem证书文件
//     mode:    加密模式，取值: MODE_RSA_PUB、MODE_RSA_PRI
//   返回
//     成功返回解密结果，失败返回错误信息
func RsaDecode(data []byte, pemFile string, mode int) ([]byte, error) {
	if mode == MODE_RSA_PUB {
		// 公钥解密
		publicKey, err := ioutil.ReadFile(pemFile)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(publicKey)
		if block == nil {
			return nil, errors.New("public key error")
		}

		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pub := pubInterface.(*rsa.PublicKey)

		return DecPKCS1v15(pub, data)
	}

	// 私钥解密
	privateKey, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}

// Base64Encode Base64加密
//   参数
//     data:     要加密的数据
//     replaces: Base64加密后的替换新老数据对，必须成对，解密时要传加密数据相反
//   返回
//     返回加密结果
func Base64Encode(data []byte, replaces ...string) string {
	str := base64.StdEncoding.EncodeToString(data)
	n := len(replaces)
	if n > 1 {
		var olds []string
		var news []string
		for i := 0; i < n - 1; i += 2 {
			olds = append(olds, replaces[i])
			news = append(news, replaces[i + 1])
		}

		str = Replace(str, olds, news, -1)
	}

	return str
}

// Base64Decode Base64解密
//   参数
//     data:     要解密的数据
//     replaces: Base64加密后的替换新老数据对，必须成对，解密时要传加密数据相反
//   返回
//     成功返回解密结果，失败返回错误信息
func Base64Decode(data string, replaces ...string) ([]byte, error) {
	n := len(replaces)
	if n > 1 {
		var olds []string
		var news []string
		for i := 0; i < n - 1; i += 2 {
			olds = append(olds, replaces[i])
			news = append(news, replaces[i + 1])
		}

		data = Replace(data, olds, news, -1)
	}

	// 处理不规范加密串
	if m := len(data) % 4; m != 0 {
		data += strings.Repeat("=", 4-m)
	}

	return base64.StdEncoding.DecodeString(data)
}

// Padding 将text补充到blockSize的整数倍
// 如果text长度正好是blockSize的整数倍，则还会添加blockSize长的数据
//   参数
//     text:      原始数据
//     blockSize: 每块大小
//   返回
//     填充后的数据
func Padding(text []byte, blockSize int) []byte {
	padding := blockSize - len(text)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padtext...)
}

// UnPadding 删除Padding填充的数据
//   参数
//     text:      原始数据
//     blockSize: 每块大小
//   返回
//     原始数据
func UnPadding(text []byte) []byte {
	length := len(text)
	unpadding := int(text[length-1])

	if length - unpadding < 0 || unpadding < 0 {
		return []byte("")
	}

	return text[:(length - unpadding)]
}

// DesEncode Des加密
//   参数
//     data: 要加密的数据
//     key:  加密key，只能是8位
//   返回
//     成功时返回加密结果，失败时返回错误信息
func DesEncode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l > 8 {
		key = key[:8]
	} else if l < 8 {
		padtext := bytes.Repeat([]byte("x"), 8-l)
		key = append(key, padtext...)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data = Padding(data, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	encode := make([]byte, len(data))
	blockMode.CryptBlocks(encode, data)
	return encode, nil
}

// DesDecode Des解密
//   参数
//     data: 要解密的数据
//     key:  加密key，只能是8位
//   返回
//     成功返回解密结果，失败返回错误信息
func DesDecode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l > 8 {
		key = key[:8]
	} else if l < 8 {
		padtext := bytes.Repeat([]byte("x"), 8-l)
		key = append(key, padtext...)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, key)
	decode := data
	blockMode.CryptBlocks(decode, data)
	decode = UnPadding(decode)
	return decode, nil
}

// AesEncode Aes加密
//   参数
//     data: 要加密的数据
//     key:  加密key，密码为16的倍数
//   返回
//     成功时返回加密结果，失败时返回错误信息
func AesEncode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l < 16 {
		padtext := bytes.Repeat([]byte("x"), 16-l)
		key = append(key, padtext...)
	} else if l%16 != 0 {
		var n int = l/16*16
		key = key[:n]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	data = Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encode := make([]byte, len(data))
	blockMode.CryptBlocks(encode, data)
	return encode, nil
}

// AesDecode Aes解密
//   参数
//     data: 要解密的数据
//     key:  加密key，密码为16的倍数
//   返回
//     成功返回解密结果，失败返回错误信息
func AesDecode(data, key []byte) ([]byte, error) {
	l := len(key)
	if l < 16 {
		padtext := bytes.Repeat([]byte("x"), 16-l)
		key = append(key, padtext...)
	} else if l%16 != 0 {
		var n int = l/16*16
		key = key[:n]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decode := make([]byte, len(data))
	blockMode.CryptBlocks(decode, data)
	decode = UnPadding(decode)

	return decode, nil

}
