// model demo
//   变更历史
//     2017-02-10  lixiaoya  新建
package models

import (
	"fmt"
	"math/rand"
	"time"
)

type DemoModel struct {
	BaseModel
}

// InsertInfo 设置信息
func (m *DemoModel) InsertInfo(name string, age int, addr string) (int64, error) {
	return m.Insert("insert into t_test(name,age,addr) values(?,?,?)", name, age, addr)
}

// GetInfo 获取信息
func (m *DemoModel) GetInfo(uid int) (map[string]string, error) {
	return m.FetchOne("select * from t_test where id=?", uid)
}

// UpdateInfo 更新用户
func (m *DemoModel) UpdateInfo(uid int, addr string) (int64, error) {
	return m.Exec("update t_test set addr=? where id=?", addr, uid)
}

// DeleteInfo 删除用户
func (m *DemoModel) DeleteInfo(uid int) (int64, error) {
	return m.Exec("delete from t_test where id=?", uid)
}

// TxTest 事务处理测试
func (m *DemoModel) TxTest() {
	m.Begin()
	uid, err := m.TxInsert("insert into t_test(name,age,addr) values(?,?,?)", "txtest", 20, "北京")
	if err != nil {
		fmt.Printf("TxInsert err, %s\n", err.Error())
		m.Rollback()
		return
	}
	fmt.Printf("uid = %d\n", uid)

	res, err := m.TxFetchOne("select * from t_test where id=?", uid)
	if err != nil {
		fmt.Printf("TxFetchOne err, %s\n", err.Error())
		m.Rollback()
		return
	}
	fmt.Println(res)

	cnt, err := m.TxExec("update t_test set name=? where id=?", "txtest2", uid)
	if err != nil {
		fmt.Printf("TxExec err, %s\n", err.Error())
		m.Rollback()
		return
	}
	fmt.Printf("update cnt = %d\n", cnt)

	res, _ = m.TxFetchOne("select * from t_test where id=?", uid)
	if err != nil {
		fmt.Printf("TxFetchOne err, %s\n", err.Error())
		m.Rollback()
		return
	}
	fmt.Println(res)

	_, err = m.TxExec("delete from t_test where id=?", uid)
	if err != nil {
		fmt.Printf("TxFetchOne err, %s\n", err.Error())
		m.Rollback()
		return
	}

	m.Commit()
}

// CacheTest Cache测试
func (m *DemoModel) CacheTest(name string) (string, error) {
	adapter, err := m.Cache(name)
	if err != nil {
		return "", err
	}

	k1 := "k1"
	v1 := "Hello " + name
	err = adapter.Set(k1, v1, 10)
	if err != nil {
		return "", err
	} else {
		var v11 string
		err, exist := adapter.Get(k1, &v11)
		if err != nil {
			return "", err
		} else if !exist {
			return "", fmt.Errorf("key [%s] is not exist", k1)
		}

		return v11, nil
	}
}

// Sort 冒泡排序测试
func (m *DemoModel) Sort() [100]int {
	var arr [100]int
	for i := 0; i < 100; i++ {
		arr[i] = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100)
	}

	for i := 0; i < 100; i++ {
		for j := 0; j < 100-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	//fmt.Println(arr)

	return arr
}

//MongoTest Mongo测试
func (m *DemoModel) MongoTest(name string) (interface{}, error) {
	instance, err := m.Mongo(name)
	if err != nil {
		return nil, err
	}
	_, errUpsert := instance.Upsert("vcs", "testmongo", map[string]string{"_id": "testmongo"}, map[string]string{"_id": "testmongo", "key": "value"})
	if errUpsert != nil {
		return nil, errUpsert
	}
	ret, errFindOne := instance.FindOne("vcs", "testmongo", map[string]string{"_id": "testmongo"})
	if errFindOne != nil {
		return nil, errFindOne
	}
	return ret, nil
}
