// Db 数据库的封装测试
//   变更历史
//     2017-02-22  lixiaoya  新建
package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

// TestDbHandle 增删改查
func TestDb(t *testing.T) {
	map1 := make(map[string]interface{})
	map1["dbName"] = "cn"
	map1["driverName"] = "mysql"
	map1["maxConn"] = 200
	map1["maxIdle"] = 100
	map1["maxLife"] = 21600
	map1["master"] = "root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8"
	map1["slaves"] = []string{"root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8", "root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8"}

	map2 := make(map[string]interface{})
	map2["dbName"] = "us"
	map2["driverName"] = "mysql"
	map2["maxConn"] = 500
	map2["maxIdle"] = 200
	map2["maxLife"] = 21600
	map2["master"] = "root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8"
	map2["slaves"] = []string{"root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8", "root:root123@tcp(10.11.145.15:3306)/passport?charset=utf8"}

	configs := []map[string]interface{}{map1, map2}
	dbBase, err := NewDbBase(configs...)
	if err != nil {
		t.Errorf("NewDbBase error, [%s]", err.Error())
		return
	}

	h := dbBase.Db("cn")

	// FetchOne
	res, err := h.FetchOne("select * from t_test where name=?", "lixy")
	if err != nil {
		t.Errorf("handle.FetchOne error, [%s]", err.Error())
		return
	}
	fmt.Println(res)

	// FetchAll
	res1, err := h.FetchAll("select * from t_test")
	if err != nil {
		t.Errorf("handle.FetchAll error, [%s]", err.Error())
		return
	}
	for _, val := range *res1 {
		fmt.Println(val)
	}

	dbBase.Close()
}
