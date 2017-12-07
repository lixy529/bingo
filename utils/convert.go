// 转换相关函数
//   变更历史
//     2017-03-30  lixiaoya  新建
package utils

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
	"strconv"
	"strings"
	"fmt"
	"net/url"
)

// IpItoa IP整型转字符串
// 如: "10.58.1.29" => 171573533
//   参数
//     iIp: IP整型形式
//   返回
//     IP字符串形式
func IpItoa(iIp int64) string {
	var bytes [4]byte
	bytes[0] = byte(iIp & 0xFF)
	bytes[1] = byte((iIp >> 8) & 0xFF)
	bytes[2] = byte((iIp >> 16) & 0xFF)
	bytes[3] = byte((iIp >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String()
}

// IpAtoi IP字符串转整型
// 如: 171573533 => "10.58.1.29"
//   参数
//     sIp: IP字符串形式
//   返回
//     IP整型形式
func IpAtoi(sIp string) int64 {
	bits := strings.Split(sIp, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

// Int32ToByte int32转[]byte
//   参数
//     i:     int32数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func Int32ToByte(i int32, isBig bool) []byte {
	bBuf := bytes.NewBuffer([]byte{})
	if isBig {
		binary.Write(bBuf, binary.BigEndian, i)
	} else {
		binary.Write(bBuf, binary.LittleEndian, i)
	}

	return bBuf.Bytes()
}

// ByteToInt32 []byte转int32
//   参数
//     b:     []byte数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func ByteToInt32(b []byte, isBig bool) int32 {
	bBuf := bytes.NewBuffer(b)
	var i int32
	if isBig {
		binary.Read(bBuf, binary.BigEndian, &i)
	} else {
		binary.Read(bBuf, binary.LittleEndian, &i)
	}

	return i
}

// Int64ToByte int64转[]byte
//   参数
//     i:     int64数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func Int64ToByte(i int64, isBig bool) []byte {
	bBuf := bytes.NewBuffer([]byte{})
	if isBig {
		binary.Write(bBuf, binary.BigEndian, i)
	} else {
		binary.Write(bBuf, binary.LittleEndian, i)
	}

	return bBuf.Bytes()
}

// ByteToInt64 []byte转int64
//   参数
//     i:     []byte数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func ByteToInt64(b []byte, isBig bool) int64 {
	bBuf := bytes.NewBuffer(b)
	var i int64
	if isBig {
		binary.Read(bBuf, binary.BigEndian, &i)
	} else {
		binary.Read(bBuf, binary.LittleEndian, &i)
	}

	return i
}

// Float32ToByte foat32转[]byte
//   参数
//     f:     float32数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func Float32ToByte(f float32, isBig bool) []byte {
	bits := math.Float32bits(f)
	bytes := make([]byte, 4)
	if isBig {
		binary.BigEndian.PutUint32(bytes, bits)
	} else {
		binary.LittleEndian.PutUint32(bytes, bits)
	}


	return bytes
}

// ByteToFloat32 []byte转foat32
//   参数
//     b:     []byte数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func ByteToFloat32(b []byte, isBig bool) float32 {
	var bits uint32
	if isBig {
		bits = binary.BigEndian.Uint32(b)
	} else {
		bits = binary.LittleEndian.Uint32(b)
	}

	return math.Float32frombits(bits)
}

// Float64ToByte foat64转[]byte
//   参数
//     f:     float32数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func Float64ToByte(f float64, isBig bool) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	if isBig {
		binary.BigEndian.PutUint64(bytes, bits)
	} else {
		binary.LittleEndian.PutUint64(bytes, bits)
	}


	return bytes
}

// ByteToFloat64 []byte转foat64
//   参数
//     b:     []byte数据
//     isBig: 是否大端，true-大端 false-小端
//   返回
//     转成结果
func ByteToFloat64(b []byte, isBig bool) float64 {
	var bits uint64
	if isBig {
		bits = binary.BigEndian.Uint64(b)
	} else {
		bits = binary.LittleEndian.Uint64(b)
	}

	return math.Float64frombits(bits)
}

// StrToJSON 将字符串进行Json编码
//   参数
//     b: json串
//   返回
//     编码后的字体串
func StrToJSON(b []byte) string {
	var buf bytes.Buffer
	strLen := len(b)
	for i := 0; i < strLen; i++ {
		c1 := int64(b[i])
		// Single byte
		if c1 < 128 {
			if c1 > 31 {
				buf.WriteByte(b[i])
			} else {
				buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt(c1, 16)))
			}
			continue
		}

		// Double byte
		i++
		if i >= strLen {
			break;
		}
		c2 := int64(b[i])
		if c1 & 32 == 0 {
			buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt((c1 - 192) * 64 + c2 - 128, 16)))
			continue
		}

		// Triple
		i++
		if i >= strLen {
			break;
		}
		c3 := int64(b[i])
		if c1 & 16 == 0 {
			buf.WriteString(fmt.Sprintf("\\u%04s", strconv.FormatInt(((c1 - 224) <<12) + ((c2 - 128) << 6) + (c3 - 128), 16)))
			continue
		}

		// Quadruple
		i++
		if i >= strLen {
			break;
		}
		c4 := int64(b[i])
		if c1 & 8 == 0 {
			var u int64 = ((c1 & 15) << 2) + ((c2>>4) & 3) - 1

			var w1 int64 = (54<<10) + (u<<6) + ((c2 & 15) << 2) + ((c3>>4) & 3)
			var w2 int64 = (55<<10) + ((c3 & 15)<<6) + (c4-128)
			buf.WriteString(fmt.Sprintf("\\u%04s\\u%04s", strconv.FormatInt(w1, 16), strconv.FormatInt(w2, 16)))
		}
	}

	return buf.String()
}

// MapToHttpQuery 将map转http参数串，如:a=11&b=22&c=33，特殊字符做url转码
//   参数
//     m: 待转换的map
//   返回
//     转换后的字符串
func MapToHttpQuery(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	var buf bytes.Buffer
	i := 0
	for key, val := range m {
		if i == 0 {
			buf.WriteString(url.QueryEscape(key) + "=")
		} else {
			buf.WriteByte('&')
			buf.WriteString(url.QueryEscape(key) + "=")
		}
		buf.WriteString(url.QueryEscape(val))

		i++
	}

	return buf.String()
}

// HttpQueryToMap 将http参数串转map
//   参数
//     m: 待转换的字符串
//   返回
//     转换后的map
func HttpQueryToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	if s == "" {
		return m, nil
	}

	l := strings.Split(s, "&")
	for _, v := range l {
		t := strings.Split(v, "=")
		key, err := url.QueryUnescape(t[0])
		if err != nil {
			return nil, err
		}

		if len(t) == 1 {
			m[key] = ""
		} else if len(t) >= 2 {
			val, err := url.QueryUnescape(t[1])
			if err != nil {
				return nil, err
			}
			m[key] = val
		}
	}

	return m, nil
}
