// 内存Session测试
//   变更历史
//     2017-02-14  lixiaoya  新建
package memory

import (
	"testing"
	"time"
)

// TestMemData 测试MemData
func TestMemData(t *testing.T) {
	data := &MemData{id: "123456", accTime: time.Now()}

	id := data.Id()
	if id != "123456" {
		t.Errorf("data.Id() failed. Got %s, expected 123456.", id)
		return
	}

	data.Set("k1", "v1")
	data.Set("k2", "v2")
	data.Set("k3", "v3")
	data.Set("k4", 4444)
	v1 := data.Get("k1")
	if v1 != "v1" {
		t.Errorf("data.Get() failed. Got %s, expected v1.", v1)
		return
	}
	v2 := data.Get("k2")
	if v2 != "v2" {
		t.Errorf("data.Get() failed. Got %s, expected v2.", v2)
		return
	}
	v3 := data.Get("k3")
	if v3 != "v3" {
		t.Errorf("data.Get() failed. Got %s, expected v3.", v3)
		return
	}
	v4 := data.Get("k4")
	if v4 != 4444 {
		t.Errorf("data.Get() failed. Got %d, expected 4444.", v4)
		return
	}

	data.Delete("k1")
	v1 = data.Get("k1")
	if v1 != nil {
		t.Errorf("data.Get() failed. Got %s, expected nil.", v1)
		return
	}

	data.Flush()
	v2 = data.Get("k2")
	if v2 != nil {
		t.Errorf("data.Get() failed. Got %s, expected nil.", v2)
		return
	}
	v4 = data.Get("k4")
	if v4 != nil {
		t.Errorf("data.Get() failed. Got %d, expected nil.", v4)
		return
	}
}

// TestMemProvider 测试MemProvider
func TestMemProvider(t *testing.T) {
	memProvider.Init(7200, "")
	if memProvider.lifeTime != 7200 {
		t.Errorf("memProvider.lifeTime failed. Got %d, expected 7200.", memProvider.lifeTime)
		return
	}

	data, _ := memProvider.GetSessData("abcdef")
	id := data.Id()
	if id != "abcdef" {
		t.Errorf("data.Id() failed. Got %s, expected abcdef.", id)
		return
	}

	data.Set("k1", "v1")
	data.Set("k2", "v2")
	data.Set("k3", "v3")
	data.Set("k4", 4444)
	v1 := data.Get("k1")
	if v1 != "v1" {
		t.Errorf("data.Get() failed. Got %s, expected v1.", v1)
		return
	}
	v2 := data.Get("k2")
	if v2 != "v2" {
		t.Errorf("data.Get() failed. Got %s, expected v2.", v2)
		return
	}
	v3 := data.Get("k3")
	if v3 != "v3" {
		t.Errorf("data.Get() failed. Got %s, expected v3.", v3)
		return
	}
	data.Set("k3", "v32")
	v3 = data.Get("k3")
	if v3 != "v32" {
		t.Errorf("data.Get() failed. Got %s, expected v32.", v3)
		return
	}
	v4 := data.Get("k4")
	if v4 != 4444 {
		t.Errorf("data.Get() failed. Got %d, expected 4444.", v4)
		return
	}

	data2, _ := memProvider.GetSessData("abcdef")
	id = data2.Id()
	if id != "abcdef" {
		t.Errorf("data.Id() failed. Got %s, expected abcdef.", id)
		return
	}

	v1 = data2.Get("k1")
	if v1 != "v1" {
		t.Errorf("data.Get() failed. Got %s, expected v1.", v1)
		return
	}

	data.Delete("k1")
	v1 = data.Get("k1")
	if v1 != nil {
		t.Errorf("data.Get() failed. Got %s, expected nil.", v1)
		return
	}

	data.Flush()
	v2 = data.Get("k2")
	if v2 != nil {
		t.Errorf("data.Get() failed. Got %s, expected nil.", v2)
		return
	}
	v4 = data.Get("k4")
	if v4 != nil {
		t.Errorf("data.Get() failed. Got %d, expected nil.", v4)
		return
	}

	// 这块需要修改life_time才能测试成功
	memProvider.Gc()
	data3, _ := memProvider.GetSessData("abcdef")
	v2 = data3.Get("k2")
	if v2 != nil {
		t.Errorf("data.Get() failed. Got %s, expected nil.", v2)
		return
	}
}
