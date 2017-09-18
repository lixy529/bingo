// 解析配置文件测试
//   变更历史
//     2017-09-18  lixiaoya  新建
package config

import "testing"

// TestIniConfig 测试ini配置
func TestIniConfig(t *testing.T) {
	cfgFile := "./cfg/app.ini"
	obj, err := NewConfig(cfgFile)
	if err != nil {
		t.Errorf("GetConfig Err, err: %s", err.Error())
		return
	}

	user := obj.GetString("comm", "user", "lixy")
	if user != "Diego" {
		t.Errorf("GetString err, Got:%s expected:%s", user, "Diego")
		return
	}

	age := obj.GetInt32("comm", "age", 100)
	if age != 30 {
		t.Errorf("GetInt32 err, Got:%d expected:%d", age, 30)
		return
	}

	test03 := obj.GetBool("ext", "test03", false)
	if !test03 {
		t.Error("GetBool err, Got:true expected:false")
		return
	}
}
