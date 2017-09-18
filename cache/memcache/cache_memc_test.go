// memcache adapter测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package memcache

import (
	"fmt"
	"testing"
	"encoding/json"
)

func TestMemcCache(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","maxIdle":"10","idelTimeOut":"100","prefix":"le_"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	////////////////////////string测试////////////////////////////
	k1 := "k1"
	v1 := "HelloWorld"
	err = adapter.Set(k1, v1, 10)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var v11 string
	err, exist := adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if !exist {
		t.Errorf("Memc Get failed. %s is not exist.", k1)
		return
	} else if v11 != v1 {
		t.Errorf("Memc Get failed. Got %s, expected %s.", v11, v1)
		return
	}

	isExist, err := adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Memc IsExist failed. err: %s.", err.Error())
		return
	} else if !isExist {
		t.Error("Memc Get failed. Got false, expected true.")
		return
	}

	err = adapter.Del(k1)
	if err != nil {
		t.Errorf("Memc Delete failed. err: %s.", err.Error())
		return
	}

	isExist, err = adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Memc IsExist failed. err: %s.", err.Error())
		return
	} else if isExist {
		t.Error("Memc Get failed. Got true, expected false.")
		return
	}

	v11 = ""
	err, _ = adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if v11 != "" {
		t.Errorf("Memc Get failed. Got %s, expected \"\".", v11)
		return
	}

	////////////////////////int32测试////////////////////////////
	k2 := "k2"
	v2 := 100
	err = adapter.Set(k2, int32(v2), 30)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var v22 int
	err, _ = adapter.Get(k2, &v22)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if v22 != v2 {
		t.Errorf("Memc Get failed. Got %d, expected %d.", v22, v2)
		return
	}

	////////////////////////float64测试////////////////////////////
	k3 := "k3"
	v3 := 100.01
	err = adapter.Set(k3, v3, 30)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var v33 float64
	err, _ = adapter.Get(k3, &v33)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if v33 != v3 {
		t.Errorf("Memc Get failed. Got %f, expected %f.", v33, v3)
		return
	}

	////////////////////////Incr、Decr测试////////////////////////////
	k4 := "k4"
	v4 := 100
	err = adapter.Set(k4, v4, 30)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var v44 int
	err, _ = adapter.Get(k4, &v44)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if v44 != v4 {
		t.Errorf("Memc Get failed. Got %s, expected %s.", v44, v4)
		return
	}

	newVal, _ := adapter.Incr(k4)
	if newVal != 101 {
		t.Errorf("Memc Incr failed. Got %d, expected %d.", newVal, 101)
		return
	}

	newVal, _ = adapter.Decr(k4)
	if newVal != 100 {
		t.Errorf("Memc Incr failed. Got %d, expected %d.", newVal, 100)
		return
	}

	newVal, _ = adapter.Incr(k4, 10)
	if newVal != 110 {
		t.Errorf("Memc Incr failed. Got %d, expected %d.", newVal, 110)
		return
	}

	newVal, _ = adapter.Decr(k4, 10)
	if newVal != 100 {
		t.Errorf("Memc Incr failed. Got %d, expected %d.", newVal, 100)
		return
	}

	////////////////////////ClearAll测试////////////////////////////
	err = adapter.ClearAll()
	if err != nil {
		t.Errorf("Memc ClearAll failed. err: %s.", err.Error())
		return
	}
}

// TestMemcMulti
func TestMemcMulti(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","prefix":"le_"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	mList := make(map[string]interface{})
	mList["k1"] = "val1111"
	mList["k2"] = "val2222"
	mList["k3"] = "val3333"
	mList["k4"] = "val4444"
	err = adapter.MSet(mList, 60)
	if err != nil {
		t.Errorf("Memc MSet failed. err: %s.", err.Error())
		return
	}

	mList2, err := adapter.MGet("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Memc MGet failed. err: %s.", err.Error())
		return
	}

	var v1, v2, v3, v4 string
	if mList2["k1"] != nil {
		//json.Unmarshal(mList2["k1"].([]byte), &v1)
		v1 = string(mList2["k1"].([]byte))
	}
	if mList2["k2"] != nil {
		//json.Unmarshal(mList2["k2"].([]byte), &v2)
		v2 = string(mList2["k2"].([]byte))
	}
	if mList2["k3"] != nil {
		//json.Unmarshal(mList2["k3"].([]byte), &v3)
		v3 = string(mList2["k3"].([]byte))
	}
	if mList2["k4"] != nil {
		//json.Unmarshal(mList2["k4"].([]byte), &v4)
		v4 = string(mList2["k4"].([]byte))
	}

	if v1 != mList["k1"] || v2 != mList["k2"] || v3 != mList["k3"] || v4 != mList["k4"] {
		t.Errorf("Memc MGet failed. v1:%s v2:%s v3:%s v4:%s.", v1, v2, v3, v4)
		return
	}

	err = adapter.MDel("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Memc MDelete failed. err: %s.", err.Error())
		return
	}
}

type User struct {
	Id   int
	Name string
}

// TestStruct
func TestStruct(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","maxIdle":"10","idelTimeOut":"100","prefix":"le_","serializer":"json","compressType":"zlib","compressThreshold":"0"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := User{
		Id:   1001,
		Name: "lixioaya",
	}
	err = adapter.Set(sk1, sv1, 10)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var sv11 User
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Memc Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}
}

// TestStruct2
func TestStruct2(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","maxIdle":"10","idelTimeOut":"100","prefix":"le_","serializer":"json","compressType":"zlib","compressThreshold":"0"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := make(map[int]User)
	sv1[0] = User{
		Id:   1001,
		Name: "lixioaya1",
	}

	sv1[1] = User{
		Id:   1002,
		Name: "lixioaya2",
	}

	err = adapter.Set(sk1, sv1, 60)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	sv11 := make(map[int]User)
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}
	fmt.Println(sv11)
}

// TestEncode 加密测试
func TestMemcEncode(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","maxIdle":"10","idelTimeOut":"100","prefix":"le_","serializer":"json","compressType":"zlib","compressThreshold":"0","encodeKey":"abcdefghij123456"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := User{
		Id:   1001,
		Name: "lixioaya",
	}
	err = adapter.Set(sk1, sv1, 60, true)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var sv11 User
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Memc Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}

	///////// MSet MGet ////////

	mList := make(map[string]interface{})
	mList["k1"] = "val1111"
	mList["k2"] = "val2222"
	mList["k3"] = "val3333"
	mList["k4"] = "val4444"
	err = adapter.MSet(mList, 60, true)
	if err != nil {
		t.Errorf("Memc MSet failed. err: %s.", err.Error())
		return
	}

	mList2, err := adapter.MGet("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Memc MGet failed. err: %s.", err.Error())
		return
	}

	var v1, v2, v3, v4 string
	if mList2["k1"] != nil {
		v1 = string(mList2["k1"].([]byte))
	}
	if mList2["k2"] != nil {
		v2 = string(mList2["k2"].([]byte))
	}
	if mList2["k3"] != nil {
		v3 = string(mList2["k3"].([]byte))
	}
	if mList2["k4"] != nil {
		v4 = string(mList2["k4"].([]byte))
	}

	if v1 != mList["k1"] || v2 != mList["k2"] || v3 != mList["k3"] || v4 != mList["k4"] {
		t.Errorf("Memc MGet failed. v1:%s v2:%s v3:%s v4:%s.", v1, v2, v3, v4)
		return
	}

	//////// int ///////

	sk2 := "k2"
	sv2 := 100
	err = adapter.Set(sk2, sv2, 60, true)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	var sv22 int
	err, _ = adapter.Get(sk2, &sv22)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	} else if sv2 != sv22 {
		t.Errorf("Memc Get failed. id[%d] name[%s].", sv22, sv2)
		return
	}
}

////////// 实现IJson接口测试 /////////////
type Item struct {
	uid   int32
	name  string
}

func (this *Item)MarshalJSON() ([]byte, error) {
	fmt.Println("Item MarshalJSON")
	str := fmt.Sprintf(`{"uid":%d, "name":"%s"}`, this.uid, this.name)
	return []byte(str), nil
}

func (this *Item)UnmarshalJSON(data []byte) error {
	fmt.Println("Item UnmarshalJSON")

	val := make(map[string]interface{})
	json.Unmarshal(data, &val)
	uid, _ := val["uid"]
	this.uid = int32(uid.(float64))
	name, _ := val["name"]
	this.name = name.(string)
	return nil
}

// TestMemcIJson 实现IJson接口测试
func TestMemcIJson(t *testing.T) {
	var err error
	adapter := &MemcCache{}
	err = adapter.Init(`{"addr":"127.0.0.1:11211","maxIdle":"10","idelTimeOut":"100","prefix":"le_","serializer":"json","compressType":"zlib","compressThreshold":"0","encodeKey":"abcdefghij123456"}`)
	if err != nil {
		t.Errorf("Memc Init failed. err: %s.", err.Error())
		return
	}

	key := "k1"
	val1 := Item{
		uid: 1000,
		name: "nick",
	}
	// Set
	err = adapter.Set(key, &val1, 3600)
	if err != nil {
		t.Errorf("Memc Set failed. err: %s.", err.Error())
		return
	}

	// Get
	val2 := Item{}
	err, _ = adapter.Get(key, &val2)
	if err != nil {
		t.Errorf("Memc Get failed. err: %s.", err.Error())
		return
	} else if val1.uid != val2.uid || val1.name != val2.name  {
		t.Errorf("Memc Get failed. Got: %d-%s expected: %d-%s.", val2.uid, val2.name, val1.uid, val1.name)
		return
	}
}
