// Model类的接口及基类
//   变更历史
//     2017-02-23  lixiaoya  新建
package bingo

import (
	"database/sql"
	"errors"
	"github.com/lixy529/bingo/cache"
	"github.com/lixy529/bingo/db"
)

type Model struct {
	db *db.DbHandle
	tx *sql.Tx
}

// Db 设置Db
//   参数
//     dbName: 比如各个地区对应不同的主从库，通过dbName区分
//   返回
//     DbHandle对象地址
func (m *Model) Db(dbName ...string) *db.DbHandle {
	if GlobalDb == nil {
		return nil
	}
	m.db = GlobalDb.Db(dbName...)

	return m.db
}

// FetchOne 查询从库，返回第一行
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) FetchOne(sqlStr string, args ...interface{}) (map[string]string, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.FetchOne(sqlStr, args...)
}

// FetchOneMaster 查询主库，返回第一行
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) FetchOneMaster(sqlStr string, args ...interface{}) (map[string]string, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.FetchOneMaster(sqlStr, args...)
}

// FetchAll 查询从库，返回所有行
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) FetchAll(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.FetchAll(sqlStr, args...)
}

// FetchAllMaster 查询主库，返回所有行
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) FetchAllMaster(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.FetchAllMaster(sqlStr, args...)
}

// Insert 插入操作，不支持事务
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回自增ID，失败返回错误信息
func (m *Model) Insert(sqlStr string, args ...interface{}) (int64, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.Insert(sqlStr, args...)
}

// Exec 更新和删除操作
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回更新或删除行数，失败返回错误信息
func (m *Model) Exec(sqlStr string, args ...interface{}) (int64, error) {
	if m.db == nil {
		m.Db()
	}

	return m.db.Exec(sqlStr, args...)
}

// FetchOne 查询从库，返回第一行，支持事务
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) TxFetchOne(sqlStr string, args ...interface{}) (map[string]string, error) {
	if m.tx == nil {
		return nil, errors.New("Model: Tx is nil")
	}

	return m.db.TxFetchOne(m.tx, sqlStr, args...)
}

// FetchAll 查询从库，返回所有行
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (m *Model) TxFetchAll(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	if m.tx == nil {
		return nil, errors.New("Model: Tx is nil")
	}

	return m.db.TxFetchAll(m.tx, sqlStr, args...)
}

// TxInsert 插入操作，支持事务
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回自增ID，失败返回错误信息
func (m *Model) TxInsert(sqlStr string, args ...interface{}) (int64, error) {
	if m.tx == nil {
		return 0, errors.New("Model: Tx is nil")
	}

	return m.db.TxInsert(m.tx, sqlStr, args...)
}

// TxExec 更新和删除操作，支持事务
// TxExec 更新和删除操作，支持事务
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回更新或删除行数，失败返回错误信息
func (m *Model) TxExec(sqlStr string, args ...interface{}) (int64, error) {
	if m.tx == nil {
		return 0, errors.New("Model: Tx is nil")
	}

	return m.db.TxExec(m.tx, sqlStr, args...)
}

// Begin 开始事务
//   参数
//
//   返回
//     成功时返回事务，失败返回错误信息
func (m *Model) Begin() (*sql.Tx, error) {
	if m.db == nil {
		m.Db()
	}
	var err error
	m.tx, err = m.db.Begin()
	return m.tx, err
}

// Commit 提交事务
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (m *Model) Commit() error {
	err := m.tx.Commit()
	m.tx = nil
	return err
}

// Rollback 回滚事务
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (m *Model) Rollback() error {
	err := m.tx.Rollback()
	m.tx = nil
	return err
}

// Cache 返回一个缓存适配器
//   参数
//     adapterName: 缓存适配器名称，如：redis、memcache
//   返回
//     成功时返回缓存适配器对象，失败返回错误信息
func (m *Model) Cache(adapterName ...string) (cache.Cache, error) {
	return cache.GetCache(adapterName...)
}
