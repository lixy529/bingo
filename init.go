// 初始化相关函数
//   变更历史
//     2017-02-06  lixiaoya  新建
package bingo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"github.com/lixy529/gotools/cache"
	"github.com/lixy529/gotools/cache/memcache"
	"github.com/lixy529/gotools/cache/redis/redism"
	"github.com/lixy529/gotools/cache/redis/redisc"
	"github.com/lixy529/gotools/cache/redis/redisd"
	"github.com/lixy529/gotools/db"
	"github.com/lixy529/bingo/lang"
	"github.com/lixy529/gotools/logs"
	"github.com/lixy529/bingo/session"
	"github.com/lixy529/gotools/utils"
	"log"
	"os"
	"strconv"
)

type initfunc func() error
type unInitfunc func()

var (
	GlobalSession *session.Manager
	GlobalDb      *db.DbBase
	Flogger       logs.Logger
	Glogger       logs.Logger
	GTemplate     *Template
	GLang         *lang.Lang
	inits         = make([]initfunc, 0)
	unInits       = make([]unInitfunc, 0)
)

// init 添加需要初始化函数
//   参数
//     void
//   返回
//     void
func init() {
	// 初始化
	AddInitFunc(initFrameLog)
	AddInitFunc(initPidFile)
	AddInitFunc(initBusLog)
	AddInitFunc(initSession)
	AddInitFunc(initDb)
	AddInitFunc(initCache)
	AddInitFunc(initViews)
	AddInitFunc(initLang)
	AddInitFunc(initGzip)

	// 反初始化
	AddUnInitFunc(unInitDb)
	AddUnInitFunc(unInitFrameLog)
	AddUnInitFunc(unInitPidFile)
}

// AddInitFunc 添加init函数
//   参数
//     f: 初始化函数
//   返回
//     void
func AddInitFunc(f initfunc) {
	inits = append(inits, f)
}

// AddUnInitFunc 添加uninit函数
//   参数
//     f: 初始化函数
//   返回
//     void
func AddUnInitFunc(f unInitfunc) {
	unInits = append(unInits, f)
}

// initPidFile 生成pid文件
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initPidFile() error {
	if isShell {
		return nil
	}

	pidFile := AppCfg.ServerCfg.PidFile
	err := utils.MkDir(pidFile, 0777, true)
	if err == nil {
		pid := []byte(strconv.Itoa(os.Getpid()))
		return ioutil.WriteFile(pidFile, pid, 0666)
	}

	return err
}

// initSession 初始化Session信息
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initSession() error {
	if isShell {
		return nil
	}

	var err error
	if AppCfg.SessCfg.SessOn {
		GlobalSession, err = session.NewManager(AppCfg.SessCfg.ProviderName, AppCfg.SessCfg.ProviderConfig, AppCfg.SessCfg.CookieName, AppCfg.SessCfg.LifeTime)
		if err != nil {
			return err
		}

		go GlobalSession.SessGc()
	}

	return nil
}

// initDb 初始化Db
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initDb() error {
	var err error
	var params []map[string]interface{}
	for _, cfg := range AppCfg.DbConfigs {
		param := make(map[string]interface{})
		param["dbName"] = cfg.dbName
		param["driverName"] = cfg.driverName
		param["maxOpen"] = cfg.maxOpen
		param["maxIdle"] = cfg.maxIdle
		param["maxLife"] = cfg.maxLife
		param["master"] = cfg.master
		param["slaves"] = cfg.slaves

		params = append(params, param)
	}

	if len(params) > 0 {
		GlobalDb, err = db.NewDbBase(params...)
	}

	return err
}

// initCache 初始化Cache
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initCache() error {
	for _, cfg := range AppCfg.CacheCfgs {
		if _, ok := cache.Adapters[cfg.cacheName]; ok {
			return fmt.Errorf("Cache: Add adapter [%s] is exists", cfg.cacheName)
		}

		if cfg.cacheType == "memcache" {
			cache.Adapters[cfg.cacheName] = memcache.NewMemcCache()
		} else if cfg.cacheType == "redisc" {
			cache.Adapters[cfg.cacheName] = redisc.NewRediscCache()
		} else if cfg.cacheType == "redisd" {
			cache.Adapters[cfg.cacheName] = redisd.NewRedisdCache()
		} else if cfg.cacheType == "redism" {
			cache.Adapters[cfg.cacheName] = redism.NewRedismCache()
		} else {
			return fmt.Errorf("Cache:  Adapter type [%s] isn't redis or memcache", cfg.cacheName)
		}

		err := cache.Adapters[cfg.cacheName].Init(cfg.cacheConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// initFrameLog 初始化框架使用的日志
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initFrameLog() error {
	if isShell {
		return nil
	}

	logName := AppCfg.LogName
	logCfg := AppCfg.LogCfg

	Flogger = logs.Log(logName)
	if Flogger == nil {
		return errors.New("Flogger is nil")
	}

	return Flogger.Init(logCfg)
}

// initBusLog 初始化业务使用的日志
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initBusLog() error {
	logName := AppCfg.Log.LogName
	logCfg := AppCfg.Log.LogCfg

	Glogger = logs.Log(logName)
	if Glogger == nil {
		if isShell {
			log.Println("GLogger is nil")
		} else {
			Flogger.Error("GLogger is nil")
		}
	}

	err := Glogger.Init(logCfg)
	if err != nil {
		if isShell {
			log.Printf("GLogger Init err: %s", err.Error())
		} else {
			Flogger.Errorf("GLogger Init err: %s", err.Error())
		}
	}

	return nil
}

// initViews 初始化view文件
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initViews() error {
	if isShell {
		return nil
	}

	GTemplate = NewTemplate(AppCfg.WebCfg.ViewsDir, AppCfg.WebCfg.ViewsExt)
	if GTemplate == nil {
		return errors.New("Template is nil")
	}

	return GTemplate.buildViews()
}

// initLang 初始化语言包
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initLang() error {
	if isShell {
		return nil
	}

	if AppCfg.LangCfg.LangPath == "" {
		return nil
	}

	var err error
	GLang, err = lang.NewLang(AppCfg.LangCfg.LangPath)
	if err != nil {
		return err
	}

	return nil
}

// initGzip 初始化gzip
//   参数
//     void
//   返回
//     成功返回nil，失败返回错误信息
func initGzip() error {
	if isShell {
		return nil
	}

	gzipLevel = AppCfg.ServerCfg.GzipLevel
	gzipMinLen = AppCfg.ServerCfg.GzipMinLen

	return nil
}

// unInitDb 关闭数据库连接
//   参数
//     void
//   返回
//     void
func unInitDb() {
	if GlobalDb != nil {
		GlobalDb.Close()
	}
}

// unInitFrameLog 关闭日志
//   参数
//     void
//   返回
//     void
func unInitFrameLog() {
	if Flogger != nil {
		Flogger.Destroy()
	}
}

// unInitPidFile 删除pid文件
//   参数
//     void
//   返回
//     void
func unInitPidFile() {
	if !isShell {
		fi, err := os.Open(AppCfg.ServerCfg.PidFile)
		if err == nil {
			defer fi.Close()
			fd, err := ioutil.ReadAll(fi)
			if err == nil {
				myPid := strconv.Itoa(os.Getpid())
				if myPid == string(fd) {
					os.Remove(AppCfg.ServerCfg.PidFile)
				}
			}
		}
	}
}
