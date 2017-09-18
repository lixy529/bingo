// 文件相关函数
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"errors"
	"os"
	"path"
	"fmt"
	"syscall"
	"reflect"
)

// FileExists 返回文件是否存在
//   参数
//     name: 文件全路径
//   返回
//     true-文件存在 false-文件不存在
func FileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// FileExists name是否是目录
//   参数
//     name: 目录全路径
//   返回
//     true-是目录 false-不是目录，获取目录信息失败时会返回错误信息
func IsDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	} else if fi.IsDir() {
		return true, nil
	}

	return false, nil
}

// FileExists name是否是文件
//   参数
//     name: 文件全路径
//   返回
//     true-是文件 false-不是文件，获取文件信息失败时会返回错误信息
func IsFile(name string) (bool, error) {
	r, err := IsDir(name)
	return !r, err
}

// MkDir 建文件夹
//   参数
//     dir:       目录路径
//     perm:      新建目录的权限
//     existName: 路径中是否带文件名，如果path包含文件名，则去掉文件名
//   返回
//     成功返回nil，失败时返回错误信息
func MkDir(dir string, perm os.FileMode, existName bool) error {
	if existName {
		dir = path.Dir(dir)
	}
	if len(dir) == 0 {
		return errors.New("dir is empty")
	}

	return os.MkdirAll(dir, perm)
}

// WriteFile 写数据到文件
//   参数
//     name: 目录路径
//     data: 要写的数据
//     flag: 写文件标识，如os.O_RDWR、os.O_CREATE、os.O_APPEND等
//     perm: 新建目录的权限
//   返回
//     成功返回nil，失败时返回错误信息
func WriteFile(name string, data []byte, flag int, perm os.FileMode) (int, error) {
	fd, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	return fd.Write(data)
}

// FileCtime 获取文件的创建时间
// linux版本使用Ctim，mac版本使用Ctimespec，为发兼容只能使用反射去找对应的字段名
//   参数
//     path: 文件路径
//   返回
//     秒、纳秒、错误信息
func FileCtime(path string) (int64, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println("Stat:", err)
		return -1, -1, err
	}

	sysInfo := fi.Sys()
	if stat, ok := sysInfo.(*syscall.Stat_t); ok {
		// linux使用Ctim, mac使用Ctimespec
		//return stat.Ctimespec.Sec, stat.Ctimespec.Nsec, nil
		//return stat.Ctim.Sec, stat.Ctim.Nsec, nil
		// 为了兼容使用下面反射处理
		elem := reflect.ValueOf(stat).Elem()
		type_ := elem.Type()
		for i := 0; i < type_.NumField(); i++ {
			fieldName := type_.Field(i).Name
			if fieldName == "Ctimespec" || fieldName == "Ctim" {
				ctim := elem.Field(i).Interface().(syscall.Timespec)
				return ctim.Sec, ctim.Nsec, nil
			}
		}
		return -1, -1, errors.New("Not found create time field")
	}

	return -1, -1, errors.New("Assertion error")
}
