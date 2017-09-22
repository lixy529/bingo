// redis adapter测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package redis

import (
	"github.com/bingo/cache"
	"fmt"
	"testing"
	"encoding/json"
)

func TestRedisCache(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6378","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
		return
	}

	////////////////////////string测试////////////////////////////
	k1 := "k1"
	v1 := "HelloWorld"
	err = adapter.Set(k1, v1, 20)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var v11 string
	err, exist := adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if !exist {
		t.Errorf("Memc Get failed. %s is not exist.", k1)
		return
	} else if v11 != v1 {
		t.Errorf("Redis Get failed. Got %s, expected %s.", v11, v1)
		return
	}

	isExist, err := adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Redis Get IsExist. err: %s.", err.Error())
		return
	} else if !isExist {
		t.Error("Redis Get failed. Got false, expected true.")
		return
	}

	err = adapter.Del(k1)
	if err != nil {
		t.Errorf("Redis Delete failed. err: %s.", err.Error())
		return
	}

	isExist, err = adapter.IsExist(k1)
	if err != nil {
		t.Errorf("Redis Get IsExist. err: %s.", err.Error())
		return
	} else if isExist {
		t.Error("Redis Get failed. Got true, expected false.")
		return
	}

	v11 = ""
	err, _ = adapter.Get(k1, &v11)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if v11 != "" {
		t.Errorf("Redis Get failed. Got %s, expected nil.", v11)
		return
	}

	////////////////////////int32测试////////////////////////////
	k2 := "k2"
	v2 := 100
	err = adapter.Set(k2, int32(v2), 30)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var v22 int
	err, _ = adapter.Get(k2, &v22)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if v22 != v2 {
		t.Errorf("Redis Get failed. Got %d, expected %d.", v22, v2)
		return
	}

	////////////////////////float64测试////////////////////////////
	k3 := "k3"
	v3 := 100.01
	err = adapter.Set(k3, v3, 30)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var v33 float64
	err, _ = adapter.Get(k3, &v33)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if v33 != v3 {
		t.Errorf("Redis Get failed. Got %f, expected %f.", v33, v3)
		return
	}

	////////////////////////Incr、Decr测试////////////////////////////
	k4 := "k4"
	v4 := "100"
	err = adapter.Set(k4, v4, 30)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var v44 string
	err, _ = adapter.Get(k4, &v44)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if v44 != v4 {
		t.Errorf("Redis Get failed. Got %s, expected %s.", v44, v4)
		return
	}

	////////
	k5 := "k5"
	v5 := 100
	err = adapter.Set(k5, v5, 30)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var v55 int
	err, _ = adapter.Get(k5, &v55)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if v55 != v5 {
		t.Errorf("Redis Get failed. Got %d, expected %d.", v55, v5)
		return
	}

	newVal5, _ := adapter.Incr(k5)
	if newVal5 != 101 {
		t.Errorf("Redis Incr failed. Got %d, expected %d.", newVal5, 101)
		return
	}

	newVal5, _ = adapter.Decr(k5)
	if newVal5 != 100 {
		t.Errorf("Redis Incr failed. Got %d, expected %d.", newVal5, 100)
		return
	}

	newVal5, _ = adapter.Incr(k5, 10)
	if newVal5 != 110 {
		t.Errorf("Redis Incr failed. Got %d, expected %d.", newVal5, 110)
		return
	}

	newVal5, _ = adapter.Decr(k5, 10)
	if newVal5 != 100 {
		t.Errorf("Redis Incr failed. Got %d, expected %d.", newVal5, 100)
		return
	}

	////////////////////////哈希表测试////////////////////////////
	k6 := "addr"
	f6 := "google"
	v6 := "www.google.com"
	adapter.HSet(k6, "baidu", "www.baidu.com", 60)
	adapter.HSet(k6, "le", "www.le.com", 60)
	r, err := adapter.HSet(k6, f6, v6, 60)
	fmt.Println(r)
	if err != nil {
		t.Errorf("Redis HSet failed. err: %s.", err.Error())
		return
	}

	var v66 string
	err, _ = adapter.HGet(k6, f6, &v66)
	if err != nil {
		t.Errorf("Redis HGet failed. err: %s.", err.Error())
		return
	} else if v66 != v6 {
		t.Errorf("Redis HGet failed. Got %s, expected %s.", v66, v6)
		return
	}

	// HGetAll
	fmt.Println("=== HGetAll Begin ===")
	v77, err := adapter.HGetAll(k6)
	if err != nil {
		t.Errorf("Redis HGetAll failed. err: %s.", err.Error())
		return
	}
	for k, v := range v77 {
		var val string
		//json.Unmarshal(v.([]byte), &val)
		val = string(v.([]byte))
		fmt.Println(k, val)
	}
	fmt.Println("=== HGetAll End ===")

	// HMGet
	fmt.Println("=== HMGet Begin ===")
	v99, err := adapter.HMGet(k6, "google", "baidu", "le")
	if err != nil {
		t.Errorf("Redis HMGet failed. err: %s.", err.Error())
		return
	}
	for k, v := range v99 {
		if v == nil {
			fmt.Println(k, v)
			continue
		}
		var val string
		//json.Unmarshal(v.([]byte), &val)
		val = string(v.([]byte))
		fmt.Println(k, val)
	}
	fmt.Println("=== HMGet End ===")

	// HVals
	v88, err := adapter.HVals(k6)
	if err != nil {
		t.Errorf("Redis HVals failed. err: %s.", err.Error())
		return
	}
	for _, v := range v88 {
		//var val string
		//json.Unmarshal(v.([]byte), &val)
		//fmt.Println(val)
		fmt.Println(string(v.([]byte)))
	}

	err = adapter.HDel(k6, f6, "baidu")
	if err != nil {
		t.Errorf("Redis HDel failed. err: %s.", err.Error())
		return
	}

	////////////////////////ClearAll测试////////////////////////////
	err = adapter.ClearAll()
	if err != nil {
		t.Errorf("Redis ClearAll failed. err: %s.", err.Error())
		return
	}
}

// TestRedisMulti
func TestRedisMulti(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
		return
	}

	mList := make(map[string]interface{})
	mList["k1"] = "val1111"
	mList["k2"] = "val2222"
	mList["k3"] = "val3333"
	mList["k4"] = "val4444"
	err = adapter.MSet(mList, 60)
	if err != nil {
		t.Errorf("Redis MSet failed. err: %s.", err.Error())
		return
	}

	mList2, err := adapter.MGet("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Redis MGet failed. err: %s.", err.Error())
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
		t.Errorf("Redis MGet failed. v1:%s v2:%s v3:%s v4:%s.", v1, v2, v3, v4)
		return
	}

	err = adapter.MDel("k1", "k2", "k3", "k4")
	if err != nil {
		t.Errorf("Redis MDelete failed. err: %s.", err.Error())
		return
	}
}

// TestRedisSet 有序集合测试
func TestRedisSet(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
		return
	}

	key := "sets"
	// 添加
	n, err := adapter.ZSet(key, 60, 5, "val5", 3.5, "val3.5", 1, "100", 4, 400, 0.5, "val0.5", 1, "val1")
	fmt.Println(n)
	if err != nil {
		t.Errorf("Redis ZSet failed. err: %s.", err.Error())
		return
	}

	// 查询，递增排列
	res, err := adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// 查询，递减排列
	res, err = adapter.ZGet(key, 0, -1, true, true)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// 基数
	n, err = adapter.ZCard(key)
	if err != nil {
		t.Errorf("Redis ZCard failed. err: %s.", n)
		return
	}
	if n != 6 {
		t.Errorf("Redis ZCard failed. Got %d, expected 6.", n)
		return
	}

	// 删除
	n, err = adapter.ZDel(key, "val3.5", "400")
	if err != nil {
		t.Errorf("Redis ZDel failed. err: %s.", err.Error())
		return
	}
	fmt.Println(n)

	// 查询
	res, err = adapter.ZGet(key, 0, -1, true, false)
	if err != nil {
		t.Errorf("Redis ZGet failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)

	// 基数
	n, err = adapter.ZCard(key)
	if err != nil {
		t.Errorf("Redis ZCard failed. err: %s.", n)
		return
	} else if n != 4 {
		t.Errorf("Redis ZCard failed. Got %d, expected 4.", n)
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
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
		return
	}

	sk1 := "k1"
	sv1 := User{
		Id:   1001,
		Name: "lixioaya",
	}
	err = adapter.Set(sk1, sv1, 10)
	if err != nil {
		t.Errorf("Redis Set failed. err: %s.", err.Error())
		return
	}

	var sv11 User
	err, _ = adapter.Get(sk1, &sv11)
	if err != nil {
		t.Errorf("Redis Get failed. err: %s.", err.Error())
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Redis Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}

	///////////哈希表////////////
	k6 := "addr"
	f6 := "google"
	v6 := User{
		Id:   1001,
		Name: "lixioaya",
	}

	_, err = adapter.HSet(k6, f6, v6, 60)
	if err != nil {
		t.Errorf("Redis HSet failed. err: %s.", err.Error())
		return
	}

	var v66 User
	err, exist := adapter.HGet(k6, f6, &v66)
	if err != nil {
		t.Errorf("Redis HGet failed. err: %s.", err.Error())
		return
	} else if !exist {
		t.Errorf("Redis Get failed. %s - %s is not exist.", k6, f6)
		return
	} else if sv11.Id != sv1.Id || sv11.Name != sv1.Name {
		t.Errorf("Redis Get failed. id[%d] name[%s].", sv11.Id, sv11.Name)
		return
	}
}

// TestRedisEncode 加密测试
func TestRedisEncode(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180","encodeKey":"abcdefghij123456"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
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

// TestRedisIJson 实现IJson接口测试
func TestRedisIJson(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180","encodeKey":"abcdefghij123456"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
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

	// Mset
	mList := make(map[string]interface{})
	mList["k1"] = &Item{
		uid: 1001,
		name: "nick1",
	}
	mList["k2"] = &Item{
		uid: 1002,
		name: "nick2",
	}
	mList["k3"] = &Item{
		uid: 1003,
		name: "nick3",
	}
	err = adapter.MSet(mList, 600)
	if err != nil {
		t.Errorf("Memc MSet failed. err: %s.", err.Error())
		return
	}

	// HSet
	k6 := "addr"
	_, err = adapter.HSet(k6, "baidu", &Item{
		uid: 1001,
		name: "baidu",
	}, 60)
	if err != nil {
		t.Errorf("Memc HSet failed. err: %s.", err.Error())
		return
	}
	_, err = adapter.HSet(k6, "le", &Item{
		uid: 1002,
		name: "leeco",
	}, 60)
	if err != nil {
		t.Errorf("Memc HSet failed. err: %s.", err.Error())
		return
	}

	// HGet
	v66 := Item{}
	err, _ = adapter.HGet(k6, "baidu", &v66)
	if err != nil {
		t.Errorf("Redis HGet failed. err: %s.", err.Error())
		return
	} else if v66.uid != 1001 || v66.name != "baidu"  {
		t.Errorf("Memc Get failed. Got: %d-%s expected: %d-%s.", v66.uid, v66.name, 1001, "baidu")
		return
	}
}

// TestPipeline
func TestPipeline(t *testing.T) {
	var err error
	adapter := &RedisCache{}
	err = adapter.Init(`{"master":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"slave":{"conn":"127.0.0.1:6379","dbNum":"0","auth":"le123123"},"maxIdle":"3","maxActive":"0","idleTimeOut":"180"}`)
	if err != nil {
		t.Errorf("Redis Init failed. err: %s.", err.Error())
		return
	}

	cmd1 := cache.NewCmd("SET", "foo", "bar")
	cmd2 := cache.NewCmd("GET", "foo")
	cmd3 := cache.NewCmd("SET", "num", 0)
	cmd4 := cache.NewCmd("INCR", "num")
	cmd5 := cache.NewCmd("INCR", "num")
	res, err := adapter.Exec(cmd1, cmd2, cmd3, cmd4, cmd5)
	if err != nil {
		t.Errorf("Redis Exec failed. err: %s.", err.Error())
		return
	}
	fmt.Println(res)
	for k, r := range res.([]interface{}) {
		if v, ok := r.(string); ok {
			fmt.Println("string", k, v)
		} else if v, ok := r.([]byte); ok {
			fmt.Println("byte", k, string(v))
		} else if v, ok := r.(int64); ok {
			fmt.Println("int64", k, v)
		}
	}
}
