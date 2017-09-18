// 处理图片相关函数测试
//   变更历史
//     2017-06-15  lixiaoya  新建
package utils

import (
	"testing"
	"io/ioutil"
)

// TestIsFile IsFile函数测试
func TestGetImageType(t *testing.T) {
	imgContent, err := ioutil.ReadFile("/letv/tmp/avatar/aaa.jpeg")
	if err != nil {
		t.Errorf("ReadFile err: %s", err.Error())
		return
	}

	imgType := GetImageType(imgContent)
	if imgType != "JPEG" {
		t.Errorf("GetImageType failed: Got [%s], expected [%s].", imgType, "JPEG")
		return
	}
}

// TestTranImage 测试图片转换
func TestTranImage(t *testing.T) {
	srcFile := "./222.gif"
	dstFile := "./222.png"

	// 解析图片
	img, err := DecodeImg(srcFile, "GIF")
	if err != nil {
		t.Errorf("DecodeImg err: %s", err.Error())
		return
	}

	// 生成图片
	err =  EncodeImage(dstFile, img, "PNG")
	if err != nil {
		t.Errorf("EncodeImage err: %s", err.Error())
		return
	}
}
