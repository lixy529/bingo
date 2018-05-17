// 处理图片相关函数
//   变更历史
//     2017-06-15  lixiaoya  新建
package utils

import (
	"image"
	"os"
	"image/jpeg"
	"fmt"
	"image/png"
	"image/gif"
	"strings"
)

// GetImageType 获取图片的类型
//   参数
//     imgFile: ip段地址库文件
//   返回
//     图片类型，目前只支持PNG、JPEG、GIF、BMP，其它类型都返回空串
func GetImageType(imgContent []byte) string {
	if imgContent[0] == 137 && imgContent[1] == 80 {
		return "PNG"
	} else if imgContent[0] == 255 && imgContent[1] == 216 {
		return "JPEG"
	} else if imgContent[0] == 71 && imgContent[1] == 73 && imgContent[2] == 70 && (imgContent[4] == 55 || imgContent[4] == 57) {
		return "GIF"
	} else if imgContent[0] == 66 && imgContent[1] == 77 {
		return "BMP"
	}

	return ""
}

// DecodeImg 解析图片
//   参数
//     imgFile: 图片文件路径
//     imgType: 图片类型，目前只支持PNG、JPEG、GIF三种类型
//   返回
//     解码后的图片内容，错误信息
func DecodeImg(imgFile string, imgType string) (image.Image, error) {
	// 打开图片文件
	f, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	// 解码
	var img image.Image
	iType := strings.ToUpper(imgType)
	if iType == "JPEG" {
		img, err = jpeg.Decode(f)
	} else if iType == "PNG" {
		img, err = png.Decode(f)
	} else if iType == "GIF" {
		img, err = gif.Decode(f)
	} else {
		return nil, fmt.Errorf("Image type [%s] is not supported", imgType)
	}
	if err != nil {
		return nil, err
	}

	return img, nil
}

// EncodeImage 生成图片
//   参数
//     imgFile: 图片文件路径
//     img:     解码后的图片内容
//     imgType: 图片类型，目前只支持PNG、JPEG、GIF三种类型
//     option:  JPEG时传图片质量-[1,100]、GIF时颜色范围-[1,256]
//   返回
//     解码后的图片内容，错误信息
func EncodeImage(imgFile string, img image.Image, imgType string, option ...int) error {
	// 创建文件目录
	err := MkDir(imgFile, os.ModeDir|0755, true)
	if err != nil {
		return err
	}

	// 打开图片文件
	f, err := os.OpenFile(imgFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if nil != err {
		return err
	}
	defer f.Close()

	// 写图片文件
	imgType = strings.ToUpper(imgType)
	if imgType == "PNG" {
		return png.Encode(f, img)
	} else if imgType == "JPEG" {
		option = append(option, 90)
		op := option[0]
		return jpeg.Encode(f, img, &jpeg.Options{op})
	} else if imgType == "GIF" {
		option = append(option, 100)
		op := option[0]
		return gif.Encode(f, img, &gif.Options{NumColors: op})
	}

	return fmt.Errorf("Image type [%s] is not supported", imgType)
}
