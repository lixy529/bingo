// Db 数据库的封装
// 支持主从读写分离
//   变更历史
//     2017-02-21  lixiaoya  新建
package db

import (
	"errors"
)

// DbBase
type DbBase struct {
	dbs map[string]*DbHandle
}

// NewDbBase 新建DbBase
//   参数
//     configs: 配置信息，比如：
//       dbName => db_cn # cn将做为dbs的key值
//       driverName => mysql
//       maxConn = 200 # 最大连接数
//       maxIdle = 100 # 最大空闲连接数
//       maxLife = 21600 # 可被重新使用的最大时间间隔，如果小于0将永久重用，单位秒
//       master = user:pwd@tcp(ip:port)/dbname?charset=utf8 # 主库
//       slave1 = user:pwd@tcp(ip:port)/dbname?charset=utf8 # 从库1
//       slave2 = user:pwd@tcp(ip:port)/dbname?charset=utf8 # 从库2
//   返回
//     成功返回DbBase对象地址，失败返回错误信息
func NewDbBase(configs ...map[string]interface{}) (*DbBase, error) {
	n := len(configs)
	if n == 0 {
		return nil, errors.New("db: Db config is empty ")
	}

	dbBase := &DbBase{}
	dbBase.dbs = make(map[string]*DbHandle)

	for _, config := range configs {
		dbName := config["dbName"].(string)
		var maxConn int = 0
		var maxIdle int = -1
		var maxLife int64 = 0
		var driverName string
		var configs []string
		if val, ok := config["driverName"]; ok && val != "" {
			driverName = val.(string)
		} else {
			return dbBase, errors.New("db: Driver name is empty")
		}
		// 主库必须有
		if val, ok := config["master"]; ok {
			configs = append(configs, val.(string))
		} else {
			return dbBase, errors.New("db: Master config is empty")
		}

		// 从库可以为空
		if len(config["slaves"].([]string)) > 0 {
			configs = append(configs, config["slaves"].([]string)...)
		}

		if val, ok := config["maxConn"]; ok {
			maxConn, _ = val.(int)
		}
		if val, ok := config["maxIdle"]; ok {
			maxIdle, _ = val.(int)
		}
		if val, ok := config["maxLife"]; ok {
			maxLife, _ = val.(int64)
		}

		h := NewDbHandle()
		err := h.Open(driverName, maxConn, maxIdle, maxLife, configs...)
		if err != nil {
			return dbBase, err
		}
		dbBase.dbs[dbName] = h
	}

	return dbBase, nil
}

// Db 根据配置返回对应的数据库连接
//   参数
//     dbName: 比如各个地区对应不同的主从库，通过dbName区分
//   返回
//     DbHandle对象地址
func (b *DbBase) Db(dbName ...string) *DbHandle {
	name := "db"
	if len(dbName) > 0 && dbName[0] != "" {
		name = dbName[0]
	}

	if db, ok := b.dbs[name]; ok {
		return db
	}
	return nil
}

// Close 关闭所有数据库连接池
//   参数
//
//   返回
//
func (b *DbBase) Close() {
	for _, h := range b.dbs {
		h.Close()
	}
}
