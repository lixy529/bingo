// 文件相关函数测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"testing"
	"os"
	"fmt"
)

// TestIsFile IsFile函数测试
func TestIsFile(t *testing.T) {
	name := "./utils.go"

	if r, _ := IsFile(name); !r {
		t.Error("StarWithStr failed. Got false, expected true.")
		return
	}
}

// TestMkDir MkDir函数测试
func TestMkDir(t *testing.T) {
	name := "./test/dd/cc/aa.pid"

	if err := MkDir(name, 0777, true); err != nil {
		t.Error("MkDir failed")
		return
	}
}

// TestWriteFile WriteFile函数测试
func TestWriteFile(t *testing.T) {
	name := "/tmp/test.log"
	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("message_%02d\n", i)
		_, err := WriteFile(name, []byte(data), os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0660))
		if err != nil {
			t.Errorf("WriteFile err: %s", err.Error())
			return
		}
	}
}

// TestWriteFile WriteFile函数测试
func TestFileCtime(t *testing.T) {
	sec, nsec, err := FileCtime("/letv/tmp/static/sso_plat_from_adminsso_online.json")
	fmt.Println(sec, nsec, err)
}
