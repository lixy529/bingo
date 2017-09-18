// db的操作类
// 支持主从读写分离
//   变更历史
//     2017-02-21  lixiaoya  新建
package db

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"
)

// DbAdapter
type DbHandle struct {
	master    *sql.DB   // 主库只有一个
	slavers   []*sql.DB // 从库可以有多个
	slaverCnt int       // 从库个数
}

// NewDbHander 实例化DbHandle对象
//   参数
//
//   返回
//     DbHandle对象
func NewDbHandle() *DbHandle {
	return &DbHandle{}
}

// Open 打开DB
//   参数
//     driverName: 驱动名称
//     maxConn:    最大连接数
//     maxIdle:    最大空闲连接数
//     maxLife:    可被重新使用的最大时间间隔，如果小于0将永久重用，单位秒
//     configs:    连接串信息，第一个做为主库，其它都为从库
//   返回
//     成功返回nil，失败返回错误信息
func (h *DbHandle) Open(driverName string, maxConn, maxIdle int, maxLife int64, configs ...string) error {
	if driverName == "" {
		return errors.New("db: Driver Name is empty")
	} else if len(configs) == 0 {
		return errors.New("db: Config is empty")
	}

	if maxConn < 0 {
		maxConn = 0
	}

	var err error

	// 主库配置
	h.master, err = sql.Open(driverName, configs[0])
	if err != nil {
		return err
	}
	h.master.SetMaxOpenConns(maxConn)
	if maxIdle >= 0 {
		h.master.SetMaxIdleConns(maxIdle)
	}
	h.master.SetConnMaxLifetime(time.Duration(maxLife))

	// 从库配置
	h.slaverCnt = len(configs) - 1
	if h.slaverCnt > 0 {
		for i := 1; i <= h.slaverCnt; i++ {
			t, err1 := sql.Open(driverName, configs[i])
			if err1 != nil {
				return err1
			}
			t.SetMaxOpenConns(maxConn)
			t.SetMaxIdleConns(maxIdle)
			t.SetConnMaxLifetime(time.Duration(maxLife))

			h.slavers = append(h.slavers, t)
		}
	}

	return nil
}

// GetMaster 返回主库连接
//   参数
//
//   返回
//     主库对象地址
func (h *DbHandle) GetMaster() *sql.DB {
	return h.master
}

// GetSlave 返回一个从库连接
//   参数
//     n: 返回第n个从库
//       如果没有从库直接返回主库连接
//       如果n的从库存在则返回第n个从库
//       其它情况随机取一个从库
//   返回
//     从库对象地址
func (h *DbHandle) GetSlave(n ...int) *sql.DB {
	if h.slaverCnt <= 0 {
		return h.master
	} else if h.slaverCnt == 1 {
		return h.slavers[0]
	}

	// n存在
	if len(n) > 0 && n[0] >= 0 && n[0] < h.slaverCnt {
		return h.slavers[n[0]]
	}

	// 随机取一个从库
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(h.slaverCnt)

	return h.slavers[i]
}

// FetchOne 查询从库，返回第一行，结果值都会转为string
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) FetchOne(sqlStr string, args ...interface{}) (map[string]string, error) {
	db := h.GetSlave()
	if db == nil {
		return nil, errors.New("db: Slave DB is nil")
	}

	return h.queryOne(db, sqlStr, args...)
}

// FetchOneMaster 查询主库，返回第一行，结果值都会转为string
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) FetchOneMaster(sqlStr string, args ...interface{}) (map[string]string, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return h.queryOne(db, sqlStr, args...)
}

// queryOne 返回第一行
//   参数
//     db:     数据库连接
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) queryOne(db *sql.DB, sqlStr string, args ...interface{}) (map[string]string, error) {
	rows, err := db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 所有字段名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make(map[string]string, colCnt)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		for i, col := range values {
			if col == nil {
				res[columns[i]] = ""
			} else {
				res[columns[i]] = string(col)
			}
		}

		break
	}

	return res, nil
}

// FetchAll 查询从库，返回所有行，结果值都会转为string
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) FetchAll(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	db := h.GetSlave()
	if db == nil {
		return nil, errors.New("db: Slave DB is nil")
	}

	return h.queryAll(db, sqlStr, args...)
}

// FetchAllMaster 查询主库，返回所有行，结果值都会转为string
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) FetchAllMaster(sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return h.queryAll(db, sqlStr, args...)
}

// FetchAll 返回所有行，结果值都会转为string
//   参数
//     db:     数据库连接
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) queryAll(db *sql.DB, sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	rows, err := db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 所有字段名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		mapVal := make(map[string]string, colCnt)
		for i, col := range values {
			if col == nil {
				mapVal[columns[i]] = ""
			} else {
				mapVal[columns[i]] = string(col)
			}
		}
		res = append(res, mapVal)
	}

	return &res, nil
}

// Insert 插入操作，不支持事务
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回自增ID，失败返回错误信息
func (h *DbHandle) Insert(sqlStr string, args ...interface{}) (int64, error) {
	db := h.GetMaster()
	if db == nil {
		return -1, errors.New("db: Master DB is nil")
	}

	stmtIns, err := db.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	res, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// Exec 更新和删除操作
//   参数
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回更新或删除行数，失败返回错误信息
func (h *DbHandle) Exec(sqlStr string, args ...interface{}) (int64, error) {
	db := h.GetMaster()
	if db == nil {
		return -1, errors.New("db: Master DB is nil")
	}

	stmtIns, err := db.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

// FetchOne 查询从库，返回第一行，支持事务，结果值都会转为string
//   参数
//     tx:     事务
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) TxFetchOne(tx *sql.Tx, sqlStr string, args ...interface{}) (map[string]string, error) {
	rows, err := tx.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 所有字段名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make(map[string]string, colCnt)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		for i, col := range values {
			if col == nil {
				res[columns[i]] = ""
			} else {
				res[columns[i]] = string(col)
			}
		}

		break
	}

	return res, nil
}

// FetchAll 返回所有行，支持事务，结果值都会转为string
//   参数
//     tx:     事务
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回查询结果，失败返回错误信息
func (h *DbHandle) TxFetchAll(tx *sql.Tx, sqlStr string, args ...interface{}) (*[]map[string]string, error) {
	rows, err := tx.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 所有字段名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colCnt := len(columns)
	values := make([]sql.RawBytes, colCnt)
	scanArgs := make([]interface{}, colCnt)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	res := make([]map[string]string, 0)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		mapVal := make(map[string]string, colCnt)
		for i, col := range values {
			if col == nil {
				mapVal[columns[i]] = ""
			} else {
				mapVal[columns[i]] = string(col)
			}
		}
		res = append(res, mapVal)
	}

	return &res, nil
}

// TxInsert 插入操作，支持事务
//   参数
//     tx:     事务
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回自增ID，失败返回错误信息
func (h *DbHandle) TxInsert(tx *sql.Tx, sqlStr string, args ...interface{}) (int64, error) {
	stmtIns, err := tx.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	res, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// TxExec 更新和删除操作，支持事务
//   参数
//     tx:     事务
//     sqlStr: Sql串
//     args:   参数
//   返回
//     成功时返回更新或删除行数，失败返回错误信息
func (h *DbHandle) TxExec(tx *sql.Tx, sqlStr string, args ...interface{}) (int64, error) {
	stmtIns, err := tx.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

// Begin 开始事务，事务都从主库进行操作
//   参数
//
//   返回
//     成功时返回事务，失败返回错误信息
func (h *DbHandle) Begin() (*sql.Tx, error) {
	db := h.GetMaster()
	if db == nil {
		return nil, errors.New("db: Master DB is nil")
	}

	return db.Begin()
}

// Commit 提交事务
//   参数
//     tx: 事务
//   返回
//     成功时返回nil，失败返回错误信息
func (h *DbHandle) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

// Rollback 回滚事务
//   参数
//     tx: 事务
//   返回
//     成功时返回nil，失败返回错误信息
func (h *DbHandle) Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}

// Close 关闭所有连接
//   参数
//
//   返回
//
func (h *DbHandle) Close() {
	if h.master != nil {
		h.master.Close()
	}

	if h.slaverCnt > 0 {
		for _, db := range h.slavers {
			if db != nil {
				db.Close()
			}
		}
	}

}
