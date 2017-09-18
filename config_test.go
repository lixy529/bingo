// config测试
//   变更历史
//     2017-02-23  lixiaoya  新建
package bingo

import (
	"fmt"
	"testing"
)

// TestConfig 测试config
func TestConfig(t *testing.T) {
	AppCfg, err := newAppConfig()
	if err != nil {
		t.Errorf("newAppConfig error, [%s]", err.Error())
		return
	}
	fmt.Println(AppCfg.DbConfigs)
}

func TestGetConfig(t *testing.T) {
	str := GetString("app", "app_name", "")
	if str != "demo" {
		t.Errorf("GetString failed. Got %s, expected demo.", str)
	}

	b := GetBool("server", "secure", true)
	if b {
		t.Errorf("GetBool failed. Got true, expected false.")
	}

	i := GetInt("server", "port")
	if i != 9090 {
		t.Errorf("GetBool failed. Got %d, expected 9090.", i)
	}
}
