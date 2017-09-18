// 模板测试
//   变更历史
//     2017-02-08  lixiaoya  新建
package bingo

import (
	"fmt"
	"os"
	"testing"
)

type StData struct {
	Title  string
	Info   StInfo
	Footer string
}

type StInfo struct {
	Name  string
	Title string
	Email string
	Phone string
}

// TestTemplate
func TestTemplate(t *testing.T) {
	tp := NewTemplate("/Users/lixiaoya/goyard/goc2p/src/legitlab.letv.cn/uc_tp/goweb/demo/views", ".html")
	if tp == nil {
		t.Errorf("Template is nil")
		return
	}

	err := tp.buildViews()
	if err != nil {
		t.Errorf("buildViews failed, err: %s", err.Error())
		return
	}

	data := StData{
		Title: "this is a test page",
		Info: StInfo{
			Name:  "lixiaoya",
			Email: "lixiaoya@le.com",
			Phone: "15811112222",
		},
		Footer: "The page is end",
	}

	tp.ViewTemp.ExecuteTemplate(os.Stdout, "demo/index.html", data)
	fmt.Println()
}
