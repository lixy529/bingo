// 模板测试
//   变更历史
//     2017-02-08  lixiaoya  新建
package bingo

import (
	"fmt"
	"testing"
)

// TestGetPatternAndParam
func TestGetPatternAndParam(t *testing.T) {
	r := &RouterTab{}

	url := "/user/demo1/ver/3.0/id/10"
	pattern, param := r.getPatternAndParam(url)
	fmt.Println(pattern)
	fmt.Println(param)

	url = "/user/demo2/ver/3.0/id/10/isbool"
	pattern, param = r.getPatternAndParam(url)
	fmt.Println(pattern)
	fmt.Println(param)

	url = "/user/demo3"
	pattern, param = r.getPatternAndParam(url)
	fmt.Println(pattern)
	fmt.Println(param)

	url = "/demo4"
	pattern, param = r.getPatternAndParam(url)
	fmt.Println(pattern)
	fmt.Println(param)

	url = "/user/demo5/ver/3.0/id/10/isbool/true/aa/11/bb/22/cc/33/dd"
	pattern, param = r.getPatternAndParam(url)
	fmt.Println(pattern)
	fmt.Println(param)
}
