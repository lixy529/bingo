// http返回数据进行压缩支持
//   变更历史
//     2017-04-21  lixiaoya  新建
package bingo

import (
	"strings"
	"strconv"
	"sync"
	"io"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"bytes"
)

var (
	defGzipMinLen = 20 // 默认为20B
	gzipMinLen = defGzipMinLen
	gzipLevel int
)

var (
	noneCompressEncoder = acceptEncoder{"", nil, nil, nil}
	gzipCompressEncoder = acceptEncoder{
		name:                    "gzip",
		levelEncode:             func(level int) resetWriter { wr, _ := gzip.NewWriterLevel(nil, level); return wr },
		customCompressLevelPool: &sync.Pool{New: func() interface{} { wr, _ := gzip.NewWriterLevel(nil, gzipLevel); return wr }},
		bestCompressionPool:     &sync.Pool{New: func() interface{} { wr, _ := gzip.NewWriterLevel(nil, flate.BestCompression); return wr }},
	}

	//according to the sec :http://tools.ietf.org/html/rfc2616#section-3.5 ,the deflate compress in http is zlib indeed
	//deflate
	//The "zlib" format defined in RFC 1950 [31] in combination with
	//the "deflate" compression mechanism described in RFC 1951 [29].
	deflateCompressEncoder = acceptEncoder{
		name:                    "deflate",
		levelEncode:             func(level int) resetWriter { wr, _ := zlib.NewWriterLevel(nil, level); return wr },
		customCompressLevelPool: &sync.Pool{New: func() interface{} { wr, _ := zlib.NewWriterLevel(nil, gzipLevel); return wr }},
		bestCompressionPool:     &sync.Pool{New: func() interface{} { wr, _ := zlib.NewWriterLevel(nil, flate.BestCompression); return wr }},
	}
)

var (
	encoderMap = map[string]acceptEncoder{ // all the other compress methods will ignore
		"gzip":     gzipCompressEncoder,
		"deflate":  deflateCompressEncoder,
		"*":        gzipCompressEncoder, // * means any compress will accept,we prefer gzip
		"identity": noneCompressEncoder, // identity means none-compress
	}
)

type resetWriter interface {
	io.Writer
	Reset(w io.Writer)
}

type nopResetWriter struct {
	io.Writer
}
func (n nopResetWriter) Reset(w io.Writer) {
	//do nothing
}

type acceptEncoder struct {
	name                    string
	levelEncode             func(int) resetWriter
	customCompressLevelPool *sync.Pool
	bestCompressionPool     *sync.Pool
}

// encode
func (ac acceptEncoder) encode(wr io.Writer, level int) resetWriter {
	if ac.customCompressLevelPool == nil || ac.bestCompressionPool == nil {
		return nopResetWriter{wr}
	}
	var rwr resetWriter
	switch level {
	case flate.BestSpeed:
		rwr = ac.customCompressLevelPool.Get().(resetWriter)
	case flate.BestCompression:
		rwr = ac.bestCompressionPool.Get().(resetWriter)
	default:
		rwr = ac.levelEncode(level)
	}
	rwr.Reset(wr)
	return rwr
}

// put
func (ac acceptEncoder) put(wr resetWriter, level int) {
	if ac.customCompressLevelPool == nil || ac.bestCompressionPool == nil {
		return
	}
	wr.Reset(nil)

	//notice
	//compressionLevel==BestCompression DOES NOT MATTER
	//sync.Pool will not memory leak

	switch level {
	case gzipLevel:
		ac.customCompressLevelPool.Put(wr)
	case flate.BestCompression:
		ac.bestCompressionPool.Put(wr)
	}
}

// RspEncoding 返回响应使用的压缩格式，如: gzip、deflate
//   参数
//     reqEncoding: 请求的Accept-Encoding
//   返回
//     响应的压缩格式
func RspEncoding(reqEncoding string) string {
	if reqEncoding == "" {
		return ""
	}

	var qName string = ""
	var qValue float64 = 0.0
	for _, v := range strings.Split(reqEncoding, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		vs := strings.Split(v, ";")
		var cf acceptEncoder
		var ok bool
		if cf, ok = encoderMap[vs[0]]; !ok {
			continue
		}
		if len(vs) == 1 {
			return cf.name
		}
		if len(vs) == 2 {
			f, _ := strconv.ParseFloat(strings.Replace(vs[1], "q=", "", -1), 64)
			if f == 0 {
				continue
			}
			if f > qValue {
				qName = cf.name
				qValue = f
			}
		}
	}
	return qName
}

// Compress 对输出的数据进行压缩
//   参数
//     encoding: 响应的压缩格式，如: gzip、deflate
//     writer:   压缩后的结果数据
//     content:  要输出到页面的数据
//   返回
//     是否压缩、压缩格式、错误信息
func Compress(encoding string, writer io.Writer, content []byte) (bool, string, error) {
	if encoding == "" || len(content) < gzipMinLen {
		_, err := writer.Write(content)
		return false, "", err
	}

	var outputWriter resetWriter
	var err error
	var ce = noneCompressEncoder

	if cf, ok := encoderMap[encoding]; ok {
		ce = cf
	}
	encoding = ce.name
	outputWriter = ce.encode(writer, gzipLevel)
	defer ce.put(outputWriter, gzipLevel)

	_, err = io.Copy(outputWriter, bytes.NewReader(content))
	if err != nil {
		return false, "", err
	}

	switch outputWriter.(type) {
	case io.WriteCloser:
		outputWriter.(io.WriteCloser).Close()
	}
	return encoding != "", encoding, nil
}
