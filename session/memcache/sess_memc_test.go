// Memcache Session测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package memcache

import (
	"strings"
	"testing"
)

// TestMemData 测试MemData
func TestMemcData(t *testing.T) {
	sessId := "123456"
	data := &MemcData{id: sessId, lifeTime: 3600, isUpd: false}

	id := data.Id()
	if id != sessId {
		t.Errorf("data.Id() failed. Got %s, expected %s.", id, sessId)
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

type Info struct {
	Name  string
	Email string
	Phone string
}

func (i *Info) ToString() string {
	return i.Name + "|" + i.Email + "|" + i.Phone
	//return "{\"Name\":\""+ i.Name + "\",\"Email\":\"" + i.Email + "\",\"Phone\":\"" + i.Phone + "\"}"
}

func (i *Info) ToObj(str string) {
	l := strings.Split(str, "|")
	if len(l) > 2 {
		i.Name = l[0]
		i.Email = l[1]
		i.Phone = l[2]
	}
}

// TestMemProvider 测试MemProvider
func TestMemcProvider(t *testing.T) {
	var lifeTime int64
	lifeTime = 3600
	var err error
	memcProvider.Init(lifeTime, "127.0.0.1:11211")
	if memcProvider.lifeTime != lifeTime {
		t.Errorf("memProvider.lifeTime failed. Got %d, expected %d.", memcProvider.lifeTime, lifeTime)
		return
	}

	sessId := "123456"
	data, _ := memcProvider.GetSessData(sessId)
	id := data.Id()
	if id != sessId {
		t.Errorf("data.Id() failed. Got %s, expected %s.", id, sessId)
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

	info := Info{Name: "lixy", Email: "lixiaoya@le.com", Phone: "15811112222"}
	err = data.Set("info", info.ToString())
	if err != nil {
		t.Errorf("data.Set() failed. err: %s.", err.Error())
		return
	}

	err = data.Write()
	if err != nil {
		t.Errorf("data.Write() failed. err: %s.", err.Error())
		return
	}

	/////////模拟第二服务器//////////

	data2, _ := memcProvider.GetSessData(sessId)
	id = data2.Id()
	if id != sessId {
		t.Errorf("data2.Id() failed. Got %s, expected %s.", id, sessId)
		return
	}

	v1 = data2.Get("k1")
	if v1 != "v1" {
		t.Errorf("data2.Get() failed. Got %s, expected v1.", v1)
		return
	}

	v2 = data2.Get("k2")
	if v2 != "v2" {
		t.Errorf("data2.Get() failed. Got %s, expected v2.", v2)
		return
	}

	v3 = data2.Get("k3")
	if v3 != "v32" {
		t.Errorf("data2.Get() failed. Got %s, expected v32.", v3)
		return
	}
	err = data2.Delete("k3")
	if err != nil {
		t.Errorf("data2.Delete() failed. err: %s.", err.Error())
		return
	}

	v4 = data2.Get("k4")
	if v4 != 4444 {
		t.Errorf("data2.Get() failed. Got %d, expected 4444.", v4)
		return
	}

	info2 := data2.Get("info")
	ifo := Info{}
	ifo.ToObj(info2.(string))
	name := ifo.Name
	email := ifo.Email
	phone := ifo.Phone
	if name != "lixy" {
		t.Errorf("data2.Get() failed. Got %s, expected lixy.", name)
		return
	}
	if email != "lixiaoya@le.com" {
		t.Errorf("data2.Get() failed. Got %s, expected lixiaoya@le.com.", email)
		return
	}
	if phone != "15811112222" {
		t.Errorf("data2.Get() failed. Got %s, expected 15811112222.", phone)
		return
	}

	err = data2.Write()
	if err != nil {
		t.Errorf("data2.Write() failed. err: %s.", err.Error())
		return
	}

	/////////模拟第三服务器//////////

	data3, _ := memcProvider.GetSessData(sessId)
	id = data3.Id()
	if id != sessId {
		t.Errorf("data3.Id() failed. Got %s, expected %s.", id, sessId)
		return
	}

	v1 = data3.Get("k1")
	if v1 != "v1" {
		t.Errorf("data.Get() failed. Got %s, expected v1.", v1)
		return
	}

	v2 = data3.Get("k2")
	if v2 != "v2" {
		t.Errorf("data3.Get() failed. Got %s, expected v2.", v2)
		return
	}
	v3 = data3.Get("k3")
	if v3 != nil {
		t.Errorf("data3.Get() failed. Got %s, expected nil.", v3)
		return
	}

	v4 = data3.Get("k4")
	if v4 != 4444 {
		t.Errorf("data3.Get() failed. Got %d, expected 4444.", v4)
		return
	}
}
